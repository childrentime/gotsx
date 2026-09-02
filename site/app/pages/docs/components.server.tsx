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
import { loc } from "../../ui/i18n";
import { sampleButton, faqs } from "../../content/site.server";

const demoA = `<Button>Primary</Button>
<Button variant="secondary">Secondary</Button>
<Button variant="outline">Outline</Button>
<Button variant="ghost">Ghost</Button>
<Button size="sm">Small</Button>
<Button size="lg" href="/docs">As a link</Button>`;

const demoB = `<Badge>Default</Badge>
<Badge color="brand">brand</Badge>
<Badge color="green">green</Badge>
<Badge color="amber">amber</Badge>
<Badge color="red">red</Badge>`;

export default function Components({ locale, path }: PageProps) {
  const lc = locale !== "" ? locale : "en";
  return (
    <DocsLayout title={loc(lc, "Component library", "组件库")} active="components" locale={lc} path={path}>
      <p class="text-[15px] leading-7 text-zinc-700 dark:text-zinc-300">
        {loc(lc, "Every component on this site is written in the dialect under ", "这个站的所有组件都用方言写在 ")}<code class="font-mono text-sm">site/app/ui/</code>{loc(lc, ", styled with Tailwind classes. They are ", ", 样式是 Tailwind class。它们是")}<strong>{loc(lc, "shared components", "共享组件")}</strong>{loc(lc, ": one source, compiled to Go on the server and to DOM instructions on the client, so they can appear in pages and in islands alike. The API deliberately looks like MUI / shadcn; the implementation is entirely its own.", ": 同一份源码, 服务端编成 Go, 客户端编成 DOM 指令, 所以既能出现在页面里, 也能出现在岛里。API 刻意长得像 MUI / shadcn, 实现完全是自己的。")}
      </p>
      <Callout kind="warn" title={loc(lc, "Why not MUI", "为什么不用 MUI")}>
        {loc(lc, "MUI is dynamic JS running on the React runtime (forwardRef, Emotion, spreading {...props}); it can't pass the subset check and can't compile to Go. A dialect library must be written in the dialect — just as AssemblyScript can't use npm packages directly.", "MUI 是跑在 React 运行时上的动态 JS(forwardRef、Emotion、{...props} 铺属性), 不可能通过子集检查, 也不可能编译成 Go。方言的库必须用方言写——就像 AssemblyScript 不能直接用 npm 包。")}
      </Callout>

      <Section title="Button" lead="ui/Button.tsx · variant / size / href / disabled / onClick">
        <div class="flex flex-wrap items-center gap-3 rounded-xl border border-zinc-200 p-5 dark:border-zinc-800">
          <Button>{loc(lc, "Primary", "主要")}</Button>
          <Button variant="secondary">{loc(lc, "Secondary", "次要")}</Button>
          <Button variant="outline">{loc(lc, "Outline", "描边")}</Button>
          <Button variant="ghost">{loc(lc, "Ghost", "幽灵")}</Button>
          <Button size="sm">{loc(lc, "Small", "小")}</Button>
          <Button size="lg" href={lpath(lc, "/docs")}>{loc(lc, "As a link", "链接形态")}</Button>
          <Button disabled>{loc(lc, "Disabled", "禁用")}</Button>
        </div>
        <CodeBlock code={demoA} lang="tsx" title={loc(lc, "usage", "用法")} />
        <CodeBlock code={sampleButton} lang="tsx" title={loc(lc, "ui/Button.tsx (excerpt): destructured defaults + a ternary choosing the element type", "ui/Button.tsx(节选): 解构默认值 + 三元选择元素类型")} />
        <p class="mt-2">{loc(lc, "Use the same Button inside an island — its children are reactive text, prerendered by the server and taken over by the browser:", "在岛里用同一个 Button——它的 children 是响应式文本, 服务端预渲染, 浏览器接管:")}</p>
        <div class="rounded-xl border border-zinc-200 p-5 dark:border-zinc-800"><ButtonPlayground /></div>
      </Section>

      <Section title="Badge" lead="ui/Badge.tsx · color">
        <div class="flex flex-wrap items-center gap-2 rounded-xl border border-zinc-200 p-5 dark:border-zinc-800">
          <Badge>{loc(lc, "Default", "默认")}</Badge>
          <Badge color="brand">brand</Badge>
          <Badge color="green">green</Badge>
          <Badge color="amber">amber</Badge>
          <Badge color="red">red</Badge>
        </div>
        <CodeBlock code={demoB} lang="tsx" title={loc(lc, "usage", "用法")} />
      </Section>

      <Section title="Card / Stat / Callout" lead={loc(lc, "Three containers for server-side layout", "服务端布局用的三个容器")}>
        <div class="grid gap-4 sm:grid-cols-3">
          <Card icon="🧩" title="Card">{loc(lc, "A container with an icon and title; children is any node.", "带 icon 和 title 的容器, children 是任意节点。")}</Card>
          <Stat value="30 µs" label={loc(lc, "Stat: a number + a caption", "Stat: 数字 + 说明")} />
          <Card title={loc(lc, "Nested", "嵌套")}>
            <Badge color="green">{loc(lc, "Components nest arbitrarily", "组件可以任意嵌套")}</Badge>
          </Card>
        </div>
        <Callout kind="tip" title="Callout">{loc(lc, "kind = info | tip | warn. What you're reading is one.", "kind = info | tip | warn。你正在看的就是一个。")}</Callout>
      </Section>

      <Section title="CodeBlock / CodeTabs" lead={loc(lc, "Highlighting is done by a tokenizer written in Go (host:hl); the dialect only renders spans", "高亮由 Go 写的 tokenizer(host:hl)完成, 方言只渲染 span")}>
        <p>{loc(lc, "The server component ", "服务端组件 ")}<code class="font-mono text-sm">CodeBlock</code>{loc(lc, " calls the host directly; the island ", " 直接调宿主; 岛 ")}<code class="font-mono text-sm">CodeTabs</code>{loc(lc, " receives the token arrays the page tokenized on the server (props), so both sides only render — client code never touches Go.", " 拿的是页面在服务端切好的 token 数组(props), 两端都只负责渲染——客户端代码永远碰不到 Go。")}</p>
        <CodeTabs tabs={[
          { label: "tsx", tokens: tokens(demoA, "tsx") },
          { label: "go", tokens: tokens(`func Hello(name string) gotsx.Node {\n\treturn gotsx.El("p", nil, gotsx.Text("hi " + name))\n}`, "go") },
          { label: "css", tokens: tokens(`@import "tailwindcss";\n.card { @apply rounded-xl border; }`, "css") },
        ]} />
      </Section>

      <Section title="Accordion" lead="islands/Accordion.client.tsx · items: { q, a }[]">
        <Accordion items={faqs(lc).slice(0, 3)} />
      </Section>
    </DocsLayout>
  );
}
