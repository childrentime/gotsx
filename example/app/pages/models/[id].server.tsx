import type { PageProps } from "gotsx";
import { models } from "host:data";
import { fmtDate, fmtNumber } from "host:intl";
import LikeButton from "../../islands/LikeButton.client";
import TagPicker from "../../islands/TagPicker.client";

export default function ModelPage({ params }: PageProps) {
  const m = models.get(params.id);         // 不存在: Go 返回 error → 路由层 404(_404.server.tsx)
  return (
    <article class="stack">
      <a href="/" class="link muted">← All models</a>
      <div class="card">
        <h1>{m.title}</h1>
        <p class="muted">{m.author} · published {fmtDate(m.createdAt)} · {fmtNumber(m.likes)} likes</p>
        <p>{m.desc}</p>
        <div class="separator"></div>
        <div class="row" style="margin-top:12px">
          <TagPicker tags={m.tags} />
          <LikeButton id={m.id} likes={m.likes} />
        </div>
      </div>
    </article>
  );
}
