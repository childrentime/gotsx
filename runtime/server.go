package gotsx

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

type Route struct {
	Pattern string
	Segs    []string
	Render  func(PageProps) Node
}

// Match: 段 "{id}" 是参数; 最后一段 "{...rest}" 是 catch-all(匹配 ≥1 个剩余段, 参数值是用 / 拼起来的路径)
func (r Route) Match(path string) (map[string]string, bool) {
	p := strings.Trim(path, "/")
	var segs []string
	if p != "" {
		segs = strings.Split(p, "/")
	}
	n := len(r.Segs)
	catchAll := n > 0 && strings.HasPrefix(r.Segs[n-1], "{...")
	if catchAll {
		if len(segs) < n {
			return nil, false
		}
	} else if len(segs) != n {
		return nil, false
	}
	params := map[string]string{}
	for i, s := range r.Segs {
		switch {
		case catchAll && i == n-1:
			params[s[4:len(s)-1]] = strings.Join(segs[i:], "/")
		case strings.HasPrefix(s, "{"):
			params[s[1:len(s)-1]] = segs[i]
		case s != segs[i]:
			return nil, false
		}
	}
	return params, true
}

// IsCatchAll: 路由是否以 catch-all 段结尾
func (r Route) IsCatchAll() bool {
	return len(r.Segs) > 0 && strings.HasPrefix(r.Segs[len(r.Segs)-1], "{...")
}

type Options struct {
	Addr      string
	Routes    []Route
	ClientDir string // 生成的客户端 JS 目录, 挂在 /_gotsx/
	PublicDir string // 静态文件, 挂在 /public/
	Actions   map[string]http.HandlerFunc
	// Before 在页面渲染前调用: 可以设置响应头/种 cookie, 并把要暴露给页面的 cookie 写进 map
	Before func(http.ResponseWriter, *http.Request, map[string]string)

	// ---- 生产加固(全部可选, 有企业级默认) ----
	ClientFS        interface{}                       // fs.FS: 内嵌客户端资源(单二进制); 非 Dev 且非空时优先于 ClientDir
	PublicFS        interface{}                       // fs.FS: 内嵌静态资源
	Dev             bool                              // 开发模式: 错误页带堆栈, 资源 no-cache
	NotFound        func(PageProps) Node              // 自定义 404 页
	ErrorPage       func(PageProps, error) Node       // 自定义 500 页(err 在 Dev 下才含细节)
	SecurityHeaders map[string]string                 // 覆盖/追加安全响应头(值为 "" 则删除该默认头)
	DisableCSP      bool                              // 关掉默认 CSP(带 nonce)
	DisableGzip     bool                              // 关掉响应压缩
	DisableCSRF     bool                              // 关掉对写操作(POST/PUT/PATCH/DELETE)的同源校验
	ReadyCheck      func() error                      // /readyz 就绪探针(nil=总就绪)
	OnShutdown      func()                            // 优雅关闭钩子
	Middleware      []func(http.Handler) http.Handler // 应用级中间件(如鉴权重定向), 作用于全部路由
	// OnClientEvent 开启客户端遥测: 浏览器把 JS 错误 / 页面浏览 上报到 /_gotsx/client-log, 这里接收
	OnClientEvent func(ClientEvent, *http.Request)
	I18n          *I18n // 可选国际化: 语言解析 + 客户端目录注入 + hreflang
	QuietLogs     bool  // 关掉每请求的访问日志(高吞吐服务或自带日志中间件时)

	// ---- typed actions and sessions ----
	HostActions   []HostAction  // gen.HostActions: host methods islands may call directly (POST /_gotsx/act/<module>/<name>)
	SessionSecret string        // HMAC key for the session cookie; empty → a random key per start (sessions do not survive restarts)
	SessionMaxAge time.Duration // lifetime of the session cookie (default 30 days)
}

// ClientEvent 是浏览器上报的一条遥测(错误 / 页面浏览)
type ClientEvent struct {
	Type    string `json:"type"` // error | rejection | pageview
	Message string `json:"message,omitempty"`
	Stack   string `json:"stack,omitempty"`
	URL     string `json:"url,omitempty"`
	Ref     string `json:"ref,omitempty"`
}

