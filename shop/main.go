// gomu: Temu 风格的全栈电商示例。页面 = 方言编译成的 Go; 写操作 = 普通 Go action; 会话 = sid cookie。
package main

import (
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	gotsx "github.com/childrentime/gotsx/runtime"
	"github.com/childrentime/gotsx/shop/gen"
	"github.com/childrentime/gotsx/shop/host"
)

func sidOf(w http.ResponseWriter, r *http.Request) string {
	if ck, err := r.Cookie("sid"); err == nil && ck.Value != "" {
		return ck.Value
	}
	sid := host.NewSID()
	http.SetCookie(w, &http.Cookie{Name: "sid", Value: sid, Path: "/", MaxAge: 86400 * 30, HttpOnly: true})
	return sid
}

type body struct {
	ID      string `json:"id"`
	Variant string `json:"variant"`
	Qty     int    `json:"qty"`
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Address string `json:"address"`
}

func parse(r *http.Request) body {
	var b body
	json.NewDecoder(r.Body).Decode(&b)
	return b
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

var phoneRe = regexp.MustCompile(`^1\d{10}$`)

func htmlAttr(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	return s
}

func atoiDefault(s string, d int) int {
	n := d
	if s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			n = v
		}
	}
	return n
}

//go:embed public
var publicEmbed embed.FS

