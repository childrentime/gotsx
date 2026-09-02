package gotsx

import (
	"context"
	"encoding/json"
	"fmt"
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

// ---------- typed actions ----------

type actTodo struct {
	ID   string `json:"id"`
	Done bool   `json:"done"`
}

func actionServer(t *testing.T, acts []HostAction, secret string) http.Handler {
	t.Helper()
	opt := Options{Routes: []Route{homeRoute()}, ClientDir: t.TempDir(), HostActions: acts, SessionSecret: secret}
	return Handler(opt)
}

func post(h http.Handler, path, body string, hdr map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("POST", path, strings.NewReader(body))
	req.Host = "app.local"
	req.Header.Set("Origin", "http://app.local")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Gotsx-Action", "1")
	for k, v := range hdr {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestActionsEndToEnd(t *testing.T) {
	calls := 0
	acts := []HostAction{
		{Module: "todos", Name: "toggle", Fn: func(req *Req, args []json.RawMessage) (any, error) {
			var id string
			if err := Arg(args, 0, &id); err != nil {
				return nil, err
			}
			calls++
			if id == "missing" {
				return nil, fmt.Errorf("%w: %s", ErrNotFound, id)
			}
			if id == "" {
				return nil, Invalid(map[string]string{"id": "required"})
			}
			if id == "boom" {
				panic("kaboom")
			}
			req.Session().Set("last", id)
			req.Session().Flash("ok", "toggled "+id)
			return actTodo{ID: id, Done: true}, nil
		}},
		{Module: "todos", Name: "ping", Fn: func(req *Req, args []json.RawMessage) (any, error) { return nil, nil }},
	}
	h := actionServer(t, acts, "secret")

	rec := post(h, "/_gotsx/act/todos/toggle", `["a"]`, nil)
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), `"data":{"id":"a","done":true}`) {
		t.Fatalf("ok call: %d %s", rec.Code, rec.Body.String())
	}
	cookie := rec.Header().Get("Set-Cookie")
	if !strings.Contains(cookie, sessionCookie+"=") || !strings.Contains(cookie, "HttpOnly") {
		t.Errorf("action should save the session cookie: %q", cookie)
	}
	// the page sees session and flash, and the flash appears only once
	get := func() *httptest.ResponseRecorder {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Cookie", strings.Split(cookie, ";")[0])
		r := httptest.NewRecorder()
		h.ServeHTTP(r, req)
		return r
	}
	var seen PageProps
	route := Route{Pattern: "/", Render: func(p PageProps) Node { seen = p; return El("html", nil, El("body", nil)) }}
	h2 := Handler(Options{Routes: []Route{route}, ClientDir: t.TempDir(), HostActions: acts, SessionSecret: "secret"})
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Cookie", strings.Split(cookie, ";")[0])
	r1 := httptest.NewRecorder()
	h2.ServeHTTP(r1, req)
	if seen.Session["last"] != "a" || len(seen.Flash) != 1 || seen.Flash[0].Text != "toggled a" || seen.CSRF() == "" {
		t.Errorf("page props from session: %+v", seen)
	}
	c2 := r1.Header().Get("Set-Cookie")
	if c2 == "" {
		t.Fatal("consuming the flash must write the session cookie back")
	}
	// anonymous visit without reading csrf → no Set-Cookie (cacheable); reading csrf → cookie set
	anon := httptest.NewRecorder()
	h2.ServeHTTP(anon, httptest.NewRequest("GET", "/", nil))
	if anon.Header().Get("Set-Cookie") != "" {
		t.Errorf("anonymous page without csrf use must not set a cookie: %q", anon.Header().Get("Set-Cookie"))
	}
	if seen.CSRF() != "" { // read after rendering: the closure still works, but the cookie can no longer be set — only check it still returns a token
		t.Log("late CSRF() returns a token (documented: read it in the shell)")
	}
	route.Render = func(p PageProps) Node { seen = p; tok := p.CSRF(); return El("html", nil, El("body", nil, Text(tok))) }
	h3 := Handler(Options{Routes: []Route{route}, ClientDir: t.TempDir(), HostActions: acts, SessionSecret: "secret"})
	withTok := httptest.NewRecorder()
	h3.ServeHTTP(withTok, httptest.NewRequest("GET", "/", nil))
	if !strings.Contains(withTok.Header().Get("Set-Cookie"), sessionCookie+"=") || !strings.Contains(withTok.Body.String(), seen.CSRF()) {
		t.Errorf("reading props.csrf must create the token and set the cookie: %q", withTok.Header().Get("Set-Cookie"))
	}
	req2 := httptest.NewRequest("GET", "/", nil)
	req2.Header.Set("Cookie", strings.Split(c2, ";")[0])
	h2.ServeHTTP(httptest.NewRecorder(), req2)
	if len(seen.Flash) != 0 || seen.Session["last"] != "a" {
		t.Errorf("flash must be consumed once, session kept: %+v", seen)
	}
	_ = get

	// 422 validation error
	rec = post(h, "/_gotsx/act/todos/toggle", `[""]`, nil)
	if rec.Code != 422 || !strings.Contains(rec.Body.String(), `"fields":{"id":"required"}`) {
		t.Errorf("validation: %d %s", rec.Code, rec.Body.String())
	}
	// 404 / 500 / argument errors
	if rec = post(h, "/_gotsx/act/todos/toggle", `["missing"]`, nil); rec.Code != 404 {
		t.Errorf("not found: %d", rec.Code)
	}
	if rec = post(h, "/_gotsx/act/todos/toggle", `["boom"]`, nil); rec.Code != 500 || strings.Contains(rec.Body.String(), "kaboom") {
		t.Errorf("panic → 500 without detail in prod: %d %s", rec.Code, rec.Body.String())
	}
	if rec = post(h, "/_gotsx/act/todos/toggle", `[]`, nil); rec.Code != 400 {
		t.Errorf("missing arg: %d %s", rec.Code, rec.Body.String())
	}
	if rec = post(h, "/_gotsx/act/todos/nope", `[]`, nil); rec.Code != 404 {
		t.Errorf("unknown action: %d", rec.Code)
	}
	if rec = post(h, "/_gotsx/act/todos/ping", `[]`, nil); rec.Code != 200 || !strings.Contains(rec.Body.String(), `"data":null`) {
		t.Errorf("void action: %d %s", rec.Code, rec.Body.String())
	}
	// CSRF: missing header / cross-origin → 403
	req = httptest.NewRequest("POST", "/_gotsx/act/todos/toggle", strings.NewReader(`["a"]`))
	req.Host = "app.local"
	req.Header.Set("Origin", "http://app.local")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 403 {
		t.Errorf("missing X-Gotsx-Action should be rejected: %d", rec.Code)
	}
	req = httptest.NewRequest("POST", "/_gotsx/act/todos/toggle", strings.NewReader(`["a"]`))
	req.Host = "app.local"
	req.Header.Set("Origin", "https://evil.example")
	req.Header.Set("X-Gotsx-Action", "1")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 403 {
		t.Errorf("cross-origin should be rejected: %d", rec.Code)
	}
	if calls != 4 {
		t.Errorf("handler calls = %d", calls)
	}
}

func TestSessionSigning(t *testing.T) {
	s := &server{opt: Options{SessionSecret: "k"}}
	sess := s.loadSession(httptest.NewRequest("GET", "/", nil))
	sess.Set("u", "1")
	rec := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	r = r.WithContext(context.WithValue(r.Context(), ctxServer, s))
	sess.save(rec, r)
	ck := rec.Header().Get("Set-Cookie")
	val := strings.TrimPrefix(strings.Split(ck, ";")[0], sessionCookie+"=")
	// tampered → empty session
	tampered := httptest.NewRequest("GET", "/", nil)
	tampered.Header.Set("Cookie", sessionCookie+"="+val[:len(val)-2]+"xx")
	if s.loadSession(tampered).Get("u") != "" {
		t.Error("tampered cookie must not verify")
	}
	good := httptest.NewRequest("GET", "/", nil)
	good.Header.Set("Cookie", sessionCookie+"="+val)
	if s.loadSession(good).Get("u") != "1" {
		t.Error("valid cookie should load")
	}
	// CSRF token verify
	sess2 := s.loadSession(good)
	tok := sess2.CSRF()
	rec2 := httptest.NewRecorder()
	sess2.save(rec2, r)
	val2 := strings.TrimPrefix(strings.Split(rec2.Header().Get("Set-Cookie"), ";")[0], sessionCookie+"=")
	form := httptest.NewRequest("POST", "/x", strings.NewReader("_csrf="+tok))
	form.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	form.Header.Set("Cookie", sessionCookie+"="+val2)
	form = form.WithContext(context.WithValue(form.Context(), ctxServer, s))
	if !VerifyCSRF(form) {
		t.Error("matching _csrf should verify")
	}
	bad := httptest.NewRequest("POST", "/x", strings.NewReader("_csrf=nope"))
	bad.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	bad.Header.Set("Cookie", sessionCookie+"="+val2)
	bad = bad.WithContext(context.WithValue(bad.Context(), ctxServer, s))
	if VerifyCSRF(bad) {
		t.Error("wrong token must fail")
	}
}

func TestActionStatusesAndNilProps(t *testing.T) {
	acts := []HostAction{
		{Module: "m", Name: "auth", Fn: func(req *Req, args []json.RawMessage) (any, error) { return nil, Unauthorized("sign in") }},
		{Module: "m", Name: "role", Fn: func(req *Req, args []json.RawMessage) (any, error) { return nil, Forbidden("viewer") }},
		{Module: "m", Name: "bad", Fn: func(req *Req, args []json.RawMessage) (any, error) { return nil, Fail("nope") }},
	}
	h := actionServer(t, acts, "k")
	for name, want := range map[string]int{"auth": 401, "role": 403, "bad": 400} {
		if rec := post(h, "/_gotsx/act/m/"+name, `[]`, nil); rec.Code != want {
			t.Errorf("%s: %d %s", name, rec.Code, rec.Body.String())
		}
	}
	if (&ValidationError{Message: "validation failed", Fields: map[string]string{"b": "B", "a": "A"}}).Error() != "a: A; b: B" {
		t.Error("ValidationError.Error should list fields")
	}
	// island props: nil slice / map → [] / {}, embedded structs promoted, json tags and omitempty honored, []byte untouched
	type inner struct {
		Tags []string            `json:"tags"`
		M    map[string]int      `json:"m"`
		Skip string              `json:"-"`
		Opt  []string            `json:"opt,omitempty"`
		Raw  []byte              `json:"raw"`
		Ptr  *[]int              `json:"ptr"`
		Nest []struct{ X []int } `json:"nest"`
	}
	type outer struct {
		inner
		Name  string  `json:"name"`
		Items []inner `json:"items"`
		Flash []Flash `json:"flash"`
	}
	b, _ := json.Marshal(noNil(outer{Name: "n"}))
	got := string(b)
	for _, want := range []string{`"tags":[]`, `"m":{}`, `"name":"n"`, `"items":[]`, `"flash":[]`, `"raw":null`, `"ptr":null`, `"nest":[]`} {
		if !strings.Contains(got, want) {
			t.Errorf("noNil: want %s in %s", want, got)
		}
	}
	if strings.Contains(got, "opt") || strings.Contains(got, "Skip") {
		t.Errorf("noNil: omitempty / - ignored: %s", got)
	}
	var out strings.Builder
	n := Island("X", struct {
		Initial []Flash `json:"initial"`
	}{}, nil)
	c := &Ctx{}
	n(c)
	out.WriteString(c.b.String())
	if !strings.Contains(out.String(), `initial&#34;:[]`) && !strings.Contains(out.String(), `initial":[]`) && !strings.Contains(out.String(), `initial&quot;:[]`) {
		t.Errorf("island props should carry [] for a nil slice: %s", out.String())
	}
}

func TestReviewFixes(t *testing.T) {
	// 1. cyclic island props must not overflow the stack
	type node struct {
		Name     string  `json:"name"`
		Parent   *node   `json:"parent"`
		Children []*node `json:"children"`
	}
	root := &node{Name: "r"}
	child := &node{Name: "c", Parent: root}
	root.Children = []*node{child}
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("cyclic props panicked: %v", r)
			}
		}()
		_ = noNil(root)
	}()
	// 2. noNil parity with encoding/json on nil-free inputs (embedded pointer promotion, omitempty)
	type base struct {
		A []int `json:"a"`
	}
	type withPtr struct {
		*base
		N     int      `json:"n"`
		Empty []string `json:"empty,omitempty"`
		Zero  int      `json:"zero,omitempty"`
		S     string   `json:"s,omitempty"`
		Keep  struct{} `json:"keep,omitempty"`
	}
	for _, v := range []any{withPtr{base: &base{A: []int{1}}, N: 1, Empty: []string{}}, withPtr{N: 2, S: "x"}} {
		want, _ := json.Marshal(v)
		got, _ := json.Marshal(noNil(v))
		var w, g any
		json.Unmarshal(want, &w)
		json.Unmarshal(got, &g)
		if !reflect.DeepEqual(w, g) { // same document; key order may differ (maps marshal sorted)
			t.Errorf("noNil parity:\n json: %s\n noNil: %s", want, got)
		}
	}
	if b, _ := json.Marshal(noNil(withPtr{base: &base{}, N: 3})); !strings.Contains(string(b), `"a":[]`) {
		t.Errorf("nil slice through an embedded pointer should become []: %s", b)
	}
	// 3. the X-Gotsx-Action header is required even with DisableCSRF (HTML forms cannot set it)
	acts := []HostAction{{Module: "m", Name: "n", Fn: func(req *Req, args []json.RawMessage) (any, error) { return "ok", nil }}}
	h := Handler(Options{Routes: []Route{homeRoute()}, ClientDir: t.TempDir(), HostActions: acts, DisableCSRF: true, SessionSecret: "k"})
	form := httptest.NewRequest("POST", "/_gotsx/act/m/n", strings.NewReader(`["1"]=`))
	form.Header.Set("Content-Type", "text/plain")
	form.Header.Set("Origin", "https://evil.example")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, form)
	if rec.Code != 403 {
		t.Errorf("text/plain form post without the header must be rejected: %d", rec.Code)
	}
	if rec = post(h, "/_gotsx/act/m/n", `[]`, map[string]string{"Origin": "https://elsewhere.example"}); rec.Code != 200 {
		t.Errorf("DisableCSRF still allows cross-origin calls that carry the header: %d", rec.Code)
	}
	// 4. a custom 404 page that reads csrf gets the cookie; Unauthorized keeps its fields
	nf := func(p PageProps) Node { tok := p.CSRF(); return El("html", nil, El("body", nil, Text(tok))) }
	h404 := Handler(Options{Routes: []Route{homeRoute()}, ClientDir: t.TempDir(), NotFound: nf, SessionSecret: "k"})
	rec = httptest.NewRecorder()
	h404.ServeHTTP(rec, httptest.NewRequest("GET", "/missing", nil))
	if rec.Code != 404 || !strings.Contains(rec.Header().Get("Set-Cookie"), sessionCookie+"=") {
		t.Errorf("404 page reading csrf must persist the session: %d %q", rec.Code, rec.Header().Get("Set-Cookie"))
	}
	ve := &ValidationError{Message: "no", Fields: map[string]string{"x": "y"}, Status: 401}
	if ve.Error() != "no: x: y" {
		t.Errorf("Error(): %q", ve.Error())
	}
	// 5. Clear() removes the cookie; a hand-built Req has a detached session
	s := &server{opt: Options{SessionSecret: "k"}}
	sess := &Session{}
	sess.Set("a", "1")
	sess.Clear()
	rec = httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	sess.save(rec, r.WithContext(context.WithValue(r.Context(), ctxServer, s)))
	if ck := rec.Header().Get("Set-Cookie"); !strings.Contains(ck, "Max-Age=0") {
		t.Errorf("Clear should delete the cookie: %q", ck)
	}
	bare := &Req{}
	bare.Session().Set("k", "v")
	if bare.Session().Get("k") != "v" {
		t.Error("a Req without a server must still offer a session")
	}
	// 6. hostgen rejects *Req in a later position and unsupported kinds
	for _, v := range []any{badReqPos{}, badChan{}} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("hostgen should reject %T", v)
				}
			}()
			GenerateHost(map[string]HostModule{"x": {Value: v, Go: "host.X"}}, "host")
		}()
	}
}

