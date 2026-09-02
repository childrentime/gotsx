import type { Node } from "gotsx";

export default function Section({ id = "", title, lead = "", children }: { id?: string; title: string; lead?: string; children?: Node }) {
  return (
    <section id={id} class="mt-12">
      <h2 class="text-xl font-semibold tracking-tight">{title}</h2>
      {lead !== "" && <p class="mt-1 text-sm text-zinc-500 dark:text-zinc-400">{lead}</p>}
      <div class="mt-4 text-[15px] leading-7 text-zinc-700 dark:text-zinc-300">{children}</div>
    </section>
  );
}
