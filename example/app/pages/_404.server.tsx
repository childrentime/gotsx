import type { PageProps, Meta } from "gotsx";

/** pages/_404.server.tsx → gen.NotFound(文件约定), 由根布局包住 */
export function meta(): Meta {
  return { title: "Not found", noIndex: true };
}

export default function NotFound({ path }: PageProps) {
  return (
    <div class="empty">
      <h1>404</h1>
      <p class="muted">There is nothing at <code>{path}</code>.</p>
      <a href="/" class="btn btn-outline">Back to models</a>
    </div>
  );
}
