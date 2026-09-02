import type { Node } from "gotsx";
import type { Card } from "host:catalog";
import Stars from "./Stars";
import Img from "./Img";

/** 商品卡视觉(共享: 服务端和客户端信息流都用同一份)。右上角 wishlist 由调用方作为 children 传入 */
export default function CardShell({ card, children }: { card: Card; children?: Node }) {
  return (
    <div class="group relative">
      <a href={`/p/${card.id}`} class="card block overflow-hidden transition-colors hover:border-foreground/25">
        <div class="relative bg-muted">
          <Img src={`/img/p/${card.id}`} alt={card.title} className="aspect-square w-full object-cover" />
          <div class="absolute left-2 top-2 flex flex-col items-start gap-1">
            {card.flash && <span class="badge badge-primary">闪购</span>}
            {card.tag !== "" && <span class="badge badge-outline bg-background/90">{card.tag}</span>}
          </div>
          {card.soldOut && <div class="absolute inset-0 flex items-center justify-center bg-background/70 text-xs font-medium tracking-[.2em] text-muted-foreground">售罄</div>}
        </div>
        <div class="p-3">
          <div class="line-clamp-2 h-10 text-[13px] leading-5 text-foreground">{card.title}</div>
          <div class="mt-1.5 flex items-center gap-1.5 text-[11px] text-muted-foreground">
            <Stars rating={card.rating} />
            <span>{card.rating}</span>
            <span>·</span>
            <span>已售 {card.soldFmt}</span>
          </div>
          <div class="mt-2 flex items-baseline gap-1.5">
            <span class="text-[15px] font-semibold tracking-tight tabular-nums">{card.priceFmt}</span>
            {card.off > 0 && <span class="text-[11px] text-muted-foreground line-through tabular-nums">{card.origFmt}</span>}
            {card.off > 0 && <span class="badge badge-secondary ml-auto tabular-nums">-{card.off}%</span>}
          </div>
        </div>
      </a>
      <div class="absolute right-2 top-2">{children}</div>
    </div>
  );
}
