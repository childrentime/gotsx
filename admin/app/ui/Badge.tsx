import type { Node } from "gotsx";
function cls(tone: string): string {
  if (tone === "admin") return "bg-violet-100 text-violet-700";
  if (tone === "editor") return "bg-sky-100 text-sky-700";
  if (tone === "viewer") return "bg-slate-100 text-slate-600";
  if (tone === "active") return "bg-emerald-100 text-emerald-700";
  if (tone === "disabled") return "bg-rose-100 text-rose-600";
  return "bg-slate-100 text-slate-600";
}
export default function Badge({ tone = "", children }: { tone?: string; children?: Node }) {
  return <span class={`inline-flex items-center gap-1 rounded-full px-2.5 py-0.5 text-xs font-medium ${cls(tone)}`}>{children}</span>;
}
