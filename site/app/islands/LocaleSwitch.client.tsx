export default function LocaleSwitch({ locale, path, label }: { locale: string; path: string; label: string }) {
  const other = locale === "zh" ? "en" : "zh";
  const go = () => {
    const base = (window.__GOTSX && window.__GOTSX.base) || "";
    window.location.href = base + lpath(other, path);
  };
  return (
    <button class="nav-link shrink-0 text-xs font-medium" onClick={go} aria-label="switch language">
      {label}
    </button>
  );
}
