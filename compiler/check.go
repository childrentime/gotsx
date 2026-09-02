package compiler

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

// goRegex: JS 正则 + flags → RE2 语法(i/m/s 变成前缀; g 由运行时处理; u 无操作)
func goRegex(pattern, flags string) string {
	pre := ""
	for _, f := range flags {
		switch f {
		case 'i', 'm', 's':
			pre += string(f)
		}
	}
	if pre != "" {
		return "(?" + pre + ")" + pattern
	}
	return pattern
}

// Checker: 子集的类型检查。目标是给 Go 后端足够的静态类型, 出子集就报带位置的错误。
type Checker struct {
	Modules     map[string]*Module
	Host        *HostInfo
	Errors      []error
	global      *Scope
	mod         *Module
	scope       *Scope
	pageProps   *ObjT
	layoutProps *ObjT  // _layout.server.tsx: PageProps + children
	errorProps  *ObjT  // _error.server.tsx: PageProps + message
	HostDTS     string // app/.gen/host.d.ts(编辑器跳转宿主类型用)
}

type HostInfo struct {
	Modules map[string]*HostModT
	Types   map[string]*ObjT
}

func (c *Checker) errf(pos Pos, format string, args ...any) {
	c.Errors = append(c.Errors, &CheckError{Pos: pos, Msg: fmt.Sprintf(format, args...)})
}

// abort: 无法继续检查当前模块时抛出
type abortCheck struct{}

func (c *Checker) fatal(pos Pos, format string, args ...any) {
	c.errf(pos, format, args...)
	panic(abortCheck{})
}

func NewChecker(hostJSON []byte) (*Checker, error) {
	c := &Checker{Modules: map[string]*Module{}}
	c.global = newScope(nil)
	for _, n := range []string{"useState", "useEffect", "useMemo", "String", "Number", "Boolean", "parseInt", "parseFloat",
		"encodeURIComponent", "decodeURIComponent", "fetch", "setTimeout", "clearTimeout", "setInterval", "clearInterval",
		"emit", "on", "alert", "confirm", "isNaN", "jsonLd", "redirect", "notFound",
		"t", "tv", "plural", "fmtNum", "fmtCur", "fmtDate", "lpath"} {
		c.global.syms[n] = &Symbol{Name: n, Kind: SBuiltin}
	}
	for _, n := range []string{"console", "JSON", "Math", "Object", "Date"} {
		c.global.syms[n] = &Symbol{Name: n, Kind: SBuiltin, Type: &GlobalT{Name: n}}
	}
	c.global.syms["isoDate"] = &Symbol{Name: "isoDate", Kind: SBuiltin}
	// 浏览器对象: 只在客户端代码里可用, 类型 any(链式访问随便写)
	for _, n := range []string{"document", "window", "location", "history", "navigator", "localStorage", "sessionStorage", "requestAnimationFrame", "Array"} {
		c.global.syms[n] = &Symbol{Name: n, Kind: SBuiltin, Type: TAny}
	}
	c.pageProps = &ObjT{Name: "PageProps", GoName: "gotsx.PageProps", Fields: []*Field{
		{Name: "params", Go: "Params", Type: &MapT{Val: TString}},
		{Name: "query", Go: "Query", Type: &MapT{Val: TString}},
		{Name: "path", Go: "Path", Type: TString},
		{Name: "locale", Go: "Locale", Type: TString},
		{Name: "cookies", Go: "Cookies", Type: &MapT{Val: TString}},
	}}
	c.global.types["PageProps"] = c.pageProps
	c.global.types["Node"] = TNode
	c.layoutProps = &ObjT{Name: "LayoutProps", GoName: "gotsx.LayoutProps", Fields: append(append([]*Field{}, c.pageProps.Fields...),
		&Field{Name: "children", Go: "Children", Type: TNode})}
	c.errorProps = &ObjT{Name: "ErrorProps", GoName: "gotsx.ErrorProps", Fields: append(append([]*Field{}, c.pageProps.Fields...),
		&Field{Name: "message", Go: "Message", Type: TString})}
	c.global.types["LayoutProps"] = c.layoutProps
	c.global.types["ErrorProps"] = c.errorProps
	// <Suspense fallback={…}>…</Suspense>: 流式 SSR 边界(服务端组件专用), Go 侧是 gotsx.Suspense(fallback, thunk)
	susProps := &ObjT{Name: "SuspenseProps", GoName: "gotsx.SuspenseProps", Fields: []*Field{
		{Name: "fallback", Go: "Fallback", Type: TNode},
		{Name: "children", Go: "Children", Type: &OptT{Elem: TNode}, Optional: true},
	}}
	sus := &Symbol{Name: "Suspense", Kind: SComp, Go: "gotsx.Suspense"}
	sus.Comp = &CompT{Name: "Suspense", Props: susProps, Sym: sus}
	sus.Type = sus.Comp
	c.global.syms["Suspense"] = sus
	if err := c.loadHost(hostJSON); err != nil {
		return nil, err
	}
	return c, nil
}

// ---------- 宿主 ----------

type hostJSON struct {
	Modules map[string]struct {
		Go      string                     `json:"go"`
		Members map[string]*hostMemberJSON `json:"members"`
	} `json:"modules"`
	Types map[string]struct {
		Go     string `json:"go"`
		Fields []struct {
			Name, Go, Type string
			GoType         string `json:"gotype"`
		} `json:"fields"`
		Methods map[string]*hostMemberJSON `json:"methods"`
	} `json:"types"`
}
type hostMemberJSON struct {
	Kind   string `json:"kind"`
	Go     string `json:"go"`
	Type   string `json:"type"`
	Params []struct{ Type, Go string }
	Ret    *struct{ Type, Go string }
	Throws bool
	File   string `json:"file"`
	Line   int    `json:"line"`
}

func (c *Checker) loadHost(data []byte) error {
	c.Host = &HostInfo{Modules: map[string]*HostModT{}, Types: map[string]*ObjT{}}
	if len(data) == 0 {
		return nil
	}
	var hj hostJSON
	if err := json.Unmarshal(data, &hj); err != nil {
		return fmt.Errorf("host.json: %w", err)
	}
	for name, tj := range hj.Types {
		c.Host.Types[name] = &ObjT{Name: name, GoName: tj.Go, Host: true, Methods: map[string]*HostMember{}}
	}
	hostScope := newScope(c.global)
	for name, t := range c.Host.Types {
		hostScope.types[name] = t
	}
	c.scope = hostScope
	parseT := func(s string) Type {
		te, err := parseTypeString(s)
		if err != nil {
			return TAny
		}
		return c.resolveType(te, Pos{})
	}
	mem := func(name string, mj *hostMemberJSON) *HostMember {
		m := &HostMember{Name: name, Go: mj.Go, Kind: mj.Kind, Throws: mj.Throws, File: mj.File, Line: mj.Line}
		if mj.Kind == "field" {
			m.Type = parseT(mj.Type)
			return m
		}
		for _, p := range mj.Params {
			m.Params = append(m.Params, HostParam{Type: parseT(p.Type), Go: p.Go})
		}
		if mj.Ret != nil {
			m.Ret = parseT(mj.Ret.Type)
			m.GoRet = mj.Ret.Go
		}
		return m
	}
	for name, tj := range hj.Types {
		t := c.Host.Types[name]
		for _, f := range tj.Fields {
			t.Fields = append(t.Fields, &Field{Name: f.Name, Go: f.Go, GoType: f.GoType, Type: parseT(f.Type)})
		}
		for mn, mj := range tj.Methods {
			t.Methods[mn] = mem(mn, mj)
		}
	}
	for name, mj := range hj.Modules {
		hm := &HostModT{Name: name, Go: mj.Go, Members: map[string]*HostMember{}}
		for mn, m := range mj.Members {
			hm.Members[mn] = mem(mn, m)
		}
		c.Host.Modules[name] = hm
	}
	c.scope = nil
	return nil
}

func parseTypeString(s string) (te TypeExpr, err error) {
	defer func() {
		if r := recover(); r != nil {
			if pe, ok := r.(*parseError); ok {
				err = pe
				return
			}
			panic(r)
		}
	}()
	p := newParser(s, "host.json", Pos{})
	return p.parseType(), nil
}

// ---------- 模块 ----------

func (c *Checker) AddModule(m *Module) { c.Modules[m.File] = m }

func (c *Checker) resolveImport(from *Module, spec string, pos Pos) *Module {
	base := filepath.Join(from.Dir, spec)
	for _, cand := range []string{base + ".tsx", base + ".server.tsx", base + ".client.tsx", base, filepath.Join(base, "index.tsx")} {
		if m, ok := c.Modules[cand]; ok {
			return m
		}
	}
	c.fatal(pos, "cannot find module %q", spec)
	return nil
}

