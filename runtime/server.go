package gotsx

import (
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

func (r Route) Match(path string) (map[string]string, bool) {
	p := strings.Trim(path, "/")
	var segs []string
	if p != "" {
		segs = strings.Split(p, "/")
	}
	if len(segs) != len(r.Segs) {
		return nil, false
	}
	params := map[string]string{}
	for i, s := range r.Segs {
		if strings.HasPrefix(s, "{") {
			params[s[1:len(s)-1]] = segs[i]
		} else if s != segs[i] {
			return nil, false
		}
	}
	return params, true
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
	opt      Options
	manifest string
	loader   string
	logURL   string
}

type ctxKey int

const (
	ctxReqID ctxKey = iota
	ctxNonce
	ctxLocale
	ctxPath
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

func Serve(opt Options) error {
	s := &server{opt: opt}
	configureI18n(opt.I18n)
	s.scanClient()
	routes := append([]Route(nil), opt.Routes...)
	sort.SliceStable(routes, func(i, j int) bool {
		return strings.Count(routes[i].Pattern, "{") < strings.Count(routes[j].Pattern, "{")
	})
	s.opt.Routes = routes

	mux := http.NewServeMux()
	mux.Handle("GET /_gotsx/", s.assetCache(http.StripPrefix("/_gotsx/", http.FileServer(http.FS(s.clientFS())))))
	if pf := s.publicFS(); pf != nil {
		mux.Handle("GET /public/", s.assetCache(http.StripPrefix("/public/", http.FileServer(http.FS(pf)))))
	}
	for pattern, h := range opt.Actions {
		mux.HandleFunc(pattern, s.csrfGuard(pattern, h))
	}
	if opt.OnClientEvent != nil {
		mux.HandleFunc("POST /_gotsx/client-log", func(w http.ResponseWriter, r *http.Request) {
			if !SameOrigin(r) { // 仅同源上报
				w.WriteHeader(http.StatusNoContent)
				return
			}
			var ev ClientEvent
			json.NewDecoder(io.LimitReader(r.Body, 16*1024)).Decode(&ev)
			s.opt.OnClientEvent(ev, r)
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
	h = accessLogMW(h)
	h = requestIDMW(h)

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
		log.Printf("gotsx: 收到关闭信号, 优雅退出中…")
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
	log.Printf("gotsx: 监听 http://localhost%s  路由 %d 条  模式=%s", opt.Addr, len(routes), mode)
	err := srv.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		<-idle
		return nil
	}
	return err
}

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
			log.Printf("CSRF 拒绝 %s %s id=%s origin=%q referer=%q", r.Method, r.URL.Path, id, r.Header.Get("Origin"), r.Header.Get("Referer"))
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
	if s.opt.Before != nil {
		s.opt.Before(w, r, props.Cookies)
	}
	if route == nil {
		s.serve404(w, r, props)
		return
	}
	t := time.Now()
	html, err := s.render(func() Node { return route.Render(props) })
	us := time.Since(t).Microseconds()
	if err != nil {
		var he *HostError
		if errors.As(err, &he) && errors.Is(he.Err, ErrNotFound) {
			s.serve404(w, r, props)
			return
		}
		id, _ := r.Context().Value(ctxReqID).(string)
		log.Printf("渲染失败 %s id=%s: %v", r.URL.Path, id, err)
		s.serveError(w, r, http.StatusInternalServerError, err)
		return
	}
	w.Header().Set("Server-Timing", fmt.Sprintf("go;dur=%.3f", float64(us)/1000))
	s.writeDoc(w, r, http.StatusOK, html)
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

// writeDoc: 注入 doctype + manifest + 带 nonce 的内联 script + loader, 写出
func (s *server) writeDoc(w http.ResponseWriter, r *http.Request, status int, html string) {
	nonce, _ := r.Context().Value(ctxNonce).(string)
	na := ""
	if nonce != "" {
		na = ` nonce="` + nonce + `"`
	}
	gv := s.manifest
	if s.logURL != "" {
		gv = strings.TrimSuffix(gv, "}") + `,"log":"` + s.logURL + `"}`
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
	inject := head + fmt.Sprintf(`<script%s>window.__GOTSX=%s</script><script type="module" src="%s"></script>`, na, gv, s.loader)
	out := "<!DOCTYPE html>" + strings.Replace(html, "</head>", inject+"</head>", 1)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, ok := w.Header()["Cache-Control"]; !ok {
		w.Header().Set("Cache-Control", "no-store")
	}
	w.WriteHeader(status)
	io.WriteString(w, out)
}

func (s *server) serve404(w http.ResponseWriter, r *http.Request, props PageProps) {
	if s.opt.NotFound != nil {
		if html, err := s.render(func() Node { return s.opt.NotFound(props) }); err == nil {
			s.writeDoc(w, r, http.StatusNotFound, html)
			return
		}
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	io.WriteString(w, "<!DOCTYPE html><meta charset=utf-8><title>404</title><h1>404 · 页面不存在</h1>")
}

func (s *server) serveError(w http.ResponseWriter, r *http.Request, status int, cause error) {
	// 已经写过响应头就没法补救, 直接返回
	shown := cause
	if !s.opt.Dev {
		shown = errors.New("服务器内部错误")
	}
	if s.opt.ErrorPage != nil {
		props := s.propsFor(r, nil)
		if html, err := s.render(func() Node { return s.opt.ErrorPage(props, shown) }); err == nil {
			s.writeDoc(w, r, status, html)
			return
		}
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	msg := "服务器内部错误"
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
