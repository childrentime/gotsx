package compiler

// 编辑器查询: hover / 跳转定义。基于已检查的模块 AST(Ident.Sym / Member 类型 / JSXElem.Comp 都在检查时填好),
// 按位置找最内层的节点, 再从符号取类型与声明位置。宿主方法跳到 Go 源码(hostgen 反射出的 file:line),
// 宿主类型 / 字段跳到 app/.gen/host.d.ts 里的那一行。

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Location: 定义位置(行列从 1 开始)
type Location struct {
	File string
	Line int
	Col  int
}

// Hover: 悬停信息(Markdown)+ 可选的定义位置
type Hover struct {
	Text string
	Def  *Location
}

// Load: 解析 + 类型检查(不生成), 返回 checker 与前端诊断。LSP 用它做 hover / definition; Analyze 在它之上跑两个后端。
func Load(appDir string, overlay map[string]string) (*Checker, []Diagnostic, error) {
	return load(appDir, overlay)
}

// ---------- 定位 ----------

type hit struct {
	ident  *Ident
	pat    *Pattern
	member *Member
	jsx    *JSXElem
	tref   *TRef
	width  int
}

// covers: (line, col) 是否落在从 pos 开始、长 n 的 token 上
func covers(pos Pos, n, line, col int) bool {
	return pos.Line == line && col >= pos.Col && col <= pos.Col+n
}

// nodeAt: 该位置上的标识符 / 成员名 / JSX 标签 / 类型引用
func nodeAt(m *Module, line, col int) *hit {
	var best *hit
	take := func(h *hit) {
		if best == nil || h.width <= best.width {
			best = h
		}
	}
	walkModule(m, func(n any) {
		switch x := n.(type) {
		case *Ident:
			if covers(x.Pos, len(x.Name), line, col) {
				take(&hit{ident: x, width: len(x.Name)})
			}
		case *Pattern: // 声明处: const x / 参数 / 解构出的名字
			if x.Kind == PatIdent && x.Sym != nil && covers(x.Pos, len(x.Name), line, col) {
				take(&hit{pat: x, width: len(x.Name)})
			}
		case *Member:
			off := 1 // "."
			if x.Optional {
				off = 2 // "?."
			}
			start := Pos{File: x.Pos.File, Line: x.Pos.Line, Col: x.Pos.Col + off}
			if covers(start, len(x.Name), line, col) {
				take(&hit{member: x, width: len(x.Name)})
			}
		case *JSXElem:
			start := Pos{File: x.Pos.File, Line: x.Pos.Line, Col: x.Pos.Col + 1}
			if x.Comp != nil && covers(start, len(x.Tag), line, col) {
				take(&hit{jsx: x, width: len(x.Tag)})
			}
		case *TRef:
			if covers(x.Pos, len(x.Name), line, col) {
				take(&hit{tref: x, width: len(x.Name)})
			}
		}
	})
	return best
}

// walkModule: 深度优先遍历模块的全部语句、表达式与类型标注
func walkModule(m *Module, f func(any)) {
	for _, s := range m.Stmts {
		walkStmt(s, f)
	}
}