// CheckAll 按依赖顺序检查全部模块
func (c *Checker) CheckAll() error {
	files := make([]string, 0, len(c.Modules))
	for f := range c.Modules {
		files = append(files, f)
	}
	sort.Strings(files)
	for _, f := range files {
		c.checkModule(c.Modules[f])
	}
	if len(c.Errors) > 0 {
		var b strings.Builder
		for _, e := range c.Errors {
			b.WriteString("  " + e.Error() + "\n")
		}
		return fmt.Errorf("type check failed:\n%s", b.String())
	}
	return nil
}

func (c *Checker) checkModule(m *Module) {
	if m.Checked {
		return
	}
	if m.Checking {
		c.errf(Pos{File: m.File}, "circular module dependency")
		return
	}
	m.Checking = true
	defer func() {
		m.Checking = false
		m.Checked = true
		if r := recover(); r != nil {
			if _, ok := r.(abortCheck); !ok {
				panic(r)
			}
		}
	}()
	m.Exports = map[string]*Symbol{}
	m.Scope = newScope(c.global)
	m.GoPrefix = goPrefixOf(m)
	saveMod, saveScope := c.mod, c.scope
	c.mod, c.scope = m, m.Scope
	defer func() { c.mod, c.scope = saveMod, saveScope }()

	for _, im := range m.Imports {
		c.importInto(m, im)
	}
	// 第一遍: 声明顶层符号
	for _, s := range m.Stmts {
		c.declareTop(s)
	}
	// 第二遍: 先模块级变量(内容数据), 再函数体
	for _, s := range m.Stmts {
		if d, ok := s.(*VarDecl); ok {
			if d.Pat.Kind != PatIdent {
				c.fatal(d.Pos, "module-level variables cannot be destructured")
			}
			if !d.Const {
				c.fatal(d.Pos, "only const is allowed at module level")
			}
			c.checkStmt(d)
			sym := d.Pat.Sym
			sym.Go = m.GoPrefix + goIdent(d.Pat.Name)
			sym.Module = m
			if d.Export {
				m.Exports[d.Pat.Name] = sym
			}
		}
	}
	for _, s := range m.Stmts {
		switch d := s.(type) {
		case *FuncDecl:
			c.checkFuncBody(d)
		case *ExportDefault:
			c.check(d.X, nil)
		}
	}
}

func (c *Checker) importInto(m *Module, im *Import) {
	switch {
	case im.From == "gotsx":
		for _, n := range im.Names {
			switch n.Name {
			case "useState", "useEffect", "useMemo", "emit", "on", "Suspense":
				m.Scope.syms[n.Local] = c.global.syms[n.Name]
			case "Node", "PageProps", "LayoutProps", "ErrorProps":
				m.Scope.types[n.Local] = c.global.types[n.Name]
			default:
				c.errf(im.Pos, "module gotsx does not export %q", n.Name)
			}
		}
	case strings.HasPrefix(im.From, "host:"):
		if m.Kind == "client" {
			typeOnly := im.TypeOnly
			if !typeOnly {
				typeOnly = len(im.Names) > 0
				for _, n := range im.Names {
					if !n.TypeOnly {
						typeOnly = false
					}
				}
			}
			if !typeOnly {
				c.fatal(im.Pos, "%q can only be imported by server components; client code may only `import type`", im.From)
			}
		}
		name := strings.TrimPrefix(im.From, "host:")
		hm, ok := c.Host.Modules[name]
		if !ok {
			c.fatal(im.Pos, "no host module named %q (check host.json)", im.From)
		}
		for _, n := range im.Names {
			if t, ok := c.Host.Types[n.Name]; ok && (n.TypeOnly || im.TypeOnly || hm.Members[n.Name] == nil) {
				m.Scope.types[n.Local] = t
				continue
			}
			mem, ok := hm.Members[n.Name]
			if !ok {
				c.fatal(im.Pos, "%s does not export %q", im.From, n.Name)
			}
			sym := &Symbol{Name: n.Local, Kind: SHostMember, Host: mem, Go: hm.Go + "." + mem.Go}
			if mem.Kind == "field" {
				sym.Type = mem.Type
			} else {
				sym.Type = &HostFnT{M: mem}
			}
			m.Scope.syms[n.Local] = sym
		}
	default:
		dep := c.resolveImport(m, im.From, im.Pos)
		if m.Kind == "client" && dep.Kind == "server" {
			c.fatal(im.Pos, "client code cannot import the server module %q", im.From)
		}
		c.checkModule(dep)
		m.Deps = append(m.Deps, dep)
		if im.Default != "" {
			if dep.Default == nil {
				c.fatal(im.Pos, "%q has no default export", im.From)
			}
			m.Scope.syms[im.Default] = dep.Default
		}
		for _, n := range im.Names {
			if t, ok := dep.Scope.types[n.Name]; ok {
				m.Scope.types[n.Local] = t
				continue
			}
			sym, ok := dep.Exports[n.Name]
			if !ok {
				c.fatal(im.Pos, "%q does not export %q", im.From, n.Name)
			}
			m.Scope.syms[n.Local] = sym
		}
	}
}

func isCapitalized(s string) bool { return s != "" && unicode.IsUpper(rune(s[0])) }

func (c *Checker) declareTop(s Stmt) {
	switch d := s.(type) {
	case *InterfaceDecl:
		t := &ObjT{Name: d.Name, GoName: d.Name, JSON: c.mod.Kind == "client", Pos: d.Pos}
		if !d.Export {
			t.GoName = c.mod.GoPrefix + d.Name
		}
		c.mod.Scope.types[d.Name] = t
		for _, baseName := range d.Extends {
			bt, ok := c.scope.lookupType(baseName).(*ObjT)
			if !ok {
				if ht, isHost := c.Host.Types[baseName]; isHost {
					bt, ok = ht, true
				}
			}
			if !ok {
				c.fatal(d.Pos, "interface %s extends %q, which is not a known object type", d.Name, baseName)
			}
			for _, f := range bt.Fields {
				if t.Field(f.Name) == nil {
					cp := *f
					t.Fields = append(t.Fields, &cp)
				}
			}
		}
		c.fillObj(t, d.Fields)
		d.T = t
	case *TypeAlias:
		d.T = c.resolveType(d.Type, d.Pos)
		c.mod.Scope.types[d.Name] = d.T
	case *FuncDecl:
		sym := &Symbol{Name: d.Name, Module: c.mod, Async: d.Async, Pos: d.Pos}
		d.Sym = sym
		if isCapitalized(d.Name) && !d.Async {
			d.Comp = true
			sym.Kind = SComp
			sym.Go = d.Name
			if !d.Export || underPages(c.mod) { // 页面 / _layout / _404 只被路由表引用: 加前缀, 两个页面都叫 Home 也不冲突
				sym.Go = c.mod.GoPrefix + d.Name
			}
			comp := &CompT{Name: d.Name, Sym: sym, Island: c.mod.Kind == "client"}
			comp.Props = c.propsType(d)
			sym.Comp = comp
			sym.Type = comp
		} else {
			sym.Kind = SFunc
			sym.Go = c.mod.GoPrefix + d.Name
			if d.Export {
				sym.Go = d.Name
			}
			sym.Type = c.fnType(d.Params, d.Ret, d.Async, d.Pos)
		}
		c.mod.Scope.syms[d.Name] = sym
		if d.Export {
			c.mod.Exports[d.Name] = sym
		}
		if d.Default {
			c.mod.Default = sym
		}
	case *ExportDefault:
		if id, ok := d.X.(*Ident); ok {
			if sym := c.mod.Scope.lookup(id.Name); sym != nil {
				c.mod.Default = sym
			}
		}
	}
}

func (c *Checker) fillObj(t *ObjT, fields []*TypeField) {
	for _, f := range fields {
		var ft Type
		if f.Method {
			ft = c.fnType(f.Params, f.Type, false, Pos{})
		} else {
			ft = c.resolveType(f.Type, Pos{})
		}
		if f.Optional {
			ft = &OptT{Elem: unopt(ft)}
		}
		nf := &Field{Name: f.Name, Go: goField(f.Name), Type: ft, Optional: f.Optional}
		if old := t.Field(f.Name); old != nil { // interface extends: 子接口同名字段覆盖
			*old = *nf
			continue
		}
		t.Fields = append(t.Fields, nf)
	}
}

func (c *Checker) fnType(params []*Param, ret TypeExpr, async bool, pos Pos) *FnT {
	ft := &FnT{Async: async, Ret: TVoid}
	for _, p := range params {
		var pt Type = TAny
		if p.Type != nil {
			pt = c.resolveType(p.Type, p.Pos)
		}
		p.T = pt
		name := p.Pat.Name
		if p.Pat.Kind != PatIdent {
			name = "props"
		}
		ft.Params = append(ft.Params, &FnParam{Name: name, Type: pt, Optional: p.Optional})
	}
	if ret != nil {
		ft.Ret = c.resolveType(ret, pos)
	} else {
		ft.Ret = nil // 待推断
	}
	return ft
}

