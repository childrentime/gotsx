import { useState, useEffect } from "gotsx";
import Icon from "../ui/Icon";

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
    <button class="btn btn-ghost btn-icon-sm shrink-0 text-muted-foreground hover:text-foreground" onClick={toggle} aria-label="切换主题">
      {dark ? <Icon name="moon" /> : <Icon name="sun" />}
    </button>
  );
}
