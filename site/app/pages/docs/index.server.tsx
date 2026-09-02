import type { PageProps } from "gotsx";
import DocsLayout from "../../components/DocsLayout.server";
import Section from "../../components/Section.server";
import CodeBlock from "../../ui/CodeBlock.server";
import Callout from "../../ui/Callout";
import { loc } from "../../ui/i18n";
import { sampleRun, sampleLayout, sampleTSX, sampleIsland, sampleHost, sampleHostDTS, sampleAction, sampleTailwind } from "../../content/site.server";

export default function Docs({ locale, path }: PageProps) {
  const lc = locale !== "" ? locale : "en";
  return (
    <DocsLayout title={loc(lc, "Quick start", "快速开始")} active="docs" locale={lc} path={path}>
      <p class="text-[15px] leading-7">
        {loc(lc, "gotsx is a Go module. An app = one Go package (routes, host modules, actions) + one ", "gotsx 是一个 Go 模块。应用 = 一个 Go 包(路由、宿主模块、动作) + 一个 ")}<code class="font-mono text-sm">app/</code>{loc(lc, " directory (pages, components, and islands written in the dialect). The compiler turns ", " 目录(方言写的页面、组件、岛)。编译器把 ")}<code class="font-mono text-sm">app/</code>{loc(lc, " into Go and JS under ", " 变成 ")}<code class="font-mono text-sm">gen/</code>{loc(lc, ", then ", " 里的 Go 和 JS, 然后 ")}<code class="font-mono text-sm">go build</code>{loc(lc, ".", "。")}
      </p>

      <Section title={loc(lc, "1. Create an app", "1. 创建应用")} lead={loc(lc, "One go install, one gotsx new — the result runs right away", "一条 go install、一条 gotsx new, 出来的应用直接能跑")}>
        <CodeBlock code={sampleRun} lang="bash" title={loc(lc, "terminal", "终端")} />
        <Callout kind="info">
          <code class="font-mono">gotsx dev</code>{loc(lc, " first runs ", " 会先跑 ")}<code class="font-mono">cmd/hostgen</code>{loc(lc, " to generate host types, then compiles the dialect, then ", " 生成宿主类型, 再编译方言, 再 ")}<code class="font-mono">go build</code>{loc(lc, " and starts up; after that it watches ", " 并启动; 之后监视 ")}<code class="font-mono">app/</code>{loc(lc, " and goes live again about 2 seconds after a change — the browser reloads by itself. When a compile fails, the old version keeps running. The repo's own demos run with ", ", 改动后约 2 秒重新上线, 浏览器自动刷新。编译失败时旧版本继续运行。仓库自带的示例用 ")}<code class="font-mono">make dev-example</code>{loc(lc, " / ", " / ")}<code class="font-mono">dev-shop</code>{loc(lc, ".", "。")}
        </Callout>
      </Section>

      <Section title={loc(lc, "2. Directory conventions", "2. 目录约定")}>
        <CodeBlock code={sampleLayout} lang="bash" title="example/" />
        <p>{loc(lc, "The suffix decides the compile target: ", "后缀决定编译目标: ")}<code class="font-mono text-sm">.server.tsx</code>{loc(lc, " is Go only, ", " 只编 Go, ")}<code class="font-mono text-sm">.client.tsx</code>{loc(lc, " is Go + JS (an island), and no suffix is a shared component (both sides). Files under ", " 编 Go + JS(岛), 无后缀是共享组件(两端都编)。")}<code class="font-mono text-sm">pages/</code>{loc(lc, " are routes, ", " 下的文件是路由, ")}<code class="font-mono text-sm">[id]</code>{loc(lc, " is a path parameter and ", " 是路径参数, ")}<code class="font-mono text-sm">[...slug]</code>{loc(lc, " a catch-all. Inside a page, ", " 是 catch-all。页面里 ")}<code class="font-mono text-sm">redirect()</code>{loc(lc, " / ", " / ")}<code class="font-mono text-sm">notFound()</code>{loc(lc, " abort the render.", " 可以中断渲染。")}</p>
      </Section>

      <Section title={loc(lc, "3. Write a page", "3. 写一个页面")} lead={loc(lc, "A page is an export default component whose props are always PageProps", "页面是 export default 的组件, props 固定是 PageProps")}>
        <CodeBlock code={sampleTSX} lang="tsx" title="app/pages/index.server.tsx" />
        <p>{loc(lc, "Note there's no ", "注意没有 ")}<code class="font-mono text-sm">async</code>{loc(lc, ": host calls are synchronous, and concurrency between requests comes from goroutines.", ": 宿主调用是同步的, 请求之间的并发由 goroutine 提供。")}</p>
      </Section>

      <Section title={loc(lc, "4. Write an island", "4. 写一个岛")} lead={loc(lc, "Put the interactive parts into .client.tsx", "需要交互的部分放进 .client.tsx")}>
        <CodeBlock code={sampleIsland} lang="tsx" title="app/islands/Counter.client.tsx" />
        <p>{loc(lc, "Use it like any component inside a server component: ", "在服务端组件里像普通组件一样用: ")}<code class="font-mono text-sm">{"<Counter start={0} />"}</code>{loc(lc, ". Props must be JSON-serializable (they go into an HTML attribute); an island takes no children.", "。props 必须可 JSON 序列化(它会进 HTML 属性), 岛不接受 children。")}</p>
      </Section>

      <Section title={loc(lc, "5. Expose Go capabilities", "5. 暴露 Go 能力")} lead={loc(lc, "A host module = one Go value; fields map by json tag and methods by lowercased first letter", "宿主模块 = 一个 Go 值; 字段按 json tag、方法首字母小写映射到方言")}>
        <CodeBlock code={sampleHost} lang="go" title="host/host.go" />
        <CodeBlock code={sampleHostDTS} lang="tsx" title={loc(lc, "app/.gen/host.d.ts (generated)", "app/.gen/host.d.ts(生成)")} />
        <p>{loc(lc, "After compilation, ", "编译后 ")}<code class="font-mono text-sm">models.search(q)</code>{loc(lc, " is just ", " 就是 ")}<code class="font-mono text-sm">host.Data.Models.Search(q)</code>{loc(lc, ": no marshalling, no reflection. Go's ", ": 没有编组, 没有反射。Go 的 ")}<code class="font-mono text-sm">int</code>{loc(lc, " and the dialect's ", " 和方言的 ")}<code class="font-mono text-sm">number</code>{loc(lc, " convert automatically; for a method returning error, the error becomes a panic recovered by the request layer, and one wrapping ", " 自动转换; 返回 error 的方法, error 变成 panic 由请求层 recover, 包了 ")}<code class="font-mono text-sm">gotsx.ErrNotFound</code>{loc(lc, " turns into a 404.", " 的回 404。")}</p>
      </Section>

      <Section title={loc(lc, "6. From an island back to Go", "6. 从岛回到 Go")} lead={loc(lc, "An island reaches the server only over HTTP; an action is a plain handler", "岛只能通过 HTTP 回到服务端, 动作就是普通 handler")}>
        <CodeBlock code={sampleAction} lang="go" title="main.go" />
      </Section>

      <Section title={loc(lc, "7. Styling: Tailwind", "7. 样式: Tailwind")} lead={loc(lc, "If app/tailwind.css exists, the Tailwind standalone CLI runs on every build", "有 app/tailwind.css 就会在每次构建时跑 Tailwind standalone CLI")}>
        <CodeBlock code={sampleTailwind} lang="css" title="app/tailwind.css" />
        <p><code class="font-mono text-sm">class="..."</code>{loc(lc, " in the dialect is just a plain string; at build time Tailwind scans ", " 在方言里就是普通字符串, Tailwind 构建期扫描 ")}<code class="font-mono text-sm">app/**/*.tsx</code>{loc(lc, " to generate ", " 生成 ")}<code class="font-mono text-sm">public/tailwind.css</code>{loc(lc, ". Binary lookup order: ", "。二进制查找顺序: ")}<code class="font-mono text-sm">$GOTSX_TAILWIND</code>{loc(lc, " → the repo's ", " → 仓库 ")}<code class="font-mono text-sm">.tools/tailwindcss</code>{loc(lc, " → PATH. Also no Node. Every class on this site came from exactly this.", " → PATH。同样不需要 Node。这个站的每一个 class 都是这么来的。")}</p>
      </Section>
    </DocsLayout>
  );
}
