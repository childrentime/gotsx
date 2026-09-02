import type { PageProps } from "gotsx";
import DocsLayout from "../../components/DocsLayout.server";
import Section from "../../components/Section.server";
import Callout from "../../ui/Callout";
import SyntaxSearch from "../../islands/SyntaxSearch.client";
import { loc } from "../../ui/i18n";
import { syntaxRows } from "../../content/site.server";

export default function Language({ locale, path }: PageProps) {
  const lc = locale !== "" ? locale : "en";
  const rows = syntaxRows(lc);
  const total = rows.length;
  const ok = rows.filter((r) => r.ok).length;
  return (
    <DocsLayout title={loc(lc, "Language reference", "语言参考")} active="language" locale={lc} path={path}>
      <p class="text-[15px] leading-7">
        {loc(lc, "The gotsx dialect is ", "gotsx 的方言是一门")}<strong>{loc(lc, "a static language that borrows TSX syntax", "借 TSX 语法的静态语言")}</strong>{loc(lc, ", not an implementation of a TypeScript subset — it has no React runtime, and its type system is bounded by what Go can represent. The subset is defined by the type system: if every expression has a static type that falls inside the allowed set, it compiles; otherwise it's a compile error with a location.", ", 不是 TypeScript 的子集实现——它没有 React 运行时, 类型系统限定在 Go 能表示的集合里。子集由类型系统定义: 每个表达式都能推出一个静态类型, 且落在允许集合里, 就能编译; 否则是带位置的编译错误。")}
      </p>
      <Callout kind="info" title={loc(lc, "A few semantic conventions", "几个语义约定")}>
        <code class="font-mono">number</code>{loc(lc, " is float64 · an optional primitive represents absence with its zero value (", " 是 float64 · 可选的原始类型用零值表示缺席(")}<code class="font-mono">"" / 0 / false</code>{loc(lc, ") · strings are handled by rune on the Go side · ", ") · 字符串在 Go 侧按 rune 处理 · ")}<code class="font-mono">||</code>{loc(lc, " and ", " 和 ")}<code class="font-mono">??</code>{loc(lc, " treat an empty string as falsy · code running on the server may not contain DOM / async / Node, code running only in the browser may.", " 对字符串以空为假 · 跑在服务端的代码里不能出现 DOM / async / Node, 只跑在浏览器的可以。")}
      </Callout>

      <Section title={loc(lc, "Syntax table", "语法表")} lead={loc(lc, `${ok} supported, ${total - ok} explicitly unsupported (they error). This table is itself an island: the filtering happens in the browser.`, `${ok} 项支持, ${total - ok} 项明确不支持(会报错)。这张表本身是一个岛: 过滤在浏览器里完成。`)}>
        <SyntaxSearch rows={rows} locale={lc} />
      </Section>

      <Section title={loc(lc, "Reactivity rules", "响应性规则")} lead={loc(lc, "The client update model is decided by the compiler; there's no such thing as hooks call order", "客户端的更新模型由编译器决定, 没有 hooks 调用顺序这回事")}>
        <ul class="list-disc space-y-2 pl-5">
          <li><code class="font-mono text-sm">const [n, setN] = useState(0)</code>{loc(lc, " — n is a signal; every place that reads n compiles to ", " — n 是 signal; 读 n 的地方编译成 ")}<code class="font-mono text-sm">n()</code>{loc(lc, ".", "。")}</li>
          <li><code class="font-mono text-sm">const double = n * 2</code>{loc(lc, " — a const whose initializer reads a signal is automatically a memo.", " — 初始化表达式读到 signal 的 const 自动是 memo。")}</li>
          <li>{loc(lc, "JSX text, attributes, conditions, and lists that read a signal/memo each bind one effect; anything that doesn't is static and leaves no marker on the server either.", "JSX 里读到 signal/memo 的文本、属性、条件、列表各自绑一个 effect; 没读到的就是静态的, 服务端也不留标记。")}</li>
          <li><code class="font-mono text-sm">list.map(cb)</code>{loc(lc, " reactivity depends only on list, not on the callback body; conditions and ternaries depend only on the condition.", " 的响应性只看 list, 不看回调内部; 条件和三元只看条件。")}</li>
          <li>{loc(lc, "Component props are not reactive (like Solid, a signal's value is read once at creation); put reactive content into children or a conditional block.", "组件的 props 不是响应式的(和 Solid 一样, 传 signal 的值只在创建时读一次); 需要响应式的内容放进 children 或条件块。")}</li>
        </ul>
      </Section>

      <Section title={loc(lc, "Server / client / shared", "服务端 / 客户端 / 共享")}>
        <ul class="list-disc space-y-2 pl-5">
          <li><code class="font-mono text-sm">*.server.tsx</code>{loc(lc, ": Go only, may import ", ": 只编成 Go, 能 import ")}<code class="font-mono text-sm">host:*</code>{loc(lc, ", never reaches the browser.", ", 永不进浏览器。")}</li>
          <li><code class="font-mono text-sm">*.client.tsx</code>{loc(lc, ": compiled to Go (single-pass SSR) and JS; may only ", ": 编成 Go(SSR 单趟)和 JS; 只能 ")}<code class="font-mono text-sm">import type</code>{loc(lc, " host types; the default export is an island.", " 宿主类型; default 导出是岛。")}</li>
          <li>{loc(lc, "No suffix: a shared component, compiled to both sides, may touch neither the host nor the DOM — a pure render function.", "无后缀: 共享组件, 两端都编, 不能碰宿主也不能碰 DOM——纯渲染函数。")}</li>
          <li>{loc(lc, "The boundary is enforced by the compiler: a client importing a server component, or await appearing on the server, are both compile errors.", "边界由编译器强制: 客户端 import 服务端组件、服务端出现 await, 都是编译错误。")}</li>
        </ul>
      </Section>
    </DocsLayout>
  );
}