func walkStmt(s Stmt, f func(any)) {
	if s == nil {
		return
	}
	f(s)
	switch d := s.(type) {
	case *FuncDecl:
		for _, p := range d.Params {
			walkParam(p, f)
		}
		walkType(d.Ret, f)
		walkBlock(d.Body, f)
	case *VarDecl:
		walkPattern(d.Pat, f)
		walkType(d.Type, f)
		walkE(d.Init, f)
	case *ReturnStmt:
		walkE(d.X, f)
	case *IfStmt:
		walkE(d.Cond, f)
		walkBlock(d.Then, f)
		walkStmt(d.Else, f)
	case *ForOfStmt:
		walkPattern(d.Pat, f)
		walkE(d.Iter, f)
		walkBlock(d.Body, f)
	case *ForStmt:
		walkStmt(d.Init, f)
		walkE(d.Cond, f)
		walkE(d.Update, f)
		walkBlock(d.Body, f)
	case *WhileStmt:
		walkE(d.Cond, f)
		walkBlock(d.Body, f)
	case *SwitchStmt:
		walkE(d.Disc, f)
		for _, cs := range d.Cases {
			walkE(cs.Test, f)
			for _, st := range cs.Body {
				walkStmt(st, f)
			}
		}
	case *ExprStmt:
		walkE(d.X, f)
	case *Block:
		walkBlock(d, f)
	case *InterfaceDecl:
		for _, fld := range d.Fields {
			walkType(fld.Type, f)
			for _, p := range fld.Params {
				walkParam(p, f)
			}
		}
	case *TypeAlias:
		walkType(d.Type, f)
	case *TryStmt:
		walkBlock(d.Body, f)
		walkBlock(d.Catch, f)
		walkBlock(d.Finally, f)
	case *ThrowStmt:
		walkE(d.X, f)
	case *ExportDefault:
		walkE(d.X, f)
	}
}

func walkBlock(b *Block, f func(any)) {
	if b == nil {
		return
	}
	for _, s := range b.Stmts {
		walkStmt(s, f)
	}
}

func walkParam(p *Param, f func(any)) {
	if p == nil {
		return
	}
	walkPattern(p.Pat, f)
	walkType(p.Type, f)
	walkE(p.Default, f)
}

func walkPattern(p *Pattern, f func(any)) {
	if p == nil {
		return
	}
	f(p)
	for _, pp := range p.Props {
		walkPattern(pp.Pat, f)
		walkE(pp.Default, f)
	}
	for _, e := range p.Elems {
		walkPattern(e, f)
	}
	walkPattern(p.Rest, f)
}

func walkType(t TypeExpr, f func(any)) {
	if t == nil {
		return
	}
	f(t)
	switch x := t.(type) {
	case *TRef:
		for _, a := range x.Args {
			walkType(a, f)
		}
	case *TArr:
		walkType(x.Elem, f)
	case *TObj:
		for _, fld := range x.Fields {
			walkType(fld.Type, f)
			for _, p := range fld.Params {
				walkParam(p, f)
			}
		}
	case *TUnion:
		for _, m := range x.Members {
			walkType(m, f)
		}
	case *TFunc:
		for _, p := range x.Params {
			walkParam(p, f)
		}
		walkType(x.Ret, f)
	}
}

// walkE: 表达式(含箭头函数体、JSX 属性与子节点)
func walkE(e Expr, f func(any)) {
	if e == nil {
		return
	}
	walkExpr(e, func(x Expr) bool {
		f(x)
		switch y := x.(type) {
		case *Arrow:
			for _, p := range y.Params {
				walkParam(p, f)
			}
			walkType(y.Ret, f)
			walkBlock(y.Body, f)
		case *AsExpr:
			walkType(y.Type, f)
		case *Call:
			for _, ta := range y.TypeArgs {
				walkType(ta, f)
			}
		}
		return true
	})
}

// ---------- hover / definition ----------

// HoverAt: 位置上的符号说明(Markdown); 没有则 nil
func (c *Checker) HoverAt(file string, line, col int) *Hover {
	m := c.Modules[file]
	if m == nil {
		return nil
	}
	saveMod := c.mod // hover text depends on the side (an action is a Promise in an island)
	c.mod = m
	defer func() { c.mod = saveMod }()
	h := nodeAt(m, line, col)
	if h == nil {
		return c.importHover(m, file, line, col)
	}
	switch {
	case h.ident != nil:
		return c.symbolHover(h.ident.Sym, h.ident.Name)
	case h.pat != nil:
		return c.symbolHover(h.pat.Sym, h.pat.Name)
	case h.member != nil:
		return c.memberHover(h.member)
	case h.jsx != nil:
		return c.symbolHover(h.jsx.Comp, h.jsx.Tag)
	case h.tref != nil:
		return c.typeHover(m, h.tref)
	}
	return nil
}

