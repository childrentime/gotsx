// Package host is the shop's host module: catalog / cart / wishlist / orders / formatting.
// Everything is in memory (behind mutexes); swapping in a database is a host implementation detail the dialect never sees.
package host

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math"
	mrand "math/rand"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	gotsx "github.com/childrentime/gotsx/runtime"
)

// lag simulates backend I/O latency so loading states / skeletons wait for a real "API". SHOP_NOLAG=1 disables it.
var lagOn = os.Getenv("SHOP_NOLAG") == ""
var lagMu sync.Mutex

func lag(minMs, maxMs int) {
	if !lagOn {
		return
	}
	lagMu.Lock()
	d := minMs + mrand.Intn(maxMs-minMs+1)
	lagMu.Unlock()
	time.Sleep(time.Duration(d) * time.Millisecond)
}

// ---------- types ----------

type Category struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Emoji string `json:"emoji"`
}

type Variant struct {
	Name    string   `json:"name"`
	Options []string `json:"options"`
}

type Product struct {
	ID       string    `json:"id"`
	Title    string    `json:"title"`
	Emoji    string    `json:"emoji"`
	Hue      int       `json:"hue"`
	Price    int       `json:"price"` // cents
	Orig     int       `json:"orig"`
	Sold     int       `json:"sold"`
	Rating   float64   `json:"rating"`
	Reviews  int       `json:"reviews"`
	Category string    `json:"category"`
	Stock    int       `json:"stock"`
	Flash    bool      `json:"flash"`
	Desc     string    `json:"desc"`
	Variants []Variant `json:"variants"`
	Gallery  []string  `json:"gallery"`
}

type Paged struct {
	Items    []Product `json:"items"`
	Total    int       `json:"total"`
	Page     int       `json:"page"`
	Pages    int       `json:"pages"`
	PageList []int     `json:"pageList"`
}

type Review struct {
	User  string `json:"user"`
	Stars int    `json:"stars"`
	Text  string `json:"text"`
	Date  string `json:"date"`
}

type CartItem struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Emoji    string `json:"emoji"`
	Hue      int    `json:"hue"`
	Variant  string `json:"variant"`
	Qty      int    `json:"qty"`
	PriceFmt string `json:"priceFmt"`
	LineFmt  string `json:"lineFmt"`
}

type CartView struct {
	Items       []CartItem `json:"items"`
	Count       int        `json:"count"`
	Empty       bool       `json:"empty"`
	SubtotalFmt string     `json:"subtotalFmt"`
	ShippingFmt string     `json:"shippingFmt"`
	TotalFmt    string     `json:"totalFmt"`
	FreeShip    bool       `json:"freeShip"`
	FreeGapFmt  string     `json:"freeGapFmt"`
}

type Order struct {
	ID         string     `json:"id"`
	Items      []CartItem `json:"items"`
	Count      int        `json:"count"`
	TotalFmt   string     `json:"totalFmt"`
	Name       string     `json:"name"`
	Phone      string     `json:"phone"`
	Address    string     `json:"address"`
	CreatedFmt string     `json:"createdFmt"`
	Status     string     `json:"status"`
}

// ---------- catalog ----------

type CatalogModule struct {
	products []Product
	byID     map[string]*Product
	cats     []Category
	flashEnd time.Time
}

func (c *CatalogModule) Categories() []Category { return c.cats }

func (c *CatalogModule) CatLabel(key string) string {
	for _, x := range c.cats {
		if x.Key == key {
			return x.Label
		}
	}
	return "All"
}

func (c *CatalogModule) FlashLeftMs() int { return int(time.Until(c.flashEnd).Milliseconds()) }

func (c *CatalogModule) Flash() []Product {
	out := []Product{}
	for _, p := range c.products {
		if p.Flash {
			out = append(out, p)
		}
	}
	return out
}

func (c *CatalogModule) Get(id string) (Product, error) {
	lag(70, 150)
	if p, ok := c.byID[id]; ok {
		return *p, nil
	}
	return Product{}, fmt.Errorf("%w: product %s", gotsx.ErrNotFound, id)
}

const pageSize = 20

