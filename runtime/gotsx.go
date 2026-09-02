// Package gotsx 是生成的 Go 代码依赖的运行时: 节点模型、字符串输出、hydrate 标记、方言内建函数。
package gotsx

import (
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"log"
	"math"
	"math/rand"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

// Ctx 是一次渲染的输出缓冲。hydrate>0 表示正在岛内部, 动态部分要带标记给客户端走位。
type Ctx struct {
	b       strings.Builder
	hydrate int
}

func (c *Ctx) Raw(s string) { c.b.WriteString(s) }

// Node: 渲染到 Ctx 的闭包。nil 是空节点。
type Node func(c *Ctx)

func Render(n Node) string {
	c := &Ctx{}
	if n != nil {
		n(c)
	}
	return c.b.String()
}

type Attr struct {
	Name string
	Val  string
	Bool bool
	On   bool
}

func A(name, val string) Attr        { return Attr{Name: name, Val: val} }
func AB(name string, on bool) Attr   { return Attr{Name: name, Bool: true, On: on} }
func AN(name string, n float64) Attr { return Attr{Name: name, Val: Num(n)} }

var voidTags = map[string]bool{"area": true, "base": true, "br": true, "col": true, "embed": true, "hr": true, "img": true, "input": true, "link": true, "meta": true, "source": true, "track": true, "wbr": true}

func El(tag string, attrs []Attr, kids ...Node) Node {
	return func(c *Ctx) {
		c.b.WriteByte('<')
		c.b.WriteString(tag)
		for _, a := range attrs {
			if a.Bool {
				if a.On {
					c.b.WriteByte(' ')
					c.b.WriteString(a.Name)
				}
				continue
			}
			c.b.WriteByte(' ')
			c.b.WriteString(a.Name)
			c.b.WriteString(`="`)
			c.b.WriteString(html.EscapeString(a.Val))
			c.b.WriteByte('"')
		}
		c.b.WriteByte('>')
		if voidTags[tag] {
			return
		}
		for _, k := range kids {
			if k != nil {
				k(c)
			}
		}
		c.b.WriteString("</")
		c.b.WriteString(tag)
		c.b.WriteByte('>')
	}
}

// Text: 静态文本(无标记)
func Text(s string) Node {
	return func(c *Ctx) { c.b.WriteString(html.EscapeString(s)) }
}

// Dyn: 响应式文本, 岛内带 <!--$-->…<!--/--> 标记
func Dyn(s string) Node {
	return func(c *Ctx) {
		if c.hydrate > 0 {
			c.b.WriteString("<!--$-->")
			c.b.WriteString(html.EscapeString(s))
			c.b.WriteString("<!--/-->")
			return
		}
		c.b.WriteString(html.EscapeString(s))
	}
}

func RawNode(s string) Node { return func(c *Ctx) { c.b.WriteString(s) } }

func Frag(ns ...Node) Node {
	return func(c *Ctx) {
		for _, n := range ns {
			if n != nil {
				n(c)
			}
		}
	}
}

func marked(c *Ctx, f func()) {
	if c.hydrate > 0 {
		c.b.WriteString("<!--[-->")
		f()
		c.b.WriteString("<!--]-->")
		return
	}
	f()
}

// Nodes: 响应式列表(岛内带标记)
func Nodes(ns []Node) Node {
	return func(c *Ctx) { marked(c, func() { Frag(ns...)(c) }) }
}
func NodesPlain(ns []Node) Node { return Frag(ns...) }

// If: 响应式条件(岛内带标记); 分支惰性求值
func If(cond bool, f func() Node) Node {
	return func(c *Ctx) {
		marked(c, func() {
			if cond {
				if n := f(); n != nil {
					n(c)
				}
			}
		})
	}
}
func IfPlain(cond bool, f func() Node) Node {
	return func(c *Ctx) {
		if cond {
			if n := f(); n != nil {
				n(c)
			}
		}
	}
}

// Marked: 响应式三元的结果
func Marked(n Node) Node {
	return func(c *Ctx) {
		marked(c, func() {
			if n != nil {
				n(c)
			}
		})
	}
}

// Island: 客户端组件的服务端外壳。props 序列化进属性, 内部子树带 hydrate 标记。
func Island(name string, props any, inner Node) Node {
	return func(c *Ctx) {
		b, _ := json.Marshal(props)
		c.b.WriteString(`<gotsx-island name="`)
		c.b.WriteString(name)
		c.b.WriteString(`" props="`)
		c.b.WriteString(html.EscapeString(string(b)))
		c.b.WriteString(`" style="display:contents">`)
		c.hydrate++
		if inner != nil {
			inner(c)
		}
		c.hydrate--
		c.b.WriteString("</gotsx-island>")
	}
}

// ---------- 方言内建 ----------

// Num: JS 的 Number → string
func Num(f float64) string {
	if f == math.Trunc(f) && math.Abs(f) < 1e21 {
		return strconv.FormatInt(int64(f), 10)
	}
	if math.IsNaN(f) {
		return "NaN"
	}
	if math.IsInf(f, 0) {
		if f > 0 {
			return "Infinity"
		}
		return "-Infinity"
	}
	return strconv.FormatFloat(f, 'f', -1, 64)
}

func Str(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case float64:
		return Num(x)
	case int:
		return strconv.Itoa(x)
	case bool:
		if x {
			return "true"
		}
		return "false"
	case nil:
		return ""
	case fmt.Stringer:
		return x.String()
	}
	return fmt.Sprint(v)
}

