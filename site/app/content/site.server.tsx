import type { QA } from "../islands/Accordion.client";
import type { Row } from "../islands/SyntaxSearch.client";
import { loc } from "../ui/i18n";

export interface Feature {
  icon: string;
  title: string;
  desc: string;
}

export function features(locale: string): Feature[] {
  return [
    { icon: "⚙️", title: loc(locale, "Compiled to Go, not interpreted", "编译到 Go, 不是解释执行"), desc: loc(locale, "Server components are real Go functions: JSX becomes direct string writes, useState is the initial value, useEffect generates nothing. A page renders in tens of microseconds.", "服务端组件是真正的 Go 函数: JSX 编成字符串直写, useState 是初值, useEffect 什么都不生成。一个页面几十微秒。") },
    { icon: "⚡", title: loc(locale, "signals on the client", "signals 客户端"), desc: loc(locale, "The same .client.tsx compiles to signals + fine-grained DOM binding (the Solid / Svelte 5 model), no vdom, a 6KB runtime.", "同一份 .client.tsx 编成 signals + 精确 DOM 绑定(Solid / Svelte 5 的模型), 没有 vdom, 运行时 6KB。") },
    { icon: "🧭", title: loc(locale, "Resumable hydration", "走位 hydrate"), desc: loc(locale, "The server only marks the reactive text/conditions/lists; the client claims nodes in the same structural order the same compiler produced, no diffing, reusing existing DOM.", "服务端只在响应式的文本/条件/列表上留标记, 客户端按同一编译器给出的结构顺序认领节点, 不需要 diff, 复用现有 DOM。") },
    { icon: "🔌", title: loc(locale, "Host modules", "宿主模块"), desc: loc(locale, "import { models } from \"host:data\" is backed by Go; after compilation it's a direct call, zero marshalling; types are reflected from Go.", "import { models } from \"host:data\" 背后是 Go, 编译后是直接调用, 零编组; 类型由 Go 反射生成。") },
    { icon: "🏝️", title: loc(locale, "Islands + SPA navigation", "岛 + SPA 跳转"), desc: loc(locale, "Pages ship zero JS; islands load on demand; navigation = fetch HTML → morph, islands survive by DOM identity so state isn't lost.", "页面零 JS, 岛按需加载; 跳转 = 拉 HTML → morph, 岛按 DOM 同一性存活, 状态不丢。") },
    { icon: "🚧", title: loc(locale, "The fence is a compile error", "围栏是编译错误"), desc: loc(locale, "The client touching host:*, a missing prop, member access on any, using ==… are all errors with file:line:col, never a silent miscompile.", "客户端碰 host:*、缺 prop、any 上取成员、用了 ==……全部是带文件:行:列的错误, 不会静默编错。") },
    { icon: "🎨", title: loc(locale, "Tailwind", "Tailwind"), desc: loc(locale, "class is just a string; the Tailwind standalone CLI scans .tsx at build time to generate CSS, also without Node. This site is proof.", "class 就是字符串, Tailwind standalone CLI 构建期扫描 .tsx 生成 CSS, 同样不需要 Node。这个站就是。") },
    { icon: "🔁", title: loc(locale, "gotsx dev", "gotsx dev"), desc: loc(locale, "Edit TSX → recompile → go build → restart, ~2s; the old version keeps running while a compile fails.", "改 TSX → 重新编译 → go build → 重启, 约 2 秒; 编译失败时旧版本继续跑。") },
  ];
}

