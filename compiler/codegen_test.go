package compiler

import (
	"strings"
	"testing"
)

// 服务端(Go 后端)代码生成
func TestServerCodegen(t *testing.T) {
	cases := []struct {
		name string
		body string // 放进组件函数体
		want []string
	}{
		{"字符串加号", `return <b>{"a" + "b"}</b>;`, []string{`gotsx.El("b"`}},
		{"数字转文本", `const n = 3; return <b>{n}</b>;`, []string{"var n float64 = 3", "gotsx.Text(gotsx.Num(n))"}},
		{"模板字符串", "const s = `hi ${1}`; return <b>{s}</b>;", []string{`"hi " + gotsx.Num(1)`}},
		{"??空合并", `const q = ""; const r = q ?? "x"; return <b>{r}</b>;`, []string{"gotsx.Or("}},
		{"三元", `const n = 1; return <b>{n > 0 ? "y" : "n"}</b>;`, []string{"if (n > 0)"}},
		{"数组map", `const xs = [1, 2]; return <ul>{xs.map((x) => <li>{x}</li>)}</ul>;`, []string{"gotsx.Map(xs", "gotsx.El(\"li\""}},
		{"filter", `const xs = [1, 2]; const ys = xs.filter((x) => x > 1); return <b>{ys.length}</b>;`, []string{"gotsx.Filter(xs"}},
		{"reduce", `const xs = [1, 2]; const s = xs.reduce((a, x) => a + x, 0); return <b>{s}</b>;`, []string{"gotsx.Reduce(xs"}},
		{"sort", `const xs = [3, 1]; const ys = xs.sort((a, b) => a - b); return <b>{ys.length}</b>;`, []string{"gotsx.Sort(xs"}},
		{"includes", `const xs = ["a"]; return <b>{xs.includes("a") ? "y" : "n"}</b>;`, []string{"gotsx.Includes(xs"}},
		{"join", `const xs = ["a", "b"]; return <b>{xs.join(",")}</b>;`, []string{"gotsx.Join(xs"}},
		{"字符串方法", `const s = "AB"; return <b>{s.toLowerCase()}</b>;`, []string{"gotsx.Lower(s)"}},
		{"越界安全下标", `const xs = [1]; return <b>{xs[5]}</b>;`, []string{"gotsx.At(xs, 5)"}},
		{"Object.keys", `const m: Record<string, number> = {}; return <b>{Object.keys(m).length}</b>;`, []string{"gotsx.ObjectKeys(m)"}},
		{"if语句", `const n = 1; if (n > 0) { return <b>+</b>; } return <b>-</b>;`, []string{"if (n > 0)"}},
		{"for-of", `const xs = [1]; for (const x of xs) { const y = x; } return <b>ok</b>;`, []string{"for _, x := range xs"}},
		{"条件子节点标记", `const n = 1; return <div>{n > 0 && <b>!</b>}</div>;`, []string{"gotsx.IfPlain"}},
		{"void元素", `return <input />;`, []string{`gotsx.El("input"`}},
		{"属性", `return <a href="/x" class="c">go</a>;`, []string{`gotsx.A("href", "/x")`, `gotsx.A("class", "c")`}},
		{"宿主调用", `return <b>{models.list().length}</b>;`, []string{"host.Data.Models.List()"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			src := "import { models } from \"host:data\";\nexport default function C() {\n" + tc.body + "\n}\n"
			gs, _ := compileOne(t, "c.server.tsx", src)
			for _, w := range tc.want {
				if !strings.Contains(gs, w) {
					t.Errorf("生成 Go 缺少 %q\n---\n%s", w, gs)
				}
			}
		})
	}
}

