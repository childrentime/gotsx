package compiler

import (
	"fmt"
	"path/filepath"
	"strings"
)

type parser struct {
	lx   *lexer
	tok  Token
	file string
}

type parseError struct {
	Pos Pos
	Msg string
}

func (e *parseError) Error() string { return fmt.Sprintf("%s: %s", e.Pos, e.Msg) }

type pstate struct {
	ls  lexState
	tok Token
}

func newParser(src, file string, start Pos) *parser {
	p := &parser{lx: newLexer(src, file, start), file: file}
	p.next()
	return p
}

func (p *parser) next()              { p.tok = p.lx.next() }
func (p *parser) save() pstate       { return pstate{p.lx.state(), p.tok} }
func (p *parser) restore(s pstate)   { p.lx.restore(s.ls); p.tok = s.tok }
func (p *parser) is(v string) bool   { return p.tok.Kind == TPunct && p.tok.Val == v }
func (p *parser) isKw(v string) bool { return p.tok.Kind == TIdent && p.tok.Val == v }
func (p *parser) accept(v string) bool {
	if p.is(v) {
		p.next()
		return true
	}
	return false
}
func (p *parser) expect(v string) {
	if !p.accept(v) {
		p.fail("expected %q, got %s", v, p.describe())
	}
}
func (p *parser) describe() string {
	switch p.tok.Kind {
	case TEOF:
		return "end of file"
	case TStr:
		return fmt.Sprintf("string %q", p.tok.Val)
	case TTemplate:
		return "template string"
	case TRegex:
		return "regex literal"
	case TJSXText:
		return "JSX text"
	}
	return fmt.Sprintf("%q", p.tok.Val)
}
func (p *parser) fail(format string, args ...any) {
	panic(&parseError{Pos: p.tok.Pos, Msg: fmt.Sprintf(format, args...)})
}
func (p *parser) ident() string {
	if p.tok.Kind != TIdent {
		p.fail("expected an identifier, got %s", p.describe())
	}
	n := p.tok.Val
	p.next()
	return n
}
func (p *parser) peekKw(v string) bool {
	s := p.save()
	p.next()
	ok := p.isKw(v)
	p.restore(s)
	return ok
}
func (p *parser) peekIs(v string) bool {
	s := p.save()
	p.next()
	ok := p.is(v)
	p.restore(s)
	return ok
}

// ParseModule 解析一个 .tsx 文件
func ParseModule(src, file string) (m *Module, err error) {
	defer func() {
		if r := recover(); r != nil {
			if pe, ok := r.(*parseError); ok {
				err = pe
				return
			}
			panic(r)
		}
	}()
	p := newParser(src, file, Pos{})
	m = &Module{File: file, Dir: filepath.Dir(file)}
	base := filepath.Base(file)
	base = strings.TrimSuffix(base, ".tsx")
	switch {
	case strings.HasSuffix(base, ".server"):
		m.Kind, m.Name = "server", strings.TrimSuffix(base, ".server")
	case strings.HasSuffix(base, ".client"):
		m.Kind, m.Name = "client", strings.TrimSuffix(base, ".client")
	default:
		m.Kind, m.Name = "shared", base
	}
	for p.tok.Kind != TEOF {
		if p.isKw("import") {
			m.Imports = append(m.Imports, p.parseImport())
			continue
		}
		s := p.parseStmt()
		if _, ok := s.(*EmptyStmt); !ok {
			m.Stmts = append(m.Stmts, s)
		}
	}
	return m, nil
}