func ToNum(v any) float64 {
	switch x := v.(type) {
	case float64:
		return x
	case int:
		return float64(x)
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(x), 64)
		if err != nil {
			return math.NaN()
		}
		return f
	case bool:
		if x {
			return 1
		}
	}
	return 0
}

func JSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func Log(args ...any) { log.Println(append([]any{"[app]"}, args...)...) }

// Or: a ?? b / a || b (字符串以空为假)
func Or(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
func OrNum(a, b float64) float64 {
	if a != 0 {
		return a
	}
	return b
}

func Map[T, U any](xs []T, f func(T) U) []U {
	out := make([]U, len(xs))
	for i, x := range xs {
		out[i] = f(x)
	}
	return out
}
func MapI[T, U any](xs []T, f func(T, float64) U) []U {
	out := make([]U, len(xs))
	for i, x := range xs {
		out[i] = f(x, float64(i))
	}
	return out
}
func Filter[T any](xs []T, f func(T) bool) []T {
	out := []T{}
	for _, x := range xs {
		if f(x) {
			out = append(out, x)
		}
	}
	return out
}
func Find[T any](xs []T, f func(T) bool) T {
	for _, x := range xs {
		if f(x) {
			return x
		}
	}
	var zero T
	return zero
}
func Some[T any](xs []T, f func(T) bool) bool {
	for _, x := range xs {
		if f(x) {
			return true
		}
	}
	return false
}
func Every[T any](xs []T, f func(T) bool) bool {
	for _, x := range xs {
		if !f(x) {
			return false
		}
	}
	return true
}
func Includes[T comparable](xs []T, v T) bool {
	for _, x := range xs {
		if x == v {
			return true
		}
	}
	return false
}
func IndexOf[T comparable](xs []T, v T) float64 {
	for i, x := range xs {
		if x == v {
			return float64(i)
		}
	}
	return -1
}
func Join(xs []string, sep string) string { return strings.Join(xs, sep) }
func JoinAny[T any](xs []T, sep string) string {
	parts := make([]string, len(xs))
	for i, x := range xs {
		parts[i] = Str(x)
	}
	return strings.Join(parts, sep)
}
func Slice[T any](xs []T, from float64, to ...float64) []T {
	n := len(xs)
	a := int(from)
	if a < 0 {
		a = n + a
	}
	if a > n {
		a = n
	}
	b := n
	if len(to) > 0 {
		b = int(to[0])
		if b < 0 {
			b = n + b
		}
		if b > n {
			b = n
		}
	}
	if a > b {
		a = b
	}
	out := make([]T, b-a)
	copy(out, xs[a:b])
	return out
}
func Concat[T any](a []T, b []T) []T {
	out := make([]T, 0, len(a)+len(b))
	out = append(out, a...)
	return append(out, b...)
}
func Len[T any](xs []T) float64 { return float64(len(xs)) }

func StrLen(s string) float64          { return float64(len([]rune(s))) }
func Upper(s string) string            { return strings.ToUpper(s) }
func Lower(s string) string            { return strings.ToLower(s) }
func Trim(s string) string             { return strings.TrimSpace(s) }
func StrIncludes(s, sub string) bool   { return strings.Contains(s, sub) }
func StartsWith(s, p string) bool      { return strings.HasPrefix(s, p) }
func EndsWith(s, p string) bool        { return strings.HasSuffix(s, p) }
func Split(s, sep string) []string     { return strings.Split(s, sep) }
func Replace(s, a, b string) string    { return strings.Replace(s, a, b, 1) }
func ReplaceAll(s, a, b string) string { return strings.ReplaceAll(s, a, b) }
func Repeat(s string, n float64) string {
	if n < 0 {
		n = 0
	}
	return strings.Repeat(s, int(n))
}
func StrSlice(s string, from float64, to ...float64) string {
	r := []rune(s)
	n := len(r)
	a := int(from)
	if a < 0 {
		a = n + a
	}
	if a > n {
		a = n
	}
	b := n
	if len(to) > 0 {
		b = int(to[0])
		if b < 0 {
			b = n + b
		}
		if b > n {
			b = n
		}
	}
	if a > b {
		a = b
	}
	return string(r[a:b])
}
func StrIndexOf(s, sub string) float64 { return float64(strings.Index(s, sub)) }
func CharAt(s string, i float64) string {
	r := []rune(s)
	if int(i) < 0 || int(i) >= len(r) {
		return ""
	}
	return string(r[int(i)])
}

func EncodeURIComponent(s string) string {
	return strings.ReplaceAll(url.QueryEscape(s), "+", "%20")
}

func Floor(f float64) float64 { return math.Floor(f) }
func Ceil(f float64) float64  { return math.Ceil(f) }
func Round(f float64) float64 { return math.Floor(f + 0.5) }
func Abs(f float64) float64   { return math.Abs(f) }
func Max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
func Min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// ---------- 宿主错误 ----------

// ErrNotFound: 宿主方法用 fmt.Errorf("%w: ...", gotsx.ErrNotFound) 包装, 路由层回 404
var ErrNotFound = errors.New("not found")

type HostError struct{ Err error }

func (e *HostError) Error() string { return e.Err.Error() }
func (e *HostError) Unwrap() error { return e.Err }

// Must: 宿主方法 (T, error) 的 error 变成 panic, 由请求层 recover
func Must[T any](v T, err error) T {
	if err != nil {
		panic(&HostError{Err: err})
	}
	return v
}
func Check(err error) {
	if err != nil {
		panic(&HostError{Err: err})
	}
}

type ThrowError struct{ Val any }

func (e *ThrowError) Error() string { return fmt.Sprint(e.Val) }
func Throw(v any)                   { panic(&ThrowError{Val: v}) }

// PageProps: 页面组件的 props, 由路由层构造
type PageProps struct {
	Params  map[string]string `json:"params"`
	Query   map[string]string `json:"query"`
	Path    string            `json:"path"`
	Locale  string            `json:"locale"`
	Cookies map[string]string `json:"-"`
}

type numeric interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~float32
}