export function faqs(locale: string): QA[] {
  return [
    { q: loc(locale, "Is this React? Can I use React component libraries from npm?", "这是 React 吗? 能用 npm 上的 React 组件库吗?"), a: loc(locale, "No. gotsx borrows React's ideas and the TSX feel; the runtime is its own. React libraries on npm (including MUI) depend on the React runtime and dynamic JS — they can't pass the subset check and can't compile to Go. Component libraries are written in the dialect — like this site's.", "不是。gotsx 借的是 React 的思想和 TSX 的手感, 运行时是自己的。npm 上的 React 库(包括 MUI)依赖 React 运行时和动态 JS, 不能通过子集检查, 也不可能编译成 Go。组件库要用方言写——就像这个站的组件库。") },
    { q: loc(locale, "Why not run JS in Go with goja / V8?", "为什么不用 goja / V8 在 Go 里跑 JS?"), a: loc(locale, "Tried it. React + MUI in goja was 50ms for one list page, and V8 brings in CGO. But SSR is a single synchronous one-pass evaluation — its semantics are small enough to compile straight to Go, so the list page becomes 30µs and go test / pprof / delve all work.", "做过。React + MUI 在 goja 里一个列表页 50ms, 换 V8 也要引入 CGO。而 SSR 是一次同步的单趟求值, 它需要的语义小到可以直接编译成 Go——于是列表页变成 30µs, 而且 go test / pprof / delve 全能用。") },
    { q: loc(locale, "How small is the subset? How much must I change?", "子集有多小? 我的代码要改多少?"), a: loc(locale, "Common component code barely changes: function components, props, destructuring, map/filter, template strings, ternaries, &&, useState/useEffect are all there. Missing: class, this, member access on any, ==, custom generics, while/switch. Outside the subset is a compile error, not a silent behavior difference. The language page has the full table.", "常见的组件代码几乎不用改: 函数组件、props、解构、map/filter、模板字符串、三元、&&、useState/useEffect 都有。没有的是 class、this、any 上的成员访问、==、自定义泛型、while/switch。出了子集是编译错误, 不是静默行为差异。语言参考页有完整表。") },
    { q: loc(locale, "What's the client update model?", "客户端的更新模型是什么?"), a: loc(locale, "useState compiles to a signal, a signal-dependent const auto-compiles to a memo, and JSX text/attrs/conditions/lists that read a signal each bind one effect. There's no \"re-run the whole function and diff.\"", "useState 编译成 signal, 依赖 signal 的 const 自动编译成 memo, 读到 signal 的 JSX 文本/属性/条件/列表各自绑一个 effect。没有\"重跑整个函数再 diff\"这回事。") },
    { q: loc(locale, "How does data come in?", "数据怎么进来?"), a: loc(locale, "Server components call host modules (Go) synchronously — no async/await, concurrency comes from goroutines. Islands take props (JSON in an HTML attribute); to get back to Go they use an HTTP action.", "服务端组件同步调用宿主模块(Go), 没有 async/await——并发由 goroutine 提供。岛拿 props(JSON 进 HTML 属性), 要回到 Go 就走 HTTP action。") },
    { q: loc(locale, "Is it production-ready?", "生产可用吗?"), a: loc(locale, "It's maturing: compiler + two backends + runtime + dev loop + four real apps, with production HTTP hardening, single-binary deploy, SEO, images, telemetry, PWA and i18n. Still missing: an LSP, keyed list diffing, more builtins, streaming, and an independent security audit.", "在成熟中: 编译器 + 两个后端 + 运行时 + dev 循环 + 四个真实应用, 有生产 HTTP 加固、单二进制部署、SEO、图片、遥测、PWA、i18n。缺的是 LSP、keyed 列表 diff、更多内建、流式渲染和独立安全审计。") },
  ];
}

function cat(locale: string, key: string): string {
  if (key === "module") return loc(locale, "Modules", "模块");
  if (key === "stmt") return loc(locale, "Statements", "语句");
  if (key === "expr") return loc(locale, "Expressions", "表达式");
  if (key === "jsx") return "JSX";
  if (key === "type") return loc(locale, "Types", "类型");
  if (key === "hooks") return "Hooks";
  if (key === "builtin") return loc(locale, "Builtins", "内建");
  if (key === "array") return loc(locale, "Arrays", "数组");
  return loc(locale, "Strings", "字符串");
}

