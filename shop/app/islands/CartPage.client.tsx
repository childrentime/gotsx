import { useState, emit } from "gotsx";
import type { CartView } from "host:cart";

export default function CartPage({ cart }: { cart: CartView }) {
  const [c, setC] = useState(cart);
  const [busy, setBusy] = useState(false);
  const update = async (id: string, variant: string, qty: number) => {
    setBusy(true);
    const r = await fetch("/actions/cart/set", { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ id, variant, qty }) });
    const d = await r.json();
    setC(d.cart);
    emit("cart:changed", { count: d.cart.count });
    setBusy(false);
  };
  return (
    <div class={busy ? "opacity-60 transition" : "transition"}>
      {c.empty ? (
        <div class="rounded-xl2 border border-ink-100 bg-white py-24 text-center">
          <div class="text-7xl">🛒</div>
          <p class="mt-4 text-ink-500">购物车还是空的</p>
          <a href="/" class="mt-5 inline-block rounded-full bg-brand-500 px-8 py-2.5 text-sm font-bold text-white transition hover:bg-brand-600">去挑点好物</a>
        </div>
      ) : (
        <div class="grid gap-5 lg:grid-cols-[1fr_320px]">
          <div class="divide-y divide-ink-100 overflow-hidden rounded-xl2 border border-ink-100 bg-white">
            {c.items.map((it) => (
              <div class="flex items-center gap-3.5 p-4">
                <a href={`/p/${it.id}`} class="flex h-20 w-20 shrink-0 items-center justify-center rounded-xl text-4xl" style={`background:radial-gradient(120% 120% at 50% 20%, #fff, hsl(${it.hue} 46% 95%))`}>{it.emoji}</a>
                <div class="min-w-0 flex-1">
                  <a href={`/p/${it.id}`} class="line-clamp-1 text-sm font-medium text-ink-800 transition hover:text-brand-600">{it.title}</a>
                  {it.variant !== "" && <div class="mt-1 inline-block rounded bg-ink-50 px-2 py-0.5 text-xs text-ink-500">{it.variant}</div>}
                  <div class="mt-1.5 text-[15px] font-bold text-brand-600">{it.priceFmt}</div>
                </div>
                <div class="flex items-center rounded-lg border border-ink-200">
                  <button class="h-8 w-8 text-ink-500 transition hover:text-brand-600 disabled:opacity-30" disabled={busy} onClick={() => update(it.id, it.variant, it.qty - 1)}>−</button>
                  <span class="w-9 text-center text-[13px] font-bold tabular-nums">{it.qty}</span>
                  <button class="h-8 w-8 text-ink-500 transition hover:text-brand-600 disabled:opacity-30" disabled={busy} onClick={() => update(it.id, it.variant, it.qty + 1)}>+</button>
                </div>
                <div class="w-20 text-right text-sm font-bold">{it.lineFmt}</div>
                <button class="text-ink-300 transition hover:text-brand-500" aria-label="删除" disabled={busy} onClick={() => update(it.id, it.variant, 0)}>🗑️</button>
              </div>
            ))}
          </div>
          <div class="h-fit rounded-xl2 border border-ink-100 bg-white p-5 lg:sticky lg:top-32">
            <h2 class="mb-4 text-base font-bold">订单摘要</h2>
            <div class="space-y-2.5 text-sm text-ink-600">
              <div class="flex justify-between"><span>商品小计({c.count} 件)</span><span class="font-medium text-ink-900">{c.subtotalFmt}</span></div>
              <div class="flex justify-between"><span>运费</span><span class={c.freeShip ? "font-semibold text-emerald-600" : "text-ink-900"}>{c.shippingFmt}</span></div>
              {!c.freeShip && (
                <div class="rounded-lg bg-brand-50 px-3 py-2 text-xs text-brand-700">
                  再买 <span class="font-bold">{c.freeGapFmt}</span> 即可免运费 🚚
                </div>
              )}
            </div>
            <div class="mt-4 flex items-baseline justify-between border-t border-ink-100 pt-4">
              <span class="text-sm text-ink-500">应付合计</span>
              <span class="text-2xl font-black text-brand-600">{c.totalFmt}</span>
            </div>
            <a href="/checkout" class="mt-5 block rounded-full bg-gradient-to-r from-brand-500 to-brand-600 py-3 text-center text-[15px] font-bold text-white shadow-pop transition hover:brightness-105">去结算 ({c.count})</a>
            <p class="mt-3 text-center text-xs text-ink-300">🔒 安全支付 · 支持 7 天无理由退换</p>
          </div>
        </div>
      )}
    </div>
  );
}
