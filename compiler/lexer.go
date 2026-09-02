package compiler

import (
	"strconv"
	"strings"
	"unicode"
)

type TokKind int

const (
	TEOF TokKind = iota
	TIdent
	TNum
	TStr
	TTemplate
	TPunct
	TJSXText
	TRegex
)

type Token struct {
	Kind     TokKind
	Val      string
	Num      float64
	Pos      Pos
	NL       bool // 前面有换行
	Quasis   []string
	ExprSrcs []string
	ExprPos  []Pos
	Flags    string // 正则的 flags
}

type lexState struct {
	i, line, col int
	lastKind     TokKind
	lastVal      string
}

type lexer struct {
	src       []rune
	i         int
	line, col int
	file      string
	lastKind  TokKind // 上一个 token: 决定 / 是除号还是正则开头
	lastVal   string
}

func newLexer(src, file string, start Pos) *lexer {
	l := &lexer{src: []rune(src), line: 1, col: 1, file: file}
	if start.Line > 0 {
		l.line, l.col = start.Line, start.Col
	}
	return l
}

func (l *lexer) state() lexState { return lexState{l.i, l.line, l.col, l.lastKind, l.lastVal} }
func (l *lexer) restore(s lexState) {
	l.i, l.line, l.col, l.lastKind, l.lastVal = s.i, s.line, s.col, s.lastKind, s.lastVal
}
func (l *lexer) pos() Pos  { return Pos{l.file, l.line, l.col} }
func (l *lexer) eof() bool { return l.i >= len(l.src) }
func (l *lexer) at(off int) rune {
	if l.i+off < len(l.src) {
		return l.src[l.i+off]
	}
	return 0
}
func (l *lexer) adv() rune {
	r := l.src[l.i]
	l.i++
	if r == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}
	return r
}
func (l *lexer) hasPrefix(s string) bool {
	for i, r := range []rune(s) {
		if l.at(i) != r {
			return false
		}
	}
	return true
}

func (l *lexer) skipSpace() bool {
	nl := false
	for !l.eof() {
		r := l.at(0)
		switch {
		case r == '\n':
			nl = true
			l.adv()
		case unicode.IsSpace(r):
			l.adv()
		case r == '/' && l.at(1) == '/':
			for !l.eof() && l.at(0) != '\n' {
				l.adv()
			}
		case r == '/' && l.at(1) == '*':
			l.adv()
			l.adv()
			for !l.eof() && !(l.at(0) == '*' && l.at(1) == '/') {
				if l.at(0) == '\n' {
					nl = true
				}
				l.adv()
			}
			if !l.eof() {
				l.adv()
				l.adv()
			}
		default:
			return nl
		}
	}
	return nl
}

var puncts = []string{"...", "===", "!==", "=>", "==", "!=", "<=", ">=", "&&", "||", "??", "?.", "++", "--", "+=", "-=", "*=", "/=", "%=",
	"{", "}", "(", ")", "[", "]", ";", ",", ".", "<", ">", "+", "-", "*", "/", "%", "!", "&", "|", "?", ":", "=", "~", "^"}

func isIdentStart(r rune) bool { return r == '_' || r == '$' || unicode.IsLetter(r) }
func isIdentPart(r rune) bool  { return isIdentStart(r) || unicode.IsDigit(r) }

// regexAllowed: / 出现在"值"之后是除号, 否则是正则字面量的开头
var regexAfterKw = map[string]bool{"return": true, "typeof": true, "delete": true, "case": true, "throw": true, "else": true, "in": true, "of": true, "void": true, "new": true, "await": true, "yield": true}

func (l *lexer) regexAllowed() bool {
	switch l.lastKind {
	case TIdent:
		return regexAfterKw[l.lastVal]
	case TNum, TStr, TTemplate, TRegex:
		return false
	case TPunct:
		// ) ] } 之后是除号; < 之后的 / 是 JSX 闭标签 </x>(代价: 不能写 a < /re/.test(s), 加括号即可)
		return l.lastVal != ")" && l.lastVal != "]" && l.lastVal != "}" && l.lastVal != "<"
	}
	return true
}

func (l *lexer) next() Token {
	t := l.scan()
	l.lastKind, l.lastVal = t.Kind, t.Val
	return t
}

func (l *lexer) scan() Token {
	nl := l.skipSpace()
	p := l.pos()
	if l.eof() {
		return Token{Kind: TEOF, Pos: p, NL: nl}
	}
	r := l.at(0)
	if r == '/' && l.at(1) != '/' && l.at(1) != '*' && l.regexAllowed() {
		return l.regex(p, nl)
	}
	switch {
	case isIdentStart(r):
		start := l.i
		for !l.eof() && isIdentPart(l.at(0)) {
			l.adv()
		}
		return Token{Kind: TIdent, Val: string(l.src[start:l.i]), Pos: p, NL: nl}
	case unicode.IsDigit(r) || (r == '.' && unicode.IsDigit(l.at(1))):
		return l.number(p, nl)
	case r == '"' || r == '\'':
		return l.str(p, nl)
	case r == '`':
		return l.template(p, nl)
	}
	for _, pt := range puncts {
		if l.hasPrefix(pt) {
			for range []rune(pt) {
				l.adv()
			}
			return Token{Kind: TPunct, Val: pt, Pos: p, NL: nl}
		}
	}
	l.adv()
	return Token{Kind: TPunct, Val: string(r), Pos: p, NL: nl}
}

