import type { PageProps } from "gotsx";
import DocsLayout from "../../components/DocsLayout.server";
import Section from "../../components/Section.server";
import CodeBlock from "../../ui/CodeBlock.server";
import Callout from "../../ui/Callout";
import { sampleRun, sampleLayout, sampleTSX, sampleIsland, sampleHost, sampleHostDTS, sampleAction, sampleTailwind } from "../../content/site.server";

export default function Docs({ path }: PageProps) {
  return (
    <DocsLayout title="快速开始" active="docs">
      <p class="text-[15px] leading-7 text-zinc-700 dark:text-zinc-300">
        gotsx 是一个 Go 模块。应用 = 一个 Go 包(路由、宿主模块、动作) + 一个 <code class="font-mono text-sm">app/</code> 目录(方言写的页面、组件、岛)。
        编译器把 <code class="font-mono text-sm">app/</code> 变成 <code class="font-mono text-sm">gen/</code> 里的 Go 和 JS, 然后 <code class="font-mono text-sm">go build</code>。
      </p>

      <Section title="1. 运行示例" lead="仓库自带一个 MakerWorld 式的示例应用">
        <CodeBlock code={sampleRun} lang="bash" title="终端" />
        <Callout kind="info">
          <code class="font-mono">gotsx dev</code> 会先跑 <code class="font-mono">cmd/hostgen</code> 生成宿主类型, 再编译方言, 再 <code class="font-mono">go build</code> 并启动;
          之后监视 <code class="font-mono">app/</code>, 改动后约 2 秒重新上线。编译失败时旧版本继续运行。
        </Callout>
      </Section>

      <Section title="2. 目录约定">
        <CodeBlock code={sampleLayout} lang="bash" title="example/" />
        <p>后缀决定编译目标: <code class="font-mono text-sm">.server.tsx</code> 只编 Go, <code class="font-mono text-sm">.client.tsx</code> 编 Go + JS(岛), 无后缀是共享组件(两端都编)。
          <code class="font-mono text-sm">pages/</code> 下的文件是路由, <code class="font-mono text-sm">[id]</code> 是路径参数。</p>
      </Section>

      <Section title="3. 写一个页面" lead="页面是 export default 的组件, props 固定是 PageProps">
        <CodeBlock code={sampleTSX} lang="tsx" title="app/pages/index.server.tsx" />
        <p>注意没有 <code class="font-mono text-sm">async</code>: 宿主调用是同步的, 请求之间的并发由 goroutine 提供。</p>
      </Section>

      <Section title="4. 写一个岛" lead="需要交互的部分放进 .client.tsx">
        <CodeBlock code={sampleIsland} lang="tsx" title="app/islands/Counter.client.tsx" />
        <p>在服务端组件里像普通组件一样用: <code class="font-mono text-sm">{"<Counter start={0} />"}</code>。
          props 必须可 JSON 序列化(它会进 HTML 属性), 岛不接受 children。</p>
      </Section>

      <Section title="5. 暴露 Go 能力" lead="宿主模块 = 一个 Go 值; 字段按 json tag、方法首字母小写映射到方言">
        <CodeBlock code={sampleHost} lang="go" title="host/host.go" />
        <CodeBlock code={sampleHostDTS} lang="tsx" title="app/.gen/host.d.ts(生成)" />
        <p>编译后 <code class="font-mono text-sm">models.search(q)</code> 就是 <code class="font-mono text-sm">host.Data.Models.Search(q)</code>: 没有编组, 没有反射。
          Go 的 <code class="font-mono text-sm">int</code> 和方言的 <code class="font-mono text-sm">number</code> 自动转换; 返回 error 的方法, error 变成 panic 由请求层 recover, 包了 <code class="font-mono text-sm">gotsx.ErrNotFound</code> 的回 404。</p>
      </Section>

      <Section title="6. 从岛回到 Go" lead="岛只能通过 HTTP 回到服务端, 动作就是普通 handler">
        <CodeBlock code={sampleAction} lang="go" title="main.go" />
      </Section>

      <Section title="7. 样式: Tailwind" lead="有 app/tailwind.css 就会在每次构建时跑 Tailwind standalone CLI">
        <CodeBlock code={sampleTailwind} lang="css" title="app/tailwind.css" />
        <p>方言里的 <code class="font-mono text-sm">class="..."</code> 就是普通字符串, Tailwind 构建期扫描 <code class="font-mono text-sm">app/**/*.tsx</code> 生成 <code class="font-mono text-sm">public/tailwind.css</code>。
          二进制查找顺序: <code class="font-mono text-sm">$GOTSX_TAILWIND</code> → 仓库 <code class="font-mono text-sm">.tools/tailwindcss</code> → PATH。同样不需要 Node。这个站的每一个 class 都是这么来的。</p>
      </Section>
    </DocsLayout>
  );
}
