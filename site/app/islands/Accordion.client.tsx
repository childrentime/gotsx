import { useState } from "gotsx";

export interface QA {
  q: string;
  a: string;
}

export default function Accordion({ items }: { items: QA[] }) {
  const [open, setOpen] = useState(-1);
  return (
    <div class="divide-y divide-zinc-200 rounded-xl border border-zinc-200 bg-white dark:divide-zinc-800 dark:border-zinc-800 dark:bg-zinc-900">
      {items.map((it, i) => (
        <div>
          <button
            class="flex w-full items-center justify-between px-5 py-3.5 text-left text-sm font-medium hover:bg-zinc-50 dark:hover:bg-zinc-800/60"
            aria-expanded={open === i}
            onClick={() => setOpen(open === i ? -1 : i)}
          >
            <span>{it.q}</span>
            <span class="text-zinc-400">{open === i ? "−" : "+"}</span>
          </button>
          {open === i && <p class="px-5 pb-4 text-sm leading-6 text-zinc-600 dark:text-zinc-300">{it.a}</p>}
        </div>
      ))}
    </div>
  );
}
