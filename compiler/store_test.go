package compiler

import (
	"strings"
	"testing"
)

// The store module, an island reading it and a page seeding it — the shape every app uses
func storeApp(extra map[string]string) map[string]string {
	files := map[string]string{
		"/x/app/stores/cart.client.tsx": `import { createStore } from "gotsx";
import type { Model } from "host:data";
export interface CartState { count: number; items: Model[]; open: boolean; note: string }
export const cart = createStore<CartState>({ count: 0, items: [] });
export function addItem(m: Model) {
  cart.set((s) => { s.items.push(m); s.count += 1; });
}
export function replaceAll(v: CartState) { cart.set(v); }
`,
		"/x/app/islands/Badge.client.tsx": `import { useState, useEffect } from "gotsx";
import { cart, addItem } from "../stores/cart";
import type { Model } from "host:data";
function Row({ m }: { m: Model }) { return <li>{m.title} / {cart.note}</li>; }
export default function Badge({ label }: { label: string }) {
  const { count, items } = cart;
  const [beat, setBeat] = useState(false);
  const double = count * 2;
  useEffect(() => { if (count > 0) setBeat(true); });
  return <div>
    <span class={beat ? "hot" : ""}>{label} {count} {double} {cart.open ? "open" : "closed"}</span>
    <ul>{items.map((m) => <Row key={m.id} m={m} />)}</ul>
    <button onClick={() => addItem({ id: "x", title: "X", price: 1, tags: [] })}>add</button>
  </div>;
}
`,
		"/x/app/pages/index.server.tsx": `import type { PageProps } from "gotsx";
import { seed } from "gotsx";
import { models } from "host:data";
import { cart } from "../stores/cart";
import Badge from "../islands/Badge.client";
export default function Home({ query }: PageProps) {
  const list = models.list();
  seed(cart, { count: list.length, items: list, open: query.open === "1", note: "hi" });
  return <html><head><title>x</title></head><body><Badge label="Cart" /></body></html>;
}
`,
	}
	for k, v := range extra {
		files[k] = v
	}
	return files
}

func TestStoreCodegen(t *testing.T) {
	gs, js := compileMany(t, storeApp(nil))
	store, island, page := gs["/x/app/stores/cart.client.tsx"], gs["/x/app/islands/Badge.client.tsx"], gs["/x/app/pages/index.server.tsx"]
	storeJS, islandJS := js["/x/app/stores/cart.client.tsx"], js["/x/app/islands/Badge.client.tsx"]
	// Go: the store is a package var holding the initial state; missing literal fields are zero values
	for _, w := range []string{
		`var stores_cart_cart = gotsx.NewStore("stores_cart_cart", CartState{Count: 0, Items: []host.Model{}})`,
	} {
		if !strings.Contains(store, w) {
			t.Errorf("store Go 缺 %q\n%s", w, store)
		}
	}
	if strings.Contains(store, "func addItem") || strings.Contains(store, "func replaceAll") {
		t.Errorf("store.set 只进 JS 后端, Go 不应生成 addItem/replaceAll\n%s", store)
	}
	// Go: the island renders with the request's value: the wrapper defers the body until the island is written (context in hand)
	for _, w := range []string{
		`return gotsx.Island("Badge", props, func(_ctx *gotsx.Ctx) { _ctx.N(Badge_ssr(_ctx, props)) })`,
		"func Badge_ssr(_ctx *gotsx.Ctx, props BadgeProps) gotsx.Node {",
		"_t1 := stores_cart_cart.Get(_ctx)", "count := _t1.Count", "items := _t1.Items",
		"stores_cart_cart.Get(_ctx).Open",
		"func islands_Badge_Row(_ctx *gotsx.Ctx, props islands_Badge_RowProps) gotsx.Node {",
		"islands_Badge_Row(_ctx, islands_Badge_RowProps{M: m})",
		"stores_cart_cart.Get(_ctx).Note",
	} {
		if !strings.Contains(island, w) {
			t.Errorf("island Go 缺 %q\n%s", w, island)
		}
	}
	// Go: seed compiles to gotsx.Seed(props, store, value) in the page body
	if !strings.Contains(page, `gotsx.Seed(props, stores_cart_cart, CartState{Count: float64(len(list)), Items: list, Open: (query["open"] == "1"), Note: "hi"})`) {
		t.Errorf("page Go 缺 Seed\n%s", page)
	}
	// JS: G.store with every field present; set compiles as a method call; reads are signal calls
	for _, w := range []string{
		`export const cart = G.store("stores_cart_cart", { count: 0, items: [], open: false, note: "" });`,
		"cart.set((s) => {", "s.items.push(m);", "s.count += 1;",
		"export function replaceAll(v) {", "cart.set(v);",
	} {
		if !strings.Contains(storeJS, w) {
			t.Errorf("store JS 缺 %q\n%s", w, storeJS)
		}
	}
	for _, w := range []string{
		`import { cart, addItem } from "./cart.js";`,
		"const { count, items } = cart;",
		"const double = G.memo(() => (count() * 2));",
		"G.text(() => count())", "G.text(() => double())",
		`G.text(() => (cart.open() ? "open" : "closed"))`,
		"G.each(() => items(), (m) => Row({ m: m }), (m) => m.id)",
		"G.text(() => cart.note())",
		"if ((count() > 0)) {",
	} {
		if !strings.Contains(islandJS, w) {
			t.Errorf("island JS 缺 %q\n%s", w, islandJS)
		}
	}
}