func (p *parser) parseImport() *Import {
	im := &Import{Pos: p.tok.Pos}
	p.next() // import
	if p.isKw("type") && !p.peekIs(",") && !p.peekKw("from") {
		im.TypeOnly = true
		p.next()
	}
	if p.tok.Kind == TStr { // import "x"
		im.From = p.tok.Val
		p.next()
		p.accept(";")
		return im
	}
	if p.tok.Kind == TIdent && !p.is("{") {
		im.Default = p.ident()
		p.accept(",")
	}
	if p.accept("{") {
		for !p.is("}") {
			spec := ImportSpec{}
			if p.isKw("type") && !p.peekIs(",") && !p.peekIs("}") && !p.peekKw("as") {
				spec.TypeOnly = true
				p.next()
			}
			spec.Name = p.ident()
			spec.Local = spec.Name
			if p.isKw("as") {
				p.next()
				spec.Local = p.ident()
			}
			im.Names = append(im.Names, spec)
			if !p.accept(",") {
				break
			}
		}
		p.expect("}")
	}
	if !p.isKw("from") {
		p.fail("import is missing 'from'")
	}
	p.next()
	if p.tok.Kind != TStr {
		p.fail("import is missing the module path")
	}
	im.From = p.tok.Val
	p.next()
	p.accept(";")
	return im
}

func (p *parser) parseStmt() Stmt {
	pos := p.tok.Pos
	export := false
	if p.isKw("export") {
		p.next()
		export = true
		if p.isKw("default") {
			p.next()
			if p.isKw("function") || (p.isKw("async") && p.peekKw("function")) {
				f := p.parseFunc()
				f.Export, f.Default = true, true
				return f
			}
			x := p.parseExpr()
			p.accept(";")
			return &ExportDefault{Pos: pos, X: x}
		}
	}
	switch {
	case p.isKw("function") || (p.isKw("async") && p.peekKw("function")):
		f := p.parseFunc()
		f.Export = export
		return f
	case p.isKw("const") || p.isKw("let") || p.isKw("var"):
		d := p.parseVarDecl()
		d.Export = export
		return d
	case p.isKw("interface"):
		p.next()
		d := &InterfaceDecl{Pos: pos, Name: p.ident(), Export: export}
		if p.isKw("extends") {
			p.next()
			for {
				d.Extends = append(d.Extends, p.ident())
				if !p.accept(",") {
					break
				}
			}
		}
		d.Fields = p.parseObjTypeFields()
		return d
	case p.isKw("type") && p.peekIsIdentThenEq():
		p.next()
		d := &TypeAlias{Pos: pos, Name: p.ident(), Export: export}
		p.expect("=")
		d.Type = p.parseType()
		p.accept(";")
		return d
	case export:
		p.fail("export must be followed by function / const / interface / type")
	case p.isKw("return"):
		p.next()
		r := &ReturnStmt{Pos: pos}
		if !p.is(";") && !p.is("}") && !p.tok.NL && p.tok.Kind != TEOF {
			r.X = p.parseExpr()
		}
		p.accept(";")
		return r
	case p.isKw("if"):
		p.next()
		p.expect("(")
		s := &IfStmt{Pos: pos, Cond: p.parseExpr()}
		p.expect(")")
		s.Then = p.parseBlockOrStmt()
		if p.isKw("else") {
			p.next()
			if p.isKw("if") {
				s.Else = p.parseStmt()
			} else {
				s.Else = p.parseBlockOrStmt()
			}
		}
		return s
	case p.isKw("for"):
		return p.parseFor(pos)
	case p.isKw("while"):
		p.next()
		p.expect("(")
		s := &WhileStmt{Pos: pos, Cond: p.parseExpr()}
		p.expect(")")
		s.Body = p.parseBlockOrStmt()
		return s
	case p.isKw("do"):
		p.fail("do-while is not in the subset; rewrite it as a while loop")
	case p.isKw("break"):
		p.next()
		p.accept(";")
		return &BreakStmt{Pos: pos}
	case p.isKw("continue"):
		p.next()
		p.accept(";")
		return &ContinueStmt{Pos: pos}
	case p.isKw("switch"):
		return p.parseSwitch(pos)
	case p.isKw("try"):
		p.next()
		s := &TryStmt{Pos: pos, Body: p.parseBlock()}
		if p.isKw("catch") {
			p.next()
			if p.accept("(") {
				s.CatchName = p.ident()
				if p.accept(":") {
					p.parseType()
				}
				p.expect(")")
			}
			s.Catch = p.parseBlock()
		}
		if p.isKw("finally") {
			p.next()
			s.Finally = p.parseBlock()
		}
		return s
	case p.isKw("throw"):
		p.next()
		s := &ThrowStmt{Pos: pos, X: p.parseExpr()}
		p.accept(";")
		return s
	case p.is("{"):
		return p.parseBlock()
	case p.is(";"):
		p.next()
		return &EmptyStmt{Pos: pos}
	}
	x := p.parseExpr()
	p.accept(";")
	return &ExprStmt{Pos: pos, X: x}
}