type server struct {
	opt        Options
	manifest   string
	loader     string
	logURL     string
	bootID     string // 每次进程启动不同: dev 模式下浏览器据此判断服务重启过 → 自动刷新
	autoSecret string // random key used when SessionSecret is empty
}

type ctxKey int

const (
	ctxReqID ctxKey = iota
	ctxNonce
	ctxLocale
	ctxPath
	ctxServer
)

// ---------- 启动 ----------

func asFS(v interface{}) fs.FS {
	if f, ok := v.(fs.FS); ok {
		return f
	}
	return nil
}

func (s *server) clientFS() fs.FS {
	if !s.opt.Dev {
		if f := asFS(s.opt.ClientFS); f != nil {
			return f
		}
	}
	return os.DirFS(s.opt.ClientDir)
}

func (s *server) publicFS() fs.FS {
	if !s.opt.Dev {
		if f := asFS(s.opt.PublicFS); f != nil {
			return f
		}
	}
	if s.opt.PublicDir == "" {
		return nil
	}
	return os.DirFS(s.opt.PublicDir)
}

// Handler builds the complete gotsx HTTP handler (pages, islands, actions, middleware chain) without starting a
// server — mount it in your own http.Server or mux, or drive it from tests / benchmarks with httptest.
func Handler(opt Options) http.Handler {
	h, _ := newServer(opt)
	return h
}

func newServer(opt Options) (http.Handler, *server) {
	s := &server{opt: opt}
	s.bootID = randomHex(6)
	if opt.SessionSecret == "" {
		s.autoSecret = randomHex(32)
		if len(opt.HostActions) > 0 || opt.Dev {
			log.Printf("gotsx: Options.SessionSecret is empty: sessions, flash messages and CSRF tokens will not survive a restart (set SESSION_SECRET in production)")
		}
	}
	configureI18n(opt.I18n)
	s.scanClient()
	s.opt.Routes = sortRoutes(opt.Routes)

	mux := http.NewServeMux()
	if opt.Dev {
		mux.HandleFunc("GET /_gotsx/dev", s.devEvents)
	}
	mux.Handle("GET /_gotsx/", s.assetCache(http.StripPrefix("/_gotsx/", http.FileServer(http.FS(s.clientFS())))))
	if pf := s.publicFS(); pf != nil {
		mux.Handle("GET /public/", s.assetCache(http.StripPrefix("/public/", http.FileServer(http.FS(pf)))))
	}
	for pattern, h := range opt.Actions {
		mux.HandleFunc(pattern, s.csrfGuard(pattern, h))
	}
	if len(opt.HostActions) > 0 {
		acts := map[string]HostAction{}
		for _, a := range opt.HostActions {
			if _, dup := acts[a.Module+"/"+a.Name]; dup {
				log.Printf("gotsx: duplicate action %s/%s in Options.HostActions (the last one wins)", a.Module, a.Name)
			}
			acts[a.Module+"/"+a.Name] = a
		}
		mux.HandleFunc("POST "+actionPrefix, s.actionHandler(acts))
	}
	if opt.OnClientEvent != nil || opt.Dev {
		mux.HandleFunc("POST /_gotsx/client-log", func(w http.ResponseWriter, r *http.Request) {
			if !SameOrigin(r) { // 仅同源上报
				w.WriteHeader(http.StatusNoContent)
				return
			}
			var ev ClientEvent
			json.NewDecoder(io.LimitReader(r.Body, 16*1024)).Decode(&ev)
			if s.opt.OnClientEvent != nil {
				s.opt.OnClientEvent(ev, r)
			} else if ev.Type != "pageview" { // dev: browser errors / console.error show up in the terminal
				msg := ev.Message
				if ev.Stack != "" {
					if i := strings.Index(ev.Stack, "\n"); i > 0 {
						msg += "  " + strings.TrimSpace(strings.SplitN(ev.Stack, "\n", 3)[1])
					}
				}
				clean := strings.NewReplacer("\n", " ", "\r", " ")
				log.Printf("[browser] %s: %s (%s)", clean.Replace(ev.Type), clean.Replace(msg), clean.Replace(ev.URL))
			}
			w.WriteHeader(http.StatusNoContent)
		})
		s.logURL = "/_gotsx/client-log"
	}
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok")) })
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		if opt.ReadyCheck != nil {
			if err := opt.ReadyCheck(); err != nil {
				http.Error(w, "not ready: "+err.Error(), http.StatusServiceUnavailable)
				return
			}
		}
		w.Write([]byte("ready"))
	})
	mux.HandleFunc("GET /favicon.ico", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) })
	mux.HandleFunc("/", s.page)

	// 中间件链(由外到内): requestID → 访问日志 → panic 恢复 → 安全头 → 应用中间件 → gzip
	var h http.Handler = mux
	if !opt.DisableGzip {
		h = gzipMW(h)
	}
	for i := len(opt.Middleware) - 1; i >= 0; i-- {
		h = opt.Middleware[i](h)
	}
	h = s.secHeadersMW(h)
	h = s.recoverMW(h)
	if !opt.QuietLogs {
		h = accessLogMW(h)
	}
	h = requestIDMW(h)
	h = s.serverCtxMW(h)
	return h, s
}

