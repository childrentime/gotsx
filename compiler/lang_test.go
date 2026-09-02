package compiler

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// 0.5 语言长尾: 循环 / switch / 原地修改 / 可选对象真值 / redirect / catch-all —— 断言两个后端的生成结构
func TestLongTailServerCodegen(t *testing.T) {
	cases := []struct {
		name string
		body string
		want []string
	}{
		{"while", `let i = 0; while (i < 3) { i++; } return <b>{i}</b>;`, []string{"for (i < 3) {", "i++"}},
		{"for经典", `let s = 0; for (let i = 0; i < 3; i++) { s += i; } return <b>{s}</b>;`, []string{"for i := float64(0); (i < 3); i++ {", "s += i"}},
		{"break/continue", `let c = 0; for (const x of [1, 2, 3]) { if (x === 2) continue; if (x === 3) break; c++; } return <b>{c}</b>;`, []string{"continue", "break"}},
		{"switch合并空case+fallthrough", `const k = "a"; let r = ""; switch (k) { case "a": case "b": r = "ab"; break; case "c": r = "c"; default: r = "d"; } return <b>{r}</b>;`,
			[]string{`case "a", "b":`, "fallthrough", "default:"}},
		{"switch(true)", `const n = 5; let g = ""; switch (true) { case n > 3: g = "big"; break; default: g = "small"; } return <b>{g}</b>;`, []string{"switch true {", "case (n > 3):"}},
		{"push", `let xs = [1]; xs.push(2, 3); return <b>{xs.length}</b>;`, []string{"gotsx.Push(&xs, 2, 3)"}},
		{"pop", `let xs = [1]; const v = xs.pop(); return <b>{v}</b>;`, []string{"gotsx.Pop(&xs)"}},
		{"shift/unshift", `let xs = ["a"]; const v = xs.shift() ?? ""; xs.unshift(v); return <b>{xs.length}</b>;`, []string{"gotsx.Shift(&xs)", "gotsx.Unshift(&xs, v)"}},
		{"splice", `let xs = [1, 2, 3]; const r = xs.splice(1, 1); return <b>{r.length}</b>;`, []string{"gotsx.Splice(&xs, 1, 1)"}},
		{"splice省略deleteCount", `let xs = [1, 2, 3]; xs.splice(1); return <b>{xs.length}</b>;`, []string{"gotsx.Splice(&xs, 1, -1)"}},
		{"字段上的push", `const o = { xs: [1] }; o.xs.push(2); return <b>{o.xs.length}</b>;`, []string{"gotsx.Push(&o.Xs, 2)"}},
		{"findIndex", `const xs = [1, 2]; return <b>{xs.findIndex((x) => x === 2)}</b>;`, []string{"gotsx.FindIndex(xs"}},
		{"lastIndexOf", `const xs = [1, 2, 1]; return <b>{xs.lastIndexOf(1)}</b>;`, []string{"gotsx.LastIndexOf(xs, 1)"}},
		{"localeCompare", `const xs = ["b", "a"]; const ys = xs.sort((a, b) => a.localeCompare(b)); return <b>{ys.join(",")}</b>;`, []string{"gotsx.Compare(a, b)"}},
		{"slice无参", `const xs = [1]; const ys = xs.slice(); return <b>{ys.length}</b>;`, []string{"gotsx.Slice(xs, 0)"}},
		{"number.toString", `const n = 5; return <b>{n.toString()}</b>;`, []string{"gotsx.Num(n)"}},
		{"string方法", `const s = " x "; return <b>{s.trimStart()}{s.trimEnd()}{s.at(-1)}{s.lastIndexOf("x")}</b>;`,
			[]string{"gotsx.TrimStart(s)", "gotsx.TrimEnd(s)", "gotsx.StrAt(s, -1)", "gotsx.StrLastIndexOf(s"}},
		{"++语句", `let n = 0; n++; --n; return <b>{n}</b>;`, []string{"n++", "n--"}},
		{"++表达式", `let n = 0; const a = n++; const b = ++n; return <b>{a + b}</b>;`, []string{"func() float64 { v := n; n++; return v }()", "func() float64 { n++; return n }()"}},
		{"%=", `let n = 7; n %= 3; return <b>{n}</b>;`, []string{"n = gotsx.Mod(n, 3)"}},
		{"字符串+=", `let s = "a"; const n = 1; s += n; return <b>{s}</b>;`, []string{"s += gotsx.Num(n)"}},
		{"下标赋值", `let xs = [1]; xs[0] = 5; return <b>{xs[0]}</b>;`, []string{"xs[int(0)] = 5"}},
		{"Record赋值", `const m: Record<string, number> = {}; m["a"] = 1; m.b = 2; return <b>{m.a}</b>;`, []string{`m["a"] = 1`, `m["b"] = 2`}},
		{"find真值", `const xs = [{ id: "a" }]; const f = xs.find((x) => x.id === "a"); return <b>{f ? "y" : "n"}{!f && <i>none</i>}</b>;`, []string{"gotsx.NonZero(f)", "!(gotsx.NonZero(f))"}},
		{"find===undefined", `const xs = [{ id: "a" }]; const f = xs.find((x) => x.id === "a"); return <b>{f === undefined ? "n" : "y"}{f !== undefined ? "y" : "n"}</b>;`, []string{"gotsx.IsZero(f)", "!gotsx.IsZero(f)"}},
		{"find??", `const xs = [{ id: "a" }]; const f = xs.find((x) => x.id === "a") ?? { id: "z" }; return <b>{f.id}</b>;`, []string{"gotsx.NonZero(l)"}},
		{"interface extends", `interface A { a: string } interface B extends A { b: number } const x: B = { a: "x", b: 1 }; return <b>{x.a}{x.b}</b>;`, []string{"A string `json:\"a\"`", "B float64 `json:\"b\"`"}},
		{"map继承期望类型", `interface P { id: number } const xs = [1, 2]; const ps: P[] = xs.map((n) => ({ id: n })); return <b>{ps.length}</b>;`, []string{"[]c_P"}},
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

func TestRedirectNotFound(t *testing.T) {
	src := `import type { PageProps } from "gotsx";
export default function P({ query, params }: PageProps) {
  if (query.go !== "") return redirect(query.go);
  if (query.perm !== "") return redirect("/new", 301);
  if (params.id === "") notFound();
  return <b>ok</b>;
}`
	gs, _ := compileOne(t, "p.server.tsx", src)
	for _, w := range []string{`gotsx.Redirect(query["go"], 302)`, `gotsx.Redirect("/new", 301)`, "gotsx.NotFound()"} {
		if !strings.Contains(gs, w) {
			t.Errorf("缺 %s\n%s", w, gs)
		}
	}
}

func TestLongTailClientCodegen(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want []string
	}{
		{"keyed each", `import { useState } from "gotsx";
export default function C() { const [xs, setXs] = useState<string[]>([]); return <ul>{xs.map((x) => <li key={x}>{x}</li>)}</ul>; }`,
			[]string{"G.each(() => xs(), (x) => G.el(\"li\"", ", (x) => x)"}},
		{"keyed each 对象+块体", `import { useState } from "gotsx";
interface Item { id: number; text: string }
export default function C() { const [xs, setXs] = useState<Item[]>([]); return <ul>{xs.map((x, i) => { const y = x; return <li key={y.id}>{y.text}</li>; })}</ul>; }`,
			[]string{", (x, i) => y.id)"}},
		{"无key不传keyFn", `import { useState } from "gotsx";
export default function C() { const [xs, setXs] = useState<string[]>([]); return <ul>{xs.map((x) => <li>{x}</li>)}</ul>; }`,
			[]string{"G.el(\"li\", null, () => [G.t(String(x))]))"}},
		{"??与Go同语义", `import { useState } from "gotsx";
export default function C() { const [q, setQ] = useState(""); const r = q ?? "x"; const n = 0; const m = n ?? 5; return <b>{r}{m}</b>; }`,
			[]string{`(q() || "x")`, "(n || 5)"}},
		{"??对象保持", `import { useState } from "gotsx";
interface P { id: string }
export default function C() { const [xs, setXs] = useState<P[]>([]); const f = xs.find((x) => x.id === "a") ?? { id: "z" }; return <b>{f.id}</b>; }`,
			[]string{`?? { id: "z" }`}},
		{"while/for/switch", `import { useState } from "gotsx";
export default function C() { const [k, setK] = useState("a"); let i = 0; while (i < 3) { i++; } for (let j = 0; j < 2; j++) { if (j === 1) break; } let r = ""; switch (k) { case "a": r = "A"; break; default: r = "?"; } return <b>{r}</b>; }`,
			[]string{"while ((i < 3)) {", "for (let j = 0; (j < 2); j++) {", "switch (k()) {", "case \"a\":", "default:", "break;"}},
		{"localeCompare→cmp", `import { useState } from "gotsx";
export default function C() { const [xs, setXs] = useState<string[]>([]); const ys = xs.sort((a, b) => a.localeCompare(b)); return <b>{ys.length}</b>; }`,
			[]string{"G.cmp(a, b)"}},
		{"push在普通变量上", `import { useState } from "gotsx";
export default function C() { const [n, setN] = useState(0); const xs: number[] = []; for (let i = 0; i < n; i++) xs.push(i); return <b>{xs.length}</b>; }`,
			[]string{"xs.push(i)"}},
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

// 新语法的围栏
func TestLongTailFence(t *testing.T) {
	cases := []struct {
		name, file, src, want string
	}{
		{"do-while", "c.server.tsx", `export default function C() { do { } while (true); return <b>1</b>; }`, "do-while"},
		{"for-in", "c.server.tsx", `export default function C() { const o: Record<string, number> = {}; for (const k in o) { } return <b>1</b>; }`, "for-in"},
		{"break在循环外", "c.server.tsx", `export default function C() { break; return <b>1</b>; }`, "break is only allowed"},
		{"continue在switch里", "c.server.tsx", `export default function C() { const k = "a"; switch (k) { case "a": continue; } return <b>1</b>; }`, "continue is only allowed"},
		{"push改state", "c.client.tsx", `import { useState } from "gotsx";
export default function C() { const [xs, setXs] = useState<number[]>([]); xs.push(1); return <b>{xs.length}</b>; }`, "cannot mutate state"},
		{"push在临时值上", "c.server.tsx", `export default function C() { [1].push(2); return <b>1</b>; }`, "must be called on a variable"},
		{"++给const", "c.server.tsx", `export default function C() { const n = 0; n++; return <b>{n}</b>; }`, "cannot be modified with"},
		{"++给字符串", "c.server.tsx", `export default function C() { let s = "a"; s++; return <b>{s}</b>; }`, "only works on number"},
		{"%=给字符串", "c.server.tsx", `export default function C() { let s = "a"; s %= 2; return <b>{s}</b>; }`, "only works on number"},
		{"redirect在客户端", "c.client.tsx", `export default function C() { return redirect("/"); }`, "server components"},
		{"notFound带参数", "c.server.tsx", `export default function C() { notFound(1); return <b>1</b>; }`, "takes no arguments"},
		{"extends未知类型", "c.server.tsx", `interface B extends Nope { b: number }
export default function C() { return <b>1</b>; }`, "not a known object type"},
		{"switch两个default", "c.server.tsx", `export default function C() { const k = 1; switch (k) { default: break; default: break; } return <b>1</b>; }`, "only one default"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			msg, failed := compileErr(tc.file, tc.src)
			if !failed {
				t.Fatalf("期望编译失败, 但通过了")
			}
			if !strings.Contains(msg, tc.want) {
				t.Errorf("错误信息应含 %q, 实际: %s", tc.want, msg)
			}
			if !posRe.MatchString(msg) {
				t.Errorf("错误信息应带 文件:行:列 位置, 实际: %s", msg)
			}
		})
	}
}

func TestRouteOfCatchAll(t *testing.T) {
	pages := filepath.Join("x", "pages")
	pat, segs := routeOf(pages, filepath.Join(pages, "docs", "[...slug].server.tsx"))
	if pat != "/docs/{...slug}" || len(segs) != 2 || segs[1] != "{...slug}" {
		t.Errorf("catch-all 路由: %q %v", pat, segs)
	}
	pat, segs = routeOf(pages, filepath.Join(pages, "p", "[id].server.tsx"))
	if pat != "/p/{id}" || segs[1] != "{id}" {
		t.Errorf("参数路由: %q %v", pat, segs)
	}
	pat, segs = routeOf(pages, filepath.Join(pages, "index.server.tsx"))
	if pat != "/" || segs != nil {
		t.Errorf("首页: %q %v", pat, segs)
	}
}

// Analyze: 只检查不落盘, 诊断带位置, 编辑器缓冲区(overlay)优先于磁盘
func TestAnalyzeWithOverlay(t *testing.T) {
	root := t.TempDir()
	app := filepath.Join(root, "app")
	pages := filepath.Join(app, "pages")
	os.MkdirAll(pages, 0o755)
	good := `import type { PageProps } from "gotsx";
export default function Home({ query }: PageProps) { return <b>{query.q ?? ""}</b>; }
`
	bad := `import type { PageProps } from "gotsx";
export default function Home({ query }: PageProps) { const x = ghost; return <b>{x}</b>; }
`
	file := filepath.Join(pages, "index.server.tsx")
	os.WriteFile(file, []byte(bad), 0o644)

	diags, err := Analyze(app, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(diags) != 1 || diags[0].File != file || diags[0].Line != 2 || diags[0].Col == 0 || !strings.Contains(diags[0].Msg, "ghost") {
		t.Fatalf("期望一条指向第 2 行的诊断, 得 %+v", diags)
	}
	// 编辑器里已改好(未保存): overlay 覆盖磁盘
	diags, err = Analyze(app, map[string]string{file: good})
	if err != nil {
		t.Fatal(err)
	}
	if len(diags) != 0 {
		t.Fatalf("overlay 修好后不应有诊断, 得 %+v", diags)
	}
	// 页面约束也在 Analyze 里报(磁盘上的 index 先修好)
	os.WriteFile(file, []byte(good), 0o644)
	os.WriteFile(filepath.Join(pages, "about.server.tsx"), []byte(`export function About() { return <b>x</b>; }`), 0o644)
	diags, _ = Analyze(app, nil)
	found := false
	for _, d := range diags {
		if strings.HasSuffix(d.File, "about.server.tsx") && strings.Contains(d.Msg, "export default") {
			found = true
		}
	}
	if !found {
		t.Errorf("about.server.tsx 缺 default 导出应被诊断: %+v", diags)
	}
}

// keyed 列表与非 keyed 列表的服务端标记一致(key 只影响客户端复用策略)
func TestKeyedMarkerParity(t *testing.T) {
	src := `import { useState } from "gotsx";
export default function C() { const [xs, setXs] = useState<string[]>([]); const [n, setN] = useState(0);
  return <ul>{xs.map((x) => <li key={x}><input /><b>{x}</b>{n > 0 && <i>!</i>}</li>)}</ul>; }`
	gs, js := compileOne(t, "C.client.tsx", src)
	if strings.Count(gs, "_ctx.ListStart(")+strings.Count(gs, "gotsx.Nodes(") != 1 || strings.Count(js, "G.each(") != 1 {
		t.Errorf("列表标记不一致\n%s\n%s", gs, js)
	}
	if strings.Count(gs, "gotsx.If(") != 1 || strings.Count(js, "G.cond(") != 1 {
		t.Errorf("嵌套条件标记不一致\n%s\n%s", gs, js)
	}
	if strings.Contains(gs, "key") {
		t.Errorf("key 不应出现在 Go 输出里\n%s", gs)
	}
}

// Suspense: fallback 进外壳, children 进 thunk; 只允许服务端组件
func TestSuspenseCodegen(t *testing.T) {
	src := `import type { PageProps } from "gotsx";
import { Suspense } from "gotsx";
import { models } from "host:data";
export default function P({ params }: PageProps) {
  return <main><Suspense fallback={<i>loading</i>}><b>{models.list().length}</b><em>x</em></Suspense></main>;
}`
	gs, _ := compileOne(t, "p.server.tsx", src)
	for _, w := range []string{"gotsx.Suspense(func(_ctx *gotsx.Ctx) {", `_ctx.W("<i>loading</i>")`, "func() gotsx.Node {", `_ctx.W("<b>")`, "host.Data.Models.List()"} {
		if !strings.Contains(gs, w) {
			t.Errorf("缺 %s\n%s", w, gs)
		}
	}
	if msg, failed := compileErr("c.client.tsx", `import { Suspense } from "gotsx";
export default function C() { return <Suspense fallback={<i>x</i>}><b>y</b></Suspense>; }`); !failed || !strings.Contains(msg, "server components") {
		t.Errorf("客户端 Suspense 应报错: %v %s", failed, msg)
	}
	if msg, failed := compileErr("c.server.tsx", `import { Suspense } from "gotsx";
export default function C() { return <Suspense><b>y</b></Suspense>; }`); !failed || !strings.Contains(msg, "missing prop") {
		t.Errorf("缺 fallback 应报错: %v %s", failed, msg)
	}
}

// 文件约定: _layout 包住页面(嵌套外层在外), _404 / _error 生成 gen.NotFound / gen.ErrorPage, 其它 _ 文件忽略
func TestLayoutConventions(t *testing.T) {
	root := t.TempDir()
	app := filepath.Join(root, "app")
	pages := filepath.Join(app, "pages")
	os.MkdirAll(filepath.Join(pages, "docs"), 0o755)
	write := func(rel, src string) { os.WriteFile(filepath.Join(pages, rel), []byte(src), 0o644) }
	write("_layout.server.tsx", `import type { LayoutProps } from "gotsx";
export default function Root({ children, path }: LayoutProps) { return <html><body data-path={path}>{children}</body></html>; }`)
	write("docs/_layout.server.tsx", `import type { LayoutProps } from "gotsx";
export default function Docs({ children }: LayoutProps) { return <section>{children}</section>; }`)
	write("index.server.tsx", `import type { PageProps } from "gotsx";
export default function Home({ query }: PageProps) { return <h1>home</h1>; }`)
	write("docs/[...slug].server.tsx", `import type { PageProps } from "gotsx";
export default function Doc({ params }: PageProps) { return <h1>{params.slug}</h1>; }`)
	write("_404.server.tsx", `import type { PageProps } from "gotsx";
export default function NotFound({ path }: PageProps) { return <h1>404 {path}</h1>; }`)
	write("_error.server.tsx", `import type { ErrorProps } from "gotsx";
export default function Oops({ message }: ErrorProps) { return <h1>{message}</h1>; }`)
	write("_draft.server.tsx", `import type { PageProps } from "gotsx";
export default function Draft({ path }: PageProps) { return <h1>draft</h1>; }`)
	out := filepath.Join(root, "gen")
	rep, err := Build(Config{AppDir: app, OutDir: out, RuntimePkg: "rt", ClientFS: fakeClientFS()})
	if err != nil {
		t.Fatal(err)
	}
	if len(rep.Routes) != 2 {
		t.Fatalf("应只有两条路由(_ 文件不是路由): %v", rep.Routes)
	}
	routes, _ := os.ReadFile(filepath.Join(out, "routes_gen.go"))
	rs := string(routes)
	for _, w := range []string{
		`Render: func(p gotsx.PageProps) gotsx.Node { return pages__layout_Root(gotsx.LayoutProps{PageProps: p, Children: pages_index_Home(p)}) }`,
		`return pages__layout_Root(gotsx.LayoutProps{PageProps: p, Children: pages_docs__layout_Docs(gotsx.LayoutProps{PageProps: p, Children: pages_docs_____slug__Doc(p)})})`,
		`var NotFound func(gotsx.PageProps) gotsx.Node = func(p gotsx.PageProps) gotsx.Node { return pages__layout_Root(gotsx.LayoutProps{PageProps: p, Children: pages__404_NotFound(p)}) }`,
		`pages__error_Oops(gotsx.ErrorProps{PageProps: p, Message: err.Error()})`,
	} {
		if !strings.Contains(rs, w) {
			t.Errorf("routes_gen.go 缺 %s\n%s", w, rs)
		}
	}
	if strings.Contains(rs, "Draft") {
		t.Errorf("_draft 不应出现在路由里\n%s", rs)
	}
	// 约束: _layout 必须用 LayoutProps
	write("docs/_layout.server.tsx", `import type { PageProps } from "gotsx";
export default function Docs({ path }: PageProps) { return <section>{path}</section>; }`)
	diags, _ := Analyze(app, nil)
	found := false
	for _, d := range diags {
		if strings.Contains(d.Msg, "LayoutProps") {
			found = true
		}
	}
	if !found {
		t.Errorf("_layout 用错 props 应被诊断: %v", diags)
	}
}

// Record 语义: 缺席的键读出零值(两端一致), Object.hasOwn 判断存在, delete 删除
func TestRecordSemantics(t *testing.T) {
	gs, _ := compileOne(t, "c.server.tsx", `import { models } from "host:data";
export default function C() {
  const m: Record<string, number> = { a: 1 };
  const has = Object.hasOwn(m, "a");
  delete m["a"];
  delete m.b;
  return <b>{has ? "y" : "n"}{m.a}</b>;
}`)
	for _, w := range []string{`gotsx.HasKey(m, "a")`, `delete(m, "a")`, `delete(m, "b")`} {
		if !strings.Contains(gs, w) {
			t.Errorf("Go 缺 %s\n%s", w, gs)
		}
	}
	_, js := compileOne(t, "c.client.tsx", `import { useState } from "gotsx";
export default function C() {
  const [n, setN] = useState(0);
  const m: Record<string, string> = { a: "x" };
  const cnt: Record<string, number> = {};
  const has = Object.hasOwn(m, "a");
  m.b = "y";
  m["c"] = "z";
  cnt.k++;
  delete m.a;
  const miss = m.zzz === undefined;
  const q = m.a;
  return <b>{m.a}{cnt.k}{has ? "y" : "n"}{miss ? "m" : "-"}{q}{n}</b>;
}`)
	for _, w := range []string{`Object.hasOwn(m, "a")`, `m.b = "y"`, `m["c"] = "z"`, `cnt.k++`, `delete m.a`, `(m.a ?? "")`, `(cnt.k ?? 0)`, `((m.zzz ?? "") === "")`} {
		if !strings.Contains(js, w) {
			t.Errorf("JS 缺 %s\n%s", w, js)
		}
	}
	if strings.Contains(js, `(m.b ?? "") =`) || strings.Contains(js, `(cnt.k ?? 0)++`) {
		t.Errorf("赋值目标不应带零值补齐\n%s", js)
	}
	if msg, failed := compileErr("c.server.tsx", `export default function C() { const xs = [1]; delete xs[0]; return <b>1</b>; }`); !failed || !strings.Contains(msg, "Record key") {
		t.Errorf("delete 非 Record 应报错: %v %s", failed, msg)
	}
}

// 正则字面量(RE2 子集)与 Date 内建: 两端 codegen + 编译期校验
func TestRegexAndDate(t *testing.T) {
	src := `import { models } from "host:data";
export default function C() {
  const s = "a1b22";
  const re = /\d+/g;
  const ok = /^a/i.test(s);
  const digits = s.match(re);
  const first = s.match(/(\d)/);
  const parts = s.split(/\d+/);
  const idx = s.search(/b/);
  const r1 = s.replace(/\d/, "#");
  const r2 = s.replaceAll(/\d/g, "$&$&");
  const half = 10 / 2;
  const t0 = Date.now();
  const t1 = Date.parse("2026-01-02T03:04:05Z");
  const iso = isoDate(t1);
  return <b>{ok ? "y" : "n"}{digits.length}{first.length}{parts.length}{idx}{r1}{r2}{half}{t0 > 0 ? "t" : "f"}{iso}</b>;
}`
	gs, _ := compileOne(t, "c.server.tsx", src)
	for _, w := range []string{`gotsx.Re("\\d+", "g")`, `gotsx.Re("(?i)^a", "i")`, "gotsx.ReTest(", "gotsx.ReMatch(s, re)", `gotsx.ReSplit(s, gotsx.Re("\\d+", "")`, "gotsx.ReSearch(s", `gotsx.ReReplace(s, gotsx.Re("\\d", ""), "#")`, `"$&$&")`, "(10 / 2)", "gotsx.Now()", `gotsx.DateParse("2026-01-02T03:04:05Z")`, "gotsx.IsoDate(t1)"} {
		if !strings.Contains(gs, w) {
			t.Errorf("Go 缺 %s\n%s", w, gs)
		}
	}
	_, js := compileOne(t, "c.client.tsx", strings.Replace(src, `import { models } from "host:data";`, "", 1))
	for _, w := range []string{"const re = /\\d+/g;", "/^a/i.test(s)", "G.match(s, re)", "s.split(/\\d+/)", `s.replace(/\d/g, "$&$&")`, "Date.now()", "G.isoDate(t1)"} {
		if !strings.Contains(js, w) {
			t.Errorf("JS 缺 %s\n%s", w, js)
		}
	}
	for _, tc := range []struct{ src, want string }{
		{`export default function C() { const ok = /(?<=a)b/.test("ab"); return <b>{ok ? "y" : "n"}</b>; }`, "RE2"},
		{`export default function C() { const ok = /a/x.test("ab"); return <b>{ok ? "y" : "n"}</b>; }`, "flag"},
		{`export default function C() { const r = "aa".replaceAll(/a/, "b"); return <b>{r}</b>; }`, "g flag"},
		{`export default function C() { const d = Date.today(); return <b>{d}</b>; }`, "Date.now()"},
	} {
		if msg, failed := compileErr("c.server.tsx", tc.src); !failed || !strings.Contains(msg, tc.want) {
			t.Errorf("应报错含 %q: %v %s", tc.want, failed, msg)
		}
	}
}

// 编辑器查询: hover 给类型与说明, definition 跳到声明(含跨文件与宿主方法的 Go 源码位置)
func TestHoverAndDefinition(t *testing.T) {
	root := t.TempDir()
	app := filepath.Join(root, "app")
	os.MkdirAll(filepath.Join(app, "pages"), 0o755)
	os.MkdirAll(filepath.Join(app, "components"), 0o755)
	os.MkdirAll(filepath.Join(app, ".gen"), 0o755)
	os.WriteFile(filepath.Join(app, ".gen", "host.json"), testHostJSON, 0o644)
	os.WriteFile(filepath.Join(app, ".gen", "host.d.ts"), []byte("declare module \"host:data\" {\n  export interface Model { id: string }\n  export const models: Store;\n}\n"), 0o644)
	card := `import type { Model } from "host:data";
export default function Card({ model }: { model: Model }) { return <b>{model.title}</b>; }
`
	page := `import type { PageProps } from "gotsx";
import { useState } from "gotsx";
import { models } from "host:data";
import Card from "../components/Card.server";

interface Row { id: string; n: number }

export default function Home({ query }: PageProps) {
  const list = models.search(query.q ?? "");
  const rows: Row[] = [];
  const first = list.find((m) => m.id === "a");
  return <ul>{list.map((m) => <Card model={m} />)}<i>{rows.length}</i></ul>;
}
`
	cardFile := filepath.Join(app, "components", "Card.server.tsx")
	pageFile := filepath.Join(app, "pages", "index.server.tsx")
	os.WriteFile(cardFile, []byte(card), 0o644)
	os.WriteFile(pageFile, []byte(page), 0o644)
	c, diags, err := Load(app, nil)
	if err != nil || len(diags) != 0 {
		t.Fatalf("load: %v %v", err, diags)
	}
	at := func(line int, snippet, word string) (int, int) { // 找到片段所在列, 返回 word 的列
		src := page
		if line > 100 {
			src = card
		}
		ln := strings.Split(src, "\n")[line-1]
		i := strings.Index(ln, snippet)
		if i < 0 {
			t.Fatalf("行 %d 没有 %q", line, snippet)
		}
		return line, i + strings.LastIndex(snippet, word) + 1
	}
	show := func(h *Hover) string {
		if h == nil {
			return "<nil>"
		}
		return fmt.Sprintf("%q def=%+v", h.Text, h.Def)
	}
	// const list → 类型 + 定义在本行
	l, col := at(9, "const list", "list")
	h := c.HoverAt(pageFile, l, col)
	if h == nil || !strings.Contains(h.Text, "const list: Model[]") || h.Def == nil || h.Def.Line != 9 {
		t.Errorf("hover const: %s", show(h))
	}
	// models.search → 宿主方法 + Go 源码位置(hostgen 反射)
	l, col = at(9, "models.search(", "search")
	h = c.HoverAt(pageFile, l, col)
	if h == nil || !strings.Contains(h.Text, "(host method) Store.search(arg0: string): Model[]") || h.Def == nil || !strings.HasSuffix(h.Def.File, "helper_test.go") {
		t.Errorf("hover host method: %s", show(h))
	}
	// query.q → Record 键
	l, col = at(9, "query.q", "q")
	if h = c.HoverAt(pageFile, l, col); h == nil || !strings.Contains(h.Text, "(record key) q: string") {
		t.Errorf("hover record key: %s", show(h))
	}
	// <Card> → 组件签名 + 跨文件定义
	l, col = at(12, "<Card model", "Card")
	h = c.HoverAt(pageFile, l, col)
	if h == nil || !strings.Contains(h.Text, "component Card({ model: Model })") || h.Def == nil || h.Def.File != cardFile || h.Def.Line != 2 {
		t.Errorf("hover component: %s", show(h))
	}
	// m.id inside the find callback → 宿主字段 → host.d.ts
	l, col = at(11, "m.id", "id")
	h = c.HoverAt(pageFile, l, col)
	if h == nil || !strings.Contains(h.Text, "(field) Model.id: string") || h.Def == nil || !strings.HasSuffix(h.Def.File, "host.d.ts") || h.Def.Line != 2 {
		t.Errorf("hover host field: %s", show(h))
	}
	// 类型标注 Row[] → interface 定义
	l, col = at(10, "rows: Row[]", "Row")
	h = c.HoverAt(pageFile, l, col)
	if h == nil || !strings.Contains(h.Text, "interface Row {") || h.Def == nil || h.Def.Line != 6 {
		t.Errorf("hover type ref: %s", show(h))
	}
	// useState (import 行上的内建) → 文档
	l, col = at(2, "{ useState }", "useState")
	if h = c.HoverAt(pageFile, l, col); h == nil || !strings.Contains(h.Text, "useState<T>(init: T)") {
		t.Errorf("hover builtin on import line: %s", show(h))
	}
	// import 行的 Card → 跨文件定义
	l, col = at(4, "import Card", "Card")
	if d := c.DefinitionAt(pageFile, l, col); d == nil || d.File != cardFile {
		t.Errorf("definition via import: %+v", d)
	}
	if c.HoverAt(pageFile, 7, 1) != nil {
		t.Error("空行不应有 hover")
	}
}
