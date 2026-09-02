package gotsx

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func testServer(t *testing.T, opt Options) http.Handler {
	t.Helper()
	dir := t.TempDir()
	os := opt
	os.ClientDir = dir
	os.Addr = ":0"
	s := &server{opt: os}
	s.scanClient()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok")) })
	mux.HandleFunc("/", s.page)
	for p, h := range os.Actions {
		mux.HandleFunc(p, s.csrfGuard(p, h))
	}
	var h http.Handler = mux
	if !os.DisableGzip {
		h = gzipMW(h)
	}
	h = s.secHeadersMW(h)
	h = s.recoverMW(h)
	h = accessLogMW(h)
	h = requestIDMW(h)
	return h
}

func homeRoute() Route {
	return Route{Pattern: "/", Segs: nil, Render: func(p PageProps) Node {
		return El("html", nil, El("head", nil), El("body", nil, El("h1", nil, Text("首页"))))
	}}
}

func TestSecurityHeaders(t *testing.T) {
	h := testServer(t, Options{Routes: []Route{homeRoute()}})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	for k, want := range map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "SAMEORIGIN",
		"Referrer-Policy":        "strict-origin-when-cross-origin",
	} {
		if got := rec.Header().Get(k); got != want {
			t.Errorf("%s = %q, want %q", k, got, want)
		}
	}
	csp := rec.Header().Get("Content-Security-Policy")
	if !strings.Contains(csp, "script-src 'self' 'nonce-") {
		t.Errorf("CSP 缺少 nonce script-src: %q", csp)
	}
	// 内联注入脚本必须带同一个 nonce
	body := rec.Body.String()
	nonce := csp[strings.Index(csp, "nonce-")+6:]
	nonce = nonce[:strings.IndexByte(nonce, '\'')]
	if !strings.Contains(body, `<script nonce="`+nonce+`">window.__GOTSX`) {
		t.Errorf("内联脚本未带 CSP nonce %q\n%s", nonce, body)
	}
}

func TestNonceIsPerResponse(t *testing.T) {
	h := testServer(t, Options{Routes: []Route{homeRoute()}})
	get := func() string {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
		return rec.Header().Get("Content-Security-Policy")
	}
	if get() == get() {
		t.Error("每个响应的 CSP nonce 应不同")
	}
}

func TestRequestID(t *testing.T) {
	h := testServer(t, Options{Routes: []Route{homeRoute()}})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	if rec.Header().Get("X-Request-Id") == "" {
		t.Error("缺少 X-Request-Id")
	}
	// 透传上游 id
	rec2 := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Request-Id", "abc123")
	h.ServeHTTP(rec2, req)
	if rec2.Header().Get("X-Request-Id") != "abc123" {
		t.Error("应透传上游 X-Request-Id")
	}
}

func TestGzip(t *testing.T) {
	big := strings.Repeat("x", 5000)
	route := Route{Pattern: "/", Render: func(p PageProps) Node {
		return El("html", nil, El("head", nil), El("body", nil, Text(big)))
	}}
	h := testServer(t, Options{Routes: []Route{route}})
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Header().Get("Content-Encoding") != "gzip" {
		t.Fatalf("期望 gzip, 头: %v", rec.Header())
	}
	gr, err := gzip.NewReader(rec.Body)
	if err != nil {
		t.Fatal(err)
	}
	out, _ := io.ReadAll(gr)
	if !strings.Contains(string(out), big) {
		t.Error("gzip 解压内容不对")
	}
}

func TestNotFound(t *testing.T) {
	h := testServer(t, Options{Routes: []Route{homeRoute()}})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/nope", nil))
	if rec.Code != 404 {
		t.Errorf("状态 = %d, want 404", rec.Code)
	}
}

func TestCustom404(t *testing.T) {
	h := testServer(t, Options{
		Routes:   []Route{homeRoute()},
		NotFound: func(p PageProps) Node { return El("html", nil, El("head", nil), El("body", nil, Text("自定义404"))) },
	})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/nope", nil))
	if rec.Code != 404 || !strings.Contains(rec.Body.String(), "自定义404") {
		t.Errorf("自定义 404 页未生效: %d %s", rec.Code, rec.Body.String())
	}
}