func (p *parser) peekIsIdentThenEq() bool {
	s := p.save()
	defer p.restore(s)
	p.next()
	if p.tok.Kind != TIdent {
		return false
	}
	p.next()
	return p.is("=") || p.is("<")
}

// for (const x of xs) | for (let i = 0; i < n; i++) | for (; cond; )
func (p *parser) parseFor(pos Pos) Stmt {
	p.next() // for
	p.expect("(")
	if p.isKw("const") || p.isKw("let") || p.isKw("var") {
		declPos := p.tok.Pos
		isConst := p.isKw("const")
		p.next()
		pat := p.parsePattern()
		if p.isKw("of") {
			p.next()
			s := &ForOfStmt{Pos: pos, Pat: pat}
			s.Iter = p.parseExpr()
			p.expect(")")
			s.Body = p.parseBlockOrStmt()
			return s
		}
		if p.isKw("in") {
			p.fail("for-in is not in the subset; use for-of over Object.keys(obj)")
		}
		d := &VarDecl{Pos: declPos, Const: isConst, Pat: pat}
		if p.accept(":") {
			d.Type = p.parseType()
		}
		if p.accept("=") {
			d.Init = p.parseExpr()
		}
		if p.is(",") {
			p.fail("declaring several variables in one statement is not in the subset")
		}
		p.expect(";")
		return p.parseForRest(pos, d)
	}
	var init Stmt
	if !p.is(";") {
		init = &ExprStmt{Pos: p.tok.Pos, X: p.parseExpr()}
	}
	p.expect(";")
	return p.parseForRest(pos, init)
}

func (p *parser) parseForRest(pos Pos, init Stmt) Stmt {
	s := &ForStmt{Pos: pos, Init: init}
	if !p.is(";") {
		s.Cond = p.parseExpr()
	}
	p.expect(";")
	if !p.is(")") {
		s.Update = p.parseExpr()
	}
	p.expect(")")
	s.Body = p.parseBlockOrStmt()
	return s
}

func (p *parser) parseSwitch(pos Pos) Stmt {
	p.next() // switch
	p.expect("(")
	s := &SwitchStmt{Pos: pos, Disc: p.parseExpr()}
	p.expect(")")
	p.expect("{")
	for !p.is("}") {
		c := &SwitchCase{Pos: p.tok.Pos}
		switch {
		case p.isKw("case"):
			p.next()
			c.Test = p.parseExpr()
		case p.isKw("default"):
			p.next()
		default:
			p.fail("expected case / default inside switch, got %s", p.describe())
		}
		p.expect(":")
		for !p.isKw("case") && !p.isKw("default") && !p.is("}") {
			if p.tok.Kind == TEOF {
				p.fail("unterminated switch")
			}
			st := p.parseStmt()
			if _, ok := st.(*EmptyStmt); !ok {
				c.Body = append(c.Body, st)
			}
		}
		s.Cases = append(s.Cases, c)
	}
	p.next() // }
	return s
}

func (p *parser) parseBlockOrStmt() *Block {
	if p.is("{") {
		return p.parseBlock()
	}
	s := p.parseStmt()
	return &Block{Pos: p.tok.Pos, Stmts: []Stmt{s}}
}

func (p *parser) parseBlock() *Block {
	b := &Block{Pos: p.tok.Pos}
	p.expect("{")
	for !p.is("}") {
		if p.tok.Kind == TEOF {
			p.fail("unterminated block")
		}
		s := p.parseStmt()
		if _, ok := s.(*EmptyStmt); !ok {
			b.Stmts = append(b.Stmts, s)
		}
	}
	p.next()
	return b
}

