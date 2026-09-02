package compiler

import "fmt"

type Pos struct {
	File string
	Line int
	Col  int
}

func (p Pos) String() string { return fmt.Sprintf("%s:%d:%d", p.File, p.Line, p.Col) }

// Module: 一个 .tsx 文件
type Module struct {
	File    string
	Dir     string
	Name    string // 文件名去掉 .server/.client/.tsx
	Kind    string // server | client | shared
	Imports []*Import
	Stmts   []Stmt

	// checker 填
	Exports   map[string]*Symbol
	Default   *Symbol
	Deps      []*Module
	Scope     *Scope
	Checked   bool
	Checking  bool
	AnonTypes []*ObjT
	GoPrefix  string
}

type Import struct {
	Pos      Pos
	From     string
	Default  string
	Names    []ImportSpec
	TypeOnly bool
}
type ImportSpec struct {
	Name, Local string
	TypeOnly    bool
}

// ---------- 语句 ----------
type Stmt interface{ stmtNode() }

type FuncDecl struct {
	Pos     Pos
	Name    string
	Params  []*Param
	Ret     TypeExpr
	Body    *Block
	Async   bool
	Export  bool
	Default bool
	// checker
	Sym   *Symbol
	Comp  bool // 组件
	RetT  Type
	Scope *Scope
}
type Param struct {
	Pos      Pos
	Pat      *Pattern
	Type     TypeExpr
	Optional bool
	Default  Expr
	T        Type
}
type VarDecl struct {
	Pos    Pos
	Const  bool
	Pat    *Pattern
	Type   TypeExpr
	Init   Expr
	Export bool
	Hook   string // checker: useState / memo / ""
}
type ReturnStmt struct {
	Pos Pos
	X   Expr
}
type IfStmt struct {
	Pos  Pos
	Cond Expr
	Then *Block
	Else Stmt
}
type ForOfStmt struct {
	Pos  Pos
	Pat  *Pattern
	Iter Expr
	Body *Block
}
type ExprStmt struct {
	Pos Pos
	X   Expr
}
type Block struct {
	Pos   Pos
	Stmts []Stmt
}
type InterfaceDecl struct {
	Pos    Pos
	Name   string
	Fields []*TypeField
	Export bool
	T      *ObjT
}
type TypeAlias struct {
	Pos    Pos
	Name   string
	Type   TypeExpr
	Export bool
	T      Type
}
type TryStmt struct {
	Pos       Pos
	Body      *Block
	CatchName string
	Catch     *Block
	Finally   *Block
}
type ThrowStmt struct {
	Pos Pos
	X   Expr
}
type ExportDefault struct {
	Pos Pos
	X   Expr
}
type EmptyStmt struct{ Pos Pos }

func (*FuncDecl) stmtNode()      {}
func (*VarDecl) stmtNode()       {}
func (*ReturnStmt) stmtNode()    {}
func (*IfStmt) stmtNode()        {}
func (*ForOfStmt) stmtNode()     {}
func (*ExprStmt) stmtNode()      {}
func (*Block) stmtNode()         {}
func (*InterfaceDecl) stmtNode() {}
func (*TypeAlias) stmtNode()     {}
func (*TryStmt) stmtNode()       {}
func (*ThrowStmt) stmtNode()     {}
func (*ExportDefault) stmtNode() {}
func (*EmptyStmt) stmtNode()     {}

// 解构模式
const (
	PatIdent = iota
	PatObject
	PatArray
)

type Pattern struct {
	Pos   Pos
	Kind  int
	Name  string
	Props []*PatProp
	Elems []*Pattern
	Rest  *Pattern
	// checker
	Sym *Symbol
	T   Type
}
type PatProp struct {
	Key     string
	Pat     *Pattern
	Default Expr
}

// ---------- 表达式 ----------
type Expr interface {
	exprNode()
	GetPos() Pos
	T() Type
	SetT(Type)
}

type base struct {
	Pos Pos
	typ Type
}

func (b *base) exprNode()   {}
func (b *base) GetPos() Pos { return b.Pos }
func (b *base) T() Type     { return b.typ }
func (b *base) SetT(t Type) { b.typ = t }

type Ident struct {
	base
	Name string
	Sym  *Symbol
}
type NumLit struct {
	base
	Val float64
	Raw string
}
type StrLit struct {
	base
	Val string
}
type BoolLit struct {
	base
	Val bool
}
type NullLit struct{ base } // null / undefined
type TemplateLit struct {
	base
	Quasis []string
	Exprs  []Expr
}
type ArrayLit struct {
	base
	Elems []Expr
}
type ObjectLit struct {
	base
	Props []*ObjProp
}
type ObjProp struct {
	Key       string
	Val       Expr
	Spread    Expr
	Shorthand bool
}
type Member struct {
	base
	X        Expr
	Name     string
	Optional bool
	// checker
	Builtin string // 数组/字符串内建方法名
	GoName  string // 宿主/结构体字段的 Go 名
	MapKey  bool   // Record 的键访问
}
type Index struct {
	base
	X, I Expr
}
type Call struct {
	base
	Fn       Expr
	Args     []Expr
	TypeArgs []TypeExpr
	Optional bool
	// checker
	Kind     string // "builtin:<name>" | "host" | "fn" | "comp" | "hook:<name>" | "global:<name>"
	Host     *HostMember
	Reactive bool
}
type Unary struct {
	base
	Op string
	X  Expr
}
type Binary struct {
	base
	Op       string
	L, R     Expr
	Reactive bool
}
type CondExpr struct {
	base
	Test, Then, Else Expr
	Reactive         bool
}
type Arrow struct {
	base
	Params   []*Param
	Body     *Block
	ExprBody Expr
	Async    bool
	Ret      TypeExpr
	RetT     Type
	Scope    *Scope
}
type Assign struct {
	base
	Op          string
	Target, Val Expr
}
type AsExpr struct {
	base
	X    Expr
	Type TypeExpr
}
type AwaitExpr struct {
	base
	X Expr
}
type SpreadExpr struct {
	base
	X Expr
}
type NonNull struct {
	base
	X Expr
}
type Paren struct {
	base
	X Expr
}
type JSXElem struct {
	base
	Tag      string
	Attrs    []*JSXAttr
	Children []Expr
	Comp     *Symbol // 大写标签: 组件
}
type JSXAttr struct {
	Name     string
	Val      Expr // nil = 布尔 true
	Spread   Expr
	Reactive bool
}
type JSXFrag struct {
	base
	Children []Expr
}
type JSXText struct {
	base
	Text string
}
type JSXExprChild struct {
	base
	X        Expr // nil: {} 或注释
	Reactive bool
}

// ---------- 类型表达式 ----------
type TypeExpr interface{ typeNode() }
type TRef struct {
	Pos  Pos
	Name string
	Args []TypeExpr
}
type TArr struct{ Elem TypeExpr }
type TObj struct{ Fields []*TypeField }
type TUnion struct{ Members []TypeExpr }
type TStrLit struct{ Val string }
type TFunc struct {
	Params []*Param
	Ret    TypeExpr
}
type TypeField struct {
	Name     string
	Type     TypeExpr
	Optional bool
	Method   bool
	Params   []*Param
}

func (*TRef) typeNode()    {}
func (*TArr) typeNode()    {}
func (*TObj) typeNode()    {}
func (*TUnion) typeNode()  {}
func (*TStrLit) typeNode() {}
func (*TFunc) typeNode()   {}
