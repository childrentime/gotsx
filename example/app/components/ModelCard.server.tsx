import type { Model } from "host:data";
import LikeButton from "../islands/LikeButton.client";

export default function ModelCard({ model }: { model: Model }) {
  return (
    <article class="card">
      <a class="title" href={`/models/${model.id}`}>{model.title}</a>
      <p class="muted">{model.author} · {model.desc}</p>
      <div class="row">
        {model.tags.map((t) => <span class="chip">{t}</span>)}
        <LikeButton id={model.id} likes={model.likes} />
      </div>
    </article>
  );
}
