import { useState } from "gotsx";

export default function LoginForm() {
  const [user, setUser] = useState("admin");
  const [pass, setPass] = useState("admin123");
  const [err, setErr] = useState("");
  const [busy, setBusy] = useState(false);
  const submit = async () => {
    setBusy(true);
    setErr("");
    const r = await fetch("/auth/login", { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ user, pass }) });
    const d = await r.json();
    if (d.ok) {
      window.location.href = "/";
    } else {
      setErr(d.error);
      setBusy(false);
    }
  };
  const field = "h-11 w-full rounded-lg border border-slate-300 bg-white px-3.5 text-sm outline-none transition focus:border-brand-500 focus:ring-2 focus:ring-brand-100";
  return (
    <div class="space-y-3">
      {err !== "" && <div class="rounded-lg bg-rose-50 px-3.5 py-2.5 text-sm text-rose-600">⚠️ {err}</div>}
      <div>
        <label class="mb-1.5 block text-xs font-medium text-slate-500">用户名</label>
        <input class={field} value={user} onInput={(e) => setUser(e.target.value)} />
      </div>
      <div>
        <label class="mb-1.5 block text-xs font-medium text-slate-500">密码</label>
        <input type="password" class={field} value={pass} onInput={(e) => setPass(e.target.value)} />
      </div>
      <button class="mt-2 inline-flex h-11 w-full items-center justify-center gap-2 rounded-lg bg-brand-500 text-sm font-bold text-white transition hover:bg-brand-600 disabled:opacity-50" disabled={busy} onClick={submit}>
        {busy && <span class="inline-block h-4 w-4 animate-spin rounded-full border-2 border-white/40 border-t-white"></span>}
        {busy ? "登录中…" : "登录"}
      </button>
    </div>
  );
}
