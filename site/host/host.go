// Package host: 官网的宿主模块。host:hl 是 Go 写的语法高亮 tokenizer —— 方言拿到 token 数组自己渲染 span,
// 不需要任何"注入 HTML"的逃生舱。host:site 提供构建信息。
package host

import (
	"runtime"
	"strings"
	"time"
	"unicode"

	gotsx "github.com/childrentime/gotsx/runtime"
)

type Token struct {
	Kind string `json:"kind"` // kw str cmt num tag attr punct plain
	Text string `json:"text"`
}

type HlModule struct{}

var keywords = map[string]map[string]bool{
	"tsx":  set("import export from default function return const let var if else for of in async await try catch finally throw interface type true false null undefined new class extends as typeof"),
	"go":   set("package import func return var const type struct interface if else for range map string int float64 bool nil true false error go defer chan select switch case break continue any"),
	"css":  set("@import @theme @custom-variant @layer @apply"),
	"bash": set("cd go run build export echo curl"),
}

func set(s string) map[string]bool {
	m := map[string]bool{}
	for _, w := range strings.Fields(s) {
		m[w] = true
	}
	return m
}

// Tokens 把代码切成带类别的片段。lang: tsx | go | bash | css | json | html
func (HlModule) Tokens(code, lang string) []Token {
	kw := keywords[lang]
	if lang == "js" || lang == "ts" {
		kw = keywords["tsx"]
	}
	src := []rune(code)
	var out []Token
	emit := func(kind string, s string) {
		if s == "" {
			return
		}
		if n := len(out); n > 0 && out[n-1].Kind == kind && kind == "plain" {
			out[n-1].Text += s
			return
		}
		out = append(out, Token{Kind: kind, Text: s})
	}
	i := 0
	inTag := false // JSX/HTML 标签内部: 标识符是属性
	for i < len(src) {
		c := src[i]
		switch {
		case (c == '/' && i+1 < len(src) && src[i+1] == '/' && lang != "css" && lang != "html") || (c == '#' && lang == "bash"):
			j := i
			for j < len(src) && src[j] != '\n' {
				j++
			}
			emit("cmt", string(src[i:j]))
			i = j
		case c == '/' && i+1 < len(src) && src[i+1] == '*':
			j := i + 2
			for j+1 < len(src) && !(src[j] == '*' && src[j+1] == '/') {
				j++
			}
			j += 2
			if j > len(src) {
				j = len(src)
			}
			emit("cmt", string(src[i:j]))
			i = j
		case c == '<' && lang == "html" && i+3 < len(src) && string(src[i:i+4]) == "<!--":
			j := i + 4
			for j+2 < len(src) && string(src[j:j+3]) != "-->" {
				j++
			}
			j += 3
			if j > len(src) {
				j = len(src)
			}
			emit("cmt", string(src[i:j]))
			i = j
		case c == '"' || c == '\'' || c == '`':
			j := i + 1
			for j < len(src) && src[j] != c {
				if src[j] == '\\' {
					j++
				}
				j++
			}
			if j < len(src) {
				j++
			}
			emit("str", string(src[i:j]))
			i = j
		case unicode.IsDigit(c):
			j := i
			for j < len(src) && (unicode.IsDigit(src[j]) || src[j] == '.' || src[j] == '_') {
				j++
			}
			emit("num", string(src[i:j]))
			i = j
		case c == '<' && (lang == "tsx" || lang == "html") && i+1 < len(src) && (unicode.IsLetter(src[i+1]) || src[i+1] == '/' || src[i+1] == '>'):
			j := i + 1
			if j < len(src) && src[j] == '/' {
				j++
			}
			k := j
			for k < len(src) && (unicode.IsLetter(src[k]) || unicode.IsDigit(src[k]) || src[k] == '-' || src[k] == '.') {
				k++
			}
			emit("punct", string(src[i:j]))
			emit("tag", string(src[j:k]))
			inTag = true
			i = k
		case c == '>' || (c == '/' && i+1 < len(src) && src[i+1] == '>'):
			j := i + 1
			if c == '/' {
				j = i + 2
			}
			emit("punct", string(src[i:j]))
			inTag = false
			i = j
		case unicode.IsLetter(c) || c == '_' || c == '$' || c == '@' || (c == '-' && inTag):
			j := i + 1
			for j < len(src) && (unicode.IsLetter(src[j]) || unicode.IsDigit(src[j]) || src[j] == '_' || src[j] == '$' || src[j] == '-' || src[j] == ':') {
				j++
			}
			w := string(src[i:j])
			switch {
			case inTag && lang != "go":
				emit("attr", w)
			case kw[w]:
				emit("kw", w)
			case lang == "tsx" && unicode.IsUpper(c) && len(w) > 1:
				emit("type", w)
			default:
				emit("plain", w)
			}
			i = j
		case strings.ContainsRune("{}()[];,.:=+-*/%!&|?<>", c):
			if c == '{' {
				inTag = false
			}
			emit("punct", string(c))
			i++
		default:
			emit("plain", string(c))
			i++
		}
	}
	return out
}

type SiteModule struct{}

func (SiteModule) GoVersion() string { return runtime.Version() }
func (SiteModule) Now() string       { return time.Now().Format("2006-01-02 15:04:05") }
func (SiteModule) Version() string   { return "v0.6" }

var (
	Hl   = HlModule{}
	Site = SiteModule{}
)

var Registry = map[string]gotsx.HostModule{
	"hl":   {Value: Hl, Go: "host.Hl"},
	"site": {Value: Site, Go: "host.Site"},
}
