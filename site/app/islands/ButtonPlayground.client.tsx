import { useState } from "gotsx";
import Button from "../ui/Button";

/** 岛里用共享组件: Button 的 children 是响应式文本, hydrate 时按 thunk 顺序走位。
 *  组件的 props 不是响应式的(和 Solid 一样), 所以第三个按钮用条件块整体重建 */
export default function ButtonPlayground() {
  const [n, setN] = useState(0);
  return (
    <div class="flex flex-wrap items-center gap-3">
      <Button onClick={() => setN(n + 1)}>点了 {n} 次</Button>
      <Button variant="secondary" onClick={() => setN(0)}>清零</Button>
      {n === 0 ? (
        <Button variant="outline" size="sm" disabled>还没点过</Button>
      ) : (
        <Button variant="outline" size="sm">已点 {n} 次</Button>
      )}
    </div>
  );
}
