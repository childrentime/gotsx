import { useState } from "gotsx";
export default function Logout() {
  const [busy, setBusy] = useState(false);
  const out = async () => {
    setBusy(true);
    await fetch("/auth/logout", { method: "POST", headers: { "Content-Type": "application/json" }, body: "{}" });
    window.location.href = "/login";
  };
  return <button class="rounded-lg border border-slate-200 px-3 py-1.5 text-xs font-medium text-slate-600 hover:bg-slate-100 disabled:opacity-50" disabled={busy} onClick={out}>退出</button>;
}
