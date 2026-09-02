import type { Node } from "gotsx";

export default function Section({ id = "", title, lead = "", children }: { id?: string; title: string; lead?: string; children?: Node }) {
  return (
    <section id={id} class="mt-12">
      <h2 class="text-xl font-semibold tracking-tight">{title}</h2>
      {lead !== "" && <p class="mt-1 text-sm text-muted-foreground">{lead}</p>}
      <div class="prose-code mt-4 text-[15px] leading-7">{children}</div>
    </section>
  );
}