// serverCtxMW puts the server into the request context (so SessionOf / VerifyCSRF work from plain handlers too)
func (s *server) serverCtxMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxServer, s)))
	})
}

func Serve(opt Options) error {
	h, s := newServer(opt)
	srv := &http.Server{
		Addr:              opt.Addr,
		Handler:           h,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	// 优雅关闭
	idle := make(chan struct{})
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
		<-sig
		log.Printf("gotsx: shutdown signal received, draining…")
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if opt.OnShutdown != nil {
			opt.OnShutdown()
		}
		srv.Shutdown(ctx)
		close(idle)
	}()

	mode := "prod"
	if opt.Dev {
		mode = "dev"
	}
	log.Printf("gotsx: listening on http://localhost%s  routes=%d  mode=%s", opt.Addr, len(s.opt.Routes), mode)
	opt = s.opt
	err := srv.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		<-idle
		return nil
	}
	return err
}

// sortRoutes: 具体的优先 —— 非 catch-all 在 catch-all 之前; 参数少的在前; 参数一样多时静态段多的在前; 同级保持声明顺序
func sortRoutes(in []Route) []Route {
	routes := append([]Route(nil), in...)
	key := func(r Route) (catchAll bool, params, static int) {
		for _, s := range r.Segs {
			if strings.HasPrefix(s, "{") {
				params++
			} else {
				static++
			}
		}
		return r.IsCatchAll(), params, static
	}
	sort.SliceStable(routes, func(i, j int) bool {
		ci, pi, si := key(routes[i])
		cj, pj, sj := key(routes[j])
		if ci != cj {
			return !ci
		}
		if pi != pj {
			return pi < pj
		}
		return si > sj
	})
	return routes
}

func randomHex(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// devEvents: dev 模式的 SSE 流。每条消息带本进程的 bootID; gotsx dev 重启进程后浏览器重连拿到新 id → 整页刷新
func (s *server) devEvents(w http.ResponseWriter, r *http.Request) {
	fl, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Accel-Buffering", "no")
	http.NewResponseController(w).SetWriteDeadline(time.Time{}) // 长连接: 关掉本响应的写超时
	fmt.Fprintf(w, "retry: 300\ndata: %s\n\n", s.bootID)
	// gotsx dev writes compile errors to .gotsx/diagnostics.json (removed on success); watch it here and push changes to the browser's error overlay
	diagPath := filepath.Join(".gotsx", "diagnostics.json")
	var diagSeen time.Time
	pushDiag := func() bool {
		st, err := os.Stat(diagPath)
		switch {
		case err != nil && diagSeen.IsZero():
			return true
		case err != nil:
			diagSeen = time.Time{}
			_, werr := fmt.Fprintf(w, "event: diag\ndata: {}\n\n")
			return werr == nil
		case st.ModTime().Equal(diagSeen):
			return true
		}
		diagSeen = st.ModTime()
		raw, err := os.ReadFile(diagPath)
		if err != nil || !json.Valid(raw) {
			return true
		}
		_, werr := fmt.Fprintf(w, "event: diag\ndata: %s\n\n", bytes.ReplaceAll(raw, []byte("\n"), []byte(" ")))
		return werr == nil
	}
	pushDiag()
	fl.Flush()
	tick := time.NewTicker(15 * time.Second)
	defer tick.Stop()
	poll := time.NewTicker(500 * time.Millisecond)
	defer poll.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case <-poll.C:
			if !pushDiag() {
				return
			}
			fl.Flush()
		case <-tick.C:
			if _, err := fmt.Fprintf(w, "data: %s\n\n", s.bootID); err != nil {
				return
			}
			fl.Flush()
		}
	}
}