type badReqPos struct{}

func (badReqPos) Bad(id string, req *Req) error { return nil }

type badChan struct{}

func (badChan) Watch() chan int { return nil }

type ptrMarshal struct{ V int }

func (p *ptrMarshal) MarshalJSON() ([]byte, error) { return []byte(`"ptr-marshaled"`), nil }

type textMarshal struct{ v string }

func (t textMarshal) MarshalText() ([]byte, error) { return []byte(t.v), nil }

type deepNode struct {
	Next *deepNode `json:"next"`
	Tags []string  `json:"tags"`
}

type typeWithReq struct{}

func (typeWithReq) Touch(req *Req) {}

type modWithTypeReq struct {
	T *typeWithReq `json:"t"`
}

func TestNoNilRoundTwo(t *testing.T) {
	// marshalers pass through exactly as encoding/json would
	in := struct {
		P    *ptrMarshal            `json:"p"`
		M    map[string]*ptrMarshal `json:"m"`
		Text textMarshal            `json:"text"`
		When time.Time              `json:"when"`
		Nil  []int                  `json:"nil"`
	}{P: &ptrMarshal{2}, M: map[string]*ptrMarshal{"k": {3}}, Text: textMarshal{"hi"}, When: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)}
	got, _ := json.Marshal(noNil(in))
	for _, want := range []string{`"p":"ptr-marshaled"`, `"m":{"k":"ptr-marshaled"}`, `"text":"hi"`, `"when":"2026-01-02T03:04:05Z"`, `"nil":[]`} {
		if !strings.Contains(string(got), want) {
			t.Errorf("want %s in %s", want, got)
		}
	}
	// a deep but acyclic chain is kept whole; a cycle stops at the back-reference
	root := &deepNode{}
	cur := root
	for i := 0; i < 200; i++ {
		cur.Next = &deepNode{}
		cur = cur.Next
	}
	got, _ = json.Marshal(noNil(root))
	if strings.Count(string(got), `"tags":[]`) != 201 {
		t.Errorf("deep chain truncated: %d nodes", strings.Count(string(got), `"tags":[]`))
	}
	cyc := &deepNode{}
	cyc.Next = cyc
	got, _ = json.Marshal(noNil(cyc))
	if string(got) != `{"next":null,"tags":[]}` {
		t.Errorf("cycle: %s", got)
	}
	// hostgen: *Req on a type method is rejected
	func() {
		defer func() {
			if recover() == nil {
				t.Error("hostgen must reject *Req on type methods")
			}
		}()
		GenerateHost(map[string]HostModule{"x": {Value: modWithTypeReq{}, Go: "host.X"}}, "host")
	}()
	// IsHTTPS through proxies
	for hdr, val := range map[string]string{"X-Forwarded-Proto": "https", "Forwarded": "for=1.2.3.4;proto=https", "CF-Visitor": `{"scheme":"https"}`} {
		r := httptest.NewRequest("GET", "/", nil)
		r.Header.Set(hdr, val)
		if !IsHTTPS(r) {
			t.Errorf("%s should mean https", hdr)
		}
	}
	if IsHTTPS(httptest.NewRequest("GET", "/", nil)) {
		t.Error("plain request is not https")
	}
}
