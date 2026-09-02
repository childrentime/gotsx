import { useState } from "gotsx";
import Icon from "../ui/Icon";

/** 首页 hero 里的活体示例: 这个按钮在服务端由 Go 渲染, 到浏览器后 hydrate 成 signal */
export default function Counter({ start }: { start: number }) {
  const [n, setN] = useState(start);
  const double = n * 2;
  return (
    <button class="btn btn-outline h-11 gap-3 px-5 font-mono text-sm shadow-xs" onClick={() => setN(n + 1)}>
      <span class="text-muted-foreground">useState →</span>
      <span class="text-lg font-semibold tabular-nums">{n}</span>
      <span class="text-muted-foreground">×2 = {double}</span>
      {n > 4 && <Icon name="flame" className="icon text-warning" />}
    </button>
  );
}
