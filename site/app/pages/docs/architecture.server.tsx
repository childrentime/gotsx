import type { PageProps } from "gotsx";
import DocsLayout from "../../components/DocsLayout.server";
import Section from "../../components/Section.server";
import Card from "../../ui/Card";
import Callout from "../../ui/Callout";
import CodeBlock from "../../ui/CodeBlock.server";
import { sampleHTML, sampleIslandJS, sampleIslandGo } from "../../content/site.server";

interface Step {
  n: string;
  title: string;
  desc: string;
}

const pipeline: Step[] = [
  { n: "1", title: "解析", desc: "自己写的 TSX 子集 parser(词法 + 递归下降 + JSX 模式), 约 900 行 Go。" },
  { n: "2", title: "类型检查", desc: "局部推断 + 边界标注: props、宿主签名、hooks、内建方法表。出子集就报错。" },
  { n: "3", title: "响应性判定", desc: "哪个表达式读到了 signal/memo。这个判定同时驱动两个后端, 所以 hydrate 标记一定对得上。" },
  { n: "4", title: "Go 后端", desc: "组件 → Go 函数, JSX → gotsx.El/Text/If/Nodes, hooks → 单趟语义, host:* → 直接调用。gofmt 校验语法。" },
  { n: "5", title: "JS 后端", desc: "组件 → 函数, useState → G.signal, 依赖 signal 的 const → G.memo, JSX → G.el/t/text/cond/each。" },
  { n: "6", title: "go build", desc: "生成的 Go 和你的 main.go、host 包一起编成一个二进制; 客户端 JS 直接作为 ES module 服务。" },
];

export default function Architecture({ path }: PageProps) {
  return (
    <DocsLayout title="架构与原理" active="architecture">
      <p class="text-[15px] leading-7 text-zinc-700 dark:text-zinc-300">
        核心洞察只有一句: <strong>SSR 是一次同步的单趟求值。</strong> 没有重渲染、没有 effect、setter 永远不会被调用。
        所以服务端不需要 React 运行时, 只需要组件的"渲染切片"——而它的语义小到可以编译成 Go。
      </p>

      <Section title="编译流水线">
        <div class="grid gap-3 sm:grid-cols-2">
          {pipeline.map((s) => (
            <Card title={`${s.n}. ${s.title}`}>{s.desc}</Card>
          ))}
        </div>
      </Section>

      <Section title="hooks 在 Go 里是什么" lead="同一份 Counter.client.tsx 的 Go 输出">
        <CodeBlock code={sampleIslandGo} lang="go" title="gen/islands_Counter_client_gen.go" />
        <ul class="list-disc space-y-1 pl-5">
          <li><code class="font-mono text-sm">useState(start)</code> → <code class="font-mono text-sm">var n float64 = props.Start</code>; setter 是空函数。</li>
          <li><code class="font-mono text-sm">useEffect</code>、事件处理器、async 函数 → 不生成。</li>
          <li>响应式的 <code class="font-mono text-sm">{"{n}"}</code> → <code class="font-mono text-sm">gotsx.Dyn</code>(带标记); 静态文本 → <code class="font-mono text-sm">gotsx.Text</code>。</li>
          <li>岛的外壳 <code class="font-mono text-sm">gotsx.Island(name, props, inner)</code> 把 props 序列化进属性, 并打开标记模式。</li>
        </ul>
      </Section>

      <Section title="走位 hydrate" lead="服务端输出的 HTML(节选)">
        <CodeBlock code={sampleHTML} lang="html" title="岛内部的标记" />
        <p>客户端拿到同一份结构的 JS:</p>
        <CodeBlock code={sampleIslandJS} lang="js" title="gen/client/Counter.js" />
        <p>
          <code class="font-mono text-sm">G.el</code> 认领下一个元素, <code class="font-mono text-sm">G.t</code> 认领下一个文本节点,
          <code class="font-mono text-sm">G.text</code> 认领 <code class="font-mono text-sm">{"<!--$-->…<!--/-->"}</code> 之间的文本并绑 effect,
          <code class="font-mono text-sm">G.cond / G.each</code> 认领 <code class="font-mono text-sm">{"<!--[-->…<!--]-->"}</code> 区块。
          子节点用 thunk 延迟到父元素认领之后再执行, 顺序就是源码顺序。没有 diff, 没有重建, 已有 DOM 原样复用。
        </p>
        <Callout kind="info">为什么能这么简单: 两端的结构来自同一个编译器、同一个响应性判定。React 需要 diff 是因为它不知道服务端到底渲染了什么。</Callout>
      </Section>

      <Section title="宿主模块与围栏">
        <ul class="list-disc space-y-2 pl-5">
          <li>宿主模块是 Go 值; <code class="font-mono text-sm">hostgen</code> 反射它生成 <code class="font-mono text-sm">host.d.ts</code>(给编辑器)和 <code class="font-mono text-sm">host.json</code>(给编译器, 含 Go 名和数值类型)。</li>
          <li>方言能做的严格等于 Go 暴露的。Go 是唯一的真相源: 路由、数据、权限、动作都在 Go。</li>
          <li>围栏在类型检查阶段: 客户端 import host、客户端 import 服务端组件、服务端出现 await、未知 prop、any 上取成员——全是编译错误。</li>
        </ul>
      </Section>

      <Section title="SPA 跳转">
        <p>点击链接 → fetch 新页 HTML → idiomorph 把 body morph 过去 → 岛按 DOM 同一性存活(同名同 props 不动, props 变了重建, 消失了卸载)。
          表单 GET 同理; 后退/前进走 popstate; 并发导航只认最后一次; 非 HTML 响应退回整页加载。</p>
      </Section>

      <Section title="和 goja 路线的对比" lead="同一台机器, 同一类页面">
        <div class="overflow-x-auto rounded-xl border border-zinc-200 dark:border-zinc-800">
          <table class="w-full text-left text-sm">
            <thead class="bg-zinc-50 text-xs uppercase text-zinc-500 dark:bg-zinc-900"><tr><th class="px-4 py-2">项</th><th class="px-4 py-2">goja + React + MUI</th><th class="px-4 py-2">gotsx</th></tr></thead>
            <tbody class="divide-y divide-zinc-200 dark:divide-zinc-800">
              <tr><td class="px-4 py-2">列表页渲染</td><td class="px-4 py-2">45–60 ms</td><td class="px-4 py-2 font-semibold">~30 µs</td></tr>
              <tr><td class="px-4 py-2">吞吐</td><td class="px-4 py-2">~31 req/s(4 VM)</td><td class="px-4 py-2 font-semibold">~28k req/s</td></tr>
              <tr><td class="px-4 py-2">客户端运行时</td><td class="px-4 py-2">react-dom 62 KB gz</td><td class="px-4 py-2 font-semibold">6 KB</td></tr>
              <tr><td class="px-4 py-2">调试</td><td class="px-4 py-2">没有断点</td><td class="px-4 py-2 font-semibold">delve / pprof / go test</td></tr>
              <tr><td class="px-4 py-2">npm 生态</td><td class="px-4 py-2">纯 JS 包可用</td><td class="px-4 py-2">不可用, 库用方言写</td></tr>
            </tbody>
          </table>
        </div>
      </Section>
    </DocsLayout>
  );
}