func (p *parser) parseFunc() *FuncDecl {
	f := &FuncDecl{Pos: p.tok.Pos}
	if p.isKw("async") {
		f.Async = true
		p.next()
	}
	p.next() // function
	f.Name = p.ident()
	if p.is("<") {
		p.fail("generic functions are not in the subset")
	}
	f.Params = p.parseParams()
	if p.accept(":") {
		f.Ret = p.parseType()
	}
	f.Body = p.parseBlock()
	return f
}

func (p *parser) parseParams() []*Param {
	p.expect("(")
	var ps []*Param
	for !p.is(")") {
		pr := &Param{Pos: p.tok.Pos, Pat: p.parsePattern()}
		if p.accept("?") {
			pr.Optional = true
		}
		if p.accept(":") {
			pr.Type = p.parseType()
		}
		if p.accept("=") {
			pr.Default = p.parseAssign()
		}
		ps = append(ps, pr)
		if !p.accept(",") {
			break
		}
	}
	p.expect(")")
	return ps
}

func (p *parser) parsePattern() *Pattern {
	pat := &Pattern{Pos: p.tok.Pos}
	switch {
	case p.is("{"):
		pat.Kind = PatObject
		p.next()
		for !p.is("}") {
			if p.accept("...") {
				pat.Rest = p.parsePattern()
			} else {
				key := p.ident()
				pp := &PatProp{Key: key}
				if p.accept(":") {
					pp.Pat = p.parsePattern()
				} else {
					pp.Pat = &Pattern{Pos: pat.Pos, Kind: PatIdent, Name: key}
				}
				if p.accept("=") {
					pp.Default = p.parseAssign()
				}
				pat.Props = append(pat.Props, pp)
			}
			if !p.accept(",") {
				break
			}
		}
		p.expect("}")
	case p.is("["):
		pat.Kind = PatArray
		p.next()
		for !p.is("]") {
			if p.accept("...") {
				pat.Rest = p.parsePattern()
			} else {
				pat.Elems = append(pat.Elems, p.parsePattern())
			}
			if !p.accept(",") {
				break
			}
		}
		p.expect("]")
	default:
		pat.Kind = PatIdent
		pat.Name = p.ident()
	}
	return pat
}

func (p *parser) parseVarDecl() *VarDecl {
	d := &VarDecl{Pos: p.tok.Pos, Const: p.isKw("const")}
	p.next()
	d.Pat = p.parsePattern()
	if p.accept(":") {
		d.Type = p.parseType()
	}
	if p.accept("=") {
		d.Init = p.parseExpr()
	}
	if p.is(",") {
		p.fail("declaring several variables in one statement is not in the subset")
	}
	p.accept(";")
	return d
}

// ---------- 类型 ----------

func (p *parser) parseType() TypeExpr {
	p.accept("|")
	first := p.parsePostfixType()
	if !p.is("|") {
		return first
	}
	u := &TUnion{Members: []TypeExpr{first}}
	for p.accept("|") {
		u.Members = append(u.Members, p.parsePostfixType())
	}
	return u
}

func (p *parser) parsePostfixType() TypeExpr {
	t := p.parsePrimaryType()
	for p.is("[") && p.peekIs("]") {
		p.next()
		p.next()
		t = &TArr{Elem: t}
	}
	return t
}

func (p *parser) parsePrimaryType() TypeExpr {
	pos := p.tok.Pos
	switch {
	case p.tok.Kind == TStr:
		v := p.tok.Val
		p.next()
		return &TStrLit{Val: v}
	case p.tok.Kind == TIdent:
		if p.isKw("typeof") || p.isKw("keyof") {
			p.fail("%s types are not in the subset", p.tok.Val)
		}
		r := &TRef{Pos: pos, Name: p.ident()}
		for p.accept(".") {
			r.Name += "." + p.ident()
		}
		if p.accept("<") {
			for !p.is(">") {
				r.Args = append(r.Args, p.parseType())
				if !p.accept(",") {
					break
				}
			}
			p.expect(">")
		}
		return r
	case p.is("{"):
		return &TObj{Fields: p.parseObjTypeFields()}
	case p.is("("):
		s := p.save()
		p.next()
		isFn := p.is(")")
		if !isFn && p.tok.Kind == TIdent {
			p.next()
			isFn = p.is(":") || p.is(",") || p.is("?") || (p.is(")") && p.peekIs("=>"))
		}
		p.restore(s)
		if isFn {
			f := &TFunc{Params: p.parseParams()}
			p.expect("=>")
			f.Ret = p.parseType()
			return f
		}
		p.next()
		t := p.parseType()
		p.expect(")")
		return t
	}
	p.fail("cannot parse a type here, got %s", p.describe())
	return nil
}

