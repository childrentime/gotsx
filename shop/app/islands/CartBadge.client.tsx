import { useState, useEffect, on } from "gotsx";

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
    <a href="/cart" class="relative flex shrink-0 items-center gap-1.5 rounded-full bg-ink-900 px-4 py-2 text-sm font-semibold text-white transition hover:bg-brand-600">
      <span class="text-base">🛒</span>
      <span class="hidden sm:inline">{label}</span>
      {n > 0 && <span class={beat ? "heart-pop absolute -right-1.5 -top-1.5 flex h-5 min-w-5 items-center justify-center rounded-full bg-brand-500 px-1 text-[11px] font-bold text-white ring-2 ring-white" : "absolute -right-1.5 -top-1.5 flex h-5 min-w-5 items-center justify-center rounded-full bg-brand-500 px-1 text-[11px] font-bold text-white ring-2 ring-white"}>{n}</span>}
    </a>
  );
}