// 组件 props 类型: 第一个参数的类型; 内联对象类型命名为 <Comp>Props
func (c *Checker) propsType(d *FuncDecl) *ObjT {
	if len(d.Params) == 0 {
		t := &ObjT{Name: d.Name + "Props", GoName: d.Name + "Props", JSON: c.mod.Kind == "client"}
		if !d.Export {
			t.GoName = c.mod.GoPrefix + t.Name
		}
		c.mod.AnonTypes = append(c.mod.AnonTypes, t)
		return t
	}
	p := d.Params[0]
	if p.Type == nil {
		c.fatal(p.Pos, "component %s needs a type annotation on its props", d.Name)
	}
	var t *ObjT
	switch te := p.Type.(type) {
	case *TObj:
		t = &ObjT{Name: d.Name + "Props", GoName: d.Name + "Props"}
		if !d.Export {
			t.GoName = c.mod.GoPrefix + t.Name
		}
		c.fillObj(t, te.Fields)
		c.mod.AnonTypes = append(c.mod.AnonTypes, t)
	default:
		rt := c.resolveType(p.Type, p.Pos)
		ot, ok := rt.(*ObjT)
		if !ok {
			c.fatal(p.Pos, "component props must be an object type, got %s", rt)
		}
		t = ot
	}
	if c.mod.Kind == "client" {
		t.JSON = true
		if d.Default {
			for _, f := range t.Fields {
				if isNode(f.Type) {
					c.fatal(p.Pos, "island props must be JSON-serializable and cannot contain a Node (including children); pass data props and render inside the island, or use a shared component there")
				}
			}
		}
	}
	p.T = t
	return t
}

// ---------- 类型表达式 → 类型 ----------

func (c *Checker) resolveType(te TypeExpr, pos Pos) Type {
	switch t := te.(type) {
	case *TRef:
		switch t.Name {
		case "string":
			return TString
		case "number":
			return TNumber
		case "boolean":
			return TBool
		case "void":
			return TVoid
		case "any", "unknown":
			return TAny
		case "RegExp":
			return TRegExp
		case "undefined", "null":
			return TUndef
		case "never":
			return TVoid
		case "Array":
			if len(t.Args) == 1 {
				return &ArrT{Elem: c.resolveType(t.Args[0], pos)}
			}
		case "Record":
			if len(t.Args) == 2 {
				return &MapT{Val: c.resolveType(t.Args[1], pos)}
			}
		case "Promise":
			if len(t.Args) == 1 {
				return c.resolveType(t.Args[0], pos)
			}
		}
		sc := c.scope
		if sc == nil {
			sc = c.global
		}
		if rt := sc.lookupType(t.Name); rt != nil {
			return rt
		}
		if ht, ok := c.Host.Types[t.Name]; ok {
			return ht
		}
		c.fatal(t.Pos, "unknown type %q", t.Name)
	case *TArr:
		return &ArrT{Elem: c.resolveType(t.Elem, pos)}
	case *TObj:
		o := &ObjT{Anon: true, JSON: c.mod != nil && c.mod.Kind == "client"}
		c.fillObj(o, t.Fields)
		if c.mod != nil {
			o.GoName = fmt.Sprintf("%sAnon%d", c.mod.GoPrefix, len(c.mod.AnonTypes))
			c.mod.AnonTypes = append(c.mod.AnonTypes, o)
		}
		return o
	case *TUnion:
		allLit := true
		var nonNil []Type
		for _, m := range t.Members {
			if _, ok := m.(*TStrLit); ok {
				continue
			}
			allLit = false
			if r, ok := m.(*TRef); ok && (r.Name == "undefined" || r.Name == "null") {
				continue
			}
			nonNil = append(nonNil, c.resolveType(m, pos))
		}
		if allLit {
			return TString
		}
		if len(nonNil) == 1 {
			return &OptT{Elem: unopt(nonNil[0])}
		}
		if len(nonNil) == 0 {
			return TUndef
		}
		c.fatal(pos, "only literal unions and T | undefined are in the subset")
	case *TStrLit:
		return TString
	case *TFunc:
		return c.fnType(t.Params, t.Ret, false, pos)
	}
	c.fatal(pos, "unresolvable type")
	return nil
}

// ---------- 函数体 ----------

func (c *Checker) checkFuncBody(d *FuncDecl) {
	sc := newScope(c.scope)
	fc := &fnCtx{component: d.Comp, async: d.Async}
	sc.fn = fc
	d.Scope = sc
	save := c.scope
	c.scope = sc
	defer func() { c.scope = save }()
	if d.Comp {
		fc.want = TNode
		if len(d.Params) > 0 {
			c.bindPattern(d.Params[0].Pat, d.Params[0].T, SParam)
		}
	} else {
		ft := d.Sym.Type.(*FnT)
		for i, p := range d.Params {
			c.bindPattern(p.Pat, ft.Params[i].Type, SParam)
		}
		fc.want = ft.Ret
	}
	c.checkBlockStmts(d.Body)
	if d.Comp {
		d.RetT = TNode
	} else {
		ft := d.Sym.Type.(*FnT)
		if ft.Ret == nil {
			ft.Ret = c.inferRet(fc)
		}
		if d.Async {
			ft.Ret = TAny
		}
		d.RetT = ft.Ret
	}
}

func (c *Checker) inferRet(fc *fnCtx) Type {
	if len(fc.rets) == 0 {
		return TVoid
	}
	return fc.rets[0]
}

func (c *Checker) bindPattern(pat *Pattern, t Type, kind SymKind) {
	pat.T = t
	switch pat.Kind {
	case PatIdent:
		sym := &Symbol{Name: pat.Name, Kind: kind, Type: t, Go: goIdent(pat.Name), Module: c.mod, Pos: pat.Pos}
		pat.Sym = sym
		c.scope.syms[pat.Name] = sym
	case PatObject:
		for _, pp := range pat.Props {
			var ft Type = TAny
			switch o := unopt(t).(type) {
			case *ObjT:
				f := o.Field(pp.Key)
				if f == nil {
					c.fatal(pat.Pos, "type %s has no field %q", o, pp.Key)
				}
				ft = f.Type
			case *MapT:
				ft = &OptT{Elem: o.Val}
			}
			if pp.Default != nil {
				ft = unopt(ft)
				c.check(pp.Default, ft)
			}
			c.bindPattern(pp.Pat, ft, kind)
		}
	case PatArray:
		for i, el := range pat.Elems {
			var et Type = TAny
			switch a := unopt(t).(type) {
			case *ArrT:
				et = a.Elem
			}
			_ = i
			c.bindPattern(el, et, kind)
		}
	}
}

func (c *Checker) checkBlockStmts(b *Block) {
	for _, s := range b.Stmts {
		c.checkStmt(s)
	}
}

func (c *Checker) checkBlock(b *Block) {
	sc := newScope(c.scope)
	save := c.scope
	c.scope = sc
	defer func() { c.scope = save }()
	c.checkBlockStmts(b)
}

