import { useState } from "gotsx";
import { login } from "host:auth";               // typed action: AuthModule.Login(req, user, pass) → Promise<Profile>
import Icon from "../ui/Icon";

export default function LoginForm() {
  const [user, setUser] = useState("admin");
  const [pass, setPass] = useState("admin123");
  const [err, setErr] = useState("");
  const [busy, setBusy] = useState(false);
  const submit = async () => {
    setBusy(true);
    setErr("");
    try {
      await login(user, pass);                     // the session cookie is set by the response
      window.location.href = "/";
    } catch (e) {
      setErr(e.fields.password ?? e.message);      // 422: field message; anything else: the error text
      setBusy(false);
    }
  };
  return (
    <div class="space-y-4">
      {err !== "" && <div class="alert alert-error flex items-center gap-2"><Icon name="alert" />{err}</div>}
      <div class="space-y-2">
        <label class="label">Username</label>
        <input class="input" name="user" value={user} onInput={(e) => setUser(e.target.value)} />
      </div>
      <div class="space-y-2">
        <label class="label">Password</label>
        <input type="password" class="input" name="pass" value={pass} onInput={(e) => setPass(e.target.value)} onKeyDown={(e) => { if (e.key === "Enter") submit(); }} />
      </div>
      <button class="btn btn-primary w-full" disabled={busy} onClick={submit}>
        {busy && <Icon name="spinner" cls="icon animate-spin" />}
        {busy ? "Signing in…" : "Sign in"}
      </button>
    </div>
  );
}