// DefinitionAt: 位置上符号的声明位置; 没有则 nil
func (c *Checker) DefinitionAt(file string, line, col int) *Location {
	if h := c.HoverAt(file, line, col); h != nil {
		return h.Def
	}
	return nil
}

func code(s string) string { return "```ts\n" + s + "\n```" }

var builtinDocs = map[string]string{
	"useState":  "useState<T>(init: T): [T, (v: T | ((prev: T) => T)) => void]\n\nServer: the initial value (single-pass SSR). Client: a signal; a const that reads it is a memo.",
	"useEffect": "useEffect(fn: () => void, deps?: []): void\n\nClient only. Dependencies are tracked automatically; deps [] runs once on mount.",
	"useMemo":   "useMemo<T>(fn: () => T): T\n\nExplicit memo. Usually unnecessary: a signal-dependent const is memoized automatically.",
	"emit":      "emit(name: string, detail?: unknown): void\n\nCross-island event bus (client only).",
	"on":        "on(name: string, fn: (detail: any) => void): void\n\nSubscribe to a cross-island event (client only).",
	"Suspense":  "<Suspense fallback={Node}>children</Suspense>\n\nStreaming boundary (server only): the fallback ships with the shell, children render in their own goroutine and are streamed in.",
	"redirect":  "redirect(url: string, status?: number): never\n\nServer pages only: abort the render and answer with a 3xx (default 302).",
	"notFound":  "notFound(): never\n\nServer pages only: abort the render and answer with the 404 page.",
	"jsonLd":    "jsonLd(json: string): Node\n\nA safe <script type=\"application/ld+json\"> (pass JSON.stringify(...)).",
	"t":         "t(locale: string, key: string): string\n\ni18n lookup with fallback to the default locale.",
	"tv":        "tv(locale: string, key: string, vars: Record<string, string>): string\n\ni18n lookup with {name} interpolation.",
	"plural":    "plural(locale: string, key: string, n: number): string\n\nCLDR-lite plural (\"one|other\" forms, {n} placeholder).",
	"fmtNum":    "fmtNum(locale: string, n: number): string",
	"fmtCur":    "fmtCur(locale: string, cents: number): string",
	"fmtDate":   "fmtDate(locale: string, iso: string): string",
	"lpath":     "lpath(locale: string, path: string): string\n\nAdds the locale prefix in URL-prefix mode.",
	"isoDate":   "isoDate(ms: number): string\n\nMilliseconds → RFC 3339 (UTC), identical on both sides.",
}

func (c *Checker) symbolHover(s *Symbol, name string) *Hover {
	if s == nil {
		return nil
	}
	var text string
	switch s.Kind {
	case SBuiltin:
		if d, ok := builtinDocs[s.Name]; ok {
			text = code(d)
		} else if s.Type != nil {
			text = code("(global) " + s.Name + ": " + s.Type.String())
		} else {
			text = code("(builtin) " + s.Name)
		}
		return &Hover{Text: text}
	case SComp:
		if s.Comp != nil && s.Comp.Props != nil {
			var ps []string
			for _, f := range s.Comp.Props.Fields {
				opt := ""
				if f.Optional {
					opt = "?"
				}
				ps = append(ps, f.Name+opt+": "+f.Type.String())
			}
			kind := "component"
			if s.Comp.Island {
				kind = "island"
			}
			text = code(fmt.Sprintf("%s %s({ %s })", kind, s.Name, strings.Join(ps, "; ")))
		} else {
			text = code("component " + s.Name)
		}
	case SFunc:
		text = code("function " + s.Name + typeSuffix(s.Type))
	case SHostMember:
		if s.Host != nil && s.Host.Kind == "method" {
			text = code("(host method) " + s.Name + c.hostSig(s.Host))
		} else {
			text = code("(host) " + s.Name + ": " + typeStr(s.Type))
		}
	case SSignal:
		text = code("(state) " + s.Name + ": " + typeStr(s.Type) + "   // useState: a signal on the client, the initial value on the server")
	case SSetter:
		text = code("(setter) " + s.Name + ": " + typeStr(s.Type))
	case SMemo:
		text = code("(memo) " + s.Name + ": " + typeStr(s.Type) + "   // reads a signal → recomputed automatically")
	case SParam:
		text = code("(parameter) " + s.Name + ": " + typeStr(s.Type))
	case SConst:
		text = code("const " + s.Name + ": " + typeStr(s.Type))
	default:
		text = code("let " + s.Name + ": " + typeStr(s.Type))
	}
	return &Hover{Text: text, Def: c.symbolDef(s)}
}

