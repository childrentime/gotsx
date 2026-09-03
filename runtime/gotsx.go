// Package gotsx 是生成的 Go 代码依赖的运行时: 节点模型、字符串输出、hydrate 标记、方言内建函数。
package gotsx

import (
	"bytes"
	"encoding"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"log"
	"math"
	"math/rand"
	"net/url"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Ctx 是一次渲染的输出缓冲。hydrate>0 表示正在岛内部, 动态部分要带标记给客户端走位。
type Ctx struct {
	b       strings.Builder
	hydrate int
	pending []*Pending // 本次渲染登记的 Suspense 边界(流式: 外壳先发, 这些随后并发填充)
	seq     *uint32    // 边界 id 计数器, 嵌套渲染共享
	inject  string     // 请求层要塞进 </head> 前的引导脚本; 写完后清空(没有 head 就塞在 </body> 前, 再没有就追加到末尾)
	scope   *reqScope  // per-request store seeds (seed(store, value) in a page); nil outside a request
}

// writeInject: the bootstrap script plus the store seeds, once, at </head> (or </body>, or the end of the document)
func (c *Ctx) writeInject() {
	if c.inject == "" {
		return
	}
	c.b.WriteString(c.inject)
	c.inject = ""
	c.b.WriteString(c.scope.script())
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

// Pending: 一个尚未填充的 Suspense 边界
type Pending struct {
	ID string
	Fn func() Node
}

// RenderPending: 渲染并返回登记的 Suspense 边界(请求层用它做流式)。seq 在嵌套填充之间共享; scope 是本次请求的 store 种子。
func RenderPending(n Node, seq *uint32, scope *reqScope) (string, []*Pending) {
	c := &Ctx{seq: seq, scope: scope}
	if n != nil {
		n(c)
	}
	return c.b.String(), c.pending
}

// RenderDoc: 渲染整个文档 —— 一次性写入 doctype, 在 </head>(或 </body>)前塞入 inject, 不做事后的字符串替换/拼接。
func RenderDoc(n Node, seq *uint32, inject string, scope *reqScope) (string, []*Pending) {
	c := &Ctx{seq: seq, inject: inject, scope: scope}
	c.b.Grow(16 * 1024)
	c.b.WriteString("<!DOCTYPE html>")
	if n != nil {
		n(c)
	}
	c.writeInject() // 既没有 head 也没有 body
	return c.b.String(), c.pending
}

// RenderIn: Render with a request scope (404 / error pages: the seeds are appended to the bootstrap by the caller)
func RenderIn(scope *reqScope, n Node) string {
	c := &Ctx{scope: scope}
	if n != nil {
		n(c)
	}
	return c.b.String()
}

func (c *Ctx) nextID() string {
	if c.seq == nil {
		c.seq = new(uint32)
	}
	*c.seq++
	return "gs" + strconv.FormatUint(uint64(*c.seq), 10)
}

// Suspense: 流式 SSR 的边界。外壳里先输出 fallback; content 在外壳发出后、在自己的 goroutine 里求值,
// 结果作为 <template> + 一小段脚本追加到响应末尾, 浏览器把它换进去。多个边界并发, 谁先完成先发谁。
// 编译器把 <Suspense fallback={…}>children</Suspense> 编成 gotsx.Suspense(fallback, func() gotsx.Node { return children })。
func Suspense(fallback Node, content func() Node) Node {
	return func(c *Ctx) {
		id := c.nextID()
		c.pending = append(c.pending, &Pending{ID: id, Fn: content})
		c.b.WriteString(`<gotsx-suspense id="`)
		c.b.WriteString(id)
		c.b.WriteString(`" style="display:contents">`)
		if fallback != nil {
			fallback(c)
		}
		c.b.WriteString("</gotsx-suspense>")
	}
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
		if tag == "head" || tag == "body" {
			c.writeInject()
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
// noNil encodes nil slices / nil maps in island props as [] / {} rather than null — in the dialect an empty array is an
// empty array, and island code calls .map on it directly. Props whose type contains no slice / map / pointer at all
// (the common case: a few scalars) are encoded as-is without reflection.
func noNil(v any) any {
	if v == nil {
		return nil
	}
	t := reflect.TypeOf(v)
	if c, ok := nilable.Load(t); ok {
		if !c.(bool) {
			return v
		}
	} else {
		n := hasNilable(t, map[reflect.Type]bool{})
		nilable.Store(t, n)
		if !n {
			return v
		}
	}
	return noNilValue(reflect.ValueOf(v), &nilWalk{path: map[uintptr]bool{}})
}

// nilWalk: state of one noNil traversal — the pointers on the current path (cycle detection) and the depth
type nilWalk struct {
	path  map[uintptr]bool
	depth int
}

var (
	jsonMarshalerT = reflect.TypeOf((*json.Marshaler)(nil)).Elem()
	textMarshalerT = reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()
)

func isMarshaler(rv reflect.Value) bool {
	t := rv.Type()
	return t.Implements(jsonMarshalerT) || t.Implements(textMarshalerT)
}

// jsonEmpty: encoding/json's omitempty rule (false, 0, "", nil pointer/interface, empty array/slice/map/string; structs never)
func jsonEmpty(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr, reflect.Float32, reflect.Float64:
		return v.IsZero()
	case reflect.Interface, reflect.Pointer:
		return v.IsNil()
	}
	return false
}

var nilable sync.Map // reflect.Type → bool

func hasNilable(t reflect.Type, seen map[reflect.Type]bool) bool {
	switch t.Kind() {
	case reflect.Slice, reflect.Map, reflect.Pointer, reflect.Interface:
		return true
	case reflect.Array:
		return hasNilable(t.Elem(), seen)
	case reflect.Struct:
		if seen[t] {
			return false
		}
		seen[t] = true
		for i := 0; i < t.NumField(); i++ {
			if hasNilable(t.Field(i).Type, seen) {
				return true
			}
		}
	}
	return false
}

func noNilValue(rv reflect.Value, w *nilWalk) any {
	if w.depth > 1000 { // safety net; cycles are caught by the path set below
		return nil
	}
	w.depth++
	defer func() { w.depth-- }()
	if rv.IsValid() && rv.CanInterface() && isMarshaler(rv) { // json.Marshaler / encoding.TextMarshaler (time.Time, netip.Addr, …): as encoding/json would
		return rv.Interface()
	}
	if rv.IsValid() && rv.CanAddr() && isMarshaler(rv.Addr()) { // pointer-receiver MarshalJSON on an addressable value
		return rv.Addr().Interface()
	}
	switch rv.Kind() {
	case reflect.Pointer, reflect.Map, reflect.Slice:
		if !rv.IsNil() {
			ptr := rv.Pointer()
			if w.path[ptr] { // a cycle (child.Parent == root): stop here instead of recursing forever
				return nil
			}
			w.path[ptr] = true
			defer delete(w.path, ptr)
		}
	}
	switch rv.Kind() {
	case reflect.Invalid:
		return nil
	case reflect.Pointer, reflect.Interface:
		if rv.IsNil() {
			return nil
		}
		return noNilValue(rv.Elem(), w)
	case reflect.Slice:
		if rv.Type().Elem().Kind() == reflect.Uint8 { // []byte keeps its base64 encoding
			return rv.Interface()
		}
		out := make([]any, rv.Len())
		for i := range out {
			out[i] = noNilValue(rv.Index(i), w)
		}
		return out
	case reflect.Map:
		out := make(map[string]any, rv.Len())
		iter := rv.MapRange()
		for iter.Next() {
			out[fmt.Sprint(iter.Key().Interface())] = noNilValue(iter.Value(), w)
		}
		return out
	case reflect.Struct:
		out := map[string]any{}
		t := rv.Type()
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			fv := rv.Field(i)
			if f.Anonymous && fv.Kind() == reflect.Pointer && fv.Type().Elem().Kind() == reflect.Struct { // embedded *T: json promotes its fields, skips nil
				if fv.IsNil() {
					continue
				}
				fv = fv.Elem()
			}
			embedded := f.Anonymous && fv.Kind() == reflect.Struct
			if !f.IsExported() && !embedded { // an unexported embedded struct promotes its exported fields, as encoding/json does
				continue
			}
			name, opts, _ := strings.Cut(f.Tag.Get("json"), ",")
			if name == "-" {
				continue
			}
			if embedded && name == "" { // embedded struct: promote its fields
				if m, ok := noNilValue(fv, w).(map[string]any); ok {
					for k, v := range m {
						out[k] = v
					}
				}
				continue
			}
			if name == "" {
				name = f.Name
			}
			if strings.Contains(opts, "omitempty") && jsonEmpty(fv) {
				continue
			}
			out[name] = noNilValue(fv, w)
		}
		return out
	}
	if !rv.CanInterface() {
		return nil
	}
	return rv.Interface()
}

func Island(name string, props any, inner Node) Node {
	return func(c *Ctx) {
		b, _ := json.Marshal(props)
		if bytes.Contains(b, []byte("null")) { // only then can a nil slice/map be hiding in there: re-encode with [] / {}
			if nb, err := json.Marshal(noNil(props)); err == nil {
				b = nb
			}
		}
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

// ---------- Stores: shared client state, seeded per request by the page ----------

// Store is what `export const cart = createStore(init)` in a client module compiles to. The browser holds the
// state as signals; the Go side only knows the initial value and the value a page seeded for the current request
// (seed(store, value)), which is what the islands of that request render with and what the browser starts from.
// store.set(...) never runs on the server, so a Store is never mutated here and is safe to share across goroutines.
type Store[T any] struct {
	Name string // the generated Go identifier: unique per app, also the key in the seed JSON the browser reads
	Init T
}

func NewStore[T any](name string, init T) *Store[T] { return &Store[T]{Name: name, Init: init} }

// Get: the value this request renders with — the seed when the page set one, else the initial value.
func (s *Store[T]) Get(c *Ctx) T {
	if c != nil && c.scope != nil {
		if v, ok := c.scope.get(s.Name); ok {
			if t, ok := v.(T); ok {
				return t
			}
		}
	}
	return s.Init
}

// Seeder: PageProps (and LayoutProps / ErrorProps through embedding) — what seed(store, value) compiles against.
type Seeder interface{ reqScope() *reqScope }

// Seed: `seed(store, value)` in the body of a page or layout component. It runs before any HTML is written (page
// and layout bodies are evaluated first), so every island of the request renders with value, and the value is
// serialized into <head> (a <script type="application/json" data-gotsx-stores> block) for the browser's store.
// A later call for the same store wins. Outside a request (tests, Render) it is a no-op.
func Seed[T any](p Seeder, s *Store[T], v T) {
	if p == nil {
		return
	}
	if sc := p.reqScope(); sc != nil {
		sc.set(s.Name, v)
	}
}

// reqScope: the seeds of one request. Written by page/layout bodies (single goroutine, before the render), read
// by island renders — also from Suspense goroutines, hence the mutex.
type reqScope struct {
	mu    sync.Mutex
	seeds map[string]any
}

func newReqScope() *reqScope { return &reqScope{seeds: map[string]any{}} }

func (sc *reqScope) set(name string, v any) {
	if sc == nil {
		return
	}
	sc.mu.Lock()
	sc.seeds[name] = v
	sc.mu.Unlock()
}

func (sc *reqScope) get(name string) (any, bool) {
	if sc == nil {
		return nil, false
	}
	sc.mu.Lock()
	v, ok := sc.seeds[name]
	sc.mu.Unlock()
	return v, ok
}

// script: the seed block for <head> ("" when nothing was seeded). A JSON data block is not executed, so it needs no
// CSP nonce; encoding/json escapes < > & so the document cannot be broken out of. Nil slices / maps become [] / {}
// like island props (an island doing items.map(...) must not meet null).
func (sc *reqScope) script() string {
	if sc == nil {
		return ""
	}
	sc.mu.Lock()
	defer sc.mu.Unlock()
	if len(sc.seeds) == 0 {
		return ""
	}
	seeds := make(map[string]any, len(sc.seeds))
	for k, v := range sc.seeds {
		seeds[k] = noNil(v)
	}
	b, err := json.Marshal(seeds)
	if err != nil {
		return ""
	}
	return `<script type="application/json" data-gotsx-stores>` + string(b) + `</script>`
}

// PageProps: 页面组件的 props, 由路由层构造
type PageProps struct {
	Params  map[string]string `json:"params"`
	Query   map[string]string `json:"query"`
	Path    string            `json:"path"`
	Locale  string            `json:"locale"`
	Cookies map[string]string `json:"-"`
	Session map[string]string `json:"-"` // session values (read-only here; write them in actions via req.Session())
	Flash   []Flash           `json:"-"` // one-shot messages (cleared after this render)
	csrf    func() string     // lazy: the token is generated and the session cookie set only when a page actually reads props.csrf (anonymous pages send no Set-Cookie and stay CDN-cacheable)
	scope   *reqScope         // store seeds of this request (seed(store, value) writes, island renders read); nil outside a request
}

// reqScope implements Seeder; LayoutProps / ErrorProps embed PageProps, so a layout's props seed as well.
func (p PageProps) reqScope() *reqScope { return p.scope }

// CSRF returns this session's CSRF token (for classic forms: <input type="hidden" name="_csrf" value={csrf} />); props.csrf in the
// dialect compiles to this call. Read it outside <Suspense> boundaries (once the shell is flushed a cookie can no longer be set).
func (p PageProps) CSRF() string {
	if p.csrf == nil {
		return ""
	}
	return p.csrf()
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

// ---------- 原地修改的数组方法(需要地址; JS 语义) ----------

// Push: xs.push(a, b) → 新长度
func Push[T any](xs *[]T, vs ...T) float64 {
	*xs = append(*xs, vs...)
	return float64(len(*xs))
}

// Pop: 弹出末尾(空数组返回零值, 对应 undefined)
func Pop[T any](xs *[]T) T {
	var zero T
	n := len(*xs)
	if n == 0 {
		return zero
	}
	v := (*xs)[n-1]
	*xs = (*xs)[:n-1]
	return v
}

// Shift: 弹出开头
func Shift[T any](xs *[]T) T {
	var zero T
	if len(*xs) == 0 {
		return zero
	}
	v := (*xs)[0]
	*xs = append([]T(nil), (*xs)[1:]...)
	return v
}

// Unshift: 头插 → 新长度
func Unshift[T any](xs *[]T, vs ...T) float64 {
	out := make([]T, 0, len(*xs)+len(vs))
	out = append(out, vs...)
	*xs = append(out, *xs...)
	return float64(len(*xs))
}

// Splice: xs.splice(start, deleteCount, ...items) → 被删除的元素; deleteCount<0 表示省略(删到末尾)
func Splice[T any](xs *[]T, start, del float64, items ...T) []T {
	n := len(*xs)
	a := int(start)
	if a < 0 {
		a = n + a
		if a < 0 {
			a = 0
		}
	}
	if a > n {
		a = n
	}
	d := n - a
	if del >= 0 && int(del) < d {
		d = int(del)
	}
	if d < 0 {
		d = 0
	}
	removed := make([]T, d)
	copy(removed, (*xs)[a:a+d])
	out := make([]T, 0, n-d+len(items))
	out = append(out, (*xs)[:a]...)
	out = append(out, items...)
	out = append(out, (*xs)[a+d:]...)
	*xs = out
	return removed
}

func FindIndex[T any](xs []T, f func(T) bool) float64 {
	for i, x := range xs {
		if f(x) {
			return float64(i)
		}
	}
	return -1
}
func FindIndexI[T any](xs []T, f func(T, float64) bool) float64 {
	for i, x := range xs {
		if f(x, float64(i)) {
			return float64(i)
		}
	}
	return -1
}
func LastIndexOf[T comparable](xs []T, v T) float64 {
	for i := len(xs) - 1; i >= 0; i-- {
		if xs[i] == v {
			return float64(i)
		}
	}
	return -1
}

// ---------- 更多字符串方法 ----------

func TrimStart(s string) string { return strings.TrimLeft(s, " \t\n\r\v\f\u00a0\ufeff") }
func TrimEnd(s string) string   { return strings.TrimRight(s, " \t\n\r\v\f\u00a0\ufeff") }
func StrLastIndexOf(s, sub string) float64 {
	i := strings.LastIndex(s, sub)
	if i < 0 {
		return -1
	}
	return float64(len([]rune(s[:i])))
}

// Compare: localeCompare 的两端一致版本 —— 按码点比较(不做本地化排序), 客户端 G.cmp 同语义
func Compare(a, b string) float64 {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	}
	return 0
}

// StrAt: 字符串 .at(i), 负数从末尾数
func StrAt(s string, i float64) string {
	r := []rune(s)
	k := int(i)
	if k < 0 {
		k += len(r)
	}
	if k < 0 || k >= len(r) {
		return ""
	}
	return string(r[k])
}

// ---------- 零值 = undefined ----------

// IsZero: obj === undefined 对复合类型的语义(Go 里"缺席"就是零值)
func IsZero(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Slice, reflect.Map, reflect.Func, reflect.Pointer, reflect.Interface:
		return rv.IsNil()
	}
	return rv.IsZero()
}

// NonZero: 可能缺席的对象在条件里的真值(find 没找到 → 零值 → false)
func NonZero(v any) bool { return !IsZero(v) }

// ---------- 页面级控制流: redirect / notFound ----------

// RedirectError: 方言里 redirect(url, status) 抛出, 请求层变成 HTTP 重定向
type RedirectError struct {
	URL    string
	Status int
}

func (e *RedirectError) Error() string { return fmt.Sprintf("redirect %d → %s", e.Status, e.URL) }

// Redirect: 中断本次渲染, 让请求层回 3xx。返回 Node 只是为了能写 return redirect("/")
func Redirect(url string, status float64) Node {
	code := int(status)
	if code < 300 || code > 399 {
		code = 302
	}
	panic(&RedirectError{URL: url, Status: code})
}

// NotFound: 中断本次渲染, 让请求层回 404 页
func NotFound() Node {
	panic(&HostError{Err: ErrNotFound})
}

// ---------- 文件约定: _layout / _404 / _error ----------

// LayoutProps: pages/**/_layout.server.tsx 的 props —— 页面 props 加上被包裹的内容
type LayoutProps struct {
	PageProps
	Meta     Meta `json:"meta"` // the result of the page's export function meta(props) (zero value when absent)
	Children Node `json:"-"`
}

// Meta is the page metadata convention — a page exports `function meta(props: PageProps): Meta`, and layouts render <title> etc. from LayoutProps.meta
type Meta struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Canonical   string `json:"canonical"`
	Image       string `json:"image"`
	NoIndex     bool   `json:"noIndex"`
}

// ErrorProps: pages/_error.server.tsx 的 props —— 页面 props 加上错误信息(生产模式下是通用文案)
type ErrorProps struct {
	PageProps
	Message string `json:"message"`
}

// HasKey: Object.hasOwn(record, key) —— 区分"键不存在"与"值是零值"
func HasKey[V any](m map[string]V, k string) bool {
	_, ok := m[k]
	return ok
}

// ---------- 正则(RE2 子集, 编译期已校验) ----------

// Regex: 方言的 RegExp; Global 对应 g 标志(replace 全部 / match 全部)
type Regex struct {
	*regexp.Regexp
	Global bool
}

var reCache sync.Map

// Re: 正则字面量 → 编译一次缓存(模式已由编译器转成 RE2 语法)
func Re(pattern, flags string) *Regex {
	key := flags + "/" + pattern
	if v, ok := reCache.Load(key); ok {
		return v.(*Regex)
	}
	r := &Regex{Regexp: regexp.MustCompile(pattern), Global: strings.Contains(flags, "g")}
	reCache.Store(key, r)
	return r
}

func ReTest(re *Regex, s string) bool { return re.MatchString(s) }

// jsRepl: JS 的替换模板($& $1 $$)→ Go 的 ${0} ${1} $$
func jsRepl(repl string) string {
	var b strings.Builder
	for i := 0; i < len(repl); i++ {
		if repl[i] != '$' || i+1 >= len(repl) {
			b.WriteByte(repl[i])
			continue
		}
		switch n := repl[i+1]; {
		case n == '&':
			b.WriteString("${0}")
			i++
		case n == '$':
			b.WriteString("$$")
			i++
		case n >= '0' && n <= '9':
			j := i + 1
			for j < len(repl) && repl[j] >= '0' && repl[j] <= '9' {
				j++
			}
			b.WriteString("${" + repl[i+1:j] + "}")
			i = j - 1
		default:
			b.WriteString("$$")
		}
	}
	return b.String()
}

// ReReplace: s.replace(re, repl) —— g 标志替换全部, 否则只替换第一个
func ReReplace(s string, re *Regex, repl string) string {
	tmpl := jsRepl(repl)
	if re.Global {
		return re.ReplaceAllString(s, tmpl)
	}
	loc := re.FindStringSubmatchIndex(s)
	if loc == nil {
		return s
	}
	var dst []byte
	dst = re.ExpandString(dst, tmpl, s, loc)
	return s[:loc[0]] + string(dst) + s[loc[1]:]
}

// ReMatch: s.match(re) —— g: 全部匹配; 否则 [整体, 分组...]; 没匹配 → 空数组(客户端 G.match 同语义)
func ReMatch(s string, re *Regex) []string {
	if re.Global {
		out := re.FindAllString(s, -1)
		if out == nil {
			return []string{}
		}
		return out
	}
	m := re.FindStringSubmatch(s)
	if m == nil {
		return []string{}
	}
	return m
}

func ReSplit(s string, re *Regex) []string { return re.Split(s, -1) }

// ReSearch: 第一个匹配的 rune 下标, 没有 → -1
func ReSearch(s string, re *Regex) float64 {
	loc := re.FindStringIndex(s)
	if loc == nil {
		return -1
	}
	return float64(len([]rune(s[:loc[0]])))
}

// ---------- 时间 ----------

// Now: Date.now() —— 毫秒
func Now() float64 { return float64(time.Now().UnixMilli()) }

// DateParse: Date.parse(iso) —— RFC3339 / 2006-01-02 → 毫秒, 失败 NaN
func DateParse(s string) float64 {
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02T15:04:05", "2006-01-02"} {
		if t, err := time.Parse(layout, s); err == nil {
			return float64(t.UnixMilli())
		}
	}
	return math.NaN()
}

// IsoDate: 毫秒 → RFC3339(UTC, 毫秒精度), 与客户端 toISOString 一致
func IsoDate(ms float64) string {
	if math.IsNaN(ms) || math.IsInf(ms, 0) {
		return ""
	}
	return time.UnixMilli(int64(ms)).UTC().Format("2006-01-02T15:04:05.000Z")
}

// ---------- 流式写入(生成代码用): 静态 HTML 合并成整段写入, 动态部分就地转义 ----------

// W: 原样写入(编译器已在编译期转义好的静态 HTML)
func (c *Ctx) W(s string) { c.b.WriteString(s) }

// Esc: 转义后写入(静态文本表达式)
func (c *Ctx) Esc(s string) { c.b.WriteString(html.EscapeString(s)) }

// Dyn: 响应式文本 —— 岛内带 <!--$-->…<!--/--> 标记
func (c *Ctx) Dyn(s string) { Dyn(s)(c) }

// Attr: 动态字符串属性 ` name="value"`(转义)
func (c *Ctx) Attr(name, val string) {
	c.b.WriteByte(' ')
	c.b.WriteString(name)
	c.b.WriteString(`="`)
	c.b.WriteString(html.EscapeString(val))
	c.b.WriteByte('"')
}

// N: 渲染一个可能为 nil 的节点
func (c *Ctx) N(n Node) {
	if n != nil {
		n(c)
	}
}

// Close: 闭合 head / body(引导脚本注入点); 其它标签由编译器直接写 </tag>
func (c *Ctx) Close(tag string) {
	if tag == "head" || tag == "body" {
		c.writeInject()
	}
	c.b.WriteString("</")
	c.b.WriteString(tag)
	c.b.WriteByte('>')
}

// ListStart / ListEnd: 响应式列表的块标记(岛内 <!--[--> … <!--]-->), 供编译器把 map 内联成 for 循环时使用
func (c *Ctx) ListStart() {
	if c.hydrate > 0 {
		c.b.WriteString("<!--[-->")
	}
}
func (c *Ctx) ListEnd() {
	if c.hydrate > 0 {
		c.b.WriteString("<!--]-->")
	}
}