func (c *CatalogModule) List(cat, q, sortBy string, page int) Paged {
	lag(120, 240)
	q = strings.ToLower(strings.TrimSpace(q))
	var hit []Product
	for _, p := range c.products {
		if cat != "" && p.Category != cat {
			continue
		}
		if q != "" && !strings.Contains(strings.ToLower(p.Title+p.Desc), q) {
			continue
		}
		hit = append(hit, p)
	}
	switch sortBy {
	case "sales":
		sort.SliceStable(hit, func(i, j int) bool { return hit[i].Sold > hit[j].Sold })
	case "price":
		sort.SliceStable(hit, func(i, j int) bool { return hit[i].Price < hit[j].Price })
	case "priceDesc":
		sort.SliceStable(hit, func(i, j int) bool { return hit[i].Price > hit[j].Price })
	}
	total := len(hit)
	pages := (total + pageSize - 1) / pageSize
	if pages == 0 {
		pages = 1
	}
	if page < 1 {
		page = 1
	}
	if page > pages {
		page = pages
	}
	a := (page - 1) * pageSize
	b := a + pageSize
	if b > total {
		b = total
	}
	var list []int
	for i := 1; i <= pages; i++ {
		if pages <= 7 || i == 1 || i == pages || (i >= page-2 && i <= page+2) {
			list = append(list, i)
		}
	}
	return Paged{Items: hit[a:b], Total: total, Page: page, Pages: pages, PageList: list}
}

var reviewPool = []string{
	"Better quality than I expected, and it shipped fast. Would buy again.",
	"Exactly as pictured, solid build. The whole family likes it.",
	"Great value for the price — already recommended it to a coworker.",
	"Carefully packed, helpful support. Five stars.",
	"Waited a week before reviewing: no issues at all.",
	"Much cheaper than in stores, delivery just took a little while.",
	"Second time ordering, as good as the first.",
	"Nice attention to detail; nothing to complain about at this price.",
}

var reviewerNames = []string{"Maya", "Liam", "Sofia", "Noah", "Ava", "Ethan", "Chloe", "Lucas", "Mia", "Oliver", "Zoe", "Leo"}

func (c *CatalogModule) ProductReviews(id string) []Review {
	lag(160, 320)
	p, ok := c.byID[id]
	if !ok {
		return []Review{}
	}
	r := mrand.New(mrand.NewSource(int64(hash(id))))
	n := 4 + r.Intn(3)
	out := make([]Review, n)
	for i := range out {
		stars := 4 + r.Intn(2)
		if i == n-1 && p.Rating < 4.5 {
			stars = 3
		}
		out[i] = Review{
			User:  fmt.Sprintf("%s %c.", reviewerNames[r.Intn(len(reviewerNames))], 'A'+r.Intn(26)),
			Stars: stars,
			Text:  reviewPool[r.Intn(len(reviewPool))],
			Date:  time.Now().AddDate(0, 0, -r.Intn(60)).Format("2006-01-02"),
		}
	}
	return out
}

func hash(s string) uint32 {
	var h uint32 = 2166136261
	for i := 0; i < len(s); i++ {
		h = (h ^ uint32(s[i])) * 16777619
	}
	return h
}

// ---------- cart ----------

type cartLine struct {
	ID      string
	Variant string
	Qty     int
}

type CartModule struct {
	mu    sync.Mutex
	carts map[string][]cartLine
}

const freeShipOver = 6900
const shipFee = 500

func (m *CartModule) view(sid string) CartView {
	lines := m.carts[sid]
	v := CartView{Items: []CartItem{}}
	subtotal := 0
	for _, l := range lines {
		p, ok := Catalog.byID[l.ID]
		if !ok {
			continue
		}
		line := p.Price * l.Qty
		subtotal += line
		v.Count += l.Qty
		v.Items = append(v.Items, CartItem{
			ID: l.ID, Title: p.Title, Emoji: p.Emoji, Hue: p.Hue, Variant: l.Variant, Qty: l.Qty,
			PriceFmt: Intl.FmtPrice(p.Price), LineFmt: Intl.FmtPrice(line),
		})
	}
	v.Empty = v.Count == 0
	ship := 0
	if !v.Empty && subtotal < freeShipOver {
		ship = shipFee
	}
	v.FreeShip = ship == 0 && !v.Empty
	if ship > 0 {
		v.FreeGapFmt = Intl.FmtPrice(freeShipOver - subtotal)
	}
	v.SubtotalFmt = Intl.FmtPrice(subtotal)
	if v.Empty {
		v.ShippingFmt = "—"
	} else if ship == 0 {
		v.ShippingFmt = "Free"
	} else {
		v.ShippingFmt = Intl.FmtPrice(ship)
	}
	v.TotalFmt = Intl.FmtPrice(subtotal + ship)
	return v
}

