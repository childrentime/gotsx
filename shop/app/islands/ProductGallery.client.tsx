import { useState } from "gotsx";

/** 商品图廊: 主图 + 缩略图切换。用真实 <img>(服务端生成的 SVG 棚拍图),懒加载 + alt */
export default function ProductGallery({ id, hue, count }: { id: string; hue: number; count: number }) {
  const [cur, setCur] = useState(0);
  const idx = [0, 1, 2, 3].slice(0, count);
  const src = (i: number) => (i === 0 ? `/img/p/${id}` : `/img/p/${id}?g=${i - 1}`);
  return (
    <div data-hue={hue}>
      <img src={src(cur)} alt="商品图" width={400} height={400} decoding="async" class="aspect-square w-full rounded-md border border-border bg-muted object-cover" />
      <div class="mt-3 flex gap-2">
        {idx.map((i) => (
          <button
            class={cur === i ? "h-16 w-16 overflow-hidden rounded-md border-2 border-foreground bg-muted" : "h-16 w-16 overflow-hidden rounded-md border border-border bg-muted transition-colors hover:border-foreground/40"}
            onClick={() => setCur(i)}
          >
            <img src={src(i)} alt="缩略图" width={64} height={64} loading="lazy" decoding="async" class="h-full w-full object-cover" />
          </button>
        ))}
      </div>
    </div>
  );
}
