import { useState } from "gotsx";
import { logout } from "host:auth";              // typed action: clears the session and queues a flash
import Icon from "../ui/Icon";
export default function Logout() {
  const [busy, setBusy] = useState(false);
  const out = async () => {
    setBusy(true);
    await logout();
    window.location.href = "/login";
  };
  return <button class="btn btn-ghost btn-icon-sm text-muted-foreground" title="Sign out" aria-label="Sign out" disabled={busy} onClick={out}><Icon name="logout" /></button>;
}