// fillScript: 流式 Suspense 的客户端一半 —— 把 <template data-gotsx-fill> 换进对应的 <gotsx-suspense>;
// 目标还不在文档里(父边界尚未填充)就先留着, 每次填充后再扫一遍。岛由 loader 的 reconcile 挂载。
const fillScript = `window.__gotsxFill=function(id){var t=document.querySelector('template[data-gotsx-fill="'+id+'"]'),e=document.getElementById(id);if(!t||!e)return;e.replaceChildren(t.content);e.setAttribute("data-ready","");t.remove();if(window.__gotsxReconcile)window.__gotsxReconcile();document.querySelectorAll("template[data-gotsx-fill]").forEach(function(x){window.__gotsxFill(x.getAttribute("data-gotsx-fill"))})}`

// ---------- 中间件 ----------

func requestIDMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-Id")
		if id == "" {
			b := make([]byte, 8)
			rand.Read(b)
			id = hex.EncodeToString(b)
		}
		w.Header().Set("X-Request-Id", id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxReqID, id)))
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (s *statusRecorder) WriteHeader(c int) { s.status = c; s.ResponseWriter.WriteHeader(c) }
func (s *statusRecorder) Write(b []byte) (int, error) {
	if s.status == 0 {
		s.status = 200
	}
	n, err := s.ResponseWriter.Write(b)
	s.bytes += n
	return n, err
}
func (s *statusRecorder) Flush() {
	if f, ok := s.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}
func (s *statusRecorder) Unwrap() http.ResponseWriter { return s.ResponseWriter }

func accessLogMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t := time.Now()
		rec := &statusRecorder{ResponseWriter: w}
		next.ServeHTTP(rec, r)
		if rec.status == 0 {
			rec.status = 200
		}
		id, _ := r.Context().Value(ctxReqID).(string)
		log.Printf("%s %s %d %dB %s id=%s", r.Method, r.URL.Path, rec.status, rec.bytes, time.Since(t).Round(time.Microsecond), id)
	})
}

func (s *server) recoverMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				id, _ := r.Context().Value(ctxReqID).(string)
				log.Printf("PANIC %s %s id=%s: %v\n%s", r.Method, r.URL.Path, id, rec, debug.Stack())
				err := fmt.Errorf("%v", rec)
				s.serveError(w, r, http.StatusInternalServerError, err)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (s *server) secHeadersMW(next http.Handler) http.Handler {
	defaults := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "SAMEORIGIN",
		"Referrer-Policy":        "strict-origin-when-cross-origin",
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		for k, v := range defaults {
			h.Set(k, v)
		}
		for k, v := range s.opt.SecurityHeaders {
			if v == "" {
				h.Del(k)
			} else {
				h.Set(k, v)
			}
		}
		// CSP + 每响应 nonce(注入到内联 <script>)
		if !s.opt.DisableCSP {
			nb := make([]byte, 12)
			rand.Read(nb)
			nonce := hex.EncodeToString(nb)
			r = r.WithContext(context.WithValue(r.Context(), ctxNonce, nonce))
			if _, ok := s.opt.SecurityHeaders["Content-Security-Policy"]; !ok {
				h.Set("Content-Security-Policy",
					"default-src 'self'; base-uri 'self'; object-src 'none'; frame-ancestors 'self'; "+
						"script-src 'self' 'nonce-"+nonce+"'; style-src 'self' 'unsafe-inline'; "+
						"img-src 'self' data:; font-src 'self' data:; connect-src 'self'")
			}
		}
		next.ServeHTTP(w, r)
	})
}

