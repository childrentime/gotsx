import type { Node } from "gotsx";
/** role / status badge: roles use neutral tiers (primary / secondary / outline), statuses use success / secondary */
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
