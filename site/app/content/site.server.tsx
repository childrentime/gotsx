import type { QA } from "../islands/Accordion.client";
import type { Row } from "../islands/SyntaxSearch.client";

export interface Feature {
  icon: string;
  title: string;
  desc: string;
}

export const features: Feature[] = [
  { icon: "⚙️", title: "编译到 Go, 不是解释执行", desc: "服务端组件是真正的 Go 函数: JSX 编成字符串直写, useState 是初值, useEffect 什么都不生成。一个页面几十微秒。" },
  { icon: "⚡", title: "signals 客户端", desc: "同一份 .client.tsx 编成 signals + 精确 DOM 绑定(Solid / Svelte 5 的模型), 没有 vdom, 运行时 6KB。" },
  { icon: "🧭", title: "走位 hydrate", desc: "服务端只在响应式的文本/条件/列表上留标记, 客户端按同一编译器给出的结构顺序认领节点, 不需要 diff, 复用现有 DOM。" },
  { icon: "🔌", title: "宿主模块", desc: "import { models } from \"host:data\" 背后是 Go, 编译后是直接调用, 零编组; 类型由 Go 反射生成。" },
  { icon: "🏝️", title: "岛 + SPA 跳转", desc: "页面零 JS, 岛按需加载; 跳转 = 拉 HTML → morph, 岛按 DOM 同一性存活, 状态不丢。" },
  { icon: "🚧", title: "围栏是编译错误", desc: "客户端碰 host:*、缺 prop、any 上取成员、用了 ==……全部是带文件:行:列的错误, 不会静默编错。" },
  { icon: "🎨", title: "Tailwind", desc: "class 就是字符串, Tailwind standalone CLI 构建期扫描 .tsx 生成 CSS, 同样不需要 Node。这个站就是。" },
  { icon: "🔁", title: "gotsx dev", desc: "改 TSX → 重新编译 → go build → 重启, 约 2 秒; 编译失败时旧版本继续跑。" },
];

export const faqs: QA[] = [
  { q: "这是 React 吗? 能用 npm 上的 React 组件库吗?", a: "不是。gotsx 借的是 React 的思想和 TSX 的手感, 运行时是自己的。npm 上的 React 库(包括 MUI)依赖 React 运行时和动态 JS, 不能通过子集检查, 也不可能编译成 Go。组件库要用方言写——就像这个站的组件库。" },
  { q: "为什么不用 goja / V8 在 Go 里跑 JS?", a: "做过。React + MUI 在 goja 里一个列表页 50ms, 换 V8 也要引入 CGO。而 SSR 是一次同步的单趟求值, 它需要的语义小到可以直接编译成 Go——于是列表页变成 30µs, 而且 go test / pprof / delve 全能用。" },
  { q: "子集有多小? 我的代码要改多少?", a: "常见的组件代码几乎不用改: 函数组件、props、解构、map/filter、模板字符串、三元、&&、useState/useEffect 都有。没有的是 class、this、any 上的成员访问、==、自定义泛型、while/switch。出了子集是编译错误, 不是静默行为差异。语言参考页有完整表。" },
  { q: "客户端的更新模型是什么?", a: "useState 编译成 signal, 依赖 signal 的 const 自动编译成 memo, 读到 signal 的 JSX 文本/属性/条件/列表各自绑一个 effect。没有\"重跑整个函数再 diff\"这回事。" },
  { q: "数据怎么进来?", a: "服务端组件同步调用宿主模块(Go), 没有 async/await——并发由 goroutine 提供。岛拿 props(JSON 进 HTML 属性), 要回到 Go 就走 HTTP action。" },
  { q: "生产可用吗?", a: "这是 PoC 2: 编译器 + 两个后端 + 运行时 + dev 循环 + 这个站, 约 7000 行 Go/JS。缺的是 //line 错误映射、keyed 列表 diff、更多内建、错误边界和一年的打磨。" },
];

