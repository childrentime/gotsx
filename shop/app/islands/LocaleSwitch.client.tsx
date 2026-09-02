import { useState } from "gotsx";

/** 语言切换: 前缀模式下,导航到目标语言的同一路径 */
export default function LocaleSwitch({ locale, path, other, label }: { locale: string; path: string; other: string; label: string }) {
  const [busy, setBusy] = useState(false);
  const go = () => {
    setBusy(true);
    const target = lpath(other, path);   // 目标语言前缀 + 当前(去前缀)路径
    window.location.href = target;        // 整页切换,拿到目标语言的服务端渲染
  };
  return (
    <button class="shrink-0 rounded-full border border-white/60 px-2.5 py-1 text-xs font-medium text-white/90 transition hover:bg-white/10 disabled:opacity-50" disabled={busy} onClick={go} aria-label="switch language">
      🌐 {label}
    </button>
  );
}
