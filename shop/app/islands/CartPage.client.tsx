import { useState } from "gotsx";
import { setQty } from "host:cart";              // typed action: CartModule.SetQty
import { cart } from "../stores/cart";
import Icon from "../ui/Icon";

// The cart page reads the shared store (seeded by the layout, so the first paint is the server's cart) and replaces
// it with the CartView every quantity change returns; the header badge updates with it.
export default function CartPage({ locale = "" }: { locale?: string }) {
  const { items, count, empty, subtotalFmt, shippingFmt, totalFmt, freeShip, freeGapFmt } = cart;
  const [busy, setBusy] = useState(false);
  const update = async (id: string, variant: string, qty: number) => {
    setBusy(true);
    try {
      cart.set(await setQty(id, variant, qty));
    } finally {
      setBusy(false);
    }
  };
  return (
    <div class={busy ? "opacity-60 transition-opacity" : "transition-opacity"}>
      {empty ? (
        <div class="card flex flex-col items-center py-24 text-center">
          <span class="flex h-12 w-12 items-center justify-center rounded-full bg-muted text-muted-foreground"><Icon name="cart" className="h-5 w-5" /></span>
          <p class="mt-4 font-medium">{t(locale, "cart.empty")}</p>
          <a href="/" class="btn btn-primary mt-6">{t(locale, "cart.browse")}</a>
        </div>
      ) : (
        <div class="grid gap-5 lg:grid-cols-[1fr_320px]">
          <div class="card divide-y divide-border overflow-hidden">
            {items.map((it) => (
              <div key={it.id + "/" + it.variant} class="flex items-center gap-4 p-4">
                <a href={`/p/${it.id}`} class="shot flex h-20 w-20 shrink-0 items-center justify-center rounded-md text-4xl"><span class="emoji">{it.emoji}</span></a>
                <div class="min-w-0 flex-1">
                  <a href={`/p/${it.id}`} class="line-clamp-1 text-sm font-medium transition-colors hover:text-muted-foreground">{it.title}</a>
                  {it.variant !== "" && <div class="mt-1 inline-block rounded-sm bg-muted px-1.5 py-0.5 text-xs text-muted-foreground">{it.variant}</div>}
                  <div class="mt-1.5 text-sm tabular-nums">{it.priceFmt}</div>
                </div>
                <div class="inline-flex items-center rounded-md border border-input bg-background">
                  <button class="btn btn-ghost btn-icon-sm rounded-r-none" disabled={busy} onClick={() => update(it.id, it.variant, it.qty - 1)} aria-label={t(locale, "add.decrease")}><Icon name="minus" /></button>
                  <span class="w-9 text-center text-[13px] font-medium tabular-nums">{it.qty}</span>
                  <button class="btn btn-ghost btn-icon-sm rounded-l-none" disabled={busy} onClick={() => update(it.id, it.variant, it.qty + 1)} aria-label={t(locale, "add.increase")}><Icon name="plus" /></button>
                </div>
                <div class="w-20 text-right text-sm font-medium tabular-nums">{it.lineFmt}</div>
                <button class="btn btn-ghost btn-icon-sm text-muted-foreground hover:text-destructive" aria-label={t(locale, "cart.remove")} disabled={busy} onClick={() => update(it.id, it.variant, 0)}><Icon name="trash" /></button>
              </div>
            ))}
          </div>
          <div class="card h-fit p-5 lg:sticky lg:top-32">
            <h2 class="mb-4 text-base font-semibold">{t(locale, "cart.summary")}</h2>
            <div class="space-y-2.5 text-sm text-muted-foreground">
              <div class="flex justify-between"><span>{plural(locale, "cart.subtotal", count)}</span><span class="text-foreground tabular-nums">{subtotalFmt}</span></div>
              <div class="flex justify-between"><span>{t(locale, "cart.shipping")}</span><span class={freeShip ? "text-success" : "text-foreground tabular-nums"}>{shippingFmt}</span></div>
              {!freeShip && (
                <div class="rounded-md bg-muted px-3 py-2 text-xs">
                  {tv(locale, "cart.freeGap", { n: freeGapFmt })}
                </div>
              )}
            </div>
            <div class="mt-4 flex items-baseline justify-between border-t border-border pt-4">
              <span class="text-sm text-muted-foreground">{t(locale, "cart.total")}</span>
              <span class="text-2xl font-semibold tracking-tight tabular-nums">{totalFmt}</span>
            </div>
            <a href="/checkout" class="btn btn-primary btn-lg mt-5 w-full">{tv(locale, "cart.checkout", { n: String(count) })}</a>
            <p class="mt-3 flex items-center justify-center gap-1.5 text-xs text-muted-foreground"><Icon name="shield" className="h-3.5 w-3.5" />{t(locale, "cart.secure")}</p>
          </div>
        </div>
      )}
    </div>
  );
}
