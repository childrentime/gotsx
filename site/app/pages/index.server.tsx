import type { PageProps } from "gotsx";
import { tokens } from "host:hl";
import Layout from "../components/Layout.server";
import Button from "../ui/Button";
import Card from "../ui/Card";
import Stat from "../ui/Stat";
import Badge from "../ui/Badge";
import Callout from "../ui/Callout";
import Counter from "../islands/Counter.client";
import CodeTabs from "../islands/CodeTabs.client";
import Accordion from "../islands/Accordion.client";
import { features, faqs, sampleTSX, sampleGo, sampleIsland, sampleIslandJS, sampleIslandGo } from "../content/site.server";

export default function Home({ query }: PageProps) {
  const start = query.start !== "" ? Number(query.start) : 0;
  return (
    <Layout title="TSX 方言, Go 原生的全栈框架">
      {/* hero */}
      <section class="mx-auto max-w-6xl px-5 pb-16 pt-20">
        <div class="mb-5 flex flex-wrap items-center gap-2">
          <Badge color="brand">PoC 2</Badge>
          <Badge>零 Node</Badge>
          <Badge>零 JS 引擎</Badge>
          <Badge>零 npm</Badge>
        </div>
        <h1 class="max-w-3xl text-4xl font-bold leading-tight tracking-tight md:text-6xl">
          借 React 的思想,<br />
          <span class="text-brand-600 dark:text-brand-300">编译到 Go</span> 的全栈框架
        </h1>
        <p class="mt-6 max-w-2xl text-lg leading-8 text-zinc-600 dark:text-zinc-300">
          一份 TSX 源码, 两个编译器: 服务端编成 Go 函数, 客户端编成 signals。
          没有 vdom, 没有 JS 引擎, 没有 Node。工具链只有 Go。
        </p>
        <div class="mt-8 flex flex-wrap items-center gap-3">
          <Button href="/docs" size="lg">开始使用</Button>
          <Button href="/docs/architecture" size="lg" variant="outline">它是怎么工作的</Button>
          <span class="ml-2 text-sm text-zinc-500">← 这些按钮是方言写的组件库</span>
        </div>
        <div class="mt-10 flex flex-wrap items-center gap-4">
          <Counter start={start} />
          <span class="text-sm text-zinc-500">这个按钮是一个岛: Go 渲染了它, 浏览器 hydrate 后它是一个 signal。点它。</span>
        </div>
      </section>

      {/* numbers */}
      <section class="border-y border-zinc-200 bg-zinc-50 dark:border-zinc-800 dark:bg-zinc-900/40">
        <div class="mx-auto grid max-w-6xl gap-4 px-5 py-10 sm:grid-cols-2 lg:grid-cols-4">
          <Stat value="30 µs" label="列表页服务端渲染(goja 版是 50 ms)" />
          <Stat value="28k req/s" label="ab -c 64, 单进程, p99 6 ms" />
          <Stat value="6 KB" label="客户端运行时(未压缩)" />
          <Stat value="1.8 s" label="改 TSX 到新版本上线" />
        </div>
      </section>

      {/* one source two compilers */}
      <section class="mx-auto max-w-6xl px-5 pt-20">
        <h2 class="text-2xl font-bold tracking-tight">一份源码, 两个编译器</h2>
        <p class="mt-2 max-w-2xl text-zinc-600 dark:text-zinc-300">
          <code class="font-mono text-sm">.server.tsx</code> 只编成 Go。
          <code class="font-mono text-sm">.client.tsx</code> 同时编成 Go(单趟求值做 SSR)和 JS(signals 做交互)。
          切换标签看同一份源码变成什么。
        </p>
        <div class="mt-6 grid gap-6 lg:grid-cols-2">
          <div>
            <div class="mb-2 text-sm font-semibold">服务端页面 → Go</div>
            <CodeTabs tabs={[{ label: "index.server.tsx", tokens: tokens(sampleTSX, "tsx") }, { label: "生成的 Go", tokens: tokens(sampleGo, "go") }]} />
          </div>
          <div>
            <div class="mb-2 text-sm font-semibold">岛 → Go + JS</div>
            <CodeTabs tabs={[{ label: "Counter.client.tsx", tokens: tokens(sampleIsland, "tsx") }, { label: "Go(SSR)", tokens: tokens(sampleIslandGo, "go") }, { label: "JS(signals)", tokens: tokens(sampleIslandJS, "js") }]} />
          </div>
        </div>
        <Callout kind="tip" title="为什么 TSX 能编成 Go">
          SSR 是一次同步的单趟求值: 没有重渲染、没有 effect、setter 永远不会被调用。所以服务端不需要"React 运行时",
          只需要组件的"渲染切片"——而这个切片的语义小到可以编译成 Go。
        </Callout>
      </section>

      {/* features */}
      <section class="mx-auto max-w-6xl px-5 pt-16">
        <h2 class="text-2xl font-bold tracking-tight">特性</h2>
        <div class="mt-6 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          {features.map((f) => <Card icon={f.icon} title={f.title}>{f.desc}</Card>)}
        </div>
      </section>

      {/* honesty */}
      <section class="mx-auto max-w-6xl px-5 pt-16">
        <h2 class="text-2xl font-bold tracking-tight">它不是什么</h2>
        <div class="mt-6 grid gap-4 md:grid-cols-3">
          <Card title="不是 React">
            没有 React 运行时, 不能用 npm 上的 React 组件库。借的是思想和语法手感, 语义是静态的。
          </Card>
          <Card title="不是 TypeScript">
            是一门借 TSX 语法的静态语言(AssemblyScript 的同类): 类型系统限定在 Go 能表示的集合里, 出了子集是编译错误。
          </Card>
          <Card title="不是 Next">
            没有 RSC 协议、Suspense、流式渲染。有的是 Go 的速度、单二进制和 6KB 客户端。
          </Card>
        </div>
      </section>

      {/* faq */}
      <section class="mx-auto max-w-3xl px-5 pt-16">
        <h2 class="text-2xl font-bold tracking-tight">常见问题</h2>
        <div class="mt-6"><Accordion items={faqs} /></div>
      </section>

      <section class="mx-auto max-w-6xl px-5 pt-20 text-center">
        <h2 class="text-2xl font-bold tracking-tight">五分钟跑起来</h2>
        <div class="mt-6 flex justify-center gap-3">
          <Button href="/docs" size="lg">快速开始</Button>
          <Button href="/docs/language" size="lg" variant="secondary">语言参考</Button>
        </div>
      </section>
    </Layout>
  );
}