func (m *CartModule) View(sid string) CartView {
	lag(40, 90)
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.view(sid)
}

func (m *CartModule) Add(sid, id, variant string, qty int) (CartView, error) {
	lag(220, 420)
	m.mu.Lock()
	defer m.mu.Unlock()
	p, ok := Catalog.byID[id]
	if !ok {
		return CartView{}, fmt.Errorf("%w: product %s", gotsx.ErrNotFound, id)
	}
	if qty < 1 {
		qty = 1
	}
	have := 0
	for _, l := range m.carts[sid] {
		if l.ID == id && l.Variant == variant {
			have = l.Qty
		}
	}
	if have+qty > p.Stock {
		return CartView{}, fmt.Errorf("Only %d left in stock", p.Stock)
	}
	lines := m.carts[sid]
	found := false
	for i := range lines {
		if lines[i].ID == id && lines[i].Variant == variant {
			lines[i].Qty += qty
			found = true
		}
	}
	if !found {
		lines = append(lines, cartLine{ID: id, Variant: variant, Qty: qty})
	}
	if m.carts == nil {
		m.carts = map[string][]cartLine{}
	}
	m.carts[sid] = lines
	return m.view(sid), nil
}

func (m *CartModule) SetQty(sid, id, variant string, qty int) CartView {
	lag(160, 320)
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []cartLine
	for _, l := range m.carts[sid] {
		if l.ID == id && l.Variant == variant {
			if qty <= 0 {
				continue
			}
			if p, ok := Catalog.byID[id]; ok && qty > p.Stock {
				qty = p.Stock
			}
			l.Qty = qty
		}
		out = append(out, l)
	}
	m.carts[sid] = out
	return m.view(sid)
}

func (m *CartModule) clear(sid string) { delete(m.carts, sid) }

// ---------- wishlist ----------

type WishModule struct {
	mu   sync.Mutex
	data map[string]map[string]bool
}