type gzipWriter struct {
	http.ResponseWriter
	gz     *gzip.Writer
	on     bool
	status int
}

func (g *gzipWriter) WriteHeader(c int) {
	g.status = c
	ct := g.Header().Get("Content-Type")
	if compressible(ct) && g.Header().Get("Content-Encoding") == "" {
		g.on = true
		g.Header().Set("Content-Encoding", "gzip")
		g.Header().Add("Vary", "Accept-Encoding")
		g.Header().Del("Content-Length")
	}
	g.ResponseWriter.WriteHeader(c)
}
func (g *gzipWriter) Write(b []byte) (int, error) {
	if g.status == 0 {
		g.WriteHeader(200)
	}
	if g.on {
		return g.gz.Write(b)
	}
	return g.ResponseWriter.Write(b)
}
func (g *gzipWriter) Flush() {
	if g.on {
		g.gz.Flush()
	}
	if f, ok := g.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}
func (g *gzipWriter) Unwrap() http.ResponseWriter { return g.ResponseWriter }

func compressible(ct string) bool {
	for _, p := range []string{"text/html", "text/css", "text/javascript", "application/javascript", "application/json", "image/svg", "text/plain"} {
		if strings.HasPrefix(ct, p) {
			return true
		}
	}
	return false
}

func gzipMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}
		gz := gzip.NewWriter(w)
		defer gz.Close()
		gw := &gzipWriter{ResponseWriter: w, gz: gz}
		next.ServeHTTP(gw, r)
	})
}

// SameOrigin: 校验请求来源与 Host 一致(CSRF 防线, 同站表单/fetch 才带 Origin/Referer)
func SameOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		origin = r.Header.Get("Referer")
	}
	if origin == "" {
		return false // 写操作缺少 Origin/Referer, 保守拒绝
	}
	i := strings.Index(origin, "://")
	if i < 0 {
		return false
	}
	host := origin[i+3:]
	if j := strings.IndexAny(host, "/?#"); j >= 0 {
		host = host[:j]
	}
	return host == r.Host
}

func unsafeMethod(m string) bool {
	switch m {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	}
	return false
}

// csrfGuard: 对写操作做同源校验; GET/HEAD 等安全方法放行
func (s *server) csrfGuard(pattern string, h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !s.opt.DisableCSRF && unsafeMethod(r.Method) && !SameOrigin(r) {
			id, _ := r.Context().Value(ctxReqID).(string)
			log.Printf("CSRF rejected %s %s id=%s origin=%q referer=%q", r.Method, r.URL.Path, id, r.Header.Get("Origin"), r.Header.Get("Referer"))
			http.Error(w, "cross-origin request rejected", http.StatusForbidden)
			return
		}
		h(w, r)
	}
}

// assetCache: 客户端/静态资源缓存策略。URL 带内容哈希 → immutable 长缓存; 否则 no-cache。dev 一律 no-cache。
func (s *server) assetCache(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.opt.Dev {
			w.Header().Set("Cache-Control", "no-cache")
		} else if r.URL.Query().Get("v") != "" {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		} else {
			w.Header().Set("Cache-Control", "public, max-age=300")
		}
		next.ServeHTTP(w, r)
	})
}

// ---------- 岛 manifest(内容哈希) ----------

func (s *server) scanClient() {
	cfs := s.clientFS()
	islands := map[string]string{}
	var only map[string]bool
	if b, err := fs.ReadFile(cfs, "islands.json"); err == nil {
		var names []string
		if json.Unmarshal(b, &names) == nil {
			only = map[string]bool{}
			for _, n := range names {
				only[n] = true
			}
		}
	}
	ents, _ := fs.ReadDir(cfs, ".")
	for _, e := range ents {
		n := e.Name()
		if !strings.HasSuffix(n, ".js") {
			continue
		}
		v := s.fileHashFS(cfs, n)
		name := strings.TrimSuffix(n, ".js")
		switch {
		case n == "loader.js":
			s.loader = "/_gotsx/loader.js?v=" + v
		case n == "runtime.js" || n == "idiomorph.esm.js":
		case only != nil && !only[name]:
		default:
			islands[name] = "/_gotsx/" + n + "?v=" + v
		}
	}
	b, _ := json.Marshal(map[string]any{"islands": islands})
	s.manifest = string(b)
}