export function syntaxRows(locale: string): Row[] {
  const c = (k: string) => cat(locale, k);
  return [
    { cat: c("module"), syntax: "import X from \"./a\"", note: loc(locale, "default import; extensions .tsx / .server.tsx / .client.tsx may be omitted", "默认导入; 路径可省 .tsx / .server.tsx / .client.tsx"), ok: true },
    { cat: c("module"), syntax: "import { a, b as c } from \"./a\"", note: loc(locale, "named imports", "命名导入"), ok: true },
    { cat: c("module"), syntax: "import type { T } from \"./a\"", note: loc(locale, "type import; client code may only import type from host:*", "类型导入; 客户端对 host:* 只允许 import type"), ok: true },
    { cat: c("module"), syntax: "import { useState, useEffect, useMemo } from \"gotsx\"", note: loc(locale, "hooks", "hooks"), ok: true },
    { cat: c("module"), syntax: "import type { Node, PageProps } from \"gotsx\"", note: loc(locale, "framework types", "框架类型"), ok: true },
    { cat: c("module"), syntax: "import { x } from \"host:name\"", note: loc(locale, "host module, server only; compiles to a direct Go call", "宿主模块, 仅服务端; 编译成直接 Go 调用"), ok: true },
    { cat: c("module"), syntax: "export default function / export function", note: loc(locale, "component (capitalized) or plain function", "组件(大写)或普通函数"), ok: true },
    { cat: c("module"), syntax: "export const data: T[] = [...]", note: loc(locale, "module-level const → Go package var; no let", "模块级 const → Go 包级变量; 不允许 let"), ok: true },
    { cat: c("module"), syntax: "export interface / export type", note: loc(locale, "type export (all types can be import type'd)", "类型导出(所有类型自动可被 import type)"), ok: true },
    { cat: c("module"), syntax: "import * as ns from", note: loc(locale, "namespace import", "命名空间导入"), ok: false },
    { cat: c("stmt"), syntax: "const x = ... / let x = ...", note: loc(locale, "single declaration; type annotated or inferred", "单声明; 类型可标注可推断"), ok: true },
    { cat: c("stmt"), syntax: "const { a, b = 1 } = obj / const [x, setX] = ...", note: loc(locale, "destructuring + defaults (primitives, zero-value semantics)", "解构 + 默认值(原始类型, 零值语义)"), ok: true },
    { cat: c("stmt"), syntax: "if / else if / else", note: loc(locale, "JS truthiness (empty string, 0 are falsy)", "条件按 JS 真值语义(空串、0 为假)"), ok: true },
    { cat: c("stmt"), syntax: "for (const x of xs)", note: loc(locale, "only for-of over arrays", "只支持 for-of 遍历数组"), ok: true },
    { cat: c("stmt"), syntax: "return / throw", note: loc(locale, "throw is a panic on the Go side, recovered by the request layer", "throw 在 Go 侧是 panic, 请求层 recover"), ok: true },
    { cat: c("stmt"), syntax: "try / catch / finally", note: loc(locale, "full on the client; Go runs only the try and finally bodies", "客户端完整; Go 侧只执行 try 和 finally 体"), ok: true },
    { cat: c("stmt"), syntax: "function f() {} / async function f() {}", note: loc(locale, "nested functions; async only reaches the JS backend", "嵌套函数; async 只进 JS 后端"), ok: true },
    { cat: c("stmt"), syntax: "while / for (;;) / do / switch", note: loc(locale, "express loops with map/filter/for-of", "循环用 map/filter/for-of 表达"), ok: false },
    { cat: c("stmt"), syntax: "const a = 1, b = 2", note: loc(locale, "multiple declarations in one statement", "一条语句多个声明"), ok: false },
    { cat: c("expr"), syntax: "number / string / template string / true / null / undefined", note: loc(locale, "number is float64", "number 是 float64"), ok: true },
    { cat: c("expr"), syntax: "[a, ...b] / { a, b: 1, ...c }", note: loc(locale, "array & object literals (object spread client-only)", "数组和对象字面量(对象展开只在客户端)"), ok: true },
    { cat: c("expr"), syntax: "a.b / a?.b / a[i] / a?.[i]", note: loc(locale, "member, optional chaining, index", "成员、可选链、下标"), ok: true },
    { cat: c("expr"), syntax: "f(x) / f?.(x) / useState<T>(x)", note: loc(locale, "call, optional call, explicit type argument", "调用、可选调用、显式类型参数"), ok: true },
    { cat: c("expr"), syntax: "! - + typeof", note: loc(locale, "unary", "一元"), ok: true },
    { cat: c("expr"), syntax: "+ - * / %   === !==   < > <= >=", note: loc(locale, "arithmetic, strict equality, comparison", "算术、严格相等、比较"), ok: true },
    { cat: c("expr"), syntax: "&& || ??", note: loc(locale, "logical; empty string is falsy", "逻辑; 字符串以空为假"), ok: true },
    { cat: c("expr"), syntax: "a ? b : c", note: loc(locale, "ternary (node or value)", "三元(节点或值)"), ok: true },
    { cat: c("expr"), syntax: "(a, b) => expr / x => { ... } / async () => ...", note: loc(locale, "arrow functions; param types inferred from context", "箭头函数; 参数类型可从上下文推断"), ok: true },
    { cat: c("expr"), syntax: "x = v / += -= *= /=", note: loc(locale, "assignment (statement position)", "赋值(语句位置)"), ok: true },
    { cat: c("expr"), syntax: "x as T / x!", note: loc(locale, "type assertion, non-null assertion", "类型断言、非空断言"), ok: true },
    { cat: c("expr"), syntax: "await x", note: loc(locale, "client code only", "只在客户端代码里"), ok: true },
    { cat: c("expr"), syntax: "== !=", note: loc(locale, "only === / !== allowed", "只允许 === / !=="), ok: false },
    { cat: c("expr"), syntax: "++ --", note: loc(locale, "write x = x + 1", "写 x = x + 1"), ok: false },
    { cat: c("expr"), syntax: "new / class / this / function expression", note: loc(locale, "no classes or prototypes", "没有类和原型"), ok: false },
    { cat: c("expr"), syntax: "/regex/ / in / instanceof / delete / void", note: loc(locale, "regex goes through a host module", "正则走宿主模块"), ok: false },
    { cat: c("jsx"), syntax: "<div class=\"x\" id={v} disabled>…</div> / <br />", note: loc(locale, "elements, string/expression/boolean-shorthand attrs, self-closing", "元素、属性字符串/表达式/布尔简写、自闭合"), ok: true },
    { cat: c("jsx"), syntax: "<></>", note: loc(locale, "fragment", "fragment"), ok: true },
    { cat: c("jsx"), syntax: "onClick={fn} / onInput={(e) => ...}", note: loc(locale, "events; handler param type any; not generated on the server", "事件; 处理器参数类型 any; 服务端不生成"), ok: true },
    { cat: c("jsx"), syntax: "aria-* / data-* / role", note: loc(locale, "booleans render as \"true\" / \"false\"", "布尔值渲染成 \"true\" / \"false\""), ok: true },
    { cat: c("jsx"), syntax: "{cond && <x/>} / {a ? <x/> : <y/>} / {list.map(...)}", note: loc(locale, "conditions & lists; reactive only if a signal is read", "条件与列表; 读到 signal 的才是响应式的"), ok: true },
    { cat: c("jsx"), syntax: "<Comp prop={v}>children</Comp>", note: loc(locale, "component call; children is a Node", "组件调用; children 是 Node"), ok: true },
    { cat: c("jsx"), syntax: "{/* comment */}", note: "", ok: true },
    { cat: c("jsx"), syntax: "<div {...props} />", note: loc(locale, "attribute spread", "属性展开"), ok: false },
    { cat: c("jsx"), syntax: "passing children to an island", note: loc(locale, "island props go through JSON in an HTML attribute — no room for a Node; children of shared & server components are fine", "岛的 props 走 JSON 进 HTML 属性, 放不下 Node; 共享组件和服务端组件的 children 正常"), ok: false },
    { cat: c("jsx"), syntax: "dangerouslySetInnerHTML", note: loc(locale, "no raw-HTML injection hole; render tokens returned by a host module yourself", "没有注入 HTML 的口子; 用宿主返回 token 自己渲染"), ok: false },
    { cat: c("type"), syntax: "string number boolean void any undefined null", note: "", ok: true },
    { cat: c("type"), syntax: "T[] / Array<T> / Record<string, T>", note: loc(locale, "arrays, maps", "数组、映射"), ok: true },
    { cat: c("type"), syntax: "{ a: string; b?: number; f(x: T): R }", note: loc(locale, "object types; optional primitive = zero-value semantics", "对象类型; 可选原始类型 = 零值语义"), ok: true },
    { cat: c("type"), syntax: "interface / type alias", note: "", ok: true },
    { cat: c("type"), syntax: "\"a\" | \"b\" / T | undefined", note: loc(locale, "literal union → string; optional", "字面量联合 → string; 可选"), ok: true },
    { cat: c("type"), syntax: "(x: T) => R / Promise<T>", note: loc(locale, "function type; Promise<T> treated as T", "函数类型; Promise<T> 视为 T"), ok: true },
    { cat: c("type"), syntax: "custom generics / tuple / intersection / keyof / typeof / enum / class", note: "", ok: false },
    { cat: c("hooks"), syntax: "const [x, setX] = useState(init)", note: loc(locale, "server = initial value; client = signal", "服务端 = 初值; 客户端 = signal"), ok: true },
    { cat: c("hooks"), syntax: "setX(v) / setX(prev => ...)", note: "", ok: true },
    { cat: c("hooks"), syntax: "const y = x * 2", note: loc(locale, "a signal-dependent const is automatically a memo, no useMemo needed", "依赖 signal 的 const 自动是 memo, 不需要 useMemo"), ok: true },
    { cat: c("hooks"), syntax: "useMemo(() => ...) / useEffect(() => ...)", note: loc(locale, "useEffect is client-only, deps tracked automatically", "useEffect 只进客户端, 依赖自动追踪"), ok: true },
    { cat: c("hooks"), syntax: "useRef / useContext / useReducer", note: "", ok: false },
    { cat: c("builtin"), syntax: "console.log / JSON.stringify / Math.max min floor ceil round abs sqrt random", note: loc(locale, "on both sides", "两端都有"), ok: true },
    { cat: c("builtin"), syntax: "String() Number() Boolean() parseInt parseFloat isNaN encodeURIComponent", note: "", ok: true },
    { cat: c("builtin"), syntax: "fetch setTimeout document window location history navigator localStorage Date", note: loc(locale, "client only, type any", "只在客户端, 类型 any"), ok: true },
    { cat: c("builtin"), syntax: "Object.keys / Object.values", note: loc(locale, "keys sorted (matching a Go map, keeping hydration stable)", "键按字典序(与 Go map 一致, 保证 hydrate 稳定)"), ok: true },
    { cat: c("builtin"), syntax: "Math.pow sign trunc", note: "", ok: true },
    { cat: c("builtin"), syntax: "Array.from / Date (server) / Object.assign", note: loc(locale, "use a host module for server-side time/date", "服务端时间/日期用宿主模块"), ok: false },
    { cat: c("array"), syntax: "length map filter find some every includes indexOf join slice concat forEach", note: "", ok: true },
    { cat: c("array"), syntax: "sort reduce reverse flat at", note: loc(locale, "sort/reduce match the Go backend (copy, don't mutate)", "sort/reduce 与 Go 后端同语义(拷贝, 不改原数组)"), ok: true },
    { cat: c("array"), syntax: "push splice pop shift", note: loc(locale, "use immutable forms: [...xs, x] / filter", "改用不可变写法: [...xs, x] / filter"), ok: false },
    { cat: c("string"), syntax: "length toUpperCase toLowerCase trim includes startsWith endsWith split slice replace replaceAll repeat indexOf charAt", note: loc(locale, "Go side works by rune", "Go 侧按 rune 处理"), ok: true },
    { cat: c("string"), syntax: "padStart padEnd / number toFixed", note: loc(locale, "by rune", "按 rune 处理"), ok: true },
    { cat: c("string"), syntax: "match matchAll replace(regex)", note: loc(locale, "regex goes through a host module", "正则走宿主模块"), ok: false },
  ];
}