func typeStr(t Type) string {
	if t == nil {
		return "any"
	}
	return t.String()
}

func typeSuffix(t Type) string {
	if ft, ok := t.(*FnT); ok {
		var ps []string
		for _, p := range ft.Params {
			ps = append(ps, p.Name+": "+typeStr(p.Type))
		}
		return "(" + strings.Join(ps, ", ") + "): " + typeStr(ft.Ret)
	}
	return ": " + typeStr(t)
}

// hostSig: the signature shown on hover; in an island an action is asynchronous (Promise<T>) and its errors are HTTP statuses
func (c *Checker) hostSig(m *HostMember) string {
	var ps []string
	for i, p := range m.Params {
		ps = append(ps, fmt.Sprintf("%s: %s", hostParamName(p, i), typeStr(p.Type)))
	}
	ret := "void"
	if m.Ret != nil {
		ret = m.Ret.String()
	}
	throws := ""
	if m.Throws {
		throws = "   // (T, error): an error becomes a 500 (ErrNotFound → 404)"
	}
	if m.Action && c.mod != nil && c.mod.Kind != "server" {
		ret = "Promise<" + ret + ">"
		throws = "   // action: POST /_gotsx/act/" + m.Mod + "/" + m.Name + "; errors throw with .status / .fields"
	}
	return "(" + strings.Join(ps, ", ") + "): " + ret + throws
}

func (c *Checker) symbolDef(s *Symbol) *Location {
	if s.Host != nil {
		return c.hostDef(s.Host, s.Name)
	}
	if s.Module != nil && s.Pos.Line > 0 {
		return &Location{File: s.Pos.File, Line: s.Pos.Line, Col: s.Pos.Col}
	}
	return nil
}

// hostDef: 宿主方法 → Go 源码; 其它 → host.d.ts 里含该名字的第一行
func (c *Checker) hostDef(m *HostMember, name string) *Location {
	if m != nil && m.File != "" {
		return &Location{File: m.File, Line: m.Line, Col: 1}
	}
	return c.dtsDef(name)
}

func (c *Checker) dtsDef(name string) *Location {
	if c.HostDTS == "" {
		return nil
	}
	b, err := os.ReadFile(c.HostDTS)
	if err != nil {
		return nil
	}
	for i, ln := range strings.Split(string(b), "\n") {
		if strings.Contains(ln, " "+name+"(") || strings.Contains(ln, " "+name+":") || strings.Contains(ln, "interface "+name+" ") {
			return &Location{File: c.HostDTS, Line: i + 1, Col: strings.Index(ln, name) + 1}
		}
	}
	return nil
}

