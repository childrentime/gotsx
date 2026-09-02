import { useState } from "gotsx";

/** 首页 hero 里的活体示例: 这个按钮在服务端由 Go 渲染, 到浏览器后 hydrate 成 signal */
export default function Counter({ start }: { start: number }) {
  const [n, setN] = useState(start);
  const double = n * 2;
  return (
    <button
      class="inline-flex h-11 items-center gap-3 rounded-xl border border-brand-300 bg-white px-5 font-mono text-sm shadow-sm transition hover:border-brand-500 dark:border-brand-700 dark:bg-zinc-900"
      onClick={() => setN(n + 1)}
    >
      <span class="text-zinc-500">useState →</span>
      <span class="text-lg font-bold text-brand-600 dark:text-brand-300">{n}</span>
      <span class="text-zinc-500">×2 = {double}</span>
      {n > 4 && <span>🔥</span>}
    </button>
  );
}
