import type { PageProps } from "gotsx";
import { tokens } from "host:hl";
import DocsLayout from "../../components/DocsLayout.server";
import Section from "../../components/Section.server";
import Button from "../../ui/Button";
import Badge from "../../ui/Badge";
import Card from "../../ui/Card";
import Callout from "../../ui/Callout";
import Stat from "../../ui/Stat";
import CodeBlock from "../../ui/CodeBlock.server";
import CodeTabs from "../../islands/CodeTabs.client";
import Accordion from "../../islands/Accordion.client";
import ButtonPlayground from "../../islands/ButtonPlayground.client";
import { sampleButton, faqs } from "../../content/site.server";

const demoA = `<Button>主要</Button>
<Button variant="secondary">次要</Button>
<Button variant="outline">描边</Button>
<Button variant="ghost">幽灵</Button>
<Button size="sm">小</Button>
<Button size="lg" href="/docs">链接形态</Button>`;

const demoB = `<Badge>默认</Badge>
<Badge color="brand">brand</Badge>
<Badge color="green">green</Badge>
<Badge color="amber">amber</Badge>
<Badge color="red">red</Badge>`;

export default function Components({ path }: PageProps) {
  return (
    <DocsLayout title="组件库" active="components">
      <p class="text-[15px] leading-7 text-zinc-700 dark:text-zinc-300">
        这个站的所有组件都用方言写在 <code class="font-mono text-sm">site/app/ui/</code>, 样式是 Tailwind class。
        它们是<strong>共享组件</strong>: 同一份源码, 服务端编成 Go, 客户端编成 DOM 指令, 所以既能出现在页面里, 也能出现在岛里。
        API 刻意长得像 MUI / shadcn, 实现完全是自己的。
      </p>
      <Callout kind="warn" title="为什么不用 MUI">
        MUI 是跑在 React 运行时上的动态 JS(forwardRef、Emotion、{"{...props}"} 铺属性), 不可能通过子集检查, 也不可能编译成 Go。
        方言的库必须用方言写——就像 AssemblyScript 不能直接用 npm 包。
      </Callout>

      <Section title="Button" lead="ui/Button.tsx · variant / size / href / disabled / onClick">
        <div class="flex flex-wrap items-center gap-3 rounded-xl border border-zinc-200 p-5 dark:border-zinc-800">
          <Button>主要</Button>
          <Button variant="secondary">次要</Button>
          <Button variant="outline">描边</Button>
          <Button variant="ghost">幽灵</Button>
          <Button size="sm">小</Button>
          <Button size="lg" href="/docs">链接形态</Button>
          <Button disabled>禁用</Button>
        </div>
        <CodeBlock code={demoA} lang="tsx" title="用法" />
        <CodeBlock code={sampleButton} lang="tsx" title="ui/Button.tsx(节选): 解构默认值 + 三元选择元素类型" />
        <p class="mt-2">在岛里用同一个 Button——它的 children 是响应式文本, 服务端预渲染, 浏览器接管:</p>
        <div class="rounded-xl border border-zinc-200 p-5 dark:border-zinc-800"><ButtonPlayground /></div>
      </Section>

      <Section title="Badge" lead="ui/Badge.tsx · color">
        <div class="flex flex-wrap items-center gap-2 rounded-xl border border-zinc-200 p-5 dark:border-zinc-800">
          <Badge>默认</Badge>
          <Badge color="brand">brand</Badge>
          <Badge color="green">green</Badge>
          <Badge color="amber">amber</Badge>
          <Badge color="red">red</Badge>
        </div>
        <CodeBlock code={demoB} lang="tsx" title="用法" />
      </Section>

      <Section title="Card / Stat / Callout" lead="服务端布局用的三个容器">
        <div class="grid gap-4 sm:grid-cols-3">
          <Card icon="🧩" title="Card">带 icon 和 title 的容器, children 是任意节点。</Card>
          <Stat value="30 µs" label="Stat: 数字 + 说明" />
          <Card title="嵌套">
            <Badge color="green">组件可以任意嵌套</Badge>
          </Card>
        </div>
        <Callout kind="tip" title="Callout">kind = info | tip | warn。你正在看的就是一个。</Callout>
      </Section>

      <Section title="CodeBlock / CodeTabs" lead="高亮由 Go 写的 tokenizer(host:hl)完成, 方言只渲染 span">
        <p>服务端组件 <code class="font-mono text-sm">CodeBlock</code> 直接调宿主; 岛 <code class="font-mono text-sm">CodeTabs</code> 拿的是页面在服务端切好的 token 数组(props), 两端都只负责渲染——客户端代码永远碰不到 Go。</p>
        <CodeTabs tabs={[
          { label: "tsx", tokens: tokens(demoA, "tsx") },
          { label: "go", tokens: tokens(`func Hello(name string) gotsx.Node {\n\treturn gotsx.El("p", nil, gotsx.Text("hi " + name))\n}`, "go") },
          { label: "css", tokens: tokens(`@import "tailwindcss";\n.card { @apply rounded-xl border; }`, "css") },
        ]} />
      </Section>

      <Section title="Accordion" lead="islands/Accordion.client.tsx · items: { q, a }[]">
        <Accordion items={faqs.slice(0, 3)} />
      </Section>
    </DocsLayout>
  );
}