var hashCache sync.Map

func (s *server) fileHashFS(cfs fs.FS, name string) string {
	if !s.opt.Dev {
		if v, ok := hashCache.Load(name); ok {
			return v.(string)
		}
	}
	b, err := fs.ReadFile(cfs, name)
	if err != nil {
		return "0"
	}
	sum := sha256.Sum256(b)
	h := hex.EncodeToString(sum[:])[:12]
	hashCache.Store(name, h)
	return h
}

// ---------- 页面 ----------

func (s *server) page(w http.ResponseWriter, r *http.Request) {
	locale, path := r.URL.Path, r.URL.Path
	if i18nCfg != nil {
		lang := ""
		if ck, err := r.Cookie("lang"); err == nil {
			lang = ck.Value
		}
		locale, path = resolveLocale(r.URL.Path, lang, r.Header.Get("Accept-Language"))
		r = r.WithContext(context.WithValue(r.Context(), ctxLocale, locale))
		r = r.WithContext(context.WithValue(r.Context(), ctxPath, path))
	}
	var route *Route
	var params map[string]string
	for i := range s.opt.Routes {
		if p, ok := s.opt.Routes[i].Match(path); ok {
			route, params = &s.opt.Routes[i], p
			break
		}
	}
	props := s.propsFor(r, params)
	if i18nCfg != nil {
		props.Locale = locale
		props.Path = path
	}
	sess := s.attachSession(&props, r)
	if s.opt.Before != nil {
		s.opt.Before(w, r, props.Cookies)
	}
	if route == nil {
		s.serve404(w, r, props)
		return
	}
	t := time.Now()
	var seq uint32
	html, pending, err := s.renderDoc(r, &seq, func() Node { return route.Render(props) })
	us := time.Since(t).Microseconds()
	if err != nil {
		var he *HostError
		if errors.As(err, &he) && errors.Is(he.Err, ErrNotFound) {
			s.serve404(w, r, props)
			return
		}
		var re *RedirectError
		if errors.As(err, &re) {
			http.Redirect(w, r, re.URL, re.Status)
			return
		}
		id, _ := r.Context().Value(ctxReqID).(string)
		log.Printf("render failed %s id=%s: %v", r.URL.Path, id, err)
		s.serveError(w, r, http.StatusInternalServerError, err)
		return
	}
	w.Header().Set("Server-Timing", "go;dur="+strconv.FormatFloat(float64(us)/1000, 'f', 3, 64))
	sess.save(w, r) // consumed flashes / a CSRF token generated during render are written back before the body (no Set-Cookie when unchanged)
	s.writeHTML(w, http.StatusOK, html)
	if len(pending) > 0 {
		s.streamPending(w, r, pending, &seq)
	}
}

// renderDoc: 渲染一整页(doctype + 引导脚本注入), recover 渲染期间的 panic 成 error
func (s *server) renderDoc(r *http.Request, seq *uint32, fn func() Node) (html string, pending []*Pending, err error) {
	defer func() {
		if rec := recover(); rec != nil {
			switch e := rec.(type) {
			case *HostError:
				err = e
			case *RedirectError:
				err = e
			case *ThrowError:
				err = e
			case error:
				err = fmt.Errorf("panic: %v\n%s", e, debug.Stack())
			default:
				err = fmt.Errorf("panic: %v\n%s", rec, debug.Stack())
			}
		}
	}()
	html, pending = RenderDoc(fn(), seq, s.injectFor(r))
	return html, pending, nil
}

// writeHTML: 页面响应头 + 正文(html 已含 doctype 与注入)
func (s *server) writeHTML(w http.ResponseWriter, status int, html string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, ok := w.Header()["Cache-Control"]; !ok {
		w.Header().Set("Cache-Control", "no-store")
	}
	w.WriteHeader(status)
	io.WriteString(w, html)
}

