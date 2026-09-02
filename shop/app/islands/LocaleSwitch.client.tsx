import { useState } from "gotsx";
import Icon from "../ui/Icon";

/** 语言切换: 前缀模式下,导航到目标语言的同一路径 */
export default function LocaleSwitch({ locale, path, other, label }: { locale: string; path: string; other: string; label: string }) {
  const [busy, setBusy] = useState(false);
  const go = () => {
    setBusy(true);
    const target = lpath(other, path);   // 目标语言前缀 + 当前(去前缀)路径
    window.location.href = target;        // 整页切换,拿到目标语言的服务端渲染
  };
  return (
    <button class="btn btn-ghost btn-sm h-6 shrink-0 gap-1.5 px-2 text-xs text-muted-foreground hover:text-foreground" disabled={busy} onClick={go} aria-label="switch language">
      <Icon name="globe" className="h-3.5 w-3.5" /> {label}
    </button>
  );
}
