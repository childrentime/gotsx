import type { Node } from "gotsx";

/** color: zinc(默认) | brand | green | amber | red → 设计系统的 badge-secondary / primary / success / warning / destructive */
function colorCls(color: string): string {
  if (color === "green") return "badge-success";
  if (color === "amber") return "badge-warning";
  if (color === "red") return "badge-destructive";
  if (color === "brand") return "badge-primary";
  return "badge-secondary";
}

export default function Badge({ color = "zinc", children }: { color?: string; children?: Node }) {
  return <span class={`badge ${colorCls(color)}`}>{children}</span>;
}
