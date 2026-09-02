// Package server builds the shop's gotsx.Options (routes, i18n, actions, error pages) so that main.go,
// tests and other hosts (the Cloudflare Worker in deploy/cloudflare) share one definition. Writes are Go actions;
// sessions are a sid cookie.
package server

import (
	"encoding/json"
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

var phoneRe = regexp.MustCompile(`^\+?[0-9]{7,15}$`) // international: optional +, 7-15 digits (spaces / dashes stripped)

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

// messages: the i18n catalog. English is the primary language; zh is the secondary locale. Keys are grouped by area
// (nav / bar / footer / home / flash / feed / product / listing / sort / card / tag / add / cart / checkout / form /
// orders / order / status / error). Plural forms use "one|other"; {name} placeholders are interpolated with tv().
var messages = map[string]map[string]string{
	"en": {
		"nav.home": "Home", "nav.cart": "Cart", "nav.orders": "Orders",
		"search.placeholder": "Search 200k+ products…", "search.button": "Search", "search.results": "Results for “{q}”",
		"bar.freeship": "Free shipping over $69", "bar.return": "7-day returns", "bar.priceGuard": "90-day price guarantee", "bar.fastship": "Ships in 24h",
		"footer.tagline": "Global goods, factory-direct. A full-stack e-commerce demo compiled to Go with the gotsx dialect.",
		"footer.guide":   "Shopping guide", "footer.about": "About gomu", "footer.promise": "Our promise",
		"lang.other": "中文", "cart.count": "{n} item|{n} items",
		"meta.description": "gomu — global goods, factory-direct. 200k+ curated products, free shipping over $69, 7-day free returns.",
		"home.title":       "Global goods, factory-direct", "home.badge": "$15 off your first order",
		"home.heading1": "Shop straight from the factory,", "home.heading2": "skip the markup",
		"home.sub": "200k+ curated products · free shipping over $69 · 7-day returns · 90-day price guarantee",
		"home.cta": "Start shopping", "home.deals": "Today's deals", "home.coupons": "$5 off $29 · $20 off $99",
		"perk.ship.t": "Free shipping over $69", "perk.ship.d": "Most regions", "perk.fast.t": "Ships in 24h", "perk.fast.d": "Direct from the factory",
		"perk.return.t": "7-day returns", "perk.return.d": "No questions asked", "perk.price.t": "90-day price guarantee", "perk.price.d": "We refund the difference",
		"flash.title": "Flash sale", "flash.badge": "Up to 50% off", "flash.ends": "Ends in", "flash.claimed": "{n}% claimed",
		"feed.title": "For you", "feed.sub": "Picked from what you browsed", "feed.end": "That's everything", "feed.more": "Load more", "feed.loading": "Loading…",
		"product.reviews": "{n} reviews", "product.sold": "Sold {n}", "product.save": "Save {n}%", "product.flash": "Flash sale",
		"product.details": "Product details", "product.reviewsTitle": "Customer reviews", "product.reviewsCount": "{n} reviews", "product.stars": "{n} stars",
		"product.related": "Related products",
		"product.note":    "This page is rendered by Go on the server (the data API includes simulated latency); the add-to-cart panel, gallery and related products are islands compiled to signals. Product images are emoji studio shots for the demo.",
		"product.meta":    "{desc} Now {price}, {sold} sold, {reviews} reviews, free shipping over $69.",
		"feature.1":       "Curated materials", "feature.2": "Factory-direct", "feature.3": "Strict quality checks", "feature.4": "Eco packaging", "feature.5": "Express delivery available", "feature.6": "Hassle-free returns",
		"pperk.ship": "Free shipping over $69", "pperk.return": "7-day free returns", "pperk.auth": "Authenticity guaranteed", "pperk.fast": "Ships within 24 hours",
		"listing.count": "{n} product|{n} products", "listing.empty": "No products found", "listing.emptySub": "Try another keyword, or browse a category", "listing.back": "Back to home",
		"sort.rec": "Featured", "sort.sales": "Best sellers", "sort.price": "Price ↑", "sort.priceDesc": "Price ↓",
		"card.soldout": "Sold out", "card.sold": "Sold {n}", "card.flash": "Flash",
		"tag.hot": "Bestseller", "tag.deal": "Under $15", "tag.rated": "Top rated",
		"add.pick": "Please choose all options first", "add.added": "Added to cart", "add.qty": "Quantity", "add.stock": "{n} in stock",
		"add.soldout": "Sold out", "add.adding": "Adding…", "add.button": "Add to cart", "add.decrease": "Decrease", "add.increase": "Increase",
		"wish.label": "Add to wishlist", "gallery.alt": "Product image", "gallery.thumb": "Thumbnail",
		"cart.title": "Cart", "cart.empty": "Your cart is empty", "cart.browse": "Find something nice", "cart.summary": "Order summary",
		"cart.subtotal": "Subtotal ({n} item)|Subtotal ({n} items)", "cart.shipping": "Shipping", "cart.freeGap": "Add {n} more for free shipping",
		"cart.total": "Total", "cart.checkout": "Checkout ({n})", "cart.secure": "Secure checkout · 7-day free returns", "cart.remove": "Remove",
		"checkout.title": "Checkout", "checkout.empty": "Your cart is empty", "checkout.browse": "Start shopping", "checkout.shipping": "Shipping details",
		"checkout.items": "Order items", "checkout.count": "{n} item|{n} items", "checkout.subtotal": "Subtotal", "checkout.shippingFee": "Shipping", "checkout.due": "Total due",
		"form.name": "Full name", "form.namePh": "Your name", "form.phone": "Phone", "form.phonePh": "e.g. +1 415 555 0132",
		"form.address": "Address", "form.addressPh": "Street, city, postal code", "form.submit": "Place order · {total}", "form.submitting": "Placing order…",
		"form.note":    "By placing the order you agree to the Terms of Service · demo environment, nothing is charged",
		"orders.title": "My orders", "orders.empty": "No orders yet", "orders.first": "Place your first order", "orders.count": "{n} item|{n} items", "orders.total": "Total",
		"order.title": "Order {id}", "order.success": "Order placed", "order.number": "Order", "order.viewAll": "View all orders", "order.items": "Items",
		"order.shipping": "Shipping details", "order.paid": "Amount paid", "order.continue": "Continue shopping",
		"status.paid":    "Paid, awaiting shipment",
		"category.meta":  "{label} · curated products, {n} items, factory-direct, free shipping over $69.",
		"error.notFound": "Page not found", "error.server": "Something went wrong, please try again later", "error.home": "Back to home",
	},
	"zh": {
		"nav.home": "首页", "nav.cart": "购物车", "nav.orders": "订单",
		"search.placeholder": "搜索 20 万+ 好物…", "search.button": "搜索", "search.results": "“{q}” 的搜索结果",
		"bar.freeship": "全场满 $69 包邮", "bar.return": "7 天无理由退换", "bar.priceGuard": "90 天价保", "bar.fastship": "24 小时发货",
		"footer.tagline": "全球好物, 工厂直发。用 gotsx 方言编译到 Go 的全栈电商示例。",
		"footer.guide":   "购物指南", "footer.about": "关于 gomu", "footer.promise": "保障承诺",
		"lang.other": "English", "cart.count": "{n} 件商品",
		"meta.description": "gomu — 全球好物, 工厂直发。20 万+ 精选好物, 满 $69 包邮, 7 天无理由退换。",
		"home.title":       "全球好物 · 工厂直发", "home.badge": "新人首单立减 $15",
		"home.heading1": "像逛工厂一样,", "home.heading2": "把好物直接搬回家",
		"home.sub": "20 万+ 精选好物 · 全场满 $69 包邮 · 7 天无理由退换 · 90 天价保",
		"home.cta": "开始逛", "home.deals": "今日闪购", "home.coupons": "满 $29 减 $5 · 满 $99 减 $20",
		"perk.ship.t": "满 $69 包邮", "perk.ship.d": "大部分地区", "perk.fast.t": "24h 极速发货", "perk.fast.d": "工厂直连",
		"perk.return.t": "7 天退换", "perk.return.d": "无理由", "perk.price.t": "90 天价保", "perk.price.d": "买贵退差",
		"flash.title": "限时闪购", "flash.badge": "5 折起", "flash.ends": "距结束", "flash.claimed": "已抢 {n}%",
		"feed.title": "为你推荐", "feed.sub": "根据你的浏览猜你喜欢", "feed.end": "已经到底了", "feed.more": "加载更多", "feed.loading": "加载中…",
		"product.reviews": "{n} 条评价", "product.sold": "已售 {n}", "product.save": "立省 {n}%", "product.flash": "闪购",
		"product.details": "商品详情", "product.reviewsTitle": "用户评价", "product.reviewsCount": "{n} 条", "product.stars": "{n} 星",
		"product.related": "相关推荐",
		"product.note":    "本页面由 Go 在服务端渲染(数据接口含模拟延迟), 加购面板、图廊、相关推荐是编译成 signals 的岛。商品图为演示用 emoji 棚拍。",
		"product.meta":    "{desc} 现价 {price}, 已售 {sold}, {reviews} 条好评, 满 $69 包邮。",
		"feature.1":       "材质精选", "feature.2": "工厂直发", "feature.3": "严格质检", "feature.4": "环保包装", "feature.5": "可选快递", "feature.6": "售后无忧",
		"pperk.ship": "满 $69 免运费", "pperk.return": "7 天无理由退换", "pperk.auth": "正品保障 · 假一赔十", "pperk.fast": "24 小时内发货",
		"listing.count": "{n} 件商品", "listing.empty": "没有找到相关商品", "listing.emptySub": "换个关键词, 或看看别的分类", "listing.back": "回首页逛逛",
		"sort.rec": "综合", "sort.sales": "销量", "sort.price": "价格 ↑", "sort.priceDesc": "价格 ↓",
		"card.soldout": "售罄", "card.sold": "已售 {n}", "card.flash": "闪购",
		"tag.hot": "热销", "tag.deal": "特价", "tag.rated": "好评",
		"add.pick": "请先选择完整规格", "add.added": "已加入购物车", "add.qty": "数量", "add.stock": "库存 {n} 件",
		"add.soldout": "已售罄", "add.adding": "加入中…", "add.button": "加入购物车", "add.decrease": "减少", "add.increase": "增加",
		"wish.label": "加入心愿单", "gallery.alt": "商品图", "gallery.thumb": "缩略图",
		"cart.title": "购物车", "cart.empty": "购物车还是空的", "cart.browse": "去挑点好物", "cart.summary": "订单摘要",
		"cart.subtotal": "商品小计({n} 件)", "cart.shipping": "运费", "cart.freeGap": "再买 {n} 即可免运费",
		"cart.total": "应付合计", "cart.checkout": "去结算 ({n})", "cart.secure": "安全支付 · 7 天无理由退换", "cart.remove": "删除",
		"checkout.title": "确认订单", "checkout.empty": "购物车是空的", "checkout.browse": "去逛逛", "checkout.shipping": "收货信息",
		"checkout.items": "商品清单", "checkout.count": "{n} 件", "checkout.subtotal": "商品小计", "checkout.shippingFee": "运费", "checkout.due": "应付",
		"form.name": "收货人", "form.namePh": "请输入姓名", "form.phone": "手机号", "form.phonePh": "例如 +86 138 0000 0000",
		"form.address": "收货地址", "form.addressPh": "省 / 市 / 区 + 详细地址", "form.submit": "提交订单 · {total}", "form.submitting": "提交中…",
		"form.note":    "点击提交即代表同意《服务条款》· 演示环境, 不会真的扣款",
		"orders.title": "我的订单", "orders.empty": "还没有订单", "orders.first": "去下第一单", "orders.count": "共 {n} 件", "orders.total": "合计",
		"order.title": "订单 {id}", "order.success": "下单成功", "order.number": "订单号", "order.viewAll": "查看全部订单", "order.items": "商品清单",
		"order.shipping": "收货信息", "order.paid": "实付金额", "order.continue": "继续购物",
		"status.paid":    "已支付, 待发货",
		"category.meta":  "{label} · 精选好物, 共 {n} 件, 工厂直发, 满 $69 包邮。",
		"error.notFound": "找不到这个页面", "error.server": "服务器开小差了, 请稍后再试", "error.home": "回首页",
	},
}

// Options: the complete server configuration. dev enables detailed error pages and live reload; publicFS serves /public/*
// (the embedded public/ directory in main.go; the Cloudflare worker embeds its own copy). Addr / ClientDir / PublicDir
// are set by the caller when it runs a real server.
func Options(dev bool, publicFS fs.FS) gotsx.Options {
	panel := func(icon, msg, sid, locale string) gotsx.Node {
		return gen.Layout(gen.LayoutProps{Title: msg, Sid: sid, Locale: locale, Wide: true, Children: gotsx.El("div",
			[]gotsx.Attr{gotsx.A("class", "card flex flex-col items-center py-24 text-center")},
			gotsx.El("div", []gotsx.Attr{gotsx.A("class", "text-5xl")}, gotsx.Text(icon)),
			gotsx.El("p", []gotsx.Attr{gotsx.A("class", "mt-4 font-medium")}, gotsx.Text(msg)),
			gotsx.El("a", []gotsx.Attr{gotsx.A("href", "/"), gotsx.A("class", "btn btn-primary mt-6")}, gotsx.Text(gotsx.Tr(locale, "error.home"))))})
	}
	i18n := &gotsx.I18n{
		Locales: []string{"en", "zh"}, Default: "en", Prefix: true, // / is English, /zh/... is Chinese
		Currency: map[string]string{"en": "$", "zh": "¥"},
		Messages: messages,
	}
	return gotsx.Options{
		Dev:      dev,
		I18n:     i18n,
		Routes:   gen.Routes,
		ClientFS: gen.ClientFS,
		PublicFS: publicFS,
		NotFound: func(p gotsx.PageProps) gotsx.Node {
			return panel("🔍", gotsx.Tr(p.Locale, "error.notFound"), p.Cookies["sid"], p.Locale)
		},
		ErrorPage: func(p gotsx.PageProps, err error) gotsx.Node {
			return panel("😵", gotsx.Tr(p.Locale, "error.server"), p.Cookies["sid"], p.Locale)
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
				w.Write([]byte(`{"name":"gomu — factory-direct store","short_name":"gomu","start_url":"/","scope":"/","display":"standalone","background_color":"#ffffff","theme_color":"#18181b","description":"Global goods, factory-direct","icons":[{"src":"/icon.svg","sizes":"any","type":"image/svg+xml","purpose":"any maskable"}]}`))
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
					errs["name"] = "Name must be at least 2 characters"
				}
				phone := strings.NewReplacer(" ", "", "-", "", "(", "", ")", "").Replace(strings.TrimSpace(b.Phone))
				if !phoneRe.MatchString(phone) {
					errs["phone"] = "Enter a valid phone number (7-15 digits, optional +)"
				}
				if len(strings.TrimSpace(b.Address)) < 5 {
					errs["address"] = "Address is too short"
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
	}
}
