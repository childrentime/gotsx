import { useState } from "gotsx";
import Icon from "../ui/Icon";

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
  return (
    <div class="space-y-4">
      {err !== "" && <div class="flex items-center gap-2 rounded-md border border-destructive/30 px-3 py-2 text-sm text-destructive"><Icon name="alert" />{err}</div>}
      <div class="space-y-2">
        <label class="label">Username</label>
        <input class="input" value={user} onInput={(e) => setUser(e.target.value)} />
      </div>
      <div class="space-y-2">
        <label class="label">Password</label>
        <input type="password" class="input" value={pass} onInput={(e) => setPass(e.target.value)} onKeyDown={(e) => { if (e.key === "Enter") submit(); }} />
      </div>
      <button class="btn btn-primary w-full" disabled={busy} onClick={submit}>
        {busy && <Icon name="spinner" cls="icon animate-spin" />}
        {busy ? "Signing in…" : "Sign in"}
      </button>
    </div>
  );
}
