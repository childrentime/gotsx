import type { LayoutProps } from "gotsx";
import { now } from "host:intl";
import Counter from "../islands/Counter.client";

/** 根布局(文件约定): pages/ 下每个页面都被它包住; props 是 PageProps + children */
function titleOf(path: string): string {
  if (path === "/") return "Models";
  if (path.startsWith("/models/")) return "Model";
  if (path === "/kitchen") return "Kitchen sink";
  if (path.startsWith("/docs")) return "Docs";
  return "gotsx";
}

export default function Root({ path, children }: LayoutProps) {
  const nav = (href: string, on: boolean) => (on ? "nav-link nav-link-active" : "nav-link");
  return (
    <html lang="en" class="">
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1" />
        <title>{titleOf(path)} · gotsx example</title>
        <link rel="stylesheet" href="/public/app.css" />
      </head>
      <body>
        <div id="gotsx-bar"></div>
        <header class="page-header">
          <div class="container-page">
            <a href="/" class="brand"><span class="mark">g</span>example</a>
            <nav class="nav">
              <a href="/" class={nav("/", path === "/" || path.startsWith("/models/"))}>Models</a>
              <a href="/kitchen" class={nav("/kitchen", path === "/kitchen")}>Kitchen sink</a>
              <a href="/docs/getting/started" class={nav("/docs", path.startsWith("/docs"))}>Docs</a>
              <Counter start={0} />
            </nav>
          </div>
        </header>
        <main class="container-page fade-up">{children}</main>
        <footer class="muted">Rendered by Go, compiled from the gotsx dialect · server time {now()}</footer>
      </body>
    </html>
  );
}
