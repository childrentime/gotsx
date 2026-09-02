import { useState, useEffect } from "gotsx";

/** 岛: 服务端用 Go 单趟求值出 HTML, 浏览器端编译成 signals + 精确 DOM 绑定, 走位 hydrate */
export default function Counter({ start }: { start: number }) {
  const [n, setN] = useState(start);
  const double = n * 2;                    // 依赖 n → 编译成 memo, 不需要 useMemo
  useEffect(() => {
    console.log("count is now", n);        // 只进 JS 后端
  });
  return (
    <button class="btn ghost" onClick={() => setN(n + 1)}>
      岛状态 {n} <small>×2 = {double}</small>
      {n > 4 && <b> 🔥</b>}
    </button>
  );
}
