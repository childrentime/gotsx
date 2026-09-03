package gotsx

import (
	"net/http/httptest"
	"strings"
	"testing"
)

type cartState struct {
	Count int      `json:"count"`
	Items []string `json:"items"`
}

// A store renders its initial value outside a request and the page's seed inside one; the seed reaches the
// browser as a JSON data block in <head>, nil slices become [] like island props.
func TestStoreSeed(t *testing.T) {
	cart := NewStore("stores_cart_cart", cartState{Count: 0, Items: []string{}})
	if v := cart.Get(nil); v.Count != 0 {
		t.Errorf("no ctx: initial value, got %+v", v)
	}
	if v := cart.Get(&Ctx{}); v.Count != 0 {
		t.Errorf("ctx without scope: initial value, got %+v", v)
	}
	Seed(PageProps{}, cart, cartState{Count: 9}) // no scope (tests, Render): a no-op, must not panic

	badge := func() Node {
		return Island("Badge", struct{}{}, func(c *Ctx) { c.Dyn(Num(float64(cart.Get(c).Count))) })
	}
	route := Route{Pattern: "/", Render: func(p PageProps) Node {
		if p.Query["seed"] != "" {
			Seed(p, cart, cartState{Count: 3, Items: nil}) // the page body runs before any HTML is written
		}
		return El("html", nil, El("head", nil, El("title", nil, Text("x"))), El("body", nil, badge()))
	}}
	h := testServer(t, Options{Routes: []Route{route}})

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/?seed=1", nil))
	body := rec.Body.String()
	if !strings.Contains(body, "<!--$-->3<!--/-->") {
		t.Errorf("the island should render the seeded value\n%s", body)
	}
	seedTag := `<script type="application/json" data-gotsx-stores>{"stores_cart_cart":{"count":3,"items":[]}}</script></head>`
	if !strings.Contains(body, seedTag) {
		t.Errorf("the seed block should sit at the end of <head> with [] for the nil slice\n%s", body)
	}
	if strings.Index(body, "window.__GOTSX") > strings.Index(body, "data-gotsx-stores") {
		t.Errorf("the seed block follows the bootstrap script\n%s", body)
	}

	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	body = rec.Body.String()
	if !strings.Contains(body, "<!--$-->0<!--/-->") || strings.Contains(body, "data-gotsx-stores") {
		t.Errorf("without a seed: initial value and no seed block\n%s", body)
	}
}

// Layouts seed through the embedded PageProps; a Suspense boundary (own goroutine) sees the seed too
func TestStoreSeedLayoutAndSuspense(t *testing.T) {
	cart := NewStore("stores_cart_cart2", cartState{})
	route := Route{Pattern: "/", Render: func(p PageProps) Node {
		lp := LayoutProps{PageProps: p}
		Seed(lp, cart, cartState{Count: 7, Items: []string{"a"}})
		return El("html", nil, El("body", nil,
			Suspense(Text("wait"), func() Node { return func(c *Ctx) { c.Esc("late:" + Num(float64(cart.Get(c).Count))) } }),
		))
	}}
	h := testServer(t, Options{Routes: []Route{route}})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	body := rec.Body.String()
	if !strings.Contains(body, `{"stores_cart_cart2":{"count":7,"items":["a"]}}`) || !strings.Contains(body, "late:7") {
		t.Errorf("layout seed + streamed boundary\n%s", body)
	}
}

// The seed block is a data block: a value containing </script> cannot end it
func TestStoreSeedEscapes(t *testing.T) {
	sc := newReqScope()
	sc.set("s", map[string]string{"x": "</script><script>alert(1)</script>"})
	out := sc.script()
	if strings.Contains(out, "</script><script>") || !strings.Contains(out, `</script>`) {
		t.Errorf("unescaped seed: %s", out)
	}
	var none *reqScope
	if none.script() != "" {
		t.Error("nil scope: no block")
	}
}