func (p *parser) parseObjTypeFields() []*TypeField {
	p.expect("{")
	var fs []*TypeField
	for !p.is("}") {
		f := &TypeField{}
		if p.tok.Kind == TStr {
			f.Name = p.tok.Val
			p.next()
		} else {
			f.Name = p.ident()
		}
		if p.accept("?") {
			f.Optional = true
		}
		if p.is("(") {
			f.Method = true
			f.Params = p.parseParams()
			p.expect(":")
			f.Type = p.parseType()
		} else {
			p.expect(":")
			f.Type = p.parseType()
		}
		fs = append(fs, f)
		if !p.accept(";") && !p.accept(",") {
			if !p.is("}") && !p.tok.NL {
				p.fail("type members must be separated by ; or a newline")
			}
		}
	}
	p.next()
	return fs
}

// ---------- 表达式 ----------

func (p *parser) parseExpr() Expr { return p.parseAssign() }

func (p *parser) parseAssign() Expr {
	if a := p.tryArrow(); a != nil {
		return a
	}
	l := p.parseCond()
	if p.tok.Kind == TPunct {
		switch p.tok.Val {
		case "=", "+=", "-=", "*=", "/=", "%=":
			op := p.tok.Val
			pos := p.tok.Pos
			p.next()
			r := p.parseAssign()
			return &Assign{base: base{Pos: pos}, Op: op, Target: l, Val: r}
		}
	}
	return l
}

// 箭头函数: (a, b) => ...  |  x => ...  |  async (...) => ...
func (p *parser) tryArrow() (result Expr) {
	s := p.save()
	pos := p.tok.Pos
	async := false
	if p.isKw("async") && (p.peekIs("(") || p.peekIsIdentThenArrow()) {
		async = true
		p.next()
	}
	if p.tok.Kind == TIdent && p.peekIs("=>") {
		name := p.ident()
		p.next() // =>
		a := &Arrow{base: base{Pos: pos}, Async: async, Params: []*Param{{Pos: pos, Pat: &Pattern{Pos: pos, Kind: PatIdent, Name: name}}}}
		p.parseArrowBody(a)
		return a
	}
	if !p.is("(") {
		p.restore(s)
		return nil
	}
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(*parseError); ok {
				p.restore(s)
				result = nil
				return
			}
			panic(r)
		}
	}()
	params := p.parseParams()
	var ret TypeExpr
	if p.accept(":") {
		ret = p.parseType()
	}
	if !p.is("=>") {
		p.restore(s)
		return nil
	}
	p.next()
	a := &Arrow{base: base{Pos: pos}, Async: async, Params: params, Ret: ret}
	p.parseArrowBody(a)
	return a
}

func (p *parser) peekIsIdentThenArrow() bool {
	s := p.save()
	defer p.restore(s)
	p.next()
	if p.tok.Kind != TIdent {
		return false
	}
	p.next()
	return p.is("=>")
}

func (p *parser) parseArrowBody(a *Arrow) {
	if p.is("{") {
		a.Body = p.parseBlock()
	} else {
		a.ExprBody = p.parseAssign()
	}
}

func (p *parser) parseCond() Expr {
	c := p.parseBinary(0)
	if p.is("?") {
		pos := p.tok.Pos
		p.next()
		a := p.parseAssign()
		p.expect(":")
		b := p.parseAssign()
		return &CondExpr{base: base{Pos: pos}, Test: c, Then: a, Else: b}
	}
	return c
}