export const sampleTSX = `import type { PageProps } from "gotsx";
import { models } from "host:data";        // Go-backed, zero marshalling
import Layout from "../components/Layout.server";
import ModelCard from "../components/ModelCard.server";

export default function Home({ query }: PageProps) {
  const q = query.q ?? "";
  const list = models.search(q);           // synchronous: concurrency from goroutines
  return (
    <Layout title="Products">
      <div class="grid">{list.map((m) => <ModelCard model={m} />)}</div>
      {list.length === 0 && <p class="empty">No matching products</p>}
    </Layout>
  );
}`;

export const sampleGo = `func Home(props gotsx.PageProps) gotsx.Node {
	var q string = gotsx.Or(props.Query["q"], "")
	var list []host.Model = host.Data.Models.Search(q)
	return Layout(LayoutProps{Title: "Products", Children: gotsx.Frag(
		gotsx.El("div", []gotsx.Attr{gotsx.A("class", "grid")},
			gotsx.NodesPlain(gotsx.Map(list, func(m host.Model) gotsx.Node {
				return ModelCard(ModelCardProps{Model: m})
			}))),
		gotsx.IfPlain(float64(len(list)) == 0, func() gotsx.Node {
			return gotsx.El("p", []gotsx.Attr{gotsx.A("class", "empty")}, gotsx.Text("No matching products"))
		}),
	)})
}`;

