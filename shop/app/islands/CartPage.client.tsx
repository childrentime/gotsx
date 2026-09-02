import { useState, emit } from "gotsx";
import type { CartView } from "host:cart";
import Icon from "../ui/Icon";

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
    <div class={busy ? "opacity-60 transition-opacity" : "transition-opacity"}>
      {c.empty ? (
        <div class="card flex flex-col items-center py-24 text-center">
          <span class="flex h-12 w-12 items-center justify-center rounded-full bg-muted text-muted-foreground"><Icon name="cart" className="h-5 w-5" /></span>
          <p class="mt-4 font-medium">购物车还是空的</p>
          <a href="/" class="btn btn-primary mt-6">去挑点好物</a>
        </div>
      ) : (
        <div class="grid gap-5 lg:grid-cols-[1fr_320px]">
          <div class="card divide-y divide-border overflow-hidden">
            {c.items.map((it) => (
              <div class="flex items-center gap-4 p-4">
                <a href={`/p/${it.id}`} class="shot flex h-20 w-20 shrink-0 items-center justify-center rounded-md text-4xl"><span class="emoji">{it.emoji}</span></a>
                <div class="min-w-0 flex-1">
                  <a href={`/p/${it.id}`} class="line-clamp-1 text-sm font-medium transition-colors hover:text-muted-foreground">{it.title}</a>
                  {it.variant !== "" && <div class="mt-1 inline-block rounded-sm bg-muted px-1.5 py-0.5 text-xs text-muted-foreground">{it.variant}</div>}
                  <div class="mt-1.5 text-sm tabular-nums">{it.priceFmt}</div>
                </div>
                <div class="inline-flex items-center rounded-md border border-input bg-background">
                  <button class="btn btn-ghost btn-icon-sm rounded-r-none" disabled={busy} onClick={() => update(it.id, it.variant, it.qty - 1)} aria-label="减少"><Icon name="minus" /></button>
                  <span class="w-9 text-center text-[13px] font-medium tabular-nums">{it.qty}</span>
                  <button class="btn btn-ghost btn-icon-sm rounded-l-none" disabled={busy} onClick={() => update(it.id, it.variant, it.qty + 1)} aria-label="增加"><Icon name="plus" /></button>
                </div>
                <div class="w-20 text-right text-sm font-medium tabular-nums">{it.lineFmt}</div>
                <button class="btn btn-ghost btn-icon-sm text-muted-foreground hover:text-destructive" aria-label="删除" disabled={busy} onClick={() => update(it.id, it.variant, 0)}><Icon name="trash" /></button>
              </div>
            ))}
          </div>
          <div class="card h-fit p-5 lg:sticky lg:top-32">
            <h2 class="mb-4 text-base font-semibold">订单摘要</h2>
            <div class="space-y-2.5 text-sm text-muted-foreground">
              <div class="flex justify-between"><span>商品小计({c.count} 件)</span><span class="text-foreground tabular-nums">{c.subtotalFmt}</span></div>
              <div class="flex justify-between"><span>运费</span><span class={c.freeShip ? "text-success" : "text-foreground tabular-nums"}>{c.shippingFmt}</span></div>
              {!c.freeShip && (
                <div class="rounded-md bg-muted px-3 py-2 text-xs">
                  再买 <span class="font-medium text-foreground">{c.freeGapFmt}</span> 即可免运费
                </div>
              )}
            </div>
            <div class="mt-4 flex items-baseline justify-between border-t border-border pt-4">
              <span class="text-sm text-muted-foreground">应付合计</span>
              <span class="text-2xl font-semibold tracking-tight tabular-nums">{c.totalFmt}</span>
            </div>
            <a href="/checkout" class="btn btn-primary btn-lg mt-5 w-full">去结算 ({c.count})</a>
            <p class="mt-3 flex items-center justify-center gap-1.5 text-xs text-muted-foreground"><Icon name="shield" className="h-3.5 w-3.5" />安全支付 · 7 天无理由退换</p>
          </div>
        </div>
      )}
    </div>
  );
}
