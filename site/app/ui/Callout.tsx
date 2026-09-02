import type { Node } from "gotsx";

function kindCls(kind: string): string {
  if (kind === "warn") return "border-amber-300 bg-amber-50 text-amber-900 dark:border-amber-700 dark:bg-amber-950/40 dark:text-amber-100";
  if (kind === "tip") return "border-emerald-300 bg-emerald-50 text-emerald-900 dark:border-emerald-700 dark:bg-emerald-950/40 dark:text-emerald-100";
  return "border-brand-200 bg-brand-50 text-brand-900 dark:border-brand-800 dark:bg-brand-950/40 dark:text-brand-100";
}

export default function Callout({ kind = "info", title = "", children }: { kind?: string; title?: string; children?: Node }) {
  return (
    <div class={`my-4 rounded-lg border px-4 py-3 text-sm leading-6 ${kindCls(kind)}`}>
      {title !== "" && <div class="mb-1 font-semibold">{title}</div>}
      {children}
    </div>
  );
}
