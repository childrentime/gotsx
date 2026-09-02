import type { PageProps, Meta } from "gotsx";

/** catch-all 路由: /docs/a/b/c → params.slug = "a/b/c" */
export function meta({ params }: PageProps): Meta {
  return { title: "Docs / " + params.slug, noIndex: true };
}

export default function Doc({ params }: PageProps) {
  const parts = params.slug.split("/");
  return (
    <div class="stack">
      <h1>docs / {parts.join(" / ")}</h1>
      <p class="muted">{parts.length} segments (catch-all route <code>pages/docs/[...slug].server.tsx</code>)</p>
      <div class="row">{parts.map((p) => <span class="badge badge-secondary">{p}</span>)}</div>
    </div>
  );
}
