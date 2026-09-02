import { useState, emit } from "gotsx";
import Icon from "../ui/Icon";

export default function CheckoutForm({ totalFmt, locale = "" }: { totalFmt: string; locale?: string }) {
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
        <label class="label">{t(locale, "form.name")}</label>
        <input class={field(errs.name !== undefined)} value={name} placeholder={t(locale, "form.namePh")} onInput={(e) => setName(e.target.value)} />
        {errs.name !== undefined && <p class="text-xs text-destructive">{errs.name}</p>}
      </div>
      <div class="space-y-1.5">
        <label class="label">{t(locale, "form.phone")}</label>
        <input class={field(errs.phone !== undefined)} value={phone} placeholder={t(locale, "form.phonePh")} onInput={(e) => setPhone(e.target.value)} />
        {errs.phone !== undefined && <p class="text-xs text-destructive">{errs.phone}</p>}
      </div>
      <div class="space-y-1.5">
        <label class="label">{t(locale, "form.address")}</label>
        <input class={field(errs.address !== undefined)} value={address} placeholder={t(locale, "form.addressPh")} onInput={(e) => setAddress(e.target.value)} />
        {errs.address !== undefined && <p class="text-xs text-destructive">{errs.address}</p>}
      </div>
      <button class="btn btn-primary btn-lg mt-2 w-full" disabled={busy} onClick={submit}>
        {busy && <Icon name="loader" className="animate-spin" />}
        {busy ? t(locale, "form.submitting") : tv(locale, "form.submit", { total: totalFmt })}
      </button>
      <p class="text-center text-xs text-muted-foreground">{t(locale, "form.note")}</p>
    </div>
  );
}
