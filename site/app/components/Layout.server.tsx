import type { Node } from "gotsx";
import { goVersion, version } from "host:site";
import { loc } from "../ui/i18n";
import ThemeToggle from "../islands/ThemeToggle.client";
import LocaleSwitch from "../islands/LocaleSwitch.client";

function navCls(active: string, me: string): string {
  return active === me ? "nav-link-active" : "nav-link";
}

export default function Layout({ title, active = "", locale = "en", path = "/", children }: { title: string; active?: string; locale?: string; path?: string; children?: Node }) {
  const lc = locale !== "" ? locale : "en";
  return (
    <html lang={lc}>
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1" />
        <title>{title} · gotsx</title>
        <meta name="description" content={loc(lc, "gotsx: a full-stack framework that borrows React + TSX ideas and compiles to a Go dialect. No JS engine, no Node.", "gotsx: 借 React + TSX 的思想, 用一门编译到 Go 的方言做的全栈框架。没有 JS 引擎, 没有 Node。")} />
        <script src="/public/theme.js"></script>
        <link rel="stylesheet" href="/public/tailwind.css" />
      </head>
      <body class="min-h-screen bg-background font-sans text-foreground antialiased">
        <div id="gotsx-bar"></div>
        <header class="page-header">
          <div class="container-page flex h-14 items-center justify-between gap-4">
            <a href={lpath(lc, "/")} class="flex shrink-0 items-center gap-2 text-sm font-semibold tracking-tight">
              <span class="inline-flex h-6 w-6 items-center justify-center rounded-md bg-primary font-mono text-xs font-semibold text-primary-foreground">g</span>
              gotsx
              <span class="badge badge-secondary ml-1 font-mono font-normal">{version()}</span>
            </a>
            <nav class="no-scrollbar flex min-w-0 items-center gap-1 overflow-x-auto">
              <a href={lpath(lc, "/docs")} class={navCls(active, "docs")}>{loc(lc, "Guide", "指南")}</a>
              <a href={lpath(lc, "/docs/language")} class={navCls(active, "language")}>{loc(lc, "Language", "语言")}</a>
              <a href={lpath(lc, "/docs/components")} class={navCls(active, "components")}>{loc(lc, "Components", "组件")}</a>
              <a href={lpath(lc, "/docs/architecture")} class={navCls(active, "architecture")}>{loc(lc, "Architecture", "架构")}</a>
              <span class="mx-1 h-5 w-px shrink-0 bg-border"></span>
              <LocaleSwitch locale={lc} path={path} label={loc(lc, "中文", "EN")} />
              <ThemeToggle />
            </nav>
          </div>
        </header>
        <main>{children}</main>
        <footer class="mt-24 border-t border-border py-10">
          <div class="container-page text-center text-xs leading-6 text-muted-foreground">
            {loc(lc, "This site is itself a gotsx app: pages rendered by Go, islands compiled from the dialect to signals, styled with Tailwind. ", "这个站本身就是 gotsx 应用: 页面由 Go 渲染, 岛由方言编译成 signals, 样式是 Tailwind。")}{goVersion()}
          </div>
        </footer>
      </body>
    </html>
  );
}
