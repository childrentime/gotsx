import { useState, useEffect } from "gotsx";

/** 主题切换: 只碰 document / localStorage, 这些在客户端代码里类型是 any */
export default function ThemeToggle() {
  const [dark, setDark] = useState(false);
  useEffect(() => {
    setDark(document.documentElement.classList.contains("dark"));
  });
  const toggle = () => {
    const next = !dark;
    setDark(next);
    document.documentElement.classList.toggle("dark", next);
    localStorage.setItem("theme", next ? "dark" : "light");
  };
  return (
    <button class="rounded-lg border border-zinc-200 px-2.5 py-1.5 text-sm hover:bg-zinc-100 dark:border-zinc-700 dark:hover:bg-zinc-800" onClick={toggle} aria-label="切换主题">
      {dark ? "🌙" : "☀️"}
    </button>
  );
}
