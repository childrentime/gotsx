import { useState } from "gotsx";
import Icon from "../ui/Icon";

/** Locale switch: in URL-prefix mode, navigate to the same path under the other locale */
export default function LocaleSwitch({ locale, path, other, label }: { locale: string; path: string; other: string; label: string }) {
  const [busy, setBusy] = useState(false);
  const go = () => {
    setBusy(true);
    const target = lpath(other, path);   // other locale prefix + the current (unprefixed) path
    window.location.href = target;        // full navigation: get the server render in the other locale
  };
  return (
    <button class="btn btn-ghost btn-sm h-6 shrink-0 gap-1.5 px-2 text-xs text-muted-foreground hover:text-foreground" disabled={busy} onClick={go} aria-label="switch language">
      <Icon name="globe" className="h-3.5 w-3.5" /> {label}
    </button>
  );
}
