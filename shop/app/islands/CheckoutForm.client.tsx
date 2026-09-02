import { useState, emit } from "gotsx";

export default function CheckoutForm({ totalFmt }: { totalFmt: string }) {
  const [name, setName] = useState("");
  const [phone, setPhone] = useState("");
  const [address, setAddress] = useState("");
  const [errs, setErrs] = useState<Record<string, string>>({});
  const [busy, setBusy] = useState(false);
  const submit = async () => {
    setBusy(true);
    const r = await fetch("/actions/checkout", { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ name, phone, address }) });
    const d = await r.json();
    if (d.ok) {
      emit("cart:changed", { count: 0 });
      window.__gotsxNavigate("/orders/" + d.orderId);
    } else {
      setErrs(d.errors);
      setBusy(false);
    }
  };
  const field = (bad: boolean) => (bad ? "h-11 w-full rounded-lg border-2 border-brand-400 bg-white px-3.5 text-sm outline-none" : "h-11 w-full rounded-lg border border-ink-200 bg-white px-3.5 text-sm outline-none transition focus:border-brand-500");
  return (
    <div class="space-y-4">
      {errs._ !== undefined && <div class="rounded-lg bg-brand-50 px-3.5 py-2.5 text-sm text-brand-700">⚠️ {errs._}</div>}
      <div>
        <label class="mb-1.5 block text-[13px] font-medium text-ink-600">收货人</label>
        <input class={field(errs.name !== undefined)} value={name} placeholder="请输入姓名" onInput={(e) => setName(e.target.value)} />
        {errs.name !== undefined && <p class="mt-1.5 text-xs text-brand-600">{errs.name}</p>}
      </div>
      <div>
        <label class="mb-1.5 block text-[13px] font-medium text-ink-600">手机号</label>
        <input class={field(errs.phone !== undefined)} value={phone} placeholder="11 位手机号" onInput={(e) => setPhone(e.target.value)} />
        {errs.phone !== undefined && <p class="mt-1.5 text-xs text-brand-600">{errs.phone}</p>}
      </div>
      <div>
        <label class="mb-1.5 block text-[13px] font-medium text-ink-600">收货地址</label>
        <input class={field(errs.address !== undefined)} value={address} placeholder="省 / 市 / 区 + 详细地址" onInput={(e) => setAddress(e.target.value)} />
        {errs.address !== undefined && <p class="mt-1.5 text-xs text-brand-600">{errs.address}</p>}
      </div>
      <button class="mt-1 inline-flex h-12 w-full items-center justify-center gap-2 rounded-full bg-gradient-to-r from-brand-500 to-brand-600 text-[15px] font-bold text-white shadow-pop transition hover:brightness-105 disabled:opacity-50" disabled={busy} onClick={submit}>
        {busy && <span class="inline-block h-4 w-4 animate-spin rounded-full border-2 border-white/40 border-t-white"></span>}
        {busy ? "提交中…" : `提交订单 · ${totalFmt}`}
      </button>
      <p class="text-center text-xs text-ink-300">点击提交即代表同意《服务条款》· 演示环境, 不会真的扣款</p>
    </div>
  );
}
