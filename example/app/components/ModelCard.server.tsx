import type { Model } from "host:data";
import LikeButton from "../islands/LikeButton.client";

export default function ModelCard({ model }: { model: Model }) {
  return (
    <article class="card stack">
      <a class="title" href={`/models/${model.id}`}>{model.title}</a>
      <p class="muted">{model.author} · {model.desc}</p>
      <div class="row">
        {model.tags.map((t) => <span class="badge badge-secondary">{t}</span>)}
        <span style="flex:1"></span>
        <LikeButton id={model.id} likes={model.likes} />
      </div>
    </article>
  );
}