func (c *Checker) checkStmt(s Stmt) {
	switch d := s.(type) {
	case *VarDecl:
		c.checkVarDecl(d)
	case *ReturnStmt:
		fc := c.scope.fnCtx()
		if d.X == nil {
			return
		}
		var want Type
		if fc != nil {
			want = fc.want
		}
		t := c.check(d.X, want)
		if fc != nil {
			fc.rets = append(fc.rets, t)
			if fc.component && !isNode(t) && t != TUndef && !isAny(t) {
				c.errf(d.Pos, "a component must return JSX, this returns %s", t)
			}
		}
	case *IfStmt:
		c.check(d.Cond, nil)
		c.checkBlock(d.Then)
		if d.Else != nil {
			switch e := d.Else.(type) {
			case *Block:
				c.checkBlock(e)
			default:
				c.checkStmt(e)
			}
		}
	case *ForOfStmt:
		it := unopt(c.check(d.Iter, nil))
		sc := newScope(c.scope)
		save := c.scope
		c.scope = sc
		switch a := it.(type) {
		case *ArrT:
			c.bindPattern(d.Pat, a.Elem, SConst)
		default:
			if !isAny(it) {
				c.fatal(d.Pos, "for-of can only iterate arrays, got %s", it)
			}
			c.bindPattern(d.Pat, TAny, SConst)
		}
		c.inLoop(func() { c.checkBlockStmts(d.Body) })
		c.scope = save
	case *ForStmt:
		sc := newScope(c.scope)
		save := c.scope
		c.scope = sc
		if d.Init != nil {
			if vd, ok := d.Init.(*VarDecl); ok && vd.Pat.Kind != PatIdent {
				c.fatal(vd.Pos, "a for initializer may declare only a single variable")
			}
			c.checkStmt(d.Init)
		}
		if d.Cond != nil {
			c.check(d.Cond, nil)
		}
		if d.Update != nil {
			c.check(d.Update, nil)
		}
		c.inLoop(func() { c.checkBlockStmts(d.Body) })
		c.scope = save
	case *WhileStmt:
		c.check(d.Cond, nil)
		c.inLoop(func() { c.checkBlock(d.Body) })
	case *BreakStmt:
		fc := c.scope.fnCtx()
		if fc == nil || (fc.loops == 0 && fc.switches == 0) {
			c.errf(d.Pos, "break is only allowed inside a loop or switch")
		}
	case *ContinueStmt:
		fc := c.scope.fnCtx()
		if fc == nil || fc.loops == 0 {
			c.errf(d.Pos, "continue is only allowed inside a loop")
		}
	case *SwitchStmt:
		dt := c.check(d.Disc, nil)
		fc := c.scope.fnCtx()
		if fc != nil {
			fc.switches++
		}
		seenDefault := false
		for _, cs := range d.Cases {
			if cs.Test == nil {
				if seenDefault {
					c.errf(cs.Pos, "switch may have only one default")
				}
				seenDefault = true
			} else {
				c.check(cs.Test, unopt(dt))
			}
			sc := newScope(c.scope)
			save := c.scope
			c.scope = sc
			for _, st := range cs.Body {
				c.checkStmt(st)
			}
			c.scope = save
		}
		if fc != nil {
			fc.switches--
		}
	case *ExprStmt:
		c.check(d.X, nil)
	case *Block:
		c.checkBlock(d)
	case *TryStmt:
		c.checkBlock(d.Body)
		if d.Catch != nil {
			sc := newScope(c.scope)
			if d.CatchName != "" {
				sc.syms[d.CatchName] = &Symbol{Name: d.CatchName, Kind: SConst, Type: TAny, Go: goIdent(d.CatchName)}
			}
			save := c.scope
			c.scope = sc
			c.checkBlockStmts(d.Catch)
			c.scope = save
		}
		if d.Finally != nil {
			c.checkBlock(d.Finally)
		}
	case *ThrowStmt:
		c.check(d.X, nil)
	case *FuncDecl:
		sym := &Symbol{Name: d.Name, Kind: SFunc, Go: goIdent(d.Name), Module: c.mod, Async: d.Async, Pos: d.Pos}
		sym.Type = c.fnType(d.Params, d.Ret, d.Async, d.Pos)
		d.Sym = sym
		c.scope.syms[d.Name] = sym
		c.checkFuncBody(d)
	case *InterfaceDecl, *TypeAlias:
		c.declareTop(s)
		if d, ok := s.(*InterfaceDecl); ok && d.T != nil { // 函数体里声明的 interface 也要生成 Go struct
			c.mod.AnonTypes = append(c.mod.AnonTypes, d.T)
		}
	case *EmptyStmt:
	default:
		c.errf(Pos{}, "unsupported statement %T", s)
	}
}

func (c *Checker) inLoop(f func()) {
	fc := c.scope.fnCtx()
	if fc != nil {
		fc.loops++
		defer func() { fc.loops-- }()
	}
	f()
}

func (c *Checker) checkVarDecl(d *VarDecl) {
	kind := SVar
	if d.Const {
		kind = SConst
	}
	var declared Type
	if d.Type != nil {
		declared = c.resolveType(d.Type, d.Pos)
	}
	// hooks
	if call, ok := d.Init.(*Call); ok {
		if id, ok := call.Fn.(*Ident); ok {
			if sym := c.scope.lookup(id.Name); sym != nil && sym.Kind == SBuiltin {
				switch sym.Name {
				case "useState":
					c.checkUseState(d, call)
					return
				case "useMemo":
					if len(call.Args) < 1 {
						c.fatal(d.Pos, "useMemo needs a function")
					}
					ft, ok := c.check(call.Args[0], &FnT{Ret: nil}).(*FnT)
					if !ok {
						c.fatal(d.Pos, "the argument of useMemo must be a function")
					}
					call.Kind = "hook:useMemo"
					call.SetT(ft.Ret)
					c.bindPattern(d.Pat, ft.Ret, SMemo)
					d.Hook = "memo"
					return
				}
			}
		}
	}
	var t Type = TAny
	if d.Init != nil {
		t = c.check(d.Init, declared)
	}
	if declared != nil {
		t = declared
	}
	if d.Init == nil && declared == nil {
		c.fatal(d.Pos, "variable %s needs an initializer or a type annotation", d.Pat.Name)
	}
	if t == TUndef {
		c.fatal(d.Pos, "cannot infer the type of %s (the initializer is undefined); add an annotation", d.Pat.Name)
	}
	if d.Const && d.Pat.Kind == PatIdent && c.mod.Kind != "server" && c.scope.fn != nil && c.scope.fn.component && c.isReactive(d.Init) {
		if _, isFn := t.(*FnT); !isFn {
			kind = SMemo
			d.Hook = "memo"
		}
	}
	c.bindPattern(d.Pat, t, kind)
}

func (c *Checker) checkUseState(d *VarDecl, call *Call) {
	if d.Pat.Kind != PatArray || len(d.Pat.Elems) != 2 || d.Pat.Elems[0].Kind != PatIdent || d.Pat.Elems[1].Kind != PatIdent {
		c.fatal(d.Pos, "useState must be written as const [x, setX] = useState(init)")
	}
	var t Type
	if len(call.TypeArgs) == 1 {
		t = c.resolveType(call.TypeArgs[0], d.Pos)
	}
	if len(call.Args) != 1 {
		c.fatal(d.Pos, "useState needs an initial value")
	}
	it := c.check(call.Args[0], t)
	if t == nil {
		t = it
	}
	if t == TUndef || isAny(t) {
		c.fatal(d.Pos, "cannot infer the useState type; write useState<T>(...)")
	}
	call.Kind = "hook:useState"
	call.SetT(t)
	d.Hook = "useState"
	c.bindPattern(d.Pat.Elems[0], t, SSignal)
	c.bindPattern(d.Pat.Elems[1], &SetterT{Elem: t}, SSetter)
	d.Pat.T = t
}

// isReactive: 表达式是否读取了 signal/memo(不进入箭头函数体; 条件/三元只看条件; map 只看接收者)
func (c *Checker) isReactive(e Expr) bool {
	switch x := e.(type) {
	case nil:
		return false
	case *Ident:
		return x.Sym != nil && (x.Sym.Kind == SSignal || x.Sym.Kind == SMemo)
	case *Member:
		return c.isReactive(x.X)
	case *Index:
		return c.isReactive(x.X) || c.isReactive(x.I)
	case *Call:
		if c.isReactive(x.Fn) {
			return true
		}
		for _, a := range x.Args {
			if arrow, ok := a.(*Arrow); ok { // 回调体里读到 signal, 结果就是响应式的(如 rows.filter(r => r.ok === onlyNo))
				if c.isReactiveArrow(arrow) {
					return true
				}
				continue
			}
			if c.isReactive(a) {
				return true
			}
		}
		return false
	case *Unary:
		return c.isReactive(x.X)
	case *Update:
		return c.isReactive(x.X)
	case *Binary:
		if x.Op == "&&" || x.Op == "||" || x.Op == "??" {
			return c.isReactive(x.L)
		}
		return c.isReactive(x.L) || c.isReactive(x.R)
	case *CondExpr:
		return c.isReactive(x.Test)
	case *TemplateLit:
		for _, a := range x.Exprs {
			if c.isReactive(a) {
				return true
			}
		}
		return false
	case *Paren:
		return c.isReactive(x.X)
	case *AsExpr:
		return c.isReactive(x.X)
	case *NonNull:
		return c.isReactive(x.X)
	case *ArrayLit:
		for _, a := range x.Elems {
			if c.isReactive(a) {
				return true
			}
		}
		return false
	case *SpreadExpr:
		return c.isReactive(x.X)
	case *ObjectLit:
		for _, p := range x.Props {
			if c.isReactive(p.Val) || c.isReactive(p.Spread) {
				return true
			}
		}
		return false
	}
	return false
}

func (c *Checker) isReactiveArrow(a *Arrow) bool {
	if a.ExprBody != nil {
		return c.isReactive(a.ExprBody)
	}
	if a.Body == nil {
		return false
	}
	return c.isReactiveStmts(a.Body.Stmts)
}