func TestPanicRecovery(t *testing.T) {
	route := Route{Pattern: "/", Render: func(p PageProps) Node {
		panic("boom in render")
	}}
	h := testServer(t, Options{Routes: []Route{route}, Dev: false})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	if rec.Code != 500 {
		t.Errorf("状态 = %d, want 500", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "boom in render") {
		t.Error("prod 模式不应把 panic 细节泄露给客户端")
	}
}

func TestPanicRecoveryDevShowsDetail(t *testing.T) {
	route := Route{Pattern: "/", Render: func(p PageProps) Node { panic("boom42") }}
	h := testServer(t, Options{Routes: []Route{route}, Dev: true})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	if !strings.Contains(rec.Body.String(), "boom42") {
		t.Error("dev 模式应显示 panic 细节")
	}
}

func TestActionPanicRecovered(t *testing.T) {
	h := testServer(t, Options{
		Routes:  []Route{homeRoute()},
		Actions: map[string]http.HandlerFunc{"POST /act": func(w http.ResponseWriter, r *http.Request) { panic("action boom") }},
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/act", nil)
	req.Header.Set("Origin", "http://example.com")
	req.Host = "example.com"
	h.ServeHTTP(rec, req)
	if rec.Code != 500 {
		t.Errorf("action panic 应被恢复为 500, 得 %d", rec.Code)
	}
}

func TestHealthz(t *testing.T) {
	h := testServer(t, Options{Routes: []Route{homeRoute()}})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/healthz", nil))
	if rec.Code != 200 || rec.Body.String() != "ok" {
		t.Error("healthz")
	}
}

func TestDisableCSP(t *testing.T) {
	h := testServer(t, Options{Routes: []Route{homeRoute()}, DisableCSP: true})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	if rec.Header().Get("Content-Security-Policy") != "" {
		t.Error("DisableCSP 应关掉 CSP")
	}
}

func TestCSRFBlocksCrossOrigin(t *testing.T) {
	called := false
	h := testServer(t, Options{
		Routes:  []Route{homeRoute()},
		Actions: map[string]http.HandlerFunc{"POST /act": func(w http.ResponseWriter, r *http.Request) { called = true; w.Write([]byte("ok")) }},
	})
	// 跨源 POST → 403
	req := httptest.NewRequest("POST", "/act", nil)
	req.Header.Set("Origin", "https://evil.example")
	req.Host = "shop.local"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 403 || called {
		t.Errorf("跨源 POST 应 403 且不执行 handler, 得 %d called=%v", rec.Code, called)
	}
	// 同源 POST → 放行
	called = false
	req2 := httptest.NewRequest("POST", "/act", nil)
	req2.Header.Set("Origin", "http://shop.local")
	req2.Host = "shop.local"
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, req2)
	if rec2.Code != 200 || !called {
		t.Errorf("同源 POST 应放行, 得 %d called=%v", rec2.Code, called)
	}
	// 缺少 Origin/Referer 的 POST → 拒绝
	called = false
	rec3 := httptest.NewRecorder()
	req3 := httptest.NewRequest("POST", "/act", nil)
	req3.Host = "shop.local"
	h.ServeHTTP(rec3, req3)
	if rec3.Code != 403 || called {
		t.Errorf("无 Origin 的 POST 应拒绝, 得 %d", rec3.Code)
	}
}

func TestCSRFAllowsGET(t *testing.T) {
	called := false
	h := testServer(t, Options{
		Routes:  []Route{homeRoute()},
		Actions: map[string]http.HandlerFunc{"GET /api/x": func(w http.ResponseWriter, r *http.Request) { called = true; w.Write([]byte("ok")) }},
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/x", nil)
	req.Host = "shop.local"
	h.ServeHTTP(rec, req)
	if rec.Code != 200 || !called {
		t.Errorf("GET(读)应放行, 得 %d", rec.Code)
	}
}
