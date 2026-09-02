import type { Node } from "gotsx";
import type { Card } from "host:catalog";
import Stars from "./Stars";
import Img from "./Img";

/** 商品卡视觉(共享: 服务端和客户端信息流都用同一份)。右上角 wishlist 由调用方作为 children 传入 */
export default function CardShell({ card, children }: { card: Card; children?: Node }) {
  return (
    <div class="group relative">
      <a href={`/p/${card.id}`} class="block overflow-hidden rounded-xl2 border border-ink-100 bg-white shadow-card transition-all duration-300 hover:-translate-y-1 hover:border-brand-200 hover:shadow-cardhover">
        <div class="relative">
          <Img src={`/img/p/${card.id}`} alt={card.title} className="aspect-square w-full rounded-t-xl2 object-cover" />
          <div class="absolute left-2 top-2 flex flex-col items-start gap-1">
            {card.flash && <span class="rounded-md bg-brand-500 px-1.5 py-0.5 text-[10px] font-bold tracking-wide text-white shadow-sm">⚡ 闪购</span>}
            {card.tag !== "" && <span class="rounded-md bg-white/90 px-1.5 py-0.5 text-[10px] font-semibold text-brand-600 shadow-sm backdrop-blur">{card.tag}</span>}
          </div>
          {card.off > 0 && <span class="absolute right-2 top-2 rounded-md bg-ink-900/85 px-1.5 py-0.5 text-[10px] font-bold text-white backdrop-blur">-{card.off}%</span>}
          {card.soldOut && <div class="absolute inset-0 flex items-center justify-center bg-white/72 text-sm font-bold tracking-[.2em] text-ink-500 backdrop-blur-[1px]">已 售 罄</div>}
        </div>
        <div class="p-2.5">
          <div class="line-clamp-2 h-9 text-[13px] font-medium leading-[18px] text-ink-800">{card.title}</div>
          <div class="mt-1.5 flex items-center gap-1.5">
            <Stars rating={card.rating} />
            <span class="text-[11px] text-ink-400">{card.rating}</span>
            <span class="text-[11px] text-ink-300">·</span>
            <span class="text-[11px] text-ink-400">已售 {card.soldFmt}</span>
          </div>
          <div class="mt-1.5 flex items-end gap-1.5">
            <span class="text-[10px] font-bold text-brand-600">¥</span>
            <span class="-ml-0.5 text-[18px] font-extrabold leading-none tracking-tight text-brand-600">{card.priceFmt.slice(1)}</span>
            {card.off > 0 && <span class="mb-px text-[11px] text-ink-300 line-through">{card.origFmt}</span>}
          </div>
        </div>
      </a>
      <div class="absolute right-2.5 top-2.5">{children}</div>
    </div>
  );
}