// 客户端(JS 后端)代码生成
func TestClientCodegen(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want []string
	}{
		{"useState→signal", `import { useState } from "gotsx";
export default function C() { const [n, setN] = useState(0); return <b>{n}</b>; }`,
			[]string{"const [n, setN] = G.signal(0)", "G.text(() => n())"}},
		{"依赖signal自动memo", `import { useState } from "gotsx";
export default function C() { const [n, setN] = useState(0); const d = n * 2; return <b>{d}</b>; }`,
			[]string{"const d = G.memo(() => (n() * 2))"}},
		{"useEffect", `import { useState, useEffect } from "gotsx";
export default function C() { const [n, setN] = useState(0); useEffect(() => { console.log(n); }); return <b>{n}</b>; }`,
			[]string{"G.effect("}},
		{"useEffect空依赖→onMount", `import { useState, useEffect } from "gotsx";
export default function C() { const [n, setN] = useState(0); useEffect(() => { console.log(1); }, []); return <b>{n}</b>; }`,
			[]string{"G.onMount("}},
		{"响应式列表each", `import { useState } from "gotsx";
export default function C() { const [xs, setXs] = useState<number[]>([]); return <ul>{xs.map((x) => <li>{x}</li>)}</ul>; }`,
			[]string{"G.each(() => xs()"}},
		{"响应式条件cond", `import { useState } from "gotsx";
export default function C() { const [n, setN] = useState(0); return <div>{n > 0 && <b>!</b>}</div>; }`,
			[]string{"G.cond("}},
		{"事件绑定", `import { useState } from "gotsx";
export default function C() { const [n, setN] = useState(0); return <button onClick={() => setN(n + 1)}>{n}</button>; }`,
			[]string{"onClick:"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, js := compileOne(t, "c.client.tsx", tc.src)
			for _, w := range tc.want {
				if !strings.Contains(js, w) {
					t.Errorf("生成 JS 缺少 %q\n---\n%s", w, js)
				}
			}
		})
	}
}

// 回归: JSON 键映射成合法 Go 字段(下划线 / 数字开头 / 连字符不能崩)
func TestWeirdJSONKeys(t *testing.T) {
	src := `import type { Node } from "gotsx";
interface Errs { _?: string; "2fa"?: string; "user-name"?: string; ok?: string; }
export default function C() {
  const e: Errs = {};
  return <div>{e._ ?? ""}{e.ok ?? ""}</div>;
}`
	// compileOne 内部跑 format.Source: 若字段名是 Go 空标识符 _ 会直接 Fatal
	gs, _ := compileOne(t, "c.client.tsx", src)
	if !strings.Contains(gs, "`json:\"_\"`") {
		t.Errorf("字段 _ 应保留 json tag\n%s", gs)
	}
	for _, bad := range []string{"\n\t_ string", "\n\t2fa "} {
		if strings.Contains(gs, bad) {
			t.Errorf("非法 Go 字段名: %q\n%s", bad, gs)
		}
	}
}

// 回归: 嵌套响应式三元的内层也要 G.cond(否则 memo 未结算时会错误固定分支)
func TestNestedReactiveConditional(t *testing.T) {
	src := `import { useState } from "gotsx";
export default function C() {
  const [loading, setLoading] = useState(true);
  const [items, setItems] = useState<number[]>([]);
  return <div>{loading ? <span>loading</span> : items.length === 0 ? <span>empty</span> : <ul>{items.map((x) => <li>{x}</li>)}</ul>}</div>;
}`
	_, js := compileOne(t, "c.client.tsx", src)
	// 两层 cond 都必须是 G.cond
	if strings.Count(js, "G.cond(") < 2 {
		t.Errorf("嵌套响应式三元的内层未包 G.cond\n%s", js)
	}
	if !strings.Contains(js, "G.each(") {
		t.Errorf("列表未编成 G.each\n%s", js)
	}
}

// 回归: 组件根返回条件(如模态框 open ? A : B)必须响应式
func TestRootConditionalReactive(t *testing.T) {
	src := `import { useState } from "gotsx";
export default function Modal() {
  const [open, setOpen] = useState(false);
  return open ? <div class="m">hi</div> : <div></div>;
}`
	_, js := compileOne(t, "Modal.client.tsx", src)
	if !strings.Contains(js, "return G.cond(") {
		t.Errorf("组件根条件返回未包 G.cond\n%s", js)
	}
}

