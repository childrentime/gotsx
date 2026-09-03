package compiler

import (
	"fmt"
	"strings"
)

// ---------- 语义类型 ----------
type Type interface{ String() string }

type Prim struct{ K string }

func (p *Prim) String() string { return p.K }

var (
	TString = &Prim{"string"}
	TNumber = &Prim{"number"}
	TBool   = &Prim{"boolean"}
	TVoid   = &Prim{"void"}
	TUndef  = &Prim{"undefined"}
	TAny    = &Prim{"any"}
	TNode   = &Prim{"Node"}
	TRegExp = &Prim{"RegExp"}
)

type ArrT struct{ Elem Type }

func (t *ArrT) String() string { return t.Elem.String() + "[]" }

type MapT struct{ Val Type }

func (t *MapT) String() string { return "Record<string, " + t.Val.String() + ">" }

type Field struct {
	Name     string
	Go       string
	GoType   string // 宿主字段的 Go 类型(数值转换用)
	Type     Type
	Optional bool
}

type ObjT struct {
	Name    string // TS 名; 匿名为空
	GoName  string // Go 类型名(含包前缀)
	Fields  []*Field
	Methods map[string]*HostMember // 宿主类型的方法
	Host    bool
	Anon    bool
	JSON    bool // 生成 json tag(岛 props)
	Pos     Pos  // interface 声明位置(宿主类型为空)
}

func (t *ObjT) String() string {
	if t.Name != "" {
		return t.Name
	}
	var ps []string
	for _, f := range t.Fields {
		ps = append(ps, f.Name+": "+f.Type.String())
	}
	return "{ " + strings.Join(ps, "; ") + " }"
}
func (t *ObjT) Field(name string) *Field {
	for _, f := range t.Fields {
		if f.Name == name {
			return f
		}
	}
	return nil
}

type FnParam struct {
	Name     string
	Type     Type
	Optional bool
}
type FnT struct {
	Params []*FnParam
	Ret    Type
	Async  bool
}

func (t *FnT) String() string {
	var ps []string
	for _, p := range t.Params {
		ps = append(ps, p.Name+": "+p.Type.String())
	}
	return "(" + strings.Join(ps, ", ") + ") => " + t.Ret.String()
}

type OptT struct{ Elem Type }

func (t *OptT) String() string { return t.Elem.String() + " | undefined" }

// PromiseT: the result type of an action call (await toggle(id) inside an island)
type PromiseT struct{ Elem Type }

func (t *PromiseT) String() string {
	if t.Elem == nil {
		return "Promise<void>"
	}
	return "Promise<" + t.Elem.String() + ">"
}

type SetterT struct{ Elem Type }

func (t *SetterT) String() string { return "Setter<" + t.Elem.String() + ">" }

// StoreT: `export const cart = createStore(init)` in a client module — in the browser a set of signals (one per
// top-level field), on the server the per-request seeded value. Sym is the module-level const; its Go name is
// the store's unique key (the Go variable, the JSON key of the seed block).
type StoreT struct {
	State *ObjT
	Sym   *Symbol
}

func (t *StoreT) String() string { return "Store<" + t.State.String() + ">" }

// StoreSetT: store.set — (draft: T) => void mutator or a whole value; client handlers/effects only
type StoreSetT struct{ Store *StoreT }

func (t *StoreSetT) String() string { return "Store<" + t.Store.State.String() + ">.set" }

type CompT struct {
	Name   string
	Props  *ObjT
	Island bool
	Sym    *Symbol
}

func (t *CompT) String() string { return "Component<" + t.Name + ">" }

// 宿主
type HostParam struct {
	Name string // Go parameter name from the source (hostgen); "" when unknown
	Type Type
	Go   string
}
type HostMember struct {
	Name   string
	Go     string
	Kind   string // field | method
	Type   Type   // field 的类型
	Params []HostParam
	Ret    Type // nil = void
	GoRet  string
	Throws bool
	File   string // 方法的 Go 源码位置(hostgen 反射得到; 字段没有)
	Line   int
	Action bool   // callable from islands
	Req    bool   // the first Go parameter is *gotsx.Req (injected by the runtime)
	Mod    string // owning host module name (used in the action URL)
}
type HostFnT struct{ M *HostMember }

