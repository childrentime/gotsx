import { useState } from "gotsx";
import type { Variant } from "host:catalog";
import { add } from "host:cart";                 // typed action: CartModule.Add over a same-origin POST
import { cart } from "../stores/cart";
import Icon from "../ui/Icon";

export default function AddToCart({ id, variants, stock, locale = "" }: { id: string; variants: Variant[]; stock: number; locale?: string }) {
  const [sel, setSel] = useState<string[]>([]);
  const [qty, setQty] = useState(1);
  const [busy, setBusy] = useState(false);
  const [msg, setMsg] = useState("");
  const [good, setGood] = useState(false);
  const pick = (i: number, opt: string) => setSel(variants.map((v, j) => (j === i ? opt : sel[j] ?? "")));
  const chosen = sel.filter((x) => x !== "").length;
  const onAdd = async () => {
    if (chosen !== variants.length) {
      setGood(false);
      setMsg(t(locale, "add.pick"));
      return;
    }
    setBusy(true);
    setMsg("");
    try {
      cart.set(await add(id, sel.join(" / "), qty));   // Promise<CartView>, typed from Go: the badge (and any other island) follows the store
      setGood(true);
      setMsg(t(locale, "add.added"));
    } catch (e) {
      setGood(false);
      setMsg(e.message);                                // gotsx.Fail("Only N left in stock") → 400 with the message
    } finally {
      setBusy(false);
    }
  };
  return (
    <div class="mt-5 space-y-5">
      {variants.map((v, i) => (
        <div>
          <div class="mb-2 text-[13px] text-muted-foreground">{v.name}<span class="ml-2 font-medium text-foreground">{sel[i] ?? ""}</span></div>
          <div class="flex flex-wrap gap-2">
            {v.options.map((opt) => (
              <button class={sel[i] === opt ? "btn btn-primary btn-sm" : "btn btn-outline btn-sm"} onClick={() => pick(i, opt)}>{opt}</button>
            ))}
          </div>
        </div>
      ))}
      <div class="flex items-center gap-4">
        <span class="text-[13px] text-muted-foreground">{t(locale, "add.qty")}</span>
        <div class="inline-flex items-center rounded-md border border-input bg-background">
          <button class="btn btn-ghost btn-icon-sm rounded-r-none" disabled={qty <= 1} onClick={() => setQty(qty - 1)} aria-label={t(locale, "add.decrease")}><Icon name="minus" /></button>
          <span class="w-10 text-center text-sm font-medium tabular-nums">{qty}</span>
          <button class="btn btn-ghost btn-icon-sm rounded-l-none" disabled={qty >= stock} onClick={() => setQty(qty + 1)} aria-label={t(locale, "add.increase")}><Icon name="plus" /></button>
        </div>
        <span class="text-xs text-muted-foreground tabular-nums">{tv(locale, "add.stock", { n: String(stock) })}</span>
      </div>
      <div class="flex flex-wrap items-center gap-3 pt-1">
        <button class="btn btn-primary btn-lg min-w-44" disabled={busy || stock === 0} onClick={onAdd}>
          {busy ? <Icon name="loader" className="animate-spin" /> : <Icon name="cart" />}
          {stock === 0 ? t(locale, "add.soldout") : busy ? t(locale, "add.adding") : t(locale, "add.button")}
        </button>
        {msg !== "" && <span class={good ? "inline-flex items-center gap-1.5 text-sm font-medium text-success" : "inline-flex items-center gap-1.5 text-sm font-medium text-destructive"}>{good ? <Icon name="check" /> : <Icon name="alert" />}{msg}</span>}
      </div>
    </div>
  );
}