// Every misuse is a positioned compile error
func TestStoreFence(t *testing.T) {
	cases := []struct {
		name  string
		files map[string]string
		want  string
	}{
		{"createStore in a server module", map[string]string{"/x/app/pages/index.server.tsx": `import { createStore } from "gotsx";
const s = createStore({ n: 0 });
export default function P() { return <b>1</b>; }`}, "createStore can only be used in a client module"},
		{"createStore inside a function", map[string]string{"/x/app/islands/A.client.tsx": `import { createStore } from "gotsx";
export default function A() { const s = createStore({ n: 0 }); return <b>{s.n}</b>; }`}, "createStore must be a module-level const"},
		{"createStore as a bare call", map[string]string{"/x/app/islands/A.client.tsx": `import { createStore } from "gotsx";
export default function A() { createStore({ n: 0 }); return <b>1</b>; }`}, "createStore can only be written as a module-level const"},
		{"store state must be an object", map[string]string{"/x/app/stores/s.client.tsx": `import { createStore } from "gotsx";
export const s = createStore(1);`}, "the store state must be an object type"},
		{"field named set", map[string]string{"/x/app/stores/s.client.tsx": `import { createStore } from "gotsx";
export const s = createStore({ set: 1 });`}, `cannot be named "set"`},
		{"unserializable field", map[string]string{"/x/app/stores/s.client.tsx": `import { createStore } from "gotsx";
export const s = createStore({ f: (x: number) => x });`}, "cannot be serialized"},
		{"set during render", map[string]string{"/x/app/stores/s.client.tsx": `import { createStore } from "gotsx";
export const s = createStore({ n: 0 });`, "/x/app/islands/A.client.tsx": `import { s } from "../stores/s";
export default function A() { s.set((d) => { d.n = 1; }); return <b>{s.n}</b>; }`}, "runs in the browser after the render"},
		{"set at module level", map[string]string{"/x/app/stores/s.client.tsx": `import { createStore } from "gotsx";
export const s = createStore({ n: 0 });
s.set((d) => { d.n = 1; });`}, "allowed at module level"},
		{"set with the wrong value", map[string]string{"/x/app/stores/s.client.tsx": `import { createStore } from "gotsx";
export const s = createStore({ n: 0 });
export function f() { s.set("x"); }`}, "is not assignable to"},
		{"assign to a store field", map[string]string{"/x/app/stores/s.client.tsx": `import { createStore } from "gotsx";
export const s = createStore({ n: 0 });
export function f() { s.n = 1; }`}, "store state is read-only outside s.set"},
		{"increment a destructured field", map[string]string{"/x/app/stores/s.client.tsx": `import { createStore } from "gotsx";
export const s = createStore({ n: 0 });`, "/x/app/islands/A.client.tsx": `import { s } from "../stores/s";
export default function A() { const { n } = s; const f = () => { n++; }; return <b onClick={f}>{n}</b>; }`}, "cannot assign to store field n"},
		{"push on a store array", map[string]string{"/x/app/stores/s.client.tsx": `import { createStore } from "gotsx";
export const s = createStore<{ xs: number[] }>({ xs: [] });
export function f() { s.xs.push(1); }`}, "cannot mutate store state directly"},
		{"push on a destructured array", map[string]string{"/x/app/stores/s.client.tsx": `import { createStore } from "gotsx";
export const s = createStore<{ xs: number[] }>({ xs: [] });`, "/x/app/islands/A.client.tsx": `import { s } from "../stores/s";
export default function A() { const { xs } = s; const f = () => { xs.push(1); }; return <b onClick={f}>{xs.length}</b>; }`}, "cannot mutate store field xs directly"},
		{"unknown field", map[string]string{"/x/app/stores/s.client.tsx": `import { createStore } from "gotsx";
export const s = createStore({ n: 0 });`, "/x/app/islands/A.client.tsx": `import { s } from "../stores/s";
export default function A() { return <b>{s.m}</b>; }`}, `store s has no field "m"`},
		{"nested destructuring", map[string]string{"/x/app/stores/s.client.tsx": `import { createStore } from "gotsx";
export const s = createStore({ u: { name: "" } });`, "/x/app/islands/A.client.tsx": `import { s } from "../stores/s";
export default function A() { const { u: { name } } = s; return <b>{name}</b>; }`}, "one level deep"},
		{"read in a server page", map[string]string{"/x/app/stores/s.client.tsx": `import { createStore } from "gotsx";
export const s = createStore({ n: 0 });`, "/x/app/pages/index.server.tsx": `import type { PageProps } from "gotsx";
import { s } from "../stores/s";
export default function P(props: PageProps) { return <b>{s.n}</b>; }`}, "can only be read in client code"},
		{"seed in an island", map[string]string{"/x/app/stores/s.client.tsx": `import { createStore } from "gotsx";
export const s = createStore({ n: 0 });`, "/x/app/islands/A.client.tsx": `import { seed } from "gotsx";
import { s } from "../stores/s";
export default function A() { seed(s, { n: 1 }); return <b>1</b>; }`}, "seed() can only be called in a page or layout"},
		{"seed in a server component outside pages", map[string]string{"/x/app/stores/s.client.tsx": `import { createStore } from "gotsx";
export const s = createStore({ n: 0 });`, "/x/app/components/C.server.tsx": `import { seed } from "gotsx";
import { s } from "../stores/s";
export default function C() { seed(s, { n: 1 }); return <b>1</b>; }`}, "seed() can only be called in a page or layout"},
		{"seed in meta", map[string]string{"/x/app/stores/s.client.tsx": `import { createStore } from "gotsx";
export const s = createStore({ n: 0 });`, "/x/app/pages/index.server.tsx": `import type { PageProps, Meta } from "gotsx";
import { seed } from "gotsx";
import { s } from "../stores/s";
export function meta(props: PageProps): Meta { seed(s, { n: 1 }); return { title: "x" }; }
export default function P(props: PageProps) { return <b>1</b>; }`}, "must be called in the body of the page's default component"},
		{"seed in a callback", map[string]string{"/x/app/stores/s.client.tsx": `import { createStore } from "gotsx";
export const s = createStore({ n: 0 });`, "/x/app/pages/index.server.tsx": `import type { PageProps } from "gotsx";
import { seed } from "gotsx";
import { s } from "../stores/s";
export default function P(props: PageProps) { const xs = [1].map((n) => { seed(s, { n }); return n; }); return <b>{xs.length}</b>; }`}, "must be called in the body of the page's default component"},
		{"seed in JSX", map[string]string{"/x/app/stores/s.client.tsx": `import { createStore } from "gotsx";
export const s = createStore({ n: 0 });`, "/x/app/pages/index.server.tsx": `import type { PageProps } from "gotsx";
import { seed } from "gotsx";
import { s } from "../stores/s";
export default function P(props: PageProps) { return <b>{seed(s, { n: 1 })}</b>; }`}, "cannot be void"},
		{"seed with the wrong value", map[string]string{"/x/app/stores/s.client.tsx": `import { createStore } from "gotsx";
export const s = createStore({ n: 0 });`, "/x/app/pages/index.server.tsx": `import type { PageProps } from "gotsx";
import { seed } from "gotsx";
import { s } from "../stores/s";
export default function P(props: PageProps) { seed(s, { n: "x" }); return <b>1</b>; }`}, "not assignable"},
		{"seed of a non-store", map[string]string{"/x/app/pages/index.server.tsx": `import type { PageProps } from "gotsx";
import { seed } from "gotsx";
export default function P(props: PageProps) { const o = { n: 1 }; seed(o, { n: 1 }); return <b>1</b>; }`}, "the first argument of seed must be a store"},
		{"store passed as a prop", map[string]string{"/x/app/stores/s.client.tsx": `import { createStore } from "gotsx";
export const s = createStore({ n: 0 });`, "/x/app/islands/A.client.tsx": `import { s } from "../stores/s";
function Row({ o }: { o: { n: number } }) { return <b>{o.n}</b>; }
export default function A() { return <Row o={s} />; }`}, "not assignable"},
		{"helper component of a client module used from a page", map[string]string{"/x/app/islands/A.client.tsx": `export function Row() { return <b>1</b>; }
export default function A() { return <Row />; }`, "/x/app/pages/index.server.tsx": `import type { PageProps } from "gotsx";
import { Row } from "../islands/A.client";
export default function P(props: PageProps) { return <Row />; }`}, "helper component of a client module, not an island"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			msg, failed := compileManyErr(tc.files)
			if !failed {
				t.Fatalf("应报错(包含 %q)", tc.want)
			}
			if !strings.Contains(msg, tc.want) {
				t.Errorf("错误信息应包含 %q, 实际: %s", tc.want, msg)
			}
			if !posRe.MatchString(msg) {
				t.Errorf("错误应带位置: %s", msg)
			}
		})
	}
}

