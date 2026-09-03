import { useState } from "gotsx";
import { place } from "host:orders";            // typed action: OrdersModule.Place (validation → 422 with fields)
import { cart } from "../stores/cart";
import Icon from "../ui/Icon";

export default function CheckoutForm({ locale = "" }: { locale?: string }) {
  const { totalFmt } = cart;                     // the shared cart store (seeded by the layout): the button shows the live total
  const [name, setName] = useState("");
  const [phone, setPhone] = useState("");
  const [address, setAddress] = useState("");
  const [errs, setErrs] = useState<Record<string, string>>({});
  const [busy, setBusy] = useState(false);
  const submit = async () => {
    setBusy(true);
    try {
      const o = await place(name, phone, address);
      window.__gotsxNavigate("/orders/" + o.id);   // the order page seeds the (now empty) cart: the badge clears with the navigation
    } catch (e) {
      setErrs(e.status === 422 ? e.fields : { _: e.message });   // field errors, or the empty-cart / server message
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
