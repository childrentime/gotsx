import type { PageProps } from "gotsx";
import { models } from "host:data";
import { fmtDate, fmtNumber } from "host:intl";
import Layout from "../../components/Layout.server";
import LikeButton from "../../islands/LikeButton.client";
import TagPicker from "../../islands/TagPicker.client";

export default function ModelPage({ params }: PageProps) {
  const m = models.get(params.id);         // 不存在: Go 返回 error → 路由层 404
  return (
    <Layout title={m.title}>
      <a href="/" class="back">← 返回列表</a>
      <article class="detail">
        <h1>{m.title}</h1>
        <p class="muted">
          {m.author} · 发布于 {fmtDate(m.createdAt)} · {fmtNumber(m.likes)} 次点赞
        </p>
        <p>{m.desc}</p>
        <div class="row">
          <TagPicker tags={m.tags} />
          <LikeButton id={m.id} likes={m.likes} />
        </div>
      </article>
    </Layout>
  );
}
