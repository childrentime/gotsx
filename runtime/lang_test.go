package gotsx

import (
	"context"
	"math"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

type person struct {
	Name string
	Tags []string
}

// 原地修改: 与 JS 的 push/pop/shift/unshift/splice 逐一对齐
func TestMutatingArrayMethods(t *testing.T) {
	xs := []float64{1, 2, 3}
	if n := Push(&xs, 4, 5); n != 5 || !reflect.DeepEqual(xs, []float64{1, 2, 3, 4, 5}) {
		t.Errorf("push: %v %v", n, xs)
	}
	if v := Pop(&xs); v != 5 || len(xs) != 4 {
		t.Errorf("pop: %v %v", v, xs)
	}
	if v := Shift(&xs); v != 1 || !reflect.DeepEqual(xs, []float64{2, 3, 4}) {
		t.Errorf("shift: %v %v", v, xs)
	}
	if n := Unshift(&xs, 0, 1); n != 5 || !reflect.DeepEqual(xs, []float64{0, 1, 2, 3, 4}) {
		t.Errorf("unshift: %v %v", n, xs)
	}
	var empty []string
	if v := Pop(&empty); v != "" {
		t.Errorf("空数组 pop 应是零值")
	}
	if v := Shift(&empty); v != "" {
		t.Errorf("空数组 shift 应是零值")
	}

	// splice(start, deleteCount, ...items)
	ys := []string{"a", "b", "c", "d"}
	if r := Splice(&ys, 1, 2); !reflect.DeepEqual(r, []string{"b", "c"}) || !reflect.DeepEqual(ys, []string{"a", "d"}) {
		t.Errorf("splice 删除: %v %v", r, ys)
	}
	if r := Splice(&ys, 1, 0, "x", "y"); len(r) != 0 || !reflect.DeepEqual(ys, []string{"a", "x", "y", "d"}) {
		t.Errorf("splice 插入: %v %v", r, ys)
	}
	if r := Splice(&ys, -1, -1); !reflect.DeepEqual(r, []string{"d"}) || !reflect.DeepEqual(ys, []string{"a", "x", "y"}) {
		t.Errorf("splice 负下标 + 省略 deleteCount: %v %v", r, ys)
	}
	if r := Splice(&ys, 10, 1); len(r) != 0 || len(ys) != 3 {
		t.Errorf("splice 越界 start: %v %v", r, ys)
	}
	if FindIndex([]float64{1, 2}, func(x float64) bool { return x == 2 }) != 1 || FindIndex([]float64{1}, func(x float64) bool { return x == 9 }) != -1 {
		t.Error("findIndex")
	}
	if LastIndexOf([]string{"a", "b", "a"}, "a") != 2 || LastIndexOf([]string{"a"}, "z") != -1 {
		t.Error("lastIndexOf")
	}
}

func TestMoreStringMethods(t *testing.T) {
	if TrimStart("  x ") != "x " || TrimEnd(" x  ") != " x" {
		t.Error("trimStart/trimEnd")
	}
	if StrLastIndexOf("héllo héllo", "llo") != 8 || StrLastIndexOf("abc", "z") != -1 {
		t.Errorf("lastIndexOf 按 rune: %v", StrLastIndexOf("héllo héllo", "llo"))
	}
	if Compare("a", "b") != -1 || Compare("b", "a") != 1 || Compare("a", "a") != 0 {
		t.Error("compare")
	}
	if StrAt("héllo", -1) != "o" || StrAt("héllo", 1) != "é" || StrAt("x", 5) != "" {
		t.Error("at")
	}
}

// 零值 = undefined: find 没找到的对象在条件里是假
func TestZeroSemantics(t *testing.T) {
	if !IsZero(person{}) || IsZero(person{Name: "a"}) || !IsZero([]string(nil)) || IsZero([]string{}) {
		t.Error("IsZero 结构体/切片")
	}
	if NonZero(person{}) || !NonZero(person{Tags: []string{"x"}}) {
		t.Error("NonZero")
	}
	if !IsZero(nil) || !IsZero(map[string]string(nil)) || IsZero(map[string]string{}) {
		t.Error("IsZero nil/map")
	}
	if IsZero("x") || !IsZero("") || !IsZero(0.0) {
		t.Error("IsZero 原始类型")
	}
}

func TestRedirectPanics(t *testing.T) {
	defer func() {
		r := recover()
		re, ok := r.(*RedirectError)
		if !ok || re.URL != "/x" || re.Status != 302 {
			t.Errorf("redirect 应 panic RedirectError, 得 %#v", r)
		}
	}()
	Redirect("/x", 999) // 非 3xx 回落到 302
}

func TestNotFoundPanics(t *testing.T) {
	defer func() {
		r := recover()
		he, ok := r.(*HostError)
		if !ok || he.Err != ErrNotFound {
			t.Errorf("notFound 应 panic HostError(ErrNotFound), 得 %#v", r)
		}
	}()
	NotFound()
}

// ---------- 路由 ----------

func TestRouteMatchCatchAll(t *testing.T) {
	r := Route{Pattern: "/docs/{...slug}", Segs: []string{"docs", "{...slug}"}}
	if p, ok := r.Match("/docs/a/b/c"); !ok || p["slug"] != "a/b/c" {
		t.Errorf("catch-all 多段: %v %v", p, ok)
	}
	if p, ok := r.Match("/docs/x"); !ok || p["slug"] != "x" {
		t.Errorf("catch-all 单段: %v %v", p, ok)
	}
	if _, ok := r.Match("/docs"); ok {
		t.Error("catch-all 需要至少一段")
	}
	if _, ok := r.Match("/other/a"); ok {
		t.Error("前缀不匹配")
	}
	plain := Route{Pattern: "/p/{id}", Segs: []string{"p", "{id}"}}
	if _, ok := plain.Match("/p/a/b"); ok {
		t.Error("普通参数不吃多段")
	}
}

func TestSortRoutes(t *testing.T) {
	in := []Route{
		{Pattern: "/{...all}", Segs: []string{"{...all}"}},
		{Pattern: "/docs/{...slug}", Segs: []string{"docs", "{...slug}"}},
		{Pattern: "/p/{id}", Segs: []string{"p", "{id}"}},
		{Pattern: "/about", Segs: []string{"about"}},
		{Pattern: "/", Segs: nil},
	}
	out := sortRoutes(in)
	var got []string
	for _, r := range out {
		got = append(got, r.Pattern)
	}
	want := []string{"/about", "/", "/p/{id}", "/docs/{...slug}", "/{...all}"}
	// 静态路由在前(静态段多的更前), 参数路由其后, catch-all 最后且更具体的(静态段多)在前
	if !reflect.DeepEqual(got, want) {
		t.Errorf("路由顺序 = %v, want %v", got, want)
	}
	// 实际匹配: /docs/x 应命中 /docs/{...slug} 而不是 /{...all}
	for _, r := range out {
		if _, ok := r.Match("/docs/x"); ok {
			if r.Pattern == "/{...all}" {
				t.Error("/docs/x 被更泛的 catch-all 抢走了")
			}
			break
		}
	}
}

// ---------- 页面级控制流 + 文档注入 ----------

func TestPageRedirect(t *testing.T) {
	route := Route{Pattern: "/", Render: func(p PageProps) Node {
		if p.Query["go"] != "" {
			return Redirect(p.Query["go"], 302)
		}
		if p.Query["missing"] != "" {
			return NotFound()
		}
		return El("html", nil, El("body", nil, Text("home")))
	}}
	h := testServer(t, Options{Routes: []Route{route}})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/?go=/login", nil))
	if rec.Code != 302 || rec.Header().Get("Location") != "/login" {
		t.Errorf("redirect: %d %q", rec.Code, rec.Header().Get("Location"))
	}
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/?missing=1", nil))
	if rec.Code != 404 {
		t.Errorf("notFound: %d", rec.Code)
	}
	// 没有 <head> 的页面: 注入到 </body> 前, 岛依然能加载
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	body := rec.Body.String()
	if rec.Code != 200 || !strings.Contains(body, "window.__GOTSX") || !strings.HasSuffix(strings.TrimSpace(body), "</body></html>") {
		t.Errorf("无 head 页面的注入不对: %d\n%s", rec.Code, body)
	}
}

func TestDevEventsAndFlag(t *testing.T) {
	s := &server{opt: Options{Dev: true, Routes: []Route{homeRoute()}, ClientDir: t.TempDir()}, bootID: "abc123"}
	s.scanClient()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /_gotsx/dev", s.devEvents)
	mux.HandleFunc("/", s.page)
	var h http.Handler = mux
	h = gzipMW(h)
	h = s.secHeadersMW(h)

	// 页面注入 dev 标记
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	if !strings.Contains(rec.Body.String(), `"dev":true`) {
		t.Errorf("dev 模式应注入 dev:true\n%s", rec.Body.String())
	}

	// SSE: 首条消息就是 bootID; 请求取消后 handler 返回
	req := httptest.NewRequest("GET", "/_gotsx/dev", nil)
	req.Header.Set("Accept-Encoding", "gzip") // 事件流不能被 gzip
	ctx, cancel := context.WithCancel(req.Context())
	rec = httptest.NewRecorder()
	done := make(chan struct{})
	go func() {
		h.ServeHTTP(rec, req.WithContext(ctx))
		close(done)
	}()
	cancel()
	<-done
	if ct := rec.Header().Get("Content-Type"); ct != "text/event-stream" {
		t.Errorf("Content-Type = %q", ct)
	}
	if rec.Header().Get("Content-Encoding") != "" {
		t.Error("事件流不应被 gzip")
	}
	if !strings.Contains(rec.Body.String(), "data: abc123") {
		t.Errorf("首条消息应带 bootID: %q", rec.Body.String())
	}
}

// 流式 SSR: 外壳先写出(带 fallback), 边界并发渲染, 谁先完成先追加谁; 嵌套边界也能填; 边界内的错误不炸整页
func TestSuspenseStreaming(t *testing.T) {
	slow := func(d time.Duration, s string) func() Node {
		return func() Node { time.Sleep(d); return El("b", nil, Text(s)) }
	}
	route := Route{Pattern: "/", Render: func(p PageProps) Node {
		return El("html", nil, El("head", nil), El("body", nil,
			Suspense(Text("wait-a"), slow(40*time.Millisecond, "A")),
			Suspense(Text("wait-b"), slow(0, "B")),
			Suspense(Text("wait-c"), func() Node {
				return El("div", nil, Suspense(Text("wait-d"), slow(0, "D"))) // 嵌套
			}),
			Suspense(Text("wait-e"), func() Node { panic("boom-in-boundary") }),
		))
	}}
	for _, dev := range []bool{false, true} {
		h := testServer(t, Options{Routes: []Route{route}, Dev: dev})
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
		body := rec.Body.String()
		if rec.Code != 200 {
			t.Fatalf("status %d", rec.Code)
		}
		shellEnd := strings.Index(body, "</html>")
		firstFill := strings.Index(body, "<template data-gotsx-fill=")
		if shellEnd < 0 || firstFill < shellEnd {
			t.Fatalf("外壳应先于填充\n%s", body)
		}
		for _, want := range []string{`<gotsx-suspense id="gs1" style="display:contents">wait-a</gotsx-suspense>`, "__gotsxFill(", `data-gotsx-fill="gs1"><b>A</b>`, `data-gotsx-fill="gs2"><b>B</b>`, `<b>D</b>`, "window.__gotsxFill=function"} {
			if !strings.Contains(body, want) {
				t.Errorf("缺 %q\n%s", want, body)
			}
		}
		if strings.Index(body, `data-gotsx-fill="gs2"`) > strings.Index(body, `data-gotsx-fill="gs1"`) {
			t.Errorf("快的边界(B)应先于慢的边界(A)到达\n%s", body)
		}
		if dev != strings.Contains(body, "boom-in-boundary") {
			t.Errorf("dev=%v 时边界错误显示不对\n%s", dev, body)
		}
		if strings.Count(body, "<template data-gotsx-fill=") != 5 {
			t.Errorf("应有 5 个填充(含嵌套与出错的), 得 %d", strings.Count(body, "<template data-gotsx-fill="))
		}
	}
}

func TestHasKey(t *testing.T) {
	m := map[string]string{"a": ""}
	if !HasKey(m, "a") || HasKey(m, "b") {
		t.Error("HasKey 应区分键存在与否, 与值无关")
	}
}

func TestRegexRuntime(t *testing.T) {
	g := Re(`\d+`, "g")
	if !ReTest(g, "a12") || ReTest(g, "abc") {
		t.Error("test")
	}
	if got := ReMatch("a1b22", g); !reflect.DeepEqual(got, []string{"1", "22"}) {
		t.Errorf("match g: %v", got)
	}
	if got := ReMatch("a1b22", Re(`(\d)(\d)`, "")); !reflect.DeepEqual(got, []string{"22", "2", "2"}) {
		t.Errorf("match groups: %v", got)
	}
	if got := ReMatch("abc", g); len(got) != 0 {
		t.Errorf("no match → 空数组: %v", got)
	}
	if got := ReReplace("a1b2", Re(`\d`, ""), "#"); got != "a#b2" {
		t.Errorf("replace first: %q", got)
	}
	if got := ReReplace("a1b2", g, "[$&]"); got != "a[1]b[22]"[:0]+"a[1]b[2]" {
		t.Errorf("replace all with $&: %q", got)
	}
	if got := ReReplace("john smith", Re(`(\w+) (\w+)`, ""), "$2, $1"); got != "smith, john" {
		t.Errorf("$n groups: %q", got)
	}
	if got := ReReplace("x", Re(`x`, ""), "$$"); got != "$" {
		t.Errorf("$$: %q", got)
	}
	if got := ReSplit("a1b22c", g); !reflect.DeepEqual(got, []string{"a", "b", "c"}) {
		t.Errorf("split: %v", got)
	}
	if ReSearch("héllo1", g) != 5 || ReSearch("abc", g) != -1 {
		t.Errorf("search 按 rune: %v", ReSearch("héllo1", g))
	}
	if Re(`x`, "g") != Re(`x`, "g") {
		t.Error("同一字面量应命中缓存")
	}
}

func TestDateBuiltins(t *testing.T) {
	ms := DateParse("2026-01-02T03:04:05Z")
	if ms != 1767323045000 {
		t.Errorf("parse RFC3339: %v", ms)
	}
	if DateParse("2026-01-02") != 1767312000000 || !math.IsNaN(DateParse("nope")) {
		t.Error("parse date-only / invalid")
	}
	if IsoDate(ms) != "2026-01-02T03:04:05.000Z" || IsoDate(math.NaN()) != "" {
		t.Errorf("isoDate: %q", IsoDate(ms))
	}
	if Now() < ms {
		t.Error("now")
	}
}