// regex: /pattern/flags —— 字符类 [...] 里的 / 不需要转义
func (l *lexer) regex(p Pos, nl bool) Token {
	l.adv() // /
	var b strings.Builder
	inClass := false
	for !l.eof() && l.at(0) != '\n' {
		c := l.at(0)
		if c == '\\' {
			b.WriteRune(l.adv())
			if !l.eof() {
				b.WriteRune(l.adv())
			}
			continue
		}
		if c == '[' {
			inClass = true
		} else if c == ']' {
			inClass = false
		} else if c == '/' && !inClass {
			break
		}
		b.WriteRune(l.adv())
	}
	if l.at(0) == '/' {
		l.adv()
	}
	start := l.i
	for !l.eof() && isIdentPart(l.at(0)) {
		l.adv()
	}
	return Token{Kind: TRegex, Val: b.String(), Flags: string(l.src[start:l.i]), Pos: p, NL: nl}
}

func (l *lexer) number(p Pos, nl bool) Token {
	start := l.i
	for !l.eof() && (unicode.IsDigit(l.at(0)) || l.at(0) == '.' || l.at(0) == '_') {
		l.adv()
	}
	if l.at(0) == 'e' || l.at(0) == 'E' {
		l.adv()
		if l.at(0) == '+' || l.at(0) == '-' {
			l.adv()
		}
		for !l.eof() && unicode.IsDigit(l.at(0)) {
			l.adv()
		}
	}
	raw := strings.ReplaceAll(string(l.src[start:l.i]), "_", "")
	f, _ := strconv.ParseFloat(raw, 64)
	return Token{Kind: TNum, Val: raw, Num: f, Pos: p, NL: nl}
}

func (l *lexer) escape() string {
	e := l.adv()
	switch e {
	case 'n':
		return "\n"
	case 't':
		return "\t"
	case 'r':
		return "\r"
	case '0':
		return "\x00"
	case 'u':
		if l.at(0) == '{' {
			l.adv()
			start := l.i
			for !l.eof() && l.at(0) != '}' {
				l.adv()
			}
			hex := string(l.src[start:l.i])
			l.adv()
			v, _ := strconv.ParseInt(hex, 16, 32)
			return string(rune(v))
		}
		hex := string(l.src[l.i : l.i+4])
		for i := 0; i < 4; i++ {
			l.adv()
		}
		v, _ := strconv.ParseInt(hex, 16, 32)
		return string(rune(v))
	case '\n':
		return ""
	}
	return string(e)
}

func (l *lexer) str(p Pos, nl bool) Token {
	q := l.adv()
	var b strings.Builder
	for !l.eof() && l.at(0) != q {
		if l.at(0) == '\\' {
			l.adv()
			b.WriteString(l.escape())
			continue
		}
		b.WriteRune(l.adv())
	}
	if !l.eof() {
		l.adv()
	}
	return Token{Kind: TStr, Val: b.String(), Pos: p, NL: nl}
}

func (l *lexer) template(p Pos, nl bool) Token {
	l.adv() // `
	t := Token{Kind: TTemplate, Pos: p, NL: nl}
	var cur strings.Builder
	for !l.eof() {
		r := l.at(0)
		if r == '`' {
			l.adv()
			break
		}
		if r == '\\' {
			l.adv()
			cur.WriteString(l.escape())
			continue
		}
		if r == '$' && l.at(1) == '{' {
			l.adv()
			l.adv()
			t.Quasis = append(t.Quasis, cur.String())
			cur.Reset()
			ep := l.pos()
			start := l.i
			depth := 0
			for !l.eof() {
				c := l.at(0)
				if c == '}' && depth == 0 {
					break
				}
				switch c {
				case '{':
					depth++
					l.adv()
				case '}':
					depth--
					l.adv()
				case '"', '\'':
					l.str(l.pos(), false)
				case '`':
					l.template(l.pos(), false)
				default:
					l.adv()
				}
			}
			t.ExprSrcs = append(t.ExprSrcs, string(l.src[start:l.i]))
			t.ExprPos = append(t.ExprPos, ep)
			if !l.eof() {
				l.adv() // }
			}
			continue
		}
		cur.WriteRune(l.adv())
	}
	t.Quasis = append(t.Quasis, cur.String())
	return t
}

// JSX 子节点文本: 读到 < 或 { 为止(不消费)
func (l *lexer) jsxText() Token {
	p := l.pos()
	start := l.i
	for !l.eof() && l.at(0) != '<' && l.at(0) != '{' {
		l.adv()
	}
	return Token{Kind: TJSXText, Val: decodeEntities(string(l.src[start:l.i])), Pos: p}
}

var entities = map[string]string{"&amp;": "&", "&lt;": "<", "&gt;": ">", "&quot;": "\"", "&#39;": "'", "&apos;": "'", "&nbsp;": " ", "&middot;": "·", "&copy;": "©"}

func decodeEntities(s string) string {
	if !strings.Contains(s, "&") {
		return s
	}
	for k, v := range entities {
		s = strings.ReplaceAll(s, k, v)
	}
	return s
}