var binPrec = map[string]int{"??": 1, "||": 2, "&&": 3, "|": 4, "^": 5, "&": 6,
	"==": 7, "!=": 7, "===": 7, "!==": 7, "<": 8, ">": 8, "<=": 8, ">=": 8,
	"+": 10, "-": 10, "*": 11, "/": 11, "%": 11}

func (p *parser) parseBinary(minPrec int) Expr {
	l := p.parseUnary()
	for p.tok.Kind == TPunct {
		prec, ok := binPrec[p.tok.Val]
		if !ok || prec <= minPrec {
			break
		}
		op := p.tok.Val
		pos := p.tok.Pos
		p.next()
		r := p.parseBinary(prec)
		l = &Binary{base: base{Pos: pos}, Op: op, L: l, R: r}
	}
	return l
}

func (p *parser) parseUnary() Expr {
	pos := p.tok.Pos
	if p.tok.Kind == TPunct && (p.tok.Val == "!" || p.tok.Val == "-" || p.tok.Val == "+") {
		op := p.tok.Val
		p.next()
		return &Unary{base: base{Pos: pos}, Op: op, X: p.parseUnary()}
	}
	if p.tok.Kind == TPunct && (p.tok.Val == "++" || p.tok.Val == "--") {
		op := p.tok.Val
		p.next()
		return &Update{base: base{Pos: pos}, Op: op, X: p.parseUnary(), Prefix: true}
	}
	if p.isKw("await") {
		p.next()
		return &AwaitExpr{base: base{Pos: pos}, X: p.parseUnary()}
	}
	if p.isKw("typeof") {
		p.next()
		return &Unary{base: base{Pos: pos}, Op: "typeof", X: p.parseUnary()}
	}
	if p.isKw("delete") {
		p.next()
		return &Unary{base: base{Pos: pos}, Op: "delete", X: p.parseUnary()}
	}
	return p.parsePostfix()
}

func (p *parser) parsePostfix() Expr {
	x := p.parsePrimary()
	for {
		pos := p.tok.Pos
		switch {
		case p.is("."):
			p.next()
			x = &Member{base: base{Pos: pos}, X: x, Name: p.ident()}
		case p.is("?."):
			p.next()
			if p.is("(") {
				x = p.parseCallArgs(x, pos, true)
			} else if p.is("[") {
				p.next()
				i := p.parseExpr()
				p.expect("]")
				x = &Index{base: base{Pos: pos}, X: x, I: i}
			} else {
				x = &Member{base: base{Pos: pos}, X: x, Name: p.ident(), Optional: true}
			}
		case p.is("["):
			p.next()
			i := p.parseExpr()
			p.expect("]")
			x = &Index{base: base{Pos: pos}, X: x, I: i}
		case p.is("("):
			x = p.parseCallArgs(x, pos, false)
		case p.is("<") && p.looksLikeTypeArgs():
			s := p.save()
			p.next()
			var targs []TypeExpr
			ok := true
			func() {
				defer func() {
					if r := recover(); r != nil {
						if _, isPE := r.(*parseError); isPE {
							ok = false
							return
						}
						panic(r)
					}
				}()
				for !p.is(">") {
					targs = append(targs, p.parseType())
					if !p.accept(",") {
						break
					}
				}
				p.expect(">")
			}()
			if !ok || !p.is("(") {
				p.restore(s)
				return x
			}
			c := p.parseCallArgs(x, pos, false)
			c.(*Call).TypeArgs = targs
			x = c
		case p.is("!") && !p.tok.NL:
			p.next()
			x = &NonNull{base: base{Pos: pos}, X: x}
		case (p.is("++") || p.is("--")) && !p.tok.NL:
			op := p.tok.Val
			p.next()
			x = &Update{base: base{Pos: pos}, Op: op, X: x}
		case p.isKw("as"):
			p.next()
			x = &AsExpr{base: base{Pos: pos}, X: x, Type: p.parseType()}
		default:
			return x
		}
	}
}

