export default function LocaleSwitch({ locale, path, label }: { locale: string; path: string; label: string }) {
  const other = locale === "zh" ? "en" : "zh";
  const go = () => {
    const base = (window.__GOTSX && window.__GOTSX.base) || "";
    window.location.href = base + lpath(other, path);
  };
  return (
    <button class="rounded-md px-2.5 py-1.5 text-xs font-medium text-zinc-500 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-50" onClick={go} aria-label="switch language">
      {label}
    </button>
  );
}