func (c *Checker) isReactiveStmts(ss []Stmt) bool {
	for _, s := range ss {
		switch d := s.(type) {
		case *VarDecl:
			if c.isReactive(d.Init) {
				return true
			}
		case *ReturnStmt:
			if c.isReactive(d.X) {
				return true
			}
		case *ExprStmt:
			if c.isReactive(d.X) {
				return true
			}
		case *IfStmt:
			if c.isReactive(d.Cond) || c.isReactiveStmts(d.Then.Stmts) {
				return true
			}
			if b, ok := d.Else.(*Block); ok && c.isReactiveStmts(b.Stmts) {
				return true
			}
		case *ForOfStmt:
			if c.isReactive(d.Iter) || c.isReactiveStmts(d.Body.Stmts) {
				return true
			}
		case *ForStmt:
			if d.Init != nil && c.isReactiveStmts([]Stmt{d.Init}) {
				return true
			}
			if c.isReactive(d.Cond) || c.isReactive(d.Update) || c.isReactiveStmts(d.Body.Stmts) {
				return true
			}
		case *WhileStmt:
			if c.isReactive(d.Cond) || c.isReactiveStmts(d.Body.Stmts) {
				return true
			}
		case *SwitchStmt:
			if c.isReactive(d.Disc) {
				return true
			}
			for _, cs := range d.Cases {
				if c.isReactive(cs.Test) || c.isReactiveStmts(cs.Body) {
					return true
				}
			}
		case *Block:
			if c.isReactiveStmts(d.Stmts) {
				return true
			}
		}
	}
	return false
}

// ---------- 表达式 ----------

func (c *Checker) check(e Expr, want Type) Type {
	t := c.checkInner(e, want)
	if t == nil {
		t = TAny
	}
	e.SetT(t)
	return t
}

func (c *Checker) checkInner(e Expr, want Type) Type {
	switch x := e.(type) {
	case *Ident:
		sym := c.scope.lookup(x.Name)
		if sym == nil {
			c.fatal(x.Pos, "undefined identifier %q", x.Name)
		}
		x.Sym = sym
		if sym.Kind == SBuiltin && sym.Type == nil {
			c.fatal(x.Pos, "%s can only be called", x.Name)
		}
		if sym.Type == nil {
			return TAny
		}
		return sym.Type
	case *NumLit:
		return TNumber
	case *StrLit:
		return TString
	case *BoolLit:
		return TBool
	case *RegexLit:
		for _, f := range x.Flags {
			if !strings.ContainsRune("gimsu", f) {
				c.fatal(x.Pos, "regex flag %q is not supported (use g i m s u)", string(f))
			}
		}
		if _, err := regexp.Compile(goRegex(x.Pattern, x.Flags)); err != nil {
			c.fatal(x.Pos, "regex is not valid RE2 (lookaround and backreferences are not supported on the server): %v", err)
		}
		return TRegExp
	case *NullLit:
		return TUndef
	case *TemplateLit:
		for _, a := range x.Exprs {
			c.check(a, nil)
		}
		return TString
	case *ArrayLit:
		var elem Type
		if a, ok := unopt(want).(*ArrT); ok && want != nil {
			elem = a.Elem
		}
		for _, el := range x.Elems {
			if sp, ok := el.(*SpreadExpr); ok {
				st := c.check(sp.X, want)
				sp.SetT(st)
				if a, ok := unopt(st).(*ArrT); ok && elem == nil {
					elem = a.Elem
				}
				continue
			}
			et := c.check(el, elem)
			if elem == nil {
				elem = et
			}
		}
		if elem == nil {
			c.fatal(x.Pos, "cannot infer the type of an empty array; add an annotation (e.g. useState<string[]>([]))")
		}
		return &ArrT{Elem: elem}
	case *ObjectLit:
		if mt, ok := unopt(want).(*MapT); ok && want != nil {
			for _, p := range x.Props {
				if p.Spread != nil {
					c.check(p.Spread, want)
					continue
				}
				c.check(p.Val, mt.Val)
			}
			return mt
		}
		if o, ok := unopt(want).(*ObjT); ok && want != nil {
			for _, p := range x.Props {
				if p.Spread != nil {
					c.check(p.Spread, o)
					continue
				}
				f := o.Field(p.Key)
				if f == nil {
					c.fatal(x.Pos, "type %s has no field %q", o, p.Key)
				}
				c.check(p.Val, f.Type)
			}
			return o
		}
		o := &ObjT{Anon: true, GoName: fmt.Sprintf("%sAnon%d", c.mod.GoPrefix, len(c.mod.AnonTypes))}
		c.mod.AnonTypes = append(c.mod.AnonTypes, o)
		for _, p := range x.Props {
			if p.Spread != nil {
				st := c.check(p.Spread, nil)
				if so, ok := unopt(st).(*ObjT); ok {
					o.Fields = append(o.Fields, so.Fields...)
				}
				continue
			}
			o.Fields = append(o.Fields, &Field{Name: p.Key, Go: goField(p.Key), Type: c.check(p.Val, nil)})
		}
		return o
	case *Member:
		return c.checkMember(x)
	case *Index:
		xt := unopt(c.check(x.X, nil))
		c.check(x.I, nil)
		switch t := xt.(type) {
		case *ArrT:
			return t.Elem
		case *MapT:
			return &OptT{Elem: t.Val}
		case *Prim:
			if t == TString {
				return TString
			}
			if t == TAny {
				return TAny
			}
		}
		c.fatal(x.Pos, "cannot index into %s", xt)
	case *Call:
		return c.checkCall(x, want)
	case *Unary:
		if x.Op == "delete" {
			var recv Expr
			switch t := x.X.(type) {
			case *Index:
				recv = t.X
			case *Member:
				recv = t.X
			}
			if recv == nil {
				c.fatal(x.Pos, "delete only works on a Record key: delete m[key] / delete m.key")
			}
			c.check(x.X, nil)
			if _, ok := unopt(recv.T()).(*MapT); !ok {
				c.fatal(x.Pos, "delete only works on a Record key, got %s", recv.T())
			}
			return TBool
		}
		c.check(x.X, nil)
		switch x.Op {
		case "!":
			return TBool
		case "typeof":
			return TString
		}
		return TNumber
	case *Binary:
		return c.checkBinary(x, want)
	case *CondExpr:
		c.check(x.Test, nil)
		a := c.check(x.Then, want)
		b := c.check(x.Else, want)
		x.Reactive = c.mod.Kind != "server" && c.isReactive(x.Test)
		if isNode(a) || isNode(b) {
			return TNode
		}
		if sameType(a, b) {
			return a
		}
		if isAny(a) {
			return b
		}
		return a
	case *Arrow:
		return c.checkArrow(x, want)
	case *Assign:
		tt := c.check(x.Target, nil)
		c.check(x.Val, tt)
		if id, ok := x.Target.(*Ident); ok && id.Sym != nil && id.Sym.Kind == SConst {
			c.errf(x.Pos, "cannot assign to const %s", id.Name)
		}
		if x.Op != "=" && x.Op != "+=" && !isNumber(tt) && !isAny(tt) {
			c.errf(x.Pos, "%s only works on number, got %s", x.Op, tt)
		}
		return tt
	case *Update:
		tt := c.check(x.X, nil)
		if !isNumber(tt) && !isAny(tt) {
			c.errf(x.Pos, "%s only works on number, got %s", x.Op, tt)
		}
		if id, ok := x.X.(*Ident); ok && id.Sym != nil && id.Sym.Kind == SConst {
			c.errf(x.Pos, "const %s cannot be modified with %s", id.Name, x.Op)
		}
		return TNumber
	case *AsExpr:
		c.check(x.X, nil)
		return c.resolveType(x.Type, x.Pos)
	case *AwaitExpr:
		c.check(x.X, nil)
		return TAny
	case *SpreadExpr:
		return c.check(x.X, want)
	case *NonNull:
		return unopt(c.check(x.X, want))
	case *Paren:
		return c.check(x.X, want)
	case *JSXElem:
		return c.checkJSX(x)
	case *JSXFrag:
		for _, k := range x.Children {
			c.checkChild(k)
		}
		return TNode
	case *JSXText:
		return TString
	case *JSXExprChild:
		if x.X == nil {
			return TUndef
		}
		return c.check(x.X, nil)
	}
	c.fatal(e.GetPos(), "unsupported expression %T", e)
	return nil
}

var arrayMethods = map[string]bool{"map": true, "filter": true, "find": true, "some": true, "every": true, "includes": true,
	"indexOf": true, "join": true, "slice": true, "concat": true, "forEach": true, "length": true,
	"sort": true, "reduce": true, "reverse": true, "flat": true, "at": true,
	"push": true, "pop": true, "shift": true, "unshift": true, "splice": true, "findIndex": true, "lastIndexOf": true}
var stringMethods = map[string]bool{"toUpperCase": true, "toLowerCase": true, "trim": true, "includes": true, "startsWith": true, "padStart": true, "padEnd": true,
	"endsWith": true, "split": true, "slice": true, "replace": true, "replaceAll": true, "repeat": true, "indexOf": true, "charAt": true, "length": true,
	"trimStart": true, "trimEnd": true, "lastIndexOf": true, "localeCompare": true, "toString": true, "at": true, "match": true, "search": true}
var numberMethods = map[string]bool{"toFixed": true, "toString": true}

