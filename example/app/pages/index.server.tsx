import type { PageProps, Meta } from "gotsx";
import { Suspense } from "gotsx";
import { models } from "host:data";        // Go 实现: 编译后是直接的 Go 调用, 零编组
import { fmtNumber } from "host:intl";
import ModelCard from "../components/ModelCard.server";
import Stats from "../components/Stats.server";

/** Page metadata: the layout renders it (props.meta) into <title> / <meta name="description"> */
export function meta(): Meta {
  return { title: "Models", description: "A gotsx example: pages compiled to Go, islands compiled to signals." };
}

export default function Home({ query }: PageProps) {
  const q = query.q ?? "";
  const list = models.search(q);           // 同步: 并发由 goroutine 提供, 语言里没有 async
  return (
    <div class="stack">
      <div>
        <h1>Models</h1>
        <p class="muted">A small catalog served by a Go host module. The stats panel below is a streaming <code>Suspense</code> boundary: the page ships first, the slow query fills in.</p>
      </div>
      <Suspense fallback={<div class="card grid grid-3"><div class="skeleton skeleton-line"></div><div class="skeleton skeleton-line"></div><div class="skeleton skeleton-line"></div></div>}>
        <Stats />
      </Suspense>
      <form method="get" action="/" class="row">
        <input name="q" value={q} placeholder="Search models… (Enter submits, SPA navigation)" />
        <button class="btn">Search</button>
      </form>
      <p class="muted">
        {fmtNumber(list.length)} models{q !== "" && <span> · matching “{q}”</span>}
      </p>
      <div class="grid grid-2">{list.map((m) => <ModelCard model={m} />)}</div>
      {list.length === 0 && <p class="empty">No matching models</p>}
    </div>
  );
}