export const syntaxRows: Row[] = [
  { cat: "模块", syntax: "import X from \"./a\"", note: "默认导入; 路径可省 .tsx / .server.tsx / .client.tsx", ok: true },
  { cat: "模块", syntax: "import { a, b as c } from \"./a\"", note: "命名导入", ok: true },
  { cat: "模块", syntax: "import type { T } from \"./a\"", note: "类型导入; 客户端对 host:* 只允许 import type", ok: true },
  { cat: "模块", syntax: "import { useState, useEffect, useMemo } from \"gotsx\"", note: "hooks", ok: true },
  { cat: "模块", syntax: "import type { Node, PageProps } from \"gotsx\"", note: "框架类型", ok: true },
  { cat: "模块", syntax: "import { x } from \"host:name\"", note: "宿主模块, 仅服务端; 编译成直接 Go 调用", ok: true },
  { cat: "模块", syntax: "export default function / export function", note: "组件(大写)或普通函数", ok: true },
  { cat: "模块", syntax: "export const data: T[] = [...]", note: "模块级 const → Go 包级变量; 不允许 let", ok: true },
  { cat: "模块", syntax: "export interface / export type", note: "类型导出(所有类型自动可被 import type)", ok: true },
  { cat: "模块", syntax: "import * as ns from", note: "命名空间导入", ok: false },
  { cat: "语句", syntax: "const x = ... / let x = ...", note: "单声明; 类型可标注可推断", ok: true },
  { cat: "语句", syntax: "const { a, b = 1 } = obj / const [x, setX] = ...", note: "解构 + 默认值(原始类型, 零值语义)", ok: true },
  { cat: "语句", syntax: "if / else if / else", note: "条件按 JS 真值语义(空串、0 为假)", ok: true },
  { cat: "语句", syntax: "for (const x of xs)", note: "只支持 for-of 遍历数组", ok: true },
  { cat: "语句", syntax: "return / throw", note: "throw 在 Go 侧是 panic, 请求层 recover", ok: true },
  { cat: "语句", syntax: "try / catch / finally", note: "客户端完整; Go 侧只执行 try 和 finally 体", ok: true },
  { cat: "语句", syntax: "function f() {} / async function f() {}", note: "嵌套函数; async 只进 JS 后端", ok: true },
  { cat: "语句", syntax: "while / for (;;) / do / switch", note: "循环用 map/filter/for-of 表达", ok: false },
  { cat: "语句", syntax: "const a = 1, b = 2", note: "一条语句多个声明", ok: false },
  { cat: "表达式", syntax: "数字 / 字符串 / 模板字符串 / true / null / undefined", note: "number 是 float64", ok: true },
  { cat: "表达式", syntax: "[a, ...b] / { a, b: 1, ...c }", note: "数组和对象字面量(对象展开只在客户端)", ok: true },
  { cat: "表达式", syntax: "a.b / a?.b / a[i] / a?.[i]", note: "成员、可选链、下标", ok: true },
  { cat: "表达式", syntax: "f(x) / f?.(x) / useState<T>(x)", note: "调用、可选调用、显式类型参数", ok: true },
  { cat: "表达式", syntax: "! - + typeof", note: "一元", ok: true },
  { cat: "表达式", syntax: "+ - * / %   === !==   < > <= >=", note: "算术、严格相等、比较", ok: true },
  { cat: "表达式", syntax: "&& || ??", note: "逻辑; 字符串以空为假", ok: true },
  { cat: "表达式", syntax: "a ? b : c", note: "三元(节点或值)", ok: true },
  { cat: "表达式", syntax: "(a, b) => expr / x => { ... } / async () => ...", note: "箭头函数; 参数类型可从上下文推断", ok: true },
  { cat: "表达式", syntax: "x = v / += -= *= /=", note: "赋值(语句位置)", ok: true },
  { cat: "表达式", syntax: "x as T / x!", note: "类型断言、非空断言", ok: true },
  { cat: "表达式", syntax: "await x", note: "只在客户端代码里", ok: true },
  { cat: "表达式", syntax: "== !=", note: "只允许 === / !==", ok: false },
  { cat: "表达式", syntax: "++ --", note: "写 x = x + 1", ok: false },
  { cat: "表达式", syntax: "new / class / this / function 表达式", note: "没有类和原型", ok: false },
  { cat: "表达式", syntax: "/regex/ / in / instanceof / delete / void", note: "正则走宿主模块", ok: false },
  { cat: "JSX", syntax: "<div class=\"x\" id={v} disabled>…</div> / <br />", note: "元素、属性字符串/表达式/布尔简写、自闭合", ok: true },
  { cat: "JSX", syntax: "<></>", note: "fragment", ok: true },
  { cat: "JSX", syntax: "onClick={fn} / onInput={(e) => ...}", note: "事件; 处理器参数类型 any; 服务端不生成", ok: true },
  { cat: "JSX", syntax: "aria-* / data-* / role", note: "布尔值渲染成 \"true\" / \"false\"", ok: true },
  { cat: "JSX", syntax: "{cond && <x/>} / {a ? <x/> : <y/>} / {list.map(...)}", note: "条件与列表; 读到 signal 的才是响应式的", ok: true },
  { cat: "JSX", syntax: "<Comp prop={v}>children</Comp>", note: "组件调用; children 是 Node", ok: true },
  { cat: "JSX", syntax: "{/* 注释 */}", note: "", ok: true },
  { cat: "JSX", syntax: "<div {...props} />", note: "属性展开", ok: false },
  { cat: "JSX", syntax: "给岛传 children", note: "岛的 props 走 JSON 进 HTML 属性, 放不下 Node; 共享组件和服务端组件的 children 正常", ok: false },
  { cat: "JSX", syntax: "dangerouslySetInnerHTML", note: "没有注入 HTML 的口子; 用宿主返回 token 自己渲染", ok: false },
  { cat: "类型", syntax: "string number boolean void any undefined null", note: "", ok: true },
  { cat: "类型", syntax: "T[] / Array<T> / Record<string, T>", note: "数组、映射", ok: true },
  { cat: "类型", syntax: "{ a: string; b?: number; f(x: T): R }", note: "对象类型; 可选原始类型 = 零值语义", ok: true },
  { cat: "类型", syntax: "interface / type 别名", note: "", ok: true },
  { cat: "类型", syntax: "\"a\" | \"b\" / T | undefined", note: "字面量联合 → string; 可选", ok: true },
  { cat: "类型", syntax: "(x: T) => R / Promise<T>", note: "函数类型; Promise<T> 视为 T", ok: true },
  { cat: "类型", syntax: "自定义泛型 / 元组 / 交叉类型 / keyof / typeof / enum / class", note: "", ok: false },
  { cat: "hooks", syntax: "const [x, setX] = useState(init)", note: "服务端 = 初值; 客户端 = signal", ok: true },
  { cat: "hooks", syntax: "setX(v) / setX(prev => ...)", note: "", ok: true },
  { cat: "hooks", syntax: "const y = x * 2", note: "依赖 signal 的 const 自动是 memo, 不需要 useMemo", ok: true },
  { cat: "hooks", syntax: "useMemo(() => ...) / useEffect(() => ...)", note: "useEffect 只进客户端, 依赖自动追踪", ok: true },
  { cat: "hooks", syntax: "useRef / useContext / useReducer", note: "", ok: false },
  { cat: "内建", syntax: "console.log / JSON.stringify / Math.max min floor ceil round abs sqrt random", note: "两端都有", ok: true },
  { cat: "内建", syntax: "String() Number() Boolean() parseInt parseFloat isNaN encodeURIComponent", note: "", ok: true },
  { cat: "内建", syntax: "fetch setTimeout document window location history navigator localStorage Date", note: "只在客户端, 类型 any", ok: true },
  { cat: "内建", syntax: "Object.keys / Object.values", note: "键按字典序(与 Go map 一致, 保证 hydrate 稳定)", ok: true },
  { cat: "内建", syntax: "Math.pow sign trunc", note: "", ok: true },
  { cat: "内建", syntax: "Array.from / Date(服务端) / Object.assign", note: "服务端时间/日期用宿主模块", ok: false },
  { cat: "数组", syntax: "length map filter find some every includes indexOf join slice concat forEach", note: "", ok: true },
  { cat: "数组", syntax: "sort reduce reverse flat at", note: "sort/reduce 与 Go 后端同语义(拷贝, 不改原数组)", ok: true },
  { cat: "数组", syntax: "push splice pop shift", note: "改用不可变写法: [...xs, x] / filter", ok: false },
  { cat: "字符串", syntax: "length toUpperCase toLowerCase trim includes startsWith endsWith split slice replace replaceAll repeat indexOf charAt", note: "Go 侧按 rune 处理", ok: true },
  { cat: "字符串", syntax: "padStart padEnd / 数字 toFixed", note: "按 rune 处理", ok: true },
  { cat: "字符串", syntax: "match matchAll replace(正则)", note: "正则走宿主模块", ok: false },
];