func (c *Checker) checkMember(x *Member) Type {
	rt := c.check(x.X, nil)
	opt := false
	if o, ok := rt.(*OptT); ok {
		rt = o.Elem
		opt = true
	}
	var res Type
	switch t := rt.(type) {
	case *ArrT:
		if !arrayMethods[x.Name] {
			c.fatal(x.Pos, "array method %q is not in the subset", x.Name)
		}
		x.Builtin = x.Name
		if x.Name == "length" {
			res = TNumber
		} else {
			res = &BuiltinT{Recv: t, Name: x.Name}
		}
	case *Prim:
		switch t {
		case TString:
			if !stringMethods[x.Name] {
				c.fatal(x.Pos, "string method %q is not in the subset", x.Name)
			}
			x.Builtin = x.Name
			if x.Name == "length" {
				res = TNumber
			} else {
				res = &BuiltinT{Recv: t, Name: x.Name}
			}
		case TNumber:
			if !numberMethods[x.Name] {
				c.fatal(x.Pos, "number method %q is not in the subset", x.Name)
			}
			x.Builtin = x.Name
			res = &BuiltinT{Recv: t, Name: x.Name}
		case TRegExp:
			if x.Name != "test" {
				c.fatal(x.Pos, "RegExp only supports .test(s) (use s.match / s.replace / s.split / s.search for the rest)")
			}
			x.Builtin = x.Name
			res = &BuiltinT{Recv: t, Name: x.Name}
		case TAny:
			res = TAny
		default:
			c.fatal(x.Pos, "%s has no property %q", t, x.Name)
		}
	case *ObjT:
		if f := t.Field(x.Name); f != nil {
			x.GoName = f.Go
			res = f.Type
		} else if m, ok := t.Methods[x.Name]; ok {
			x.GoName = m.Go
			res = &HostFnT{M: m}
		} else {
			c.fatal(x.Pos, "type %s has no field %q", t, x.Name)
		}
	case *MapT:
		x.MapKey = true
		res = &OptT{Elem: t.Val}
	case *HostModT:
		m, ok := t.Members[x.Name]
		if !ok {
			c.fatal(x.Pos, "%s has no member %q", t, x.Name)
		}
		x.GoName = m.Go
		if m.Kind == "field" {
			res = m.Type
		} else {
			res = &HostFnT{M: m}
		}
	case *GlobalT:
		res = &BuiltinT{Recv: t, Name: x.Name}
	default:
		c.fatal(x.Pos, "%s has no accessible property %q", rt, x.Name)
	}
	if opt && x.Optional {
		if _, isOpt := res.(*OptT); !isOpt {
			return &OptT{Elem: res}
		}
	}
	return res
}

func (c *Checker) checkArgs(x *Call, params []*FnParam) {
	for i, a := range x.Args {
		var want Type
		if i < len(params) {
			want = params[i].Type
		}
		c.check(a, want)
	}
	if len(x.Args) < len(params) {
		for _, p := range params[len(x.Args):] {
			if !p.Optional {
				if _, isOpt := p.Type.(*OptT); !isOpt {
					c.fatal(x.Pos, "missing argument %s", p.Name)
				}
			}
		}
	}
}

func (c *Checker) checkCall(x *Call, want Type) Type {
	// 全局内建
	if id, ok := x.Fn.(*Ident); ok {
		if sym := c.scope.lookup(id.Name); sym != nil && sym.Kind == SBuiltin && sym.Type == nil {
			id.Sym = sym
			x.Kind = "global:" + sym.Name
			switch sym.Name {
			case "useState":
				c.fatal(x.Pos, "useState can only be written as const [x, setX] = useState(init)")
			case "useMemo":
				c.fatal(x.Pos, "useMemo can only be written as const x = useMemo(() => ...)")
			case "useEffect":
				if len(x.Args) < 1 {
					c.fatal(x.Pos, "useEffect needs a function")
				}
				c.check(x.Args[0], &FnT{Ret: TVoid})
				once := false
				if len(x.Args) >= 2 {
					if arr, ok := x.Args[1].(*ArrayLit); ok && len(arr.Elems) == 0 {
						once = true // useEffect(fn, []): 挂载时跑一次, 不建立 signal 依赖
					}
				}
				if once {
					x.Kind = "hook:useEffectOnce"
				} else {
					for _, a := range x.Args[1:] {
						c.check(a, nil)
					}
					x.Kind = "hook:useEffect"
				}
				return TVoid
			case "String", "encodeURIComponent", "decodeURIComponent":
				for _, a := range x.Args {
					c.check(a, nil)
				}
				return TString
			case "Number", "parseInt", "parseFloat":
				for _, a := range x.Args {
					c.check(a, nil)
				}
				return TNumber
			case "Boolean", "isNaN":
				for _, a := range x.Args {
					c.check(a, nil)
				}
				return TBool
			case "jsonLd":
				if len(x.Args) != 1 {
					c.fatal(x.Pos, "jsonLd needs one JSON string (usually JSON.stringify(...))")
				}
				c.check(x.Args[0], TString)
				return TNode
			case "redirect":
				if c.mod.Kind != "server" {
					c.fatal(x.Pos, "redirect() can only be used in server components (*.server.tsx)")
				}
				if len(x.Args) < 1 || len(x.Args) > 2 {
					c.fatal(x.Pos, "redirect(url, status?) takes 1 or 2 arguments")
				}
				c.check(x.Args[0], TString)
				if len(x.Args) == 2 {
					c.check(x.Args[1], TNumber)
				}
				return TNode
			case "notFound":
				if c.mod.Kind != "server" {
					c.fatal(x.Pos, "notFound() can only be used in server components (*.server.tsx)")
				}
				if len(x.Args) != 0 {
					c.fatal(x.Pos, "notFound() takes no arguments")
				}
				return TNode
			case "t", "fmtDate":
				if len(x.Args) != 2 {
					c.fatal(x.Pos, "%s(locale, key) takes two arguments", sym.Name)
				}
				c.check(x.Args[0], TString)
				c.check(x.Args[1], TString)
				return TString
			case "isoDate":
				if len(x.Args) != 1 {
					c.fatal(x.Pos, "isoDate(ms) takes one argument")
				}
				c.check(x.Args[0], TNumber)
				return TString
			case "lpath":
				if len(x.Args) != 2 {
					c.fatal(x.Pos, "lpath(locale, path) takes two arguments")
				}
				c.check(x.Args[0], TString)
				c.check(x.Args[1], TString)
				return TString
			case "tv":
				if len(x.Args) != 3 {
					c.fatal(x.Pos, "tv(locale, key, vars) takes three arguments (vars is a Record<string,string>)")
				}
				c.check(x.Args[0], TString)
				c.check(x.Args[1], TString)
				c.check(x.Args[2], &MapT{Val: TString})
				return TString
			case "plural":
				if len(x.Args) != 3 {
					c.fatal(x.Pos, "plural(locale, key, n) takes three arguments")
				}
				c.check(x.Args[0], TString)
				c.check(x.Args[1], TString)
				c.check(x.Args[2], TNumber)
				return TString
			case "fmtNum", "fmtCur":
				if len(x.Args) != 2 {
					c.fatal(x.Pos, "%s(locale, n) takes two arguments", sym.Name)
				}
				c.check(x.Args[0], TString)
				c.check(x.Args[1], TNumber)
				return TString
			default: // fetch / setTimeout / alert ...
				for _, a := range x.Args {
					c.check(a, nil)
				}
				return TAny
			}
		}
	}
	ft := c.check(x.Fn, nil)
	if o, ok := ft.(*OptT); ok {
		ft = o.Elem
	}
	switch t := ft.(type) {
	case *BuiltinT:
		return c.checkBuiltinCall(x, t, want)
	case *HostFnT:
		x.Kind = "host"
		x.Host = t.M
		var params []*FnParam
		for i, p := range t.M.Params {
			params = append(params, &FnParam{Name: fmt.Sprintf("arg%d", i), Type: p.Type})
		}
		c.checkArgs(x, params)
		if t.M.Ret == nil {
			return TVoid
		}
		return t.M.Ret
	case *FnT:
		x.Kind = "fn"
		c.checkArgs(x, t.Params)
		if t.Async {
			return TAny
		}
		if t.Ret == nil {
			return TAny
		}
		return t.Ret
	case *SetterT:
		x.Kind = "setter"
		if len(x.Args) != 1 {
			c.fatal(x.Pos, "a setter takes one argument")
		}
		if a, ok := x.Args[0].(*Arrow); ok {
			c.check(a, &FnT{Params: []*FnParam{{Name: "prev", Type: t.Elem}}, Ret: t.Elem})
		} else {
			c.check(x.Args[0], t.Elem)
		}
		return TVoid
	case *CompT:
		c.fatal(x.Pos, "component %s must be used as JSX: <%s />", t.Name, t.Name)
	case *Prim:
		if t == TAny {
			x.Kind = "any"
			for _, a := range x.Args {
				c.check(a, nil)
			}
			return TAny
		}
	}
	c.fatal(x.Pos, "%s is not a function", ft)
	return nil
}