// Nums: 宿主的数值切片([]int 等) → 方言的 number[]
func Nums[T numeric](xs []T) []float64 {
	out := make([]float64, len(xs))
	for i, x := range xs {
		out[i] = float64(x)
	}
	return out
}

// At: 方括号下标 xs[i] 的语义 —— 越界(含负数)返回零值(对应 JS 的 undefined)
func At[T any](xs []T, i float64) T {
	if int(i) < 0 || int(i) >= len(xs) {
		var zero T
		return zero
	}
	return xs[int(i)]
}

// AtWrap: 数组 .at(i) 方法 —— 负数从末尾回绕, 仍越界则零值
func AtWrap[T any](xs []T, i float64) T {
	k := int(i)
	if k < 0 {
		k += len(xs)
	}
	if k < 0 || k >= len(xs) {
		var zero T
		return zero
	}
	return xs[k]
}
func FilterI[T any](xs []T, f func(T, float64) bool) []T {
	out := []T{}
	for i, x := range xs {
		if f(x, float64(i)) {
			out = append(out, x)
		}
	}
	return out
}
func FindI[T any](xs []T, f func(T, float64) bool) T {
	for i, x := range xs {
		if f(x, float64(i)) {
			return x
		}
	}
	var zero T
	return zero
}
func SomeI[T any](xs []T, f func(T, float64) bool) bool {
	for i, x := range xs {
		if f(x, float64(i)) {
			return true
		}
	}
	return false
}
func EveryI[T any](xs []T, f func(T, float64) bool) bool {
	for i, x := range xs {
		if !f(x, float64(i)) {
			return false
		}
	}
	return true
}
func ForEachI[T any](xs []T, f func(T, float64)) {
	for i, x := range xs {
		f(x, float64(i))
	}
}
func Reverse[T any](xs []T) []T {
	out := make([]T, len(xs))
	for i, x := range xs {
		out[len(xs)-1-i] = x
	}
	return out
}
func Sort[T any](xs []T, cmp func(T, T) float64) []T {
	out := make([]T, len(xs))
	copy(out, xs)
	sort.SliceStable(out, func(i, j int) bool { return cmp(out[i], out[j]) < 0 })
	return out
}
func Reduce[T, U any](xs []T, f func(U, T) U, init U) U {
	acc := init
	for _, x := range xs {
		acc = f(acc, x)
	}
	return acc
}
func Flat[T any](xs [][]T) []T {
	out := []T{}
	for _, x := range xs {
		out = append(out, x...)
	}
	return out
}
func PadStart(s string, n float64, pad string) string {
	if pad == "" {
		pad = " "
	}
	for float64(len([]rune(s))) < n {
		need := int(n) - len([]rune(s))
		if len([]rune(pad)) > need {
			s = string([]rune(pad)[:need]) + s
		} else {
			s = pad + s
		}
	}
	return s
}
func PadEnd(s string, n float64, pad string) string {
	if pad == "" {
		pad = " "
	}
	for float64(len([]rune(s))) < n {
		need := int(n) - len([]rune(s))
		if len([]rune(pad)) > need {
			s = s + string([]rune(pad)[:need])
		} else {
			s = s + pad
		}
	}
	return s
}
func ObjectKeys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
func ObjectValues[V any](m map[string]V) []V {
	keys := ObjectKeys(m)
	out := make([]V, len(keys))
	for i, k := range keys {
		out[i] = m[k]
	}
	return out
}
func Pow(a, b float64) float64 { return math.Pow(a, b) }
func Trunc(f float64) float64  { return math.Trunc(f) }
func Sign(f float64) float64 {
	if f > 0 {
		return 1
	}
	if f < 0 {
		return -1
	}
	return 0
}

