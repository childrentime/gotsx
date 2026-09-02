import type { PageProps } from "gotsx";
import DocsLayout from "../../components/DocsLayout.server";
import Section from "../../components/Section.server";
import Callout from "../../ui/Callout";
import SyntaxSearch from "../../islands/SyntaxSearch.client";
import { syntaxRows } from "../../content/site.server";

export default function Language({ path }: PageProps) {
  const total = syntaxRows.length;
  const ok = syntaxRows.filter((r) => r.ok).length;
  return (
    <DocsLayout title="语言参考" active="language">
      <p class="text-[15px] leading-7 text-zinc-700 dark:text-zinc-300">
        gotsx 的方言是一门<strong>借 TSX 语法的静态语言</strong>, 不是 TypeScript 的子集实现——它没有 React 运行时, 类型系统限定在 Go 能表示的集合里。
        子集由类型系统定义: 每个表达式都能推出一个静态类型, 且落在允许集合里, 就能编译; 否则是带位置的编译错误。
      </p>
      <Callout kind="info" title="几个语义约定">
        <code class="font-mono">number</code> 是 float64 · 可选的原始类型用零值表示缺席(<code class="font-mono">"" / 0 / false</code>) ·
        字符串在 Go 侧按 rune 处理 · <code class="font-mono">||</code> 和 <code class="font-mono">??</code> 对字符串以空为假 ·
        跑在服务端的代码里不能出现 DOM / async / Node, 只跑在浏览器的可以。
      </Callout>

      <Section title="语法表" lead={`${ok} 项支持, ${total - ok} 项明确不支持(会报错)。这张表本身是一个岛: 过滤在浏览器里完成。`}>
        <SyntaxSearch rows={syntaxRows} />
      </Section>

      <Section title="响应性规则" lead="客户端的更新模型由编译器决定, 没有 hooks 调用顺序这回事">
        <ul class="list-disc space-y-2 pl-5">
          <li><code class="font-mono text-sm">const [n, setN] = useState(0)</code> — n 是 signal; 读 n 的地方编译成 <code class="font-mono text-sm">n()</code>。</li>
          <li><code class="font-mono text-sm">const double = n * 2</code> — 初始化表达式读到 signal 的 const 自动是 memo。</li>
          <li>JSX 里读到 signal/memo 的文本、属性、条件、列表各自绑一个 effect; 没读到的就是静态的, 服务端也不留标记。</li>
          <li><code class="font-mono text-sm">list.map(cb)</code> 的响应性只看 list, 不看回调内部; 条件和三元只看条件。</li>
          <li>组件的 props 不是响应式的(和 Solid 一样, 传 signal 的值只在创建时读一次); 需要响应式的内容放进 children 或条件块。</li>
        </ul>
      </Section>

      <Section title="服务端 / 客户端 / 共享">
        <ul class="list-disc space-y-2 pl-5">
          <li><code class="font-mono text-sm">*.server.tsx</code>: 只编成 Go, 能 import <code class="font-mono text-sm">host:*</code>, 永不进浏览器。</li>
          <li><code class="font-mono text-sm">*.client.tsx</code>: 编成 Go(SSR 单趟)和 JS; 只能 <code class="font-mono text-sm">import type</code> 宿主类型; default 导出是岛。</li>
          <li>无后缀: 共享组件, 两端都编, 不能碰宿主也不能碰 DOM——纯渲染函数。</li>
          <li>边界由编译器强制: 客户端 import 服务端组件、服务端出现 await, 都是编译错误。</li>
        </ul>
      </Section>
    </DocsLayout>
  );
}