export const sampleIsland = `import { useState, useEffect } from "gotsx";

export default function Counter({ start }: { start: number }) {
  const [n, setN] = useState(start);
  const double = n * 2;                    // depends on n → auto memo
  useEffect(() => { console.log(n); });    // JS backend only
  return (
    <button onClick={() => setN(n + 1)}>
      {n} ×2 = {double}{n > 4 && <b> 🔥</b>}
    </button>
  );
}`;

export const sampleIslandJS = `export default function Counter({ start }) {
  const [n, setN] = G.signal(start);
  const double = G.memo(() => (n() * 2));
  G.effect(() => { console.log(n()); });
  return G.el("button", { onClick: () => setN((n() + 1)) }, () => [
    G.text(() => n()), G.t(" ×2 = "), G.text(() => double()),
    G.cond(() => (n() > 4), () => G.el("b", null, () => [G.t(" 🔥")]))]);
}`;

export const sampleIslandGo = `func Counter(props CounterProps) gotsx.Node {
	return gotsx.Island("Counter", props, Counter_ssr(props))
}

func Counter_ssr(props CounterProps) gotsx.Node {
	var n float64 = props.Start            // useState → initial value
	setN := func(any) {}                   // setter → empty function
	var double float64 = (n * 2)
	return gotsx.El("button", nil,
		gotsx.Dyn(gotsx.Num(n)), gotsx.Text(" ×2 = "), gotsx.Dyn(gotsx.Num(double)),
		gotsx.If((n > 4), func() gotsx.Node { return gotsx.El("b", nil, gotsx.Text(" 🔥")) }))
}`;

