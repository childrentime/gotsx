// gotsx admin —— 企业后台管理示例: 认证 / 受保护路由 / 用户表格 CRUD / 服务端校验 / 模态框 / toast。
package main

import (
	"embed"
	"encoding/json"
	"flag"
	"io/fs"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/childrentime/gotsx/admin/gen"
	"github.com/childrentime/gotsx/admin/host"
	gotsx "github.com/childrentime/gotsx/runtime"
)

//go:embed public
var publicEmbed embed.FS

func sid(r *http.Request) string {
	if c, err := r.Cookie("sid"); err == nil {
		return c.Value
	}
	return ""
}

func body(r *http.Request) map[string]string {
	m := map[string]string{}
	json.NewDecoder(r.Body).Decode(&m)
	return m
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func atoi(s string, d int) int {
	if n, err := strconv.Atoi(s); err == nil {
		return n
	}
	return d
}

// 鉴权中间件: 未登录访问受保护路径 → 页面重定向 /login, 接口 → 401
func authMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		public := p == "/login" || p == "/favicon.ico" || p == "/healthz" || p == "/readyz" ||
			strings.HasPrefix(p, "/public/") || strings.HasPrefix(p, "/_gotsx/") || p == "/auth/login"
		if !public && !host.Auth.IsAuthed(sid(r)) {
			if strings.HasPrefix(p, "/api/") || strings.HasPrefix(p, "/users/") || strings.HasPrefix(p, "/auth/") {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		// 已登录还访问 /login → 回首页
		if p == "/login" && host.Auth.IsAuthed(sid(r)) {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	addr := flag.String("addr", ":3000", "监听地址")
	dev := flag.Bool("dev", false, "开发模式")
	flag.Parse()

	panel := func(icon, msg string) gotsx.Node {
		return gen.Shell(gen.ShellProps{Title: msg, Children: gotsx.El("div",
			[]gotsx.Attr{gotsx.A("class", "flex flex-col items-center justify-center py-32 text-center")},
			gotsx.El("div", []gotsx.Attr{gotsx.A("class", "text-6xl")}, gotsx.Text(icon)),
			gotsx.El("p", []gotsx.Attr{gotsx.A("class", "mt-4 text-slate-500")}, gotsx.Text(msg)),
			gotsx.El("a", []gotsx.Attr{gotsx.A("href", "/"), gotsx.A("class", "mt-5 rounded-lg bg-brand-500 px-6 py-2.5 text-sm font-semibold text-white")}, gotsx.Text("回到仪表盘")))})
	}

	err := gotsx.Serve(gotsx.Options{
		Addr:       *addr,
		Dev:        *dev,
		Routes:     gen.Routes,
		ClientDir:  gotsx.FindDir("gen/client"),
		ClientFS:   gen.ClientFS,
		PublicDir:  gotsx.FindDir("public"),
		PublicFS:   mustSub(publicEmbed, "public"),
		Middleware: []func(http.Handler) http.Handler{authMW},
		Before: func(w http.ResponseWriter, r *http.Request, cookies map[string]string) {
			s := host.Auth.Current(sid(r))
			cookies["_user"] = s.User
			cookies["_name"] = s.Name
			cookies["_role"] = s.Role
		},
		NotFound: func(p gotsx.PageProps) gotsx.Node { return panel("🔍", "找不到这个页面") },
		ErrorPage: func(p gotsx.PageProps, err error) gotsx.Node {
			return panel("😵", "服务器开小差了, 请稍后再试")
		},
		Actions: map[string]http.HandlerFunc{
			"POST /auth/login": func(w http.ResponseWriter, r *http.Request) {
				b := body(r)
				s, err := host.Auth.Login(b["user"], b["pass"])
				if err != nil {
					writeJSON(w, map[string]any{"ok": false, "error": err.Error()})
					return
				}
				http.SetCookie(w, &http.Cookie{Name: "sid", Value: s, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode, MaxAge: 86400})
				writeJSON(w, map[string]any{"ok": true})
			},
			"POST /auth/logout": func(w http.ResponseWriter, r *http.Request) {
				host.Auth.Logout(sid(r))
				http.SetCookie(w, &http.Cookie{Name: "sid", Value: "", Path: "/", MaxAge: -1})
				writeJSON(w, map[string]any{"ok": true})
			},
			"GET /api/users": func(w http.ResponseWriter, r *http.Request) {
				writeJSON(w, host.Users.All())
			},
			"POST /users/create": func(w http.ResponseWriter, r *http.Request) {
				b := body(r)
				u, errs := host.Users.Create(b["name"], b["email"], b["role"], b["dept"])
				if len(errs) > 0 {
					writeJSON(w, map[string]any{"ok": false, "errors": errs})
					return
				}
				writeJSON(w, map[string]any{"ok": true, "user": u})
			},
			"POST /users/update": func(w http.ResponseWriter, r *http.Request) {
				b := body(r)
				u, errs := host.Users.Update(b["id"], b["name"], b["email"], b["role"], b["dept"], b["status"])
				if len(errs) > 0 {
					writeJSON(w, map[string]any{"ok": false, "errors": errs})
					return
				}
				writeJSON(w, map[string]any{"ok": true, "user": u})
			},
			"POST /users/delete": func(w http.ResponseWriter, r *http.Request) {
				writeJSON(w, map[string]any{"ok": host.Users.Delete(body(r)["id"])})
			},
		},
	})
	log.Fatal(err)
}

func mustSub(f embed.FS, dir string) fs.FS {
	sub, err := fs.Sub(f, dir)
	if err != nil {
		panic(err)
	}
	return sub
}
