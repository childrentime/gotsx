import { useState } from "gotsx";

/** 数组状态: includes / filter / spread / join, 响应式 class 和条件块 */
export default function TagPicker({ tags }: { tags: string[] }) {
  const [picked, setPicked] = useState<string[]>([]);
  const toggle = (t: string) =>
    setPicked(picked.includes(t) ? picked.filter((x) => x !== t) : [...picked, t]);
  return (
    <div class="row">
      {tags.map((t) => (
        <button class={picked.includes(t) ? "chip on" : "chip"} onClick={() => toggle(t)}>{t}</button>
      ))}
      {picked.length > 0 && <span class="muted">已选 {picked.length} 个: {picked.join(", ")}</span>}
    </div>
  );
}