export const sampleHTML = `<gotsx-island name="Counter" props="{&quot;start&quot;:0}">
  <button>
    <!--$-->0<!--/--> ×2 = <!--$-->0<!--/-->
    <!--[--><!--]-->
  </button>
</gotsx-island>`;

export const sampleHost = `// host/host.go — the host module: the other side of import { models } from "host:data"
type ModelStore struct{ items []Model }

func (s *ModelStore) Search(q string) []Model { … }
func (s *ModelStore) Get(id string) (Model, error) {   // error → 404 at the request layer
	…
	return Model{}, fmt.Errorf("%w: %s", gotsx.ErrNotFound, id)
}

var Registry = map[string]gotsx.HostModule{
	"data": {Value: &DataModule{Models: store}, Go: "host.Data"},
}`;

export const sampleHostDTS = `// app/.gen/host.d.ts — reflected from Go; what TSX sees = what Go exposes
declare module "host:data" {
  export interface Model { id: string; title: string; likes: number; tags: string[] }
  export interface ModelStore {
    search(arg0: string): Model[];
    get(arg0: string): Model;
  }
  export const models: ModelStore;
}`;

export const sampleRun = `cd ~/work/gotsx
go run ./cmd/gotsx dev example -addr :3000   # compile → go build → run, rebuild on app/ change
go run ./cmd/gotsx build example              # only generate example/gen`;

export const sampleLayout = `example/
├── gotsx.json           # { "module": "gotsx/example", "hostPackage": "gotsx/example/host" }
├── main.go              # gotsx.Serve(gotsx.Options{Routes: gen.Routes, ...})
├── host/host.go         # host module (Go)
├── cmd/hostgen/         # reflect host.Registry → app/.gen/host.d.ts + host.json
├── app/
│   ├── pages/           # file routing: index.server.tsx → /, models/[id].server.tsx → /models/{id}
│   ├── components/      # *.server.tsx → Go only
│   ├── islands/         # *.client.tsx → Go (SSR) + JS (signals)
│   ├── ui/              # no suffix = shared, compiled to both
│   └── tailwind.css     # present → Tailwind runs
├── public/              # static files → /public/*
└── gen/                 # generated output: *_gen.go, routes_gen.go, client/*.js`;

export const sampleAction = `// an island's channel back to Go is HTTP: an action is a plain handler
Actions: map[string]http.HandlerFunc{
	"POST /actions/like": func(w http.ResponseWriter, r *http.Request) {
		n, err := host.Data.Models.Like(r.URL.Query().Get("id"))
		…
	},
}`;

export const sampleButton = `import type { Node } from "gotsx";

export default function Button({ variant = "primary", size = "md", href = "", onClick, children }: ButtonProps) {
  const cls = \`inline-flex items-center rounded-lg font-medium \${sizeCls(size)} \${variantCls(variant)}\`;
  return href !== ""
    ? <a href={href} class={cls}>{children}</a>
    : <button class={cls} onClick={onClick}>{children}</button>;
}`;

export const sampleTailwind = `/* app/tailwind.css */
@import "tailwindcss";
@custom-variant dark (&:where(.dark, .dark *));
@theme { --color-brand-600: #1557d6; }`;