func ForEach[T any](xs []T, f func(T)) {
	for _, x := range xs {
		f(x)
	}
}
func ToFixed(f float64, digits float64) string { return strconv.FormatFloat(f, 'f', int(digits), 64) }
func Random() float64                          { return rand.Float64() }
func Sqrt(f float64) float64                   { return math.Sqrt(f) }
func Mod(a, b float64) float64                 { return math.Mod(a, b) }
func IsNaN(v any) bool                         { return math.IsNaN(ToNum(v)) }
func DecodeURIComponent(s string) string {
	out, err := url.QueryUnescape(strings.ReplaceAll(s, "+", "%2B"))
	if err != nil {
		return s
	}
	return out
}

// Truthy: JS 真值语义(any 用)
func Truthy(v any) bool {
	switch x := v.(type) {
	case nil:
		return false
	case bool:
		return x
	case string:
		return x != ""
	case float64:
		return x != 0 && !math.IsNaN(x)
	case int:
		return x != 0
	case Node:
		return x != nil
	}
	return true
}

func htmlEscape(s string) string { return html.EscapeString(s) }

// LdScript: 结构化数据(JSON-LD)。内容应为 JSON(通常 JSON.stringify 的结果, 已转义 <>&),
// 再把 </ 变成 <\/ 防止 </script> 逃逸。
func LdScript(jsonStr string) Node {
	safe := strings.ReplaceAll(jsonStr, "</", "<\\/")
	return func(c *Ctx) {
		c.b.WriteString(`<script type="application/ld+json">`)
		c.b.WriteString(safe)
		c.b.WriteString(`</script>`)
	}
}