func main() {
	addr := flag.String("addr", ":3000", "监听地址")
	dev := flag.Bool("dev", false, "开发模式(错误页带堆栈, 资源 no-cache)")
	flag.Parse()
	panel := func(icon, msg, sid string) gotsx.Node {
		return gen.Layout(gen.LayoutProps{Title: msg, Sid: sid, Wide: true, Children: gotsx.El("div",
			[]gotsx.Attr{gotsx.A("class", "rounded-xl2 border border-ink-100 bg-white py-24 text-center")},
			gotsx.El("div", []gotsx.Attr{gotsx.A("class", "text-7xl")}, gotsx.Text(icon)),
			gotsx.El("p", []gotsx.Attr{gotsx.A("class", "mt-4 text-ink-500")}, gotsx.Text(msg)),
			gotsx.El("a", []gotsx.Attr{gotsx.A("href", "/"), gotsx.A("class", "mt-5 inline-block rounded-full bg-brand-500 px-8 py-2.5 text-sm font-bold text-white")}, gotsx.Text("回首页")))})
	}
	i18n := &gotsx.I18n{
		Locales: []string{"zh", "en"}, Default: "zh", Prefix: true,
		Currency: map[string]string{"zh": "¥", "en": "$"},
		Messages: map[string]map[string]string{
			"zh": {
				"nav.home": "首页", "nav.cart": "购物车", "nav.orders": "订单",
				"search.placeholder": "搜索 20 万+ 好物…", "search.button": "搜索",
				"bar.freeship": "全场满 ¥69 包邮", "bar.return": "7 天无理由退换", "bar.priceGuard": "90 天价保", "bar.fastship": "24 小时发货",
				"footer.tagline": "全球好物, 工厂直发。用 gotsx 方言编译到 Go 的全栈电商示例。",
				"footer.guide":   "购物指南", "footer.about": "关于 gomu", "footer.promise": "保障承诺",
				"lang.other": "English", "cart.count": "{n} 件商品",
			},
			"en": {
				"nav.home": "Home", "nav.cart": "Cart", "nav.orders": "Orders",
				"search.placeholder": "Search 200k+ products…", "search.button": "Search",
				"bar.freeship": "Free shipping over ¥69", "bar.return": "7-day returns", "bar.priceGuard": "90-day price guarantee", "bar.fastship": "Ships in 24h",
				"footer.tagline": "Global goods, factory-direct. A full-stack e-commerce demo compiled to Go with the gotsx dialect.",
				"footer.guide":   "Shopping Guide", "footer.about": "About gomu", "footer.promise": "Our Promise",
				"lang.other": "中文", "cart.count": "{n} item|{n} items",
			},
		},
	}
	err := gotsx.Serve(gotsx.Options{
		Addr:      *addr,
		Dev:       *dev,
		I18n:      i18n,
		Routes:    gen.Routes,
		ClientDir: gotsx.FindDir("gen/client"),
		ClientFS:  gen.ClientFS,
		PublicFS:  mustSub(publicEmbed, "public"),
		PublicDir: gotsx.FindDir("public"),
		NotFound:  func(p gotsx.PageProps) gotsx.Node { return panel("🔍", "找不到这个页面", p.Cookies["sid"]) },
		ErrorPage: func(p gotsx.PageProps, err error) gotsx.Node {
			return panel("😵", "服务器开小差了, 请稍后再试", p.Cookies["sid"])
		},
		OnClientEvent: func(ev gotsx.ClientEvent, r *http.Request) {
			if ev.Type == "pageview" {
				log.Printf("[telemetry] pageview %s ref=%q", ev.URL, ev.Ref)
			} else {
				log.Printf("[telemetry] %s: %s @ %s", ev.Type, ev.Message, ev.URL)
			}
		},
		Before: func(w http.ResponseWriter, r *http.Request, cookies map[string]string) {
			if cookies["sid"] == "" {
				cookies["sid"] = sidOf(w, r)
			}
		},
		Actions: map[string]http.HandlerFunc{
			"GET /manifest.webmanifest": func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/manifest+json")
				w.Write([]byte(`{"name":"gomu 好物商城","short_name":"gomu","start_url":"/","scope":"/","display":"standalone","background_color":"#ffffff","theme_color":"#18181b","description":"全球好物, 工厂直发","icons":[{"src":"/icon.svg","sizes":"any","type":"image/svg+xml","purpose":"any maskable"}]}`))
			},
			"GET /icon.svg": func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "image/svg+xml")
				w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
				w.Write([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 512 512"><rect width="512" height="512" rx="112" fill="#18181b"/><text x="256" y="300" font-size="300" font-weight="900" text-anchor="middle" fill="#fff" font-family="system-ui,sans-serif">g</text></svg>`))
			},
			"GET /img/p/{id}": func(w http.ResponseWriter, r *http.Request) {
				p, err := host.Catalog.Get(r.PathValue("id"))
				if err != nil {
					http.NotFound(w, r)
					return
				}
				glyph := p.Emoji
				if g := r.URL.Query().Get("g"); g != "" {
					if i, e := strconv.Atoi(g); e == nil && i >= 0 && i < len(p.Gallery) {
						glyph = p.Gallery[i]
					}
				}
				w.Header().Set("Content-Type", "image/svg+xml; charset=utf-8")
				w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
				svg := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 400 400" width="400" height="400" role="img" aria-label="%s">`+
					`<defs><filter id="s" x="-30%%" y="-30%%" width="160%%" height="160%%"><feDropShadow dx="0" dy="10" stdDeviation="9" flood-color="#09090b" flood-opacity="0.14"/></filter></defs>`+
					`<rect width="400" height="400" fill="#f4f4f5"/>`+
					`<ellipse cx="200" cy="318" rx="104" ry="14" fill="#09090b" opacity="0.06"/>`+
					`<text x="200" y="212" font-size="180" text-anchor="middle" dominant-baseline="central" filter="url(#s)">%s</text>`+
					`</svg>`, htmlAttr(p.Title), glyph)
				w.Write([]byte(svg))
			},
			"GET /robots.txt": func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/plain; charset=utf-8")
				w.Write([]byte("User-agent: *\nAllow: /\nDisallow: /cart\nDisallow: /checkout\nDisallow: /orders\n\nSitemap: " + host.Site.BaseUrl() + "/sitemap.xml\n"))
			},
			"GET /sitemap.xml": func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/xml; charset=utf-8")
				var b strings.Builder
				b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
				b.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n")
				for _, u := range host.Site.SitemapURLs() {
					b.WriteString("  <url><loc>" + u + "</loc></url>\n")
				}
				b.WriteString("</urlset>\n")
				w.Write([]byte(b.String()))
			},
			"GET /api/feed": func(w http.ResponseWriter, r *http.Request) {
				writeJSON(w, host.Catalog.Feed(atoiDefault(r.URL.Query().Get("page"), 1)))
			},
			"GET /api/related": func(w http.ResponseWriter, r *http.Request) {
				writeJSON(w, map[string]any{"cards": host.Catalog.Related(r.URL.Query().Get("id"))})
			},
			"POST /actions/cart/add": func(w http.ResponseWriter, r *http.Request) {
				sid, b := sidOf(w, r), parse(r)
				cv, err := host.Cart.Add(sid, b.ID, b.Variant, b.Qty)
				if err != nil {
					writeJSON(w, map[string]any{"ok": false, "error": err.Error()})
					return
				}
				writeJSON(w, map[string]any{"ok": true, "cart": cv})
			},
			"POST /actions/cart/set": func(w http.ResponseWriter, r *http.Request) {
				sid, b := sidOf(w, r), parse(r)
				writeJSON(w, map[string]any{"ok": true, "cart": host.Cart.SetQty(sid, b.ID, b.Variant, b.Qty)})
			},
			"POST /actions/wish": func(w http.ResponseWriter, r *http.Request) {
				sid, b := sidOf(w, r), parse(r)
				writeJSON(w, map[string]any{"ok": true, "wished": host.Wish.Toggle(sid, b.ID)})
			},
			"POST /actions/checkout": func(w http.ResponseWriter, r *http.Request) {
				sid, b := sidOf(w, r), parse(r)
				errs := map[string]string{}
				if len(strings.TrimSpace(b.Name)) < 2 {
					errs["name"] = "收货人至少 2 个字"
				}
				if !phoneRe.MatchString(strings.TrimSpace(b.Phone)) {
					errs["phone"] = "手机号格式不对(11 位, 1 开头)"
				}
				if len(strings.TrimSpace(b.Address)) < 5 {
					errs["address"] = "地址太短了"
				}
				if len(errs) > 0 {
					writeJSON(w, map[string]any{"ok": false, "errors": errs})
					return
				}
				o, err := host.Orders.Place(sid, strings.TrimSpace(b.Name), strings.TrimSpace(b.Phone), strings.TrimSpace(b.Address))
				if err != nil {
					writeJSON(w, map[string]any{"ok": false, "errors": map[string]string{"_": err.Error()}})
					return
				}
				writeJSON(w, map[string]any{"ok": true, "orderId": o.ID})
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