// streamPending: 外壳已经写出; 每个 Suspense 边界在自己的 goroutine 里渲染, 完成一个就追加一个
// <template data-gotsx-fill="id">…</template><script>__gotsxFill("id")</script> 并 flush。嵌套边界递归处理。
// 边界内的错误不能再变成 500(头已发出): Dev 下把错误文本填进去, 生产下留空并记日志。
func (s *server) streamPending(w http.ResponseWriter, r *http.Request, pending []*Pending, seq *uint32) {
	fl, _ := w.(http.Flusher)
	flush := func() {
		if fl != nil {
			fl.Flush()
		}
	}
	flush()
	nonce, _ := r.Context().Value(ctxNonce).(string)
	na := ""
	if nonce != "" {
		na = ` nonce="` + nonce + `"`
	}
	type fill struct{ id, html string }
	ch := make(chan fill)
	var wg sync.WaitGroup
	var resolve func(p *Pending)
	resolve = func(p *Pending) {
		defer wg.Done()
		var out string
		var nested []*Pending
		err := func() (err error) {
			defer func() {
				if rec := recover(); rec != nil {
					err = fmt.Errorf("%v", rec)
				}
			}()
			out, nested = RenderPending(p.Fn(), seq)
			return nil
		}()
		if err != nil {
			id, _ := r.Context().Value(ctxReqID).(string)
			log.Printf("suspense %s failed %s id=%s: %v", p.ID, r.URL.Path, id, err)
			out = ""
			if s.opt.Dev {
				out = `<pre class="gotsx-error" style="color:#dc2626;white-space:pre-wrap">` + htmlEscape(err.Error()) + `</pre>`
			}
		}
		for _, n := range nested {
			wg.Add(1)
			go resolve(n)
		}
		ch <- fill{p.ID, out}
	}
	for _, p := range pending {
		wg.Add(1)
		go resolve(p)
	}
	go func() {
		wg.Wait()
		close(ch)
	}()
	for f := range ch {
		if r.Context().Err() != nil {
			continue // 客户端已断开: 把剩下的 goroutine 排空
		}
		fmt.Fprintf(w, `<template data-gotsx-fill="%s">%s</template><script%s>__gotsxFill("%s")</script>`, f.id, f.html, na, f.id)
		flush()
	}
}

// attachSession: session values, the lazy CSRF token and the (consumed) flash messages for a page render.
// The caller saves the session before writing the response.
func (s *server) attachSession(props *PageProps, r *http.Request) *Session {
	sess := s.loadSession(r)
	props.Session = sess.Values()
	props.csrf = func() string { return sess.CSRF() } // generated lazily; saved once after rendering
	if len(sess.data.Flash) > 0 {
		props.Flash = sess.data.Flash
		sess.data.Flash = nil
		sess.modified = true
	}
	return sess
}

func (s *server) propsFor(r *http.Request, params map[string]string) PageProps {
	if params == nil {
		params = map[string]string{}
	}
	query := map[string]string{}
	for k, v := range r.URL.Query() {
		query[k] = v[0]
	}
	cookies := map[string]string{}
	for _, ck := range r.Cookies() {
		cookies[ck.Name] = ck.Value
	}
	return PageProps{Params: params, Query: query, Path: r.URL.Path, Cookies: cookies}
}

// writeDoc: 给已渲染好的 html(404 / 错误页)注入 doctype + 引导脚本后写出
func (s *server) writeDoc(w http.ResponseWriter, r *http.Request, status int, html string) {
	inject := s.injectFor(r)
	var out string
	switch {
	case strings.Contains(html, "</head>"):
		out = "<!DOCTYPE html>" + strings.Replace(html, "</head>", inject+"</head>", 1)
	case strings.Contains(html, "</body>"):
		out = "<!DOCTYPE html>" + strings.Replace(html, "</body>", inject+"</body>", 1)
	default:
		out = "<!DOCTYPE html>" + html + inject
	}
	s.writeHTML(w, status, out)
}