func (m *WishModule) List(sid string) []string {
	lag(30, 70)
	m.mu.Lock()
	defer m.mu.Unlock()
	out := []string{}
	for id := range m.data[sid] {
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}

func (m *WishModule) Toggle(sid, id string) bool {
	lag(150, 300)
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.data == nil {
		m.data = map[string]map[string]bool{}
	}
	if m.data[sid] == nil {
		m.data[sid] = map[string]bool{}
	}
	if m.data[sid][id] {
		delete(m.data[sid], id)
		return false
	}
	m.data[sid][id] = true
	return true
}

// ---------- orders ----------

type OrdersModule struct {
	mu   sync.Mutex
	data map[string][]Order
	seq  int
}

func (m *OrdersModule) List(sid string) []Order {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := append([]Order{}, m.data[sid]...)
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out
}

func (m *OrdersModule) Get(sid, id string) (Order, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, o := range m.data[sid] {
		if o.ID == id {
			return o, nil
		}
	}
	return Order{}, fmt.Errorf("%w: order %s", gotsx.ErrNotFound, id)
}

// Place is only called from Go actions — writes never go through the dialect
func (m *OrdersModule) Place(sid, name, phone, address string) (Order, error) {
	lag(500, 900)
	cv := Cart.View(sid)
	if cv.Empty {
		return Order{}, fmt.Errorf("Your cart is empty")
	}
	m.mu.Lock()
	m.seq++
	o := Order{
		ID: fmt.Sprintf("ORD%06d", m.seq), Items: cv.Items, Count: cv.Count, TotalFmt: cv.TotalFmt,
		Name: name, Phone: phone, Address: address,
		CreatedFmt: time.Now().Format("2006-01-02 15:04"), Status: "paid", // status is a key, translated by the UI (status.paid)
	}
	if m.data == nil {
		m.data = map[string][]Order{}
	}
	m.data[sid] = append(m.data[sid], o)
	m.mu.Unlock()
	Cart.mu.Lock()
	Cart.clear(sid)
	Cart.mu.Unlock()
	return o, nil
}

// ---------- site / formatting ----------

type SiteModule struct{}

func siteBase() string {
	if b := os.Getenv("SHOP_BASE"); b != "" {
		return strings.TrimRight(b, "/")
	}
	return "https://shop.example.com"
}

func (SiteModule) Name() string    { return "gomu" }
func (SiteModule) BaseUrl() string { return siteBase() }
func (SiteModule) Url(path string) string {
	if path == "" {
		path = "/"
	}
	if strings.HasPrefix(path, "http") {
		return path
	}
	return siteBase() + path
}

// SitemapURLs: home + categories + every product (for /sitemap.xml)
func (SiteModule) SitemapURLs() []string {
	out := []string{siteBase() + "/"}
	for _, c := range Catalog.cats {
		out = append(out, siteBase()+"/c/"+c.Key)
	}
	for _, p := range Catalog.products {
		out = append(out, siteBase()+"/p/"+p.ID)
	}
	return out
}

type IntlModule struct{}

func (IntlModule) FmtPrice(cents int) string {
	sign := ""
	if cents < 0 {
		sign = "-"
		cents = -cents
	}
	return fmt.Sprintf("%s$%d.%02d", sign, cents/100, cents%100)
}

func (IntlModule) FmtSold(n int) string {
	if n >= 1000 {
		return fmt.Sprintf("%d.%dk", n/1000, (n%1000)/100)
	}
	return fmt.Sprintf("%d", n)
}

// ---------- seed data ----------

type catSpec struct {
	key, label, emoji string
	hue               int
	variants          []Variant
	items             []itemSpec
}
type itemSpec struct {
	name, emoji string
}

var specs = []catSpec{
	{"digital", "Electronics", "📱", 210, []Variant{{Name: "Color", Options: []string{"Obsidian Black", "Pearl White", "Sierra Blue"}}},
		[]itemSpec{{"Wireless Earbuds", "🎧"}, {"Smart Watch", "⌚"}, {"20,000mAh Power Bank", "🔋"}, {"Bluetooth Speaker", "🔊"}, {"Desk Phone Stand", "📱"}, {"3-in-1 Fast-Charge Cable", "🔌"}}},
	{"home", "Home & Living", "🏠", 30, []Variant{{Name: "Style", Options: []string{"Nordic", "Minimal", "Cream"}}},
		[]itemSpec{{"Bean Bag Lounger", "🛋️"}, {"Aroma Humidifier", "💨"}, {"Blackout Curtains", "🪟"}, {"Storage Shelf", "🗄️"}, {"Bedside Night Light", "💡"}, {"Memory Foam Pillow", "🛏️"}}},
	{"fashion", "Fashion", "👗", 330, []Variant{{Name: "Color", Options: []string{"Black", "Ivory", "Khaki"}}, {Name: "Size", Options: []string{"S", "M", "L", "XL"}}},
		[]itemSpec{{"Relaxed-Fit Hoodie", "🧥"}, {"High-Waist Wide-Leg Pants", "👖"}, {"Floral Midi Dress", "👗"}, {"Chunky Sneakers", "👟"}, {"Bucket Hat", "🧢"}, {"Wool Scarf", "🧣"}}},
	{"beauty", "Beauty", "💄", 300, []Variant{{Name: "Shade", Options: []string{"01 Rosewood", "02 Maple", "03 Tomato"}}},
		[]itemSpec{{"Matte Velvet Lipstick", "💄"}, {"Hydrating Foundation", "🧴"}, {"12-Shade Eyeshadow Palette", "🎨"}, {"Cleansing Balm", "🧼"}, {"Hyaluronic Sheet Mask", "🧖"}, {"SPF50+ Sunscreen", "☀️"}}},
	{"toys", "Toys", "🧸", 45, nil,
		[]itemSpec{{"60cm Plush Bear", "🧸"}, {"980-Piece Rocket Building Set", "🚀"}, {"RC Off-Road Truck", "🏎️"}, {"Fidget Cube", "🧩"}, {"Electric Bubble Machine", "🫧"}, {"Dinosaur Figure Set", "🦖"}}},
	{"sports", "Sports", "🏀", 140, []Variant{{Name: "Spec", Options: []string{"Standard", "Weighted"}}},
		[]itemSpec{{"Thick Yoga Mat", "🧘"}, {"Counting Jump Rope", "🪢"}, {"Size 7 Basketball", "🏀"}, {"Cycling Gloves", "🧤"}, {"1L Sports Bottle", "🚰"}, {"Mini Massage Gun", "💆"}}},
	{"kitchen", "Kitchen", "🍳", 20, nil,
		[]itemSpec{{"28cm Non-Stick Skillet", "🍳"}, {"5-Piece Kitchen Shears Set", "✂️"}, {"Food Storage Container Set", "🥡"}, {"Manual Juicer Cup", "🥤"}, {"Bamboo Cutting Board", "🪵"}, {"Ceramic Dinner Plates (Set of 4)", "🍽️"}}},
	{"pets", "Pets", "🐾", 260, []Variant{{Name: "Flavor", Options: []string{"Chicken", "Salmon", "Beef"}}},
		[]itemSpec{{"Freeze-Dried Cat Treats", "🐱"}, {"Dog Dental Chews", "🦴"}, {"Automatic Pet Water Fountain", "⛲"}, {"Cat Scratcher Bed", "📦"}, {"Cat Wand Toy Set", "🪶"}, {"Pet Grooming Brush", "🐾"}}},
}

var adjectives = []string{"Bestselling", "Upgraded", "Everyday", "Compact"}

func newCatalog() *CatalogModule {
	r := mrand.New(mrand.NewSource(42))
	c := &CatalogModule{byID: map[string]*Product{}, flashEnd: time.Now().Add(4 * time.Hour)}
	n := 0
	for _, cs := range specs {
		c.cats = append(c.cats, Category{Key: cs.key, Label: cs.label, Emoji: cs.emoji})
		for _, it := range cs.items {
			for _, adj := range adjectives {
				n++
				price := 190 + r.Intn(12000)
				orig := price + price*(30+r.Intn(170))/100
				p := Product{
					ID:       fmt.Sprintf("p%03d", n),
					Title:    fmt.Sprintf("%s %s", adj, it.name),
					Emoji:    it.emoji,
					Hue:      (cs.hue + r.Intn(40)) % 360,
					Price:    price,
					Orig:     orig,
					Sold:     r.Intn(50000),
					Rating:   4.0 + float64(r.Intn(10))/10,
					Reviews:  50 + r.Intn(5000),
					Category: cs.key,
					Stock:    r.Intn(500),
					Desc:     fmt.Sprintf("%s %s — carefully sourced materials, shipped straight from the factory. 7-day free returns; free shipping on orders over $69.", adj, it.name),
					Variants: cs.variants,
					Gallery:  []string{it.emoji, cs.emoji, "🎁"},
				}
				if p.Variants == nil {
					p.Variants = []Variant{}
				}
				if n%13 == 0 {
					p.Stock = 0
				}
				if n%23 < 2 && p.Stock > 0 {
					p.Flash = true
					p.Price = p.Price * 65 / 100
				}
				c.products = append(c.products, p)
			}
		}
	}
	for i := range c.products {
		c.byID[c.products[i].ID] = &c.products[i]
	}
	return c
}

func NewSID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// ---------- card DTO for client rendering: prices / discounts pre-formatted on the server ----------

type Card struct {
	ID       string  `json:"id"`
	Title    string  `json:"title"`
	Emoji    string  `json:"emoji"`
	Hue      int     `json:"hue"`
	PriceFmt string  `json:"priceFmt"`
	OrigFmt  string  `json:"origFmt"`
	Off      int     `json:"off"`
	Rating   float64 `json:"rating"`
	Reviews  int     `json:"reviews"`
	SoldFmt  string  `json:"soldFmt"`
	Flash    bool    `json:"flash"`
	SoldOut  bool    `json:"soldOut"`
	Progress int     `json:"progress"`
	Tag      string  `json:"tag"`
}

func toCard(p Product) Card {
	off := 0
	if p.Orig > 0 {
		off = int(math.Round((1 - float64(p.Price)/float64(p.Orig)) * 100))
	}
	tag := "" // a key the UI translates (tag.hot / tag.deal / tag.rated)
	switch {
	case p.Sold > 30000:
		tag = "hot"
	case p.Price < 1500:
		tag = "deal"
	case p.Rating >= 4.8:
		tag = "rated"
	}
	return Card{
		ID: p.ID, Title: p.Title, Emoji: p.Emoji, Hue: p.Hue,
		PriceFmt: Intl.FmtPrice(p.Price), OrigFmt: Intl.FmtPrice(p.Orig), Off: off,
		Rating: p.Rating, Reviews: p.Reviews, SoldFmt: Intl.FmtSold(p.Sold),
		Flash: p.Flash, SoldOut: p.Stock == 0, Progress: 45 + int(hash(p.ID)%50), Tag: tag,
	}
}

func cards(ps []Product) []Card {
	out := make([]Card, len(ps))
	for i, p := range ps {
		out[i] = toCard(p)
	}
	return out
}

type PagedCards struct {
	Items    []Card `json:"items"`
	Total    int    `json:"total"`
	Page     int    `json:"page"`
	Pages    int    `json:"pages"`
	PageList []int  `json:"pageList"`
}

func (c *CatalogModule) ListCards(cat, q, sortBy string, page int) PagedCards {
	pg := c.List(cat, q, sortBy, page)
	return PagedCards{Items: cards(pg.Items), Total: pg.Total, Page: pg.Page, Pages: pg.Pages, PageList: pg.PageList}
}

func (c *CatalogModule) FlashCards() []Card {
	lag(90, 180)
	return cards(c.Flash())
}

type FeedResult struct {
	Cards   []Card `json:"cards"`
	HasMore bool   `json:"hasMore"`
	Page    int    `json:"page"`
}

// Feed: the home page "for you" stream, stable shuffled pagination
func (c *CatalogModule) Feed(page int) FeedResult {
	lag(450, 750)
	all := append([]Product{}, c.products...)
	r := mrand.New(mrand.NewSource(7))
	r.Shuffle(len(all), func(i, j int) { all[i], all[j] = all[j], all[i] })
	const n = 15
	if page < 1 {
		page = 1
	}
	a := (page - 1) * n
	if a > len(all) {
		a = len(all)
	}
	b := a + n
	if b > len(all) {
		b = len(all)
	}
	return FeedResult{Cards: cards(all[a:b]), HasMore: b < len(all) && page < 6, Page: page}
}

// Related: same-category recommendations by sales, excluding the product itself
func (c *CatalogModule) Related(id string) []Card {
	lag(400, 650)
	p, ok := c.byID[id]
	if !ok {
		return []Card{}
	}
	var pool []Product
	for _, x := range c.products {
		if x.Category == p.Category && x.ID != id && x.Stock > 0 {
			pool = append(pool, x)
		}
	}
	sort.SliceStable(pool, func(i, j int) bool { return pool[i].Sold > pool[j].Sold })
	if len(pool) > 6 {
		pool = pool[:6]
	}
	return cards(pool)
}

var (
	Catalog = newCatalog()
	Cart    = &CartModule{carts: map[string][]cartLine{}}
	Wish    = &WishModule{data: map[string]map[string]bool{}}
	Orders  = &OrdersModule{data: map[string][]Order{}}
	Intl    = IntlModule{}
	Site    = SiteModule{}
)

var Registry = map[string]gotsx.HostModule{
	"catalog": {Value: Catalog, Go: "host.Catalog"},
	"cart":    {Value: Cart, Go: "host.Cart"},
	"wish":    {Value: Wish, Go: "host.Wish"},
	"orders":  {Value: Orders, Go: "host.Orders"},
	"intl":    {Value: Intl, Go: "host.Intl"},
	"site":    {Value: Site, Go: "host.Site"},
}
