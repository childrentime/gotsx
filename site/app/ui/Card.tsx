import type { Node } from "gotsx";

export default function Card({ title = "", icon = "", className = "", children }: { title?: string; icon?: string; className?: string; children?: Node }) {
  return (
    <div class={`rounded-xl border border-zinc-200 bg-white p-5 shadow-sm dark:border-zinc-800 dark:bg-zinc-900 ${className}`}>
      {icon !== "" && <div class="mb-3 text-2xl">{icon}</div>}
      {title !== "" && <h3 class="mb-1.5 text-base font-semibold text-zinc-900 dark:text-zinc-50">{title}</h3>}
      <div class="text-sm leading-6 text-zinc-600 dark:text-zinc-300">{children}</div>
    </div>
  );
}