// injectFor: 这次响应要塞进文档头部的东西 —— hreflang、window.__GOTSX(manifest / i18n 目录 / dev 标记)、填充函数、loader
func (s *server) injectFor(r *http.Request) string {
	nonce, _ := r.Context().Value(ctxNonce).(string)
	na := ""
	if nonce != "" {
		na = ` nonce="` + nonce + `"`
	}
	gv := s.manifest
	if s.logURL != "" {
		gv = strings.TrimSuffix(gv, "}") + `,"log":"` + s.logURL + `"}`
	}
	if s.opt.Dev {
		gv = strings.TrimSuffix(gv, "}") + `,"dev":true}`
	}
	head := ""
	if i18nCfg != nil {
		locale, _ := r.Context().Value(ctxLocale).(string)
		if locale == "" {
			locale = i18nCfg.Default
		}
		path, _ := r.Context().Value(ctxPath).(string)
		i18nObj, _ := json.Marshal(map[string]any{
			"locale": locale, "default": i18nCfg.Default, "prefix": i18nCfg.Prefix,
			"currency": i18nCfg.Currency, "messages": i18nCfg.Messages[locale],
		})
		gv = strings.TrimSuffix(gv, "}") + `,"i18n":` + string(i18nObj) + `}`
		// hreflang 备用链接(每语言一条 + x-default)
		scheme := "http"
		if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}
		var hb strings.Builder
		for _, loc := range i18nCfg.Locales {
			href := scheme + "://" + r.Host + LPath(loc, path)
			hb.WriteString(`<link rel="alternate" hreflang="` + loc + `" href="` + href + `">`)
		}
		hb.WriteString(`<link rel="alternate" hreflang="x-default" href="` + scheme + "://" + r.Host + LPath(i18nCfg.Default, path) + `">`)
		head = hb.String()
	}
	return head + `<script` + na + `>window.__GOTSX=` + gv + `;` + fillScript + `</script><script type="module" src="` + s.loader + `"></script>`
}

func (s *server) serve404(w http.ResponseWriter, r *http.Request, props PageProps) {
	if s.opt.NotFound != nil {
		sess := s.attachSession(&props, r)
		if html, err := s.render(func() Node { return s.opt.NotFound(props) }); err == nil {
			sess.save(w, r) // a consumed flash / a token generated by the 404 page must reach the cookie too
			s.writeDoc(w, r, http.StatusNotFound, html)
			return
		}
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	io.WriteString(w, "<!DOCTYPE html><meta charset=utf-8><title>404</title><h1>404 · Not Found</h1>")
}

func (s *server) serveError(w http.ResponseWriter, r *http.Request, status int, cause error) {
	// 已经写过响应头就没法补救, 直接返回
	shown := cause
	if !s.opt.Dev {
		shown = errors.New("Internal Server Error")
	}
	if s.opt.ErrorPage != nil {
		props := s.propsFor(r, nil)
		sess := s.loadSession(r) // error pages see the session (for the header chrome) but do not consume flashes
		props.Session = sess.Values()
		props.csrf = func() string { return sess.CSRF() }
		if html, err := s.render(func() Node { return s.opt.ErrorPage(props, shown) }); err == nil {
			sess.save(w, r)
			s.writeDoc(w, r, status, html)
			return
		}
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	msg := "Internal Server Error"
	if s.opt.Dev {
		msg = cause.Error()
	}
	fmt.Fprintf(w, "<!DOCTYPE html><meta charset=utf-8><title>%d</title><h1>%d</h1><pre>%s</pre>", status, status, htmlEscape(msg))
}

func (s *server) render(fn func() Node) (out string, err error) {
	defer func() {
		if rec := recover(); rec != nil {
			switch e := rec.(type) {
			case *HostError:
				err = e
			case *RedirectError:
				err = e
			case *ThrowError:
				err = e
			case error:
				err = fmt.Errorf("panic: %v\n%s", e, debug.Stack())
			default:
				err = fmt.Errorf("panic: %v\n%s", rec, debug.Stack())
			}
		}
	}()
	return Render(fn()), nil
}

// FindDir 从可执行文件目录/工作目录往上找一个存在的相对路径
func FindDir(rel string) string {
	for _, base := range []string{".", filepath.Dir(os.Args[0])} {
		d := base
		for i := 0; i < 5; i++ {
			p := filepath.Join(d, rel)
			if st, err := os.Stat(p); err == nil && st.IsDir() {
				return p
			}
			d = filepath.Join(d, "..")
		}
	}
	return rel
}