// A module-level helper that reads a store has no render context on the server: it is browser-only (no Go definition)
func TestStoreHelperIsClientOnly(t *testing.T) {
	gs, js := compileMany(t, storeApp(map[string]string{"/x/app/stores/cart.client.tsx": `import { createStore } from "gotsx";
import type { Model } from "host:data";
export interface CartState { count: number; items: Model[]; open: boolean; note: string }
export const cart = createStore<CartState>({ count: 0, items: [] });
export function addItem(m: Model) { cart.set((s) => { s.items.push(m); s.count += 1; }); }
export function total(): number { return cart.count * 2; }
export function label(n: number): string { return "n=" + n; }
`}))
	store := gs["/x/app/stores/cart.client.tsx"]
	if strings.Contains(store, "func total") {
		t.Errorf("total reads the store: Go should not define it\n%s", store)
	}
	if !strings.Contains(store, "func label(n float64) string") {
		t.Errorf("label is plain: Go should define it\n%s", store)
	}
	if !strings.Contains(js["/x/app/stores/cart.client.tsx"], "return (cart.count() * 2);") {
		t.Errorf("JS total 应读 signal\n%s", js["/x/app/stores/cart.client.tsx"])
	}
}

// seed from a layout: LayoutProps embeds PageProps, the same gotsx.Seed(props, …) call
func TestStoreSeedInLayout(t *testing.T) {
	gs, _ := compileMany(t, storeApp(map[string]string{"/x/app/pages/_layout.server.tsx": `import type { LayoutProps } from "gotsx";
import { seed } from "gotsx";
import { cart } from "../stores/cart";
export default function Root({ children, cookies }: LayoutProps) {
  if (cookies.sid !== "") seed(cart, { count: 1, items: [], open: false, note: cookies.sid });
  return <html><body>{children}</body></html>;
}`}))
	if !strings.Contains(gs["/x/app/pages/_layout.server.tsx"], `gotsx.Seed(props, stores_cart_cart, CartState{Count: 1, Items: []host.Model{}, Open: false, Note: cookies["sid"]})`) {
		t.Errorf("layout Go 缺 Seed\n%s", gs["/x/app/pages/_layout.server.tsx"])
	}
}
