import type { LayoutProps } from "gotsx";
import { now } from "host:intl";
import Counter from "../islands/Counter.client";

/** Root layout (file convention): wraps every page under pages/; props are PageProps + meta (the page's export function meta) + children */
export default function Root({ path, meta, children }: LayoutProps) {
  const nav = (href: string, on: boolean) => (on ? "nav-link nav-link-active" : "nav-link");
  return (
    <html lang="en" class="">
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1" />
        <title>{meta.title ? meta.title + " · gotsx example" : "gotsx example"}</title>
        {meta.description && <meta name="description" content={meta.description} />}
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
