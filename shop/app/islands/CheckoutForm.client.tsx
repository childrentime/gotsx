import { useState, emit } from "gotsx";
import Icon from "../ui/Icon";

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
  const field = (bad: boolean) => (bad ? "input border-destructive focus-visible:ring-destructive/30" : "input");
  return (
    <div class="space-y-4">
      {errs._ !== undefined && <div class="flex items-center gap-2 rounded-md border border-destructive/30 bg-destructive/5 px-3 py-2.5 text-sm text-destructive"><Icon name="alert" />{errs._}</div>}
      <div class="space-y-1.5">
        <label class="label">收货人</label>
        <input class={field(errs.name !== undefined)} value={name} placeholder="请输入姓名" onInput={(e) => setName(e.target.value)} />
        {errs.name !== undefined && <p class="text-xs text-destructive">{errs.name}</p>}
      </div>
      <div class="space-y-1.5">
        <label class="label">手机号</label>
        <input class={field(errs.phone !== undefined)} value={phone} placeholder="11 位手机号" onInput={(e) => setPhone(e.target.value)} />
        {errs.phone !== undefined && <p class="text-xs text-destructive">{errs.phone}</p>}
      </div>
      <div class="space-y-1.5">
        <label class="label">收货地址</label>
        <input class={field(errs.address !== undefined)} value={address} placeholder="省 / 市 / 区 + 详细地址" onInput={(e) => setAddress(e.target.value)} />
        {errs.address !== undefined && <p class="text-xs text-destructive">{errs.address}</p>}
      </div>
      <button class="btn btn-primary btn-lg mt-2 w-full" disabled={busy} onClick={submit}>
        {busy && <Icon name="loader" className="animate-spin" />}
        {busy ? "提交中…" : `提交订单 · ${totalFmt}`}
      </button>
      <p class="text-center text-xs text-muted-foreground">点击提交即代表同意《服务条款》· 演示环境, 不会真的扣款</p>
    </div>
  );
}
