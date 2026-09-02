import { useState, emit } from "gotsx";
import type { Variant } from "host:catalog";

export default function AddToCart({ id, variants, stock }: { id: string; variants: Variant[]; stock: number }) {
  const [sel, setSel] = useState<string[]>([]);
  const [qty, setQty] = useState(1);
  const [busy, setBusy] = useState(false);
  const [msg, setMsg] = useState("");
  const [good, setGood] = useState(false);
  const pick = (i: number, opt: string) => setSel(variants.map((v, j) => (j === i ? opt : sel[j] ?? "")));
  const chosen = sel.filter((x) => x !== "").length;
  const add = async () => {
    if (chosen !== variants.length) {
      setGood(false);
      setMsg("请先选择完整规格");
      return;
    }
    setBusy(true);
    setMsg("");
    const r = await fetch("/actions/cart/add", { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ id, variant: sel.join(" / "), qty }) });
    const d = await r.json();
    setBusy(false);
    if (d.ok) {
      setGood(true);
      setMsg("已加入购物车");
      emit("cart:changed", { count: d.cart.count });
    } else {
      setGood(false);
      setMsg(d.error);
    }
  };
  return (
    <div class="mt-5 space-y-5">
      {variants.map((v, i) => (
        <div>
          <div class="mb-2 text-[13px] text-ink-500">{v.name}<span class="ml-2 font-semibold text-ink-900">{sel[i] ?? ""}</span></div>
          <div class="flex flex-wrap gap-2">
            {v.options.map((opt) => (
              <button
                class={sel[i] === opt ? "rounded-lg border-2 border-brand-500 bg-brand-50 px-4 py-2 text-[13px] font-semibold text-brand-700" : "rounded-lg border border-ink-200 bg-white px-4 py-2 text-[13px] text-ink-700 transition hover:border-brand-300"}
                onClick={() => pick(i, opt)}
              >{opt}</button>
            ))}
          </div>
        </div>
      ))}
      <div class="flex items-center gap-4">
        <span class="text-[13px] text-ink-500">数量</span>
        <div class="flex items-center rounded-lg border border-ink-200 bg-white">
          <button class="h-9 w-9 text-lg text-ink-500 transition hover:text-brand-600 disabled:opacity-30" disabled={qty <= 1} onClick={() => setQty(qty - 1)}>−</button>
          <span class="w-12 text-center text-sm font-bold tabular-nums">{qty}</span>
          <button class="h-9 w-9 text-lg text-ink-500 transition hover:text-brand-600 disabled:opacity-30" disabled={qty >= stock} onClick={() => setQty(qty + 1)}>+</button>
        </div>
        <span class="text-xs text-ink-400">库存 {stock} 件</span>
      </div>
      <div class="flex flex-wrap items-center gap-3 pt-1">
        <button
          class="inline-flex h-12 items-center gap-2 rounded-full bg-gradient-to-r from-brand-500 to-brand-600 px-12 text-[15px] font-bold text-white shadow-pop transition hover:brightness-105 disabled:cursor-not-allowed disabled:from-ink-300 disabled:to-ink-300 disabled:shadow-none"
          disabled={busy || stock === 0}
          onClick={add}
        >
          {busy && <span class="inline-block h-4 w-4 animate-spin rounded-full border-2 border-white/40 border-t-white"></span>}
          {stock === 0 ? "已售罄" : busy ? "加入中…" : "🛒 加入购物车"}
        </button>
        {msg !== "" && <span class={good ? "inline-flex items-center gap-1 text-sm font-semibold text-emerald-600" : "text-sm font-semibold text-brand-600"}>{good && <span>✓</span>}{msg}</span>}
      </div>
    </div>
  );
}
