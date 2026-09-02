import type { Node } from "gotsx";
import { goVersion, version } from "host:site";
import ThemeToggle from "../islands/ThemeToggle.client";

function navCls(active: string, me: string): string {
  return active === me
    ? "rounded-md px-3 py-1.5 text-sm font-medium text-brand-700 dark:text-brand-300"
    : "rounded-md px-3 py-1.5 text-sm text-zinc-600 hover:text-zinc-900 dark:text-zinc-300 dark:hover:text-zinc-50";
}

export default function Layout({ title, active = "", children }: { title: string; active?: string; children?: Node }) {
  return (
    <html lang="zh-CN">
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1" />
        <title>{title} · gotsx</title>
        <meta name="description" content="gotsx: 借 React + TSX 的思想, 用一门编译到 Go 的方言做的全栈框架。没有 JS 引擎, 没有 Node。" />
        <script src="/public/theme.js"></script>
        <link rel="stylesheet" href="/public/tailwind.css" />
      </head>
      <body class="min-h-screen bg-white font-sans text-zinc-900 antialiased dark:bg-zinc-950 dark:text-zinc-100">
        <header class="sticky top-0 z-20 border-b border-zinc-200/80 bg-white/80 backdrop-blur dark:border-zinc-800 dark:bg-zinc-950/80">
          <div class="mx-auto flex h-14 max-w-6xl items-center justify-between px-5">
            <a href="/" class="flex items-center gap-2 font-bold tracking-tight">
              <span class="inline-flex h-7 w-7 items-center justify-center rounded-md bg-brand-600 font-mono text-sm text-white">g</span>
              gotsx
              <span class="ml-1 rounded-full bg-zinc-100 px-2 py-0.5 font-mono text-[11px] font-normal text-zinc-500 dark:bg-zinc-800 dark:text-zinc-400">{version()}</span>
            </a>
            <nav class="flex items-center gap-1">
              <a href="/docs" class={navCls(active, "docs")}>指南</a>
              <a href="/docs/language" class={navCls(active, "language")}>语言</a>
              <a href="/docs/components" class={navCls(active, "components")}>组件</a>
              <a href="/docs/architecture" class={navCls(active, "architecture")}>架构</a>
              <span class="mx-1 h-5 w-px bg-zinc-200 dark:bg-zinc-800"></span>
              <ThemeToggle />
            </nav>
          </div>
        </header>
        <main>{children}</main>
        <footer class="mt-24 border-t border-zinc-200 py-10 text-center text-xs text-zinc-500 dark:border-zinc-800">
          这个站本身就是 gotsx 应用: 页面由 Go 渲染, 岛由方言编译成 signals, 样式是 Tailwind。{goVersion()}
        </footer>
      </body>
    </html>
  );
}