// 标记一致性: Go SSR 的 hydrate 标记必须与客户端结构逐一对应(否则 hydrate 错位)。
// Go: Dyn↔text, Nodes↔each, If/Marked↔cond。这是当初 nodeExpr 继承 reactive 引入的 bug 的护栏。
func TestMarkerParity(t *testing.T) {
	snippets := []string{
		// 嵌套响应式三元
		`import { useState } from "gotsx";
export default function C() { const [a, setA] = useState(true); const [xs, setXs] = useState<number[]>([]);
  return <div>{a ? <b>x</b> : xs.length === 0 ? <i>empty</i> : <ul>{xs.map((x) => <li>{x}</li>)}</ul>}</div>; }`,
		// 响应式条件里套非响应式列表(骨架屏模式)
		`import { useState } from "gotsx";
export default function C() { const [load, setLoad] = useState(true);
  return <div>{load && [0, 1, 2].map(() => <span>s</span>)}</div>; }`,
		// 响应式文本 + 响应式列表混合
		`import { useState } from "gotsx";
export default function C() { const [n, setN] = useState(0); const [xs, setXs] = useState<string[]>([]);
  return <div>{n}{xs.map((x) => <b>{x}</b>)}{n > 0 && <i>!</i>}</div>; }`,
	}
	count := func(s string, subs []string) int {
		n := 0
		for _, sub := range subs {
			n += strings.Count(s, sub)
		}
		return n
	}
	// DOM 层面: Go 的 Nodes/If/Marked 与 JS 的 each/cond 都发同一种块标记 <!--[-->...<!--]-->;
	// 只有 Dyn / text 发 <!--$-->。所以按"块标记"和"文本标记"两类各自对齐即可。
	for i, src := range snippets {
		gs, js := compileOne(t, "C.client.tsx", src)
		goText := count(gs, []string{"gotsx.Dyn("})
		jsText := count(js, []string{"G.text("})
		goBlock := count(gs, []string{"gotsx.Nodes(", "gotsx.If(", "gotsx.Marked("})
		jsBlock := count(js, []string{"G.each(", "G.cond("})
		if goText != jsText || goBlock != jsBlock {
			t.Errorf("片段 %d 标记不一致: 文本 %d/%d 块 %d/%d\n=== Go ===\n%s\n=== JS ===\n%s",
				i, goText, jsText, goBlock, jsBlock, gs, js)
		}
	}
}

// SEO 用得到: @ 开头 / 特殊符号的 JSON-LD 键要映射成合法 Go 字段
func TestJsonLdKeys(t *testing.T) {
	src := `import type { PageProps } from "gotsx";
export default function P({ params }: PageProps) {
  const ld = { "@context": "https://schema.org", "@type": "Product", name: params.id };
  return <head>{jsonLd(JSON.stringify(ld))}</head>;
}`
	gs, _ := compileOne(t, "p.server.tsx", src)
	if !strings.Contains(gs, "`json:\"@context\"`") || !strings.Contains(gs, "`json:\"@type\"`") {
		t.Errorf("JSON-LD @ 键未保留 json tag\n%s", gs)
	}
}

// i18n 内建映射到运行时
func TestI18nBuiltins(t *testing.T) {
	src := `import type { PageProps } from "gotsx";
export default function P({ locale }: PageProps) {
  const vars: Record<string, string> = { name: "x" };
  return <div>{t(locale, "hi")}{tv(locale, "hi", vars)}{plural(locale, "items", 3)}{fmtCur(locale, 100)}{lpath(locale, "/a")}</div>;
}`
	gs, _ := compileOne(t, "p.server.tsx", src)
	for _, w := range []string{"gotsx.Tr(", "gotsx.Trv(", "gotsx.Plural(", "gotsx.FmtCur(", "gotsx.LPath("} {
		if !strings.Contains(gs, w) {
			t.Errorf("缺 %s\n%s", w, gs)
		}
	}
}