func (c *Checker) memberHover(x *Member) *Hover {
	rt := unopt(x.X.T())
	switch t := rt.(type) {
	case *ObjT:
		if f := t.Field(x.Name); f != nil {
			var def *Location
			if t.Host {
				def = c.dtsDef(x.Name)
			} else if t.Pos.Line > 0 {
				def = &Location{File: t.Pos.File, Line: t.Pos.Line, Col: t.Pos.Col}
			}
			return &Hover{Text: code("(field) " + t.String() + "." + x.Name + ": " + typeStr(f.Type)), Def: def}
		}
		if m, ok := t.Methods[x.Name]; ok {
			return &Hover{Text: code("(host method) " + t.String() + "." + x.Name + c.hostSig(m)), Def: c.hostDef(m, x.Name)}
		}
	case *HostModT:
		if m, ok := t.Members[x.Name]; ok {
			if m.Kind == "method" {
				return &Hover{Text: code("(host method) " + x.Name + c.hostSig(m)), Def: c.hostDef(m, x.Name)}
			}
			return &Hover{Text: code("(host) " + x.Name + ": " + typeStr(m.Type)), Def: c.hostDef(m, x.Name)}
		}
	case *MapT:
		return &Hover{Text: code("(record key) " + x.Name + ": " + typeStr(t.Val) + "   // zero value when absent; Object.hasOwn(m, key) tests presence")}
	case *ArrT, *Prim:
		if x.Builtin != "" {
			return &Hover{Text: code("(builtin method) " + rt.String() + "." + x.Name + ": " + typeStr(x.T()))}
		}
	case *GlobalT:
		return &Hover{Text: code("(builtin) " + t.Name + "." + x.Name)}
	}
	return nil
}

func (c *Checker) typeHover(m *Module, r *TRef) *Hover {
	var t Type
	if m.Scope != nil {
		t = m.Scope.lookupType(r.Name)
	}
	if t == nil {
		t = c.global.lookupType(r.Name)
	}
	if t == nil {
		if ht, ok := c.Host.Types[r.Name]; ok {
			t = ht
		}
	}
	if t == nil {
		return nil
	}
	if o, ok := t.(*ObjT); ok {
		var fs []string
		for _, f := range o.Fields {
			opt := ""
			if f.Optional {
				opt = "?"
			}
			fs = append(fs, "  "+f.Name+opt+": "+typeStr(f.Type)+";")
		}
		text := code("interface " + o.Name + " {\n" + strings.Join(fs, "\n") + "\n}")
		var def *Location
		if o.Host {
			def = c.dtsDef(o.Name)
		} else if o.Pos.Line > 0 {
			def = &Location{File: o.Pos.File, Line: o.Pos.Line, Col: o.Pos.Col}
		}
		return &Hover{Text: text, Def: def}
	}
	return &Hover{Text: code("type " + r.Name + " = " + t.String())}
}

// importHover: 光标在 import 行的某个名字上 → 当作该模块作用域里的符号
func (c *Checker) importHover(m *Module, file string, line, col int) *Hover {
	for _, im := range m.Imports {
		if im.Pos.Line != line {
			continue
		}
		word := wordAt(file, line, col)
		if word == "" {
			return nil
		}
		if sym := m.Scope.lookup(word); sym != nil {
			return c.symbolHover(sym, word)
		}
		if t := m.Scope.lookupType(word); t != nil {
			return c.typeHover(m, &TRef{Name: word})
		}
	}
	return nil
}

// wordAt: 源文件该位置上的标识符(读磁盘; LSP 传的是同一份内容)
func wordAt(file string, line, col int) string {
	b, err := os.ReadFile(file)
	if err != nil {
		return ""
	}
	lines := strings.Split(string(b), "\n")
	if line < 1 || line > len(lines) {
		return ""
	}
	r := []rune(lines[line-1])
	i := col - 1
	if i < 0 || i > len(r) {
		return ""
	}
	s, e := i, i
	for s > 0 && isIdentPart(r[s-1]) {
		s--
	}
	for e < len(r) && isIdentPart(r[e]) {
		e++
	}
	return string(r[s:e])
}

// HostDTSPath: app/.gen/host.d.ts 的路径(给跳转用)
func HostDTSPath(appDir string) string { return filepath.Join(appDir, ".gen", "host.d.ts") }