export const sampleTSX = `import type { PageProps } from "gotsx";
import { models } from "host:data";        // Go 实现, 零编组
import Layout from "../components/Layout.server";
import ModelCard from "../components/ModelCard.server";

export default function Home({ query }: PageProps) {
  const q = query.q ?? "";
  const list = models.search(q);           // 同步: 并发由 goroutine 提供
  return (
    <Layout title="模型">
      <div class="grid">{list.map((m) => <ModelCard model={m} />)}</div>
      {list.length === 0 && <p class="empty">没有匹配的模型</p>}
    </Layout>
  );
}`;

export const sampleGo = `func Home(props gotsx.PageProps) gotsx.Node {
	var q string = gotsx.Or(props.Query["q"], "")
	var list []host.Model = host.Data.Models.Search(q)
	return Layout(LayoutProps{Title: "模型", Children: gotsx.Frag(
		gotsx.El("div", []gotsx.Attr{gotsx.A("class", "grid")},
			gotsx.NodesPlain(gotsx.Map(list, func(m host.Model) gotsx.Node {
				return ModelCard(ModelCardProps{Model: m})
			}))),
		gotsx.IfPlain(float64(len(list)) == 0, func() gotsx.Node {
			return gotsx.El("p", []gotsx.Attr{gotsx.A("class", "empty")}, gotsx.Text("没有匹配的模型"))
		}),
	)})
}`;

