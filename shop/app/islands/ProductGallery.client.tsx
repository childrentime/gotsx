import { useState } from "gotsx";

/** Product gallery: main image + thumbnails. Real <img> (server-generated SVG studio shots), lazy + alt */
export default function ProductGallery({ id, hue, count, locale = "" }: { id: string; hue: number; count: number; locale?: string }) {
  const [cur, setCur] = useState(0);
  const idx = [0, 1, 2, 3].slice(0, count);
  const src = (i: number) => (i === 0 ? `/img/p/${id}` : `/img/p/${id}?g=${i - 1}`);
  return (
    <div data-hue={hue}>
      <img src={src(cur)} alt={t(locale, "gallery.alt")} width={400} height={400} decoding="async" class="aspect-square w-full rounded-md border border-border bg-muted object-cover" />
      <div class="mt-3 flex gap-2">
        {idx.map((i) => (
          <button
            class={cur === i ? "h-16 w-16 overflow-hidden rounded-md border-2 border-foreground bg-muted" : "h-16 w-16 overflow-hidden rounded-md border border-border bg-muted transition-colors hover:border-foreground/40"}
            onClick={() => setCur(i)}
          >
            <img src={src(i)} alt={t(locale, "gallery.thumb")} width={64} height={64} loading="lazy" decoding="async" class="h-full w-full object-cover" />
          </button>
        ))}
      </div>
    </div>
  );
}
