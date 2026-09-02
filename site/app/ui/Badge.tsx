import type { Node } from "gotsx";

function colorCls(color: string): string {
  if (color === "green") return "bg-emerald-100 text-emerald-800 dark:bg-emerald-900/40 dark:text-emerald-200";
  if (color === "amber") return "bg-amber-100 text-amber-800 dark:bg-amber-900/40 dark:text-amber-200";
  if (color === "red") return "bg-red-100 text-red-800 dark:bg-red-900/40 dark:text-red-200";
  if (color === "brand") return "bg-brand-100 text-brand-800 dark:bg-brand-900/40 dark:text-brand-200";
  return "bg-zinc-100 text-zinc-700 dark:bg-zinc-800 dark:text-zinc-200";
}

export default function Badge({ color = "zinc", children }: { color?: string; children?: Node }) {
  return <span class={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ${colorCls(color)}`}>{children}</span>;
}