export const sampleIsland = `import { useState, useEffect } from "gotsx";

export default function Counter({ start }: { start: number }) {
  const [n, setN] = useState(start);
  const double = n * 2;                    // 依赖 n → 自动 memo
  useEffect(() => { console.log(n); });    // 只进 JS 后端
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
	var n float64 = props.Start            // useState → 初值
	setN := func(any) {}                   // setter → 空函数
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

export const sampleHost = `// host/host.go —— 宿主模块: 方言里 import { models } from "host:data" 的另一端
type ModelStore struct{ items []Model }

func (s *ModelStore) Search(q string) []Model { … }
func (s *ModelStore) Get(id string) (Model, error) {   // error → 请求层 404
	…
	return Model{}, fmt.Errorf("%w: %s", gotsx.ErrNotFound, id)
}

var Registry = map[string]gotsx.HostModule{
	"data": {Value: &DataModule{Models: store}, Go: "host.Data"},
}`;

export const sampleHostDTS = `// app/.gen/host.d.ts —— 由 Go 反射生成, TSX 能看到的 = Go 暴露的
declare module "host:data" {
  export interface Model { id: string; title: string; likes: number; tags: string[] }
  export interface ModelStore {
    search(arg0: string): Model[];
    get(arg0: string): Model;
  }
  export const models: ModelStore;
}`;

export const sampleRun = `cd ~/work/gotsx
go run ./cmd/gotsx dev example -addr :3000   # 编译 → go build → 运行, 改 app/ 自动重来
go run ./cmd/gotsx build example              # 只生成 example/gen`;

export const sampleLayout = `example/
├── gotsx.json           # { "module": "gotsx/example", "hostPackage": "gotsx/example/host" }
├── main.go              # gotsx.Serve(gotsx.Options{Routes: gen.Routes, ...})
├── host/host.go         # 宿主模块(Go)
├── cmd/hostgen/         # 反射 host.Registry → app/.gen/host.d.ts + host.json
├── app/
│   ├── pages/           # 文件路由: index.server.tsx → /, models/[id].server.tsx → /models/{id}
│   ├── components/      # *.server.tsx 只编成 Go
│   ├── islands/         # *.client.tsx 编成 Go(SSR) + JS(signals)
│   ├── ui/              # 无后缀 = 共享, 两端都编译
│   └── tailwind.css     # 有它就跑 Tailwind
├── public/              # 静态文件 → /public/*
└── gen/                 # 生成产物: *_gen.go, routes_gen.go, client/*.js`;

export const sampleAction = `// 岛回到 Go 的通道是 HTTP: 动作就是普通 handler
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
