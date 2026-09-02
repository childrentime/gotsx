import type { Node } from "gotsx";
/** 角色 / 状态徽标: 角色用中性层级(primary / secondary / outline), 状态用 success / secondary */
function cls(tone: string): string {
  if (tone === "admin") return "badge-primary";
  if (tone === "editor") return "badge-secondary";
  if (tone === "viewer") return "badge-outline";
  if (tone === "active") return "badge-success";
  if (tone === "disabled") return "badge-secondary text-muted-foreground";
  return "badge-outline";
}
export default function Badge({ tone = "", children }: { tone?: string; children?: Node }) {
  return <span class={`badge ${cls(tone)}`}>{children}</span>;
}
