import { useState, useEffect, on } from "gotsx";
import Icon from "../ui/Icon";

export default function CartBadge({ count, label = "" }: { count: number; label?: string }) {
  const [n, setN] = useState(count);
  const [beat, setBeat] = useState(false);
  useEffect(() => {
    on("cart:changed", (d: any) => {
      setN(d.count);
      setBeat(true);
      setTimeout(() => setBeat(false), 400);
    });
  });
  return (
    <a href="/cart" class="btn btn-outline relative shrink-0">
      <Icon name="cart" />
      <span class="hidden sm:inline">{label}</span>
      {n > 0 && <span class={beat ? "heart-pop absolute -right-2 -top-2 flex h-5 min-w-5 items-center justify-center rounded-full bg-primary px-1.5 text-[11px] font-medium text-primary-foreground tabular-nums" : "absolute -right-2 -top-2 flex h-5 min-w-5 items-center justify-center rounded-full bg-primary px-1.5 text-[11px] font-medium text-primary-foreground tabular-nums"}>{n}</span>}
    </a>
  );
}