func (c *Checker) checkBuiltinCall(x *Call, b *BuiltinT, want Type) Type {
	x.Kind = "builtin:" + b.Name
	args := func(n int) {
		if len(x.Args) < n {
			c.fatal(x.Pos, "%s needs %d argument(s)", b.Name, n)
		}
		for _, a := range x.Args {
			c.check(a, nil)
		}
	}
	switch r := b.Recv.(type) {
	case *ArrT:
		cb := func(ret Type) *FnT {
			return &FnT{Params: []*FnParam{{Name: "x", Type: r.Elem}, {Name: "i", Type: TNumber, Optional: true}}, Ret: ret}
		}
		switch b.Name {
		case "map":
			if len(x.Args) != 1 {
				c.fatal(x.Pos, "map needs a callback")
			}
			var wantElem Type // useState<Item[]>(xs.map(...)) / const ys: Item[] = xs.map(...): 期望类型传进回调, 对象字面量就能对上
			if wa, ok := unopt(want).(*ArrT); ok && want != nil {
				wantElem = wa.Elem
			}
			ft, ok := c.check(x.Args[0], cb(wantElem)).(*FnT)
			if !ok {
				c.fatal(x.Pos, "the argument of map must be a function")
			}
			if ft.Ret == nil {
				c.fatal(x.Pos, "cannot infer the return type of the map callback")
			}
			x.Reactive = c.mod.Kind != "server" && c.isReactive(x)
			return &ArrT{Elem: ft.Ret}
		case "filter", "find", "some", "every", "forEach", "findIndex":
			if len(x.Args) != 1 {
				c.fatal(x.Pos, "%s needs a callback", b.Name)
			}
			ret := TBool
			if b.Name == "forEach" {
				ret = TVoid
			}
			c.check(x.Args[0], cb(ret))
			switch b.Name {
			case "filter":
				return &ArrT{Elem: r.Elem}
			case "find":
				return &OptT{Elem: r.Elem}
			case "forEach":
				return TVoid
			case "findIndex":
				return TNumber
			}
			return TBool
		case "includes":
			args(1)
			return TBool
		case "indexOf", "lastIndexOf":
			args(1)
			return TNumber
		case "push", "unshift":
			if len(x.Args) < 1 {
				c.fatal(x.Pos, "%s needs at least one element", b.Name)
			}
			c.requireMutable(x.Pos, m0(x), b.Name)
			for _, a := range x.Args {
				c.check(a, r.Elem)
			}
			return TNumber
		case "pop", "shift":
			args(0)
			c.requireMutable(x.Pos, m0(x), b.Name)
			return &OptT{Elem: r.Elem}
		case "splice":
			if len(x.Args) < 1 {
				c.fatal(x.Pos, "splice takes (start, deleteCount?, ...items)")
			}
			c.requireMutable(x.Pos, m0(x), b.Name)
			c.check(x.Args[0], TNumber)
			if len(x.Args) > 1 {
				c.check(x.Args[1], TNumber)
			}
			if len(x.Args) > 2 {
				for _, a := range x.Args[2:] {
					c.check(a, r.Elem)
				}
			}
			return &ArrT{Elem: r.Elem}
		case "join":
			args(0)
			return TString
		case "slice", "concat":
			args(0)
			return &ArrT{Elem: r.Elem}
		case "reverse":
			args(0)
			return &ArrT{Elem: r.Elem}
		case "at":
			args(1)
			return r.Elem
		case "sort":
			if len(x.Args) < 1 {
				c.fatal(x.Pos, "sort needs a comparator (a, b) => number")
			}
			c.check(x.Args[0], &FnT{Params: []*FnParam{{Name: "a", Type: r.Elem}, {Name: "b", Type: r.Elem}}, Ret: TNumber})
			return &ArrT{Elem: r.Elem}
		case "flat":
			args(0)
			if inner, ok := r.Elem.(*ArrT); ok {
				return &ArrT{Elem: inner.Elem}
			}
			return &ArrT{Elem: r.Elem}
		case "reduce":
			if len(x.Args) != 2 {
				c.fatal(x.Pos, "reduce takes (callback, initialValue)")
			}
			accT := c.check(x.Args[1], nil)
			c.check(x.Args[0], &FnT{Params: []*FnParam{{Name: "acc", Type: accT}, {Name: "x", Type: r.Elem}, {Name: "i", Type: TNumber, Optional: true}}, Ret: accT})
			return accT
		}
	case *Prim:
		if r == TRegExp { // re.test(s)
			args(1)
			c.check(x.Args[0], TString)
			return TBool
		}
		if r == TString && len(x.Args) > 0 {
			switch b.Name {
			case "replace", "replaceAll", "split", "match", "search":
				if c.check(x.Args[0], nil) == TRegExp {
					x.Kind = "builtin:re_" + b.Name
					if b.Name == "replaceAll" && !strings.Contains(x.Args[0].(*RegexLit).Flags, "g") {
						c.fatal(x.Pos, "replaceAll with a regex requires the g flag")
					}
					for _, a := range x.Args[1:] {
						c.check(a, TString)
					}
					switch b.Name {
					case "match", "split":
						return &ArrT{Elem: TString}
					case "search":
						return TNumber
					}
					return TString
				}
			}
		}
		switch b.Name {
		case "match", "search":
			c.fatal(x.Pos, "%s takes a regex literal", b.Name)
		case "toUpperCase", "toLowerCase", "trim", "trimStart", "trimEnd", "slice", "replace", "replaceAll", "repeat", "charAt", "toFixed", "toString", "at":
			args(0)
			return TString
		case "includes", "startsWith", "endsWith":
			args(1)
			return TBool
		case "split":
			args(1)
			return &ArrT{Elem: TString}
		case "indexOf", "lastIndexOf", "localeCompare":
			args(1)
			return TNumber
		case "padStart", "padEnd":
			args(1)
			return TString
		}
	case *GlobalT:
		switch r.Name {
		case "console":
			args(0)
			return TVoid
		case "JSON":
			args(0)
			if b.Name == "stringify" {
				return TString
			}
			return TAny
		case "Math":
			args(0)
			if b.Name == "pow" || b.Name == "sign" || b.Name == "trunc" {
				return TNumber
			}
			return TNumber
		case "Date":
			switch b.Name {
			case "now":
				args(0)
				return TNumber
			case "parse":
				args(1)
				return TNumber
			}
			c.fatal(x.Pos, "only Date.now() / Date.parse(iso) are in the subset (format with fmtDate / isoDate)")
		case "Object":
			if b.Name == "hasOwn" {
				if len(x.Args) != 2 {
					c.fatal(x.Pos, "Object.hasOwn(record, key) takes two arguments")
				}
				if _, ok := unopt(c.check(x.Args[0], nil)).(*MapT); !ok {
					c.fatal(x.Pos, "Object.hasOwn: the first argument must be a Record")
				}
				c.check(x.Args[1], TString)
				return TBool
			}
			if len(x.Args) != 1 {
				c.fatal(x.Pos, "Object.%s takes one argument", b.Name)
			}
			at := unopt(c.check(x.Args[0], nil))
			switch b.Name {
			case "keys":
				return &ArrT{Elem: TString}
			case "values":
				if mt, ok := at.(*MapT); ok {
					return &ArrT{Elem: mt.Val}
				}
				return &ArrT{Elem: TAny}
			}
			c.fatal(x.Pos, "only Object.keys / Object.values / Object.hasOwn are in the subset")
		}
		args(0)
		return TAny
	}
	c.fatal(x.Pos, "%s is not in the subset", b)
	return nil
}

func m0(x *Call) Expr {
	if m, ok := x.Fn.(*Member); ok {
		return m.X
	}
	return nil
}

// requireMutable: push/pop/splice 这类原地修改需要一个可寻址的数组(变量 / 字段 / 下标), 不能是临时值
func (c *Checker) requireMutable(pos Pos, recv Expr, method string) {
	for {
		p, ok := recv.(*Paren)
		if !ok {
			break
		}
		recv = p.X
	}
	switch r := recv.(type) {
	case *Ident:
		if r.Sym != nil && (r.Sym.Kind == SSignal || r.Sym.Kind == SMemo) {
			c.errf(pos, "%s cannot mutate state %s; use the immutable form set%s([...%s, x])", method, r.Name, strings.ToUpper(r.Name[:1])+r.Name[1:], r.Name)
		}
		return
	case *Member, *Index:
		return
	}
	c.errf(pos, "%s must be called on a variable / field / index (it mutates an addressable array in place)", method)
}