func (p *parser) looksLikeTypeArgs() bool {
	// x<T>(...) : 只对标识符/成员调用尝试
	s := p.save()
	defer p.restore(s)
	p.next()
	return p.tok.Kind == TIdent || p.is("{") || p.is("(") || p.tok.Kind == TStr
}

func (p *parser) parseCallArgs(fn Expr, pos Pos, optional bool) Expr {
	p.expect("(")
	c := &Call{base: base{Pos: pos}, Fn: fn, Optional: optional}
	for !p.is(")") {
		if p.accept("...") {
			c.Args = append(c.Args, &SpreadExpr{base: base{Pos: pos}, X: p.parseAssign()})
		} else {
			c.Args = append(c.Args, p.parseAssign())
		}
		if !p.accept(",") {
			break
		}
	}
	p.expect(")")
	return c
}

func (p *parser) parsePrimary() Expr {
	pos := p.tok.Pos
	switch p.tok.Kind {
	case TNum:
		t := p.tok
		p.next()
		return &NumLit{base: base{Pos: pos}, Val: t.Num, Raw: t.Val}
	case TStr:
		v := p.tok.Val
		p.next()
		return &StrLit{base: base{Pos: pos}, Val: v}
	case TRegex:
		t := p.tok
		p.next()
		return &RegexLit{base: base{Pos: pos}, Pattern: t.Val, Flags: t.Flags}
	case TTemplate:
		t := p.tok
		p.next()
		tl := &TemplateLit{base: base{Pos: pos}, Quasis: t.Quasis}
		for i, src := range t.ExprSrcs {
			sub := newParser(src, p.file, t.ExprPos[i])
			e := sub.parseExpr()
			if sub.tok.Kind != TEOF {
				sub.fail("unexpected trailing content in a template-string expression")
			}
			tl.Exprs = append(tl.Exprs, e)
		}
		return tl
	case TIdent:
		switch p.tok.Val {
		case "true", "false":
			v := p.tok.Val == "true"
			p.next()
			return &BoolLit{base: base{Pos: pos}, Val: v}
		case "null", "undefined":
			p.next()
			return &NullLit{base: base{Pos: pos}}
		case "function":
			p.fail("function expressions are not in the subset; use an arrow function")
		case "new", "class", "this", "yield", "delete", "void", "in", "instanceof":
			p.fail("%s is not in the subset", p.tok.Val)
		}
		name := p.ident()
		return &Ident{base: base{Pos: pos}, Name: name}
	case TPunct:
		switch p.tok.Val {
		case "(":
			p.next()
			x := p.parseExpr()
			p.expect(")")
			return &Paren{base: base{Pos: pos}, X: x}
		case "[":
			p.next()
			a := &ArrayLit{base: base{Pos: pos}}
			for !p.is("]") {
				if p.accept("...") {
					a.Elems = append(a.Elems, &SpreadExpr{base: base{Pos: pos}, X: p.parseAssign()})
				} else {
					a.Elems = append(a.Elems, p.parseAssign())
				}
				if !p.accept(",") {
					break
				}
			}
			p.expect("]")
			return a
		case "{":
			return p.parseObjectLit()
		case "<":
			return p.parseJSX(false)
		}
	}
	p.fail("cannot parse an expression here, got %s", p.describe())
	return nil
}

func (p *parser) parseObjectLit() Expr {
	o := &ObjectLit{base: base{Pos: p.tok.Pos}}
	p.expect("{")
	for !p.is("}") {
		if p.accept("...") {
			o.Props = append(o.Props, &ObjProp{Spread: p.parseAssign()})
		} else {
			var key string
			if p.tok.Kind == TStr {
				key = p.tok.Val
				p.next()
			} else {
				key = p.ident()
			}
			if p.accept(":") {
				o.Props = append(o.Props, &ObjProp{Key: key, Val: p.parseAssign()})
			} else if p.is("(") {
				p.fail("object method shorthand is not in the subset")
			} else {
				o.Props = append(o.Props, &ObjProp{Key: key, Val: &Ident{base: base{Pos: p.tok.Pos}, Name: key}, Shorthand: true})
			}
		}
		if !p.accept(",") {
			break
		}
	}
	p.expect("}")
	return o
}

