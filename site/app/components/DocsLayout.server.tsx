import type { Node } from "gotsx";
import { loc } from "../ui/i18n";
import Layout from "./Layout.server";

interface Link { href: string; en: string; zh: string; key: string; }

const links: Link[] = [
  { href: "/docs", en: "Quick start", zh: "快速开始", key: "docs" },
  { href: "/docs/language", en: "Language reference", zh: "语言参考", key: "language" },
  { href: "/docs/components", en: "Component library", zh: "组件库", key: "components" },
  { href: "/docs/architecture", en: "Architecture", zh: "架构与原理", key: "architecture" },
];

function sideCls(active: string, me: string): string {
  return active === me
    ? "block rounded-md bg-brand-50 px-3 py-1.5 text-sm font-medium text-brand-700 dark:bg-brand-950/50 dark:text-brand-300"
    : "block rounded-md px-3 py-1.5 text-sm text-zinc-600 hover:bg-zinc-100 dark:text-zinc-300 dark:hover:bg-zinc-800";
}

export default function DocsLayout({ title, active, locale = "en", path = "/", children }: { title: string; active: string; locale?: string; path?: string; children?: Node }) {
  const lc = locale !== "" ? locale : "en";
  return (
    <Layout title={title} active={active} locale={lc} path={path}>
      <div class="mx-auto grid max-w-6xl gap-10 px-5 py-10 md:grid-cols-[210px_1fr]">
        <aside class="md:sticky md:top-20 md:self-start">
          <div class="mb-2 px-3 text-xs font-semibold uppercase tracking-wider text-zinc-400">{loc(lc, "Docs", "文档")}</div>
          <nav class="space-y-0.5">{links.map((l) => <a href={lpath(lc, l.href)} class={sideCls(active, l.key)}>{loc(lc, l.en, l.zh)}</a>)}</nav>
        </aside>
        <article class="min-w-0 max-w-3xl">
          <h1 class="mb-6 text-3xl font-bold tracking-tight">{title}</h1>
          {children}
        </article>
      </div>
    </Layout>
  );
}