func (t *HostFnT) String() string { return "host function " + t.M.Name }

type HostModT struct {
	Name    string
	Go      string
	Members map[string]*HostMember
}

func (t *HostModT) String() string { return "host:" + t.Name }

// 内建: console / JSON / Math 这类命名空间, 以及数组/字符串上的方法
type GlobalT struct{ Name string }

func (t *GlobalT) String() string { return t.Name }

type BuiltinT struct {
	Recv Type
	Name string
}

func (t *BuiltinT) String() string { return t.Recv.String() + "." + t.Name }

func unopt(t Type) Type {
	if o, ok := t.(*OptT); ok {
		return o.Elem
	}
	return t
}
func isAny(t Type) bool    { return t == TAny }
func isString(t Type) bool { return unopt(t) == TString }
func isNumber(t Type) bool { return unopt(t) == TNumber }
func isBool(t Type) bool   { return unopt(t) == TBool }
func isNode(t Type) bool   { return unopt(t) == TNode }
func sameType(a, b Type) bool {
	if a == b {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	switch x := a.(type) {
	case *ArrT:
		y, ok := b.(*ArrT)
		return ok && sameType(x.Elem, y.Elem)
	case *OptT:
		y, ok := b.(*OptT)
		return ok && sameType(x.Elem, y.Elem)
	case *MapT:
		y, ok := b.(*MapT)
		return ok && sameType(x.Val, y.Val)
	case *ObjT:
		y, ok := b.(*ObjT)
		if !ok {
			return false
		}
		if x.Name != "" || y.Name != "" {
			return x == y
		}
		if len(x.Fields) != len(y.Fields) {
			return false
		}
		for i := range x.Fields {
			if x.Fields[i].Name != y.Fields[i].Name || !sameType(x.Fields[i].Type, y.Fields[i].Type) {
				return false
			}
		}
		return true
	}
	return false
}

// ---------- 符号与作用域 ----------
type SymKind int

const (
	SVar SymKind = iota
	SConst
	SFunc
	SComp
	SType
	SHostMember
	SSignal
	SSetter
	SMemo
	SBuiltin
	SParam
)

type Symbol struct {
	Name   string
	Kind   SymKind
	Type   Type
	Go     string // Go 里的名字/表达式
	Module *Module
	Host   *HostMember
	Comp   *CompT
	Async  bool
	Pos    Pos     // 声明位置(编辑器跳转用); 内建为空
	Store  *StoreT // a field destructured from a store (const { count } = cart): a signal that only store.set may change
}

type Scope struct {
	parent *Scope
	syms   map[string]*Symbol
	types  map[string]Type
	fn     *fnCtx
}

type fnCtx struct {
	decl      *FuncDecl // the function declaration (nil for arrows)
	component bool
	async     bool
	declared  bool // the return type was written down (returns are checked against want)
	rets      []Type
	want      Type
	loops     int // 当前嵌套的循环层数(break/continue 合法性)
	switches  int // 当前嵌套的 switch 层数(break 合法性)
}

func newScope(parent *Scope) *Scope {
	return &Scope{parent: parent, syms: map[string]*Symbol{}, types: map[string]Type{}}
}
func (s *Scope) lookup(name string) *Symbol {
	for sc := s; sc != nil; sc = sc.parent {
		if sym, ok := sc.syms[name]; ok {
			return sym
		}
	}
	return nil
}
func (s *Scope) lookupType(name string) Type {
	for sc := s; sc != nil; sc = sc.parent {
		if t, ok := sc.types[name]; ok {
			return t
		}
	}
	return nil
}
func (s *Scope) fnCtx() *fnCtx {
	for sc := s; sc != nil; sc = sc.parent {
		if sc.fn != nil {
			return sc.fn
		}
	}
	return nil
}

type CheckError struct {
	Pos Pos
	Msg string
}

func (e *CheckError) Error() string { return fmt.Sprintf("%s: %s", e.Pos, e.Msg) }