// ---------- JSX ----------
// 约定: 进入子节点模式时, p.tok 是刚扫到的 ">"/"}", 词法器位置在其后; 文本由 lx.jsxText 直接读。
// inChildren=true 时, 元素结束后不扫描下一个普通 token(留给父级读 JSX text)。

func (p *parser) parseJSX(inChildren bool) Expr {
	pos := p.tok.Pos
	p.next()       // 跳过 "<", 现在 p.tok 是标签名或 ">"
	if p.is(">") { // fragment
		f := &JSXFrag{base: base{Pos: pos}}
		f.Children = p.parseJSXChildren("")
		if !inChildren {
			p.next()
		}
		return f
	}
	el := &JSXElem{base: base{Pos: pos}}
	el.Tag = p.ident()
	for p.accept(".") {
		el.Tag += "." + p.ident()
	}
	for !p.is(">") && !p.is("/") {
		if p.is("{") {
			p.next()
			p.expect("...")
			el.Attrs = append(el.Attrs, &JSXAttr{Spread: p.parseExpr()})
			p.expect("}")
			continue
		}
		if p.tok.Kind != TIdent {
			p.fail("invalid JSX attribute name, got %s", p.describe())
		}
		name := p.tok.Val
		for p.lx.at(0) == '-' || p.lx.at(0) == ':' {
			name += string(p.lx.adv())
			for isIdentPart(p.lx.at(0)) {
				name += string(p.lx.adv())
			}
		}
		p.next()
		a := &JSXAttr{Name: name}
		if p.accept("=") {
			if p.tok.Kind == TStr {
				a.Val = &StrLit{base: base{Pos: p.tok.Pos}, Val: decodeEntities(p.tok.Val)}
				p.next()
			} else if p.is("{") {
				p.next()
				a.Val = p.parseExpr()
				p.expect("}")
			} else {
				p.fail("a JSX attribute value must be a string or {expression}")
			}
		}
		el.Attrs = append(el.Attrs, a)
	}
	if p.is("/") { // 自闭合
		p.next()
		if !p.is(">") {
			p.fail("self-closing tag is missing >")
		}
		if !inChildren {
			p.next()
		}
		return el
	}
	// p.tok == ">" : 进入子节点
	el.Children = p.parseJSXChildren(el.Tag)
	if !inChildren {
		p.next()
	}
	return el
}

// 进入时 p.tok 是开标签的 ">"; 返回时 p.tok 是闭标签的 ">"
func (p *parser) parseJSXChildren(tag string) []Expr {
	var kids []Expr
	for {
		txt := p.lx.jsxText()
		if txt.Val != "" {
			kids = append(kids, &JSXText{base: base{Pos: txt.Pos}, Text: txt.Val})
		}
		p.next() // "<" 或 "{"
		switch {
		case p.is("{"):
			pos := p.tok.Pos
			p.next()
			if p.is("}") {
				continue // {} 或 {/* 注释 */}
			}
			x := p.parseExpr()
			if !p.is("}") {
				p.fail("JSX expression is missing }")
			}
			kids = append(kids, &JSXExprChild{base: base{Pos: pos}, X: x})
		case p.is("<"):
			if p.lx.at(0) == '/' { // 闭标签
				p.next() // "/"
				p.next() // 标签名 或 ">"
				if tag == "" {
					if !p.is(">") {
						p.fail("a fragment must be closed with </>")
					}
					return kids
				}
				name := p.ident()
				for p.accept(".") {
					name += "." + p.ident()
				}
				if name != tag {
					p.fail("closing tag </%s> does not match opening tag <%s>", name, tag)
				}
				if !p.is(">") {
					p.fail("closing tag is missing >")
				}
				return kids
			}
			kids = append(kids, p.parseJSX(true))
		case p.tok.Kind == TEOF:
			p.fail("JSX <%s> is not closed", tag)
		default:
			p.fail("unexpected %s inside JSX", p.describe())
		}
	}
}