func (c *Checker) checkBinary(x *Binary, want Type) Type {
	switch x.Op {
	case "&&":
		c.check(x.L, nil)
		r := c.check(x.R, want)
		x.Reactive = c.mod.Kind != "server" && c.isReactive(x.L)
		if isNode(r) {
			return TNode
		}
		return r
	case "||", "??":
		l := c.check(x.L, want)
		r := c.check(x.R, unopt(l))
		x.Reactive = c.mod.Kind != "server" && c.isReactive(x.L)
		if isAny(l) {
			return r
		}
		return unopt(l)
	}
	l := c.check(x.L, nil)
	r := c.check(x.R, nil)
	switch x.Op {
	case "+":
		if isString(l) || isString(r) {
			return TString
		}
		return TNumber
	case "-", "*", "/", "%":
		return TNumber
	case "==", "!=":
		c.errf(x.Pos, "only === / !== are allowed (== and != are not in the subset)")
		return TBool
	default:
		return TBool
	}
}

func (c *Checker) checkArrow(x *Arrow, want Type) Type {
	var wf *FnT
	if f, ok := unopt(want).(*FnT); ok && want != nil {
		wf = f
	}
	ft := &FnT{Async: x.Async}
	sc := newScope(c.scope)
	fc := &fnCtx{async: x.Async}
	sc.fn = fc
	x.Scope = sc
	save := c.scope
	c.scope = sc
	defer func() { c.scope = save }()
	for i, p := range x.Params {
		var pt Type
		if p.Type != nil {
			pt = c.resolveType(p.Type, p.Pos)
		} else if wf != nil && i < len(wf.Params) {
			pt = wf.Params[i].Type
		} else {
			c.fatal(p.Pos, "cannot infer the type of parameter %s; add an annotation", p.Pat.Name)
		}
		p.T = pt
		name := p.Pat.Name
		if p.Pat.Kind != PatIdent {
			name = fmt.Sprintf("p%d", i)
		}
		ft.Params = append(ft.Params, &FnParam{Name: name, Type: pt})
		c.bindPattern(p.Pat, pt, SParam)
	}
	if x.Ret != nil {
		ft.Ret = c.resolveType(x.Ret, x.Pos)
	} else if wf != nil && wf.Ret != nil {
		ft.Ret = wf.Ret
	}
	fc.want = ft.Ret
	if x.ExprBody != nil {
		t := c.check(x.ExprBody, ft.Ret)
		if ft.Ret == nil {
			ft.Ret = t
		}
	} else {
		c.checkBlockStmts(x.Body)
		if ft.Ret == nil {
			ft.Ret = c.inferRet(fc)
		}
	}
	if x.Async {
		ft.Ret = TAny
	}
	x.RetT = ft.Ret
	return ft
}

func (c *Checker) checkChild(k Expr) {
	switch x := k.(type) {
	case *JSXText:
		x.SetT(TString)
	case *JSXExprChild:
		if x.X == nil {
			x.SetT(TUndef)
			return
		}
		t := c.check(x.X, nil)
		x.SetT(t)
		x.Reactive = c.mod.Kind != "server" && c.isReactive(x.X)
		switch tt := unopt(t).(type) {
		case *Prim:
			if tt == TVoid {
				c.fatal(x.Pos, "a JSX child cannot be void")
			}
		case *ArrT:
			if !isNode(tt.Elem) && !isString(tt.Elem) && !isNumber(tt.Elem) && !isAny(tt.Elem) {
				c.fatal(x.Pos, "elements of a JSX child array must be nodes or text, got %s", tt.Elem)
			}
		case *ObjT, *FnT, *MapT:
			c.fatal(x.Pos, "a JSX child cannot be %s", t)
		}
	default:
		c.check(k, TNode)
	}
}

var attrNameMap = map[string]string{"className": "class", "htmlFor": "for", "charSet": "charset", "tabIndex": "tabindex"}

func (c *Checker) checkJSX(x *JSXElem) Type {
	if isCapitalized(x.Tag) {
		sym := c.scope.lookup(x.Tag)
		if sym == nil || sym.Kind != SComp {
			c.fatal(x.Pos, "%q is not a component", x.Tag)
		}
		x.Comp = sym
		if sym.Go == "gotsx.Suspense" && c.mod.Kind != "server" {
			c.fatal(x.Pos, "Suspense is only available in server components (*.server.tsx)")
		}
		props := sym.Comp.Props
		seen := map[string]bool{}
		for _, a := range x.Attrs {
			if a.Spread != nil {
				c.fatal(x.Pos, "JSX attribute spread is not in the subset")
			}
			if a.Name == "key" {
				if a.Val != nil {
					c.check(a.Val, nil)
				}
				continue
			}
			f := props.Field(a.Name)
			if f == nil {
				c.fatal(x.Pos, "component %s has no prop %q", x.Tag, a.Name)
			}
			seen[a.Name] = true
			if a.Val == nil {
				if !isBool(f.Type) {
					c.fatal(x.Pos, "prop %s needs a value", a.Name)
				}
				continue
			}
			c.check(a.Val, f.Type)
			a.Reactive = c.mod.Kind != "server" && c.isReactive(a.Val)
		}
		for _, f := range props.Fields {
			if !seen[f.Name] && !f.Optional && f.Name != "children" {
				if _, isOpt := f.Type.(*OptT); !isOpt {
					c.fatal(x.Pos, "component %s is missing prop %q", x.Tag, f.Name)
				}
			}
		}
		if len(x.Children) > 0 && props.Field("children") == nil {
			c.fatal(x.Pos, "component %s does not accept children", x.Tag)
		}
		for _, k := range x.Children {
			c.checkChild(k)
		}
		return TNode
	}
	for _, a := range x.Attrs {
		if a.Spread != nil {
			c.fatal(x.Pos, "JSX attribute spread is not in the subset")
		}
		if n, ok := attrNameMap[a.Name]; ok {
			a.Name = n
		}
		if a.Val != nil {
			var want Type
			if strings.HasPrefix(a.Name, "on") {
				want = &FnT{Params: []*FnParam{{Name: "e", Type: TAny}}, Ret: TVoid}
			}
			c.check(a.Val, want)
			a.Reactive = c.mod.Kind != "server" && c.isReactive(a.Val)
		}
	}
	for _, k := range x.Children {
		c.checkChild(k)
	}
	return TNode
}

// ---------- 命名 ----------

var goKeywords = map[string]bool{"break": true, "case": true, "chan": true, "const": true, "continue": true, "default": true,
	"defer": true, "else": true, "fallthrough": true, "for": true, "func": true, "go": true, "goto": true, "if": true,
	"import": true, "interface": true, "map": true, "package": true, "range": true, "return": true, "select": true,
	"struct": true, "switch": true, "type": true, "var": true, "len": true, "cap": true, "new": true, "make": true,
	"append": true, "copy": true, "delete": true, "panic": true, "recover": true, "print": true, "println": true,
	"string": true, "int": true, "bool": true, "error": true, "nil": true, "true": true, "false": true, "float64": true}

func goIdent(name string) string {
	name = strings.ReplaceAll(name, "$", "_")
	if goKeywords[name] || name == "gotsx" || name == "host" || name == "props" {
		return name + "_"
	}
	return name
}

// goField: 把任意 JSON 键映射成合法的导出 Go 字段名(真实键由 json tag 保留)
func goField(name string) string {
	var b strings.Builder
	for _, c := range name {
		if unicode.IsLetter(c) || unicode.IsDigit(c) || c == '_' {
			b.WriteRune(c)
		} else {
			b.WriteByte('_') // 任意符号(含 @ - . :)→ _
		}
	}
	s := b.String()
	if s == "" {
		return "F"
	}
	r := []rune(s)
	if unicode.IsLetter(r[0]) {
		return string(unicode.ToUpper(r[0])) + string(r[1:])
	}
	return "F" + s // 首字符是数字 / 下划线 / 被替换的符号: 前缀 F 保证合法且导出
}

// ReadHostJSON 读取 app/.gen/host.json(可能不存在)
func ReadHostJSON(appDir string) []byte {
	b, _ := os.ReadFile(filepath.Join(appDir, ".gen", "host.json"))
	return b
}

// goPrefixOf: 模块内非导出符号的 Go 名前缀。用 app/ 下的相对路径(pages_docs__layout_), 同名文件不同目录不冲突
func goPrefixOf(m *Module) string {
	name := m.Name
	if i := strings.LastIndex(filepath.ToSlash(m.File), "/app/"); i >= 0 {
		rel := filepath.ToSlash(m.File)[i+5:]
		for _, suf := range []string{".tsx", ".server", ".client"} {
			rel = strings.TrimSuffix(rel, suf)
		}
		name = rel
	}
	return sanitizeIdent(name) + "_"
}

func underPages(m *Module) bool {
	return strings.Contains(filepath.ToSlash(m.File), "/app/pages/")
}

func sanitizeIdent(s string) string {
	var b strings.Builder
	for i, r := range s {
		if r == '_' || unicode.IsLetter(r) || (i > 0 && unicode.IsDigit(r)) {
			b.WriteRune(r)
		} else {
			b.WriteByte('_')
		}
	}
	return b.String()
}
