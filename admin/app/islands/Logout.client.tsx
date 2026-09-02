import { useState } from "gotsx";
import Icon from "../ui/Icon";
export default function Logout() {
  const [busy, setBusy] = useState(false);
  const out = async () => {
    setBusy(true);
    await fetch("/auth/logout", { method: "POST", headers: { "Content-Type": "application/json" }, body: "{}" });
    window.location.href = "/login";
  };
  return <button class="btn btn-ghost btn-icon-sm text-muted-foreground" title="退出" aria-label="退出" disabled={busy} onClick={out}><Icon name="logout" /></button>;
}
