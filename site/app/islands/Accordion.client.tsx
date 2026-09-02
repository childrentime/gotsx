import { useState } from "gotsx";
import Icon from "../ui/Icon";

export interface QA {
  q: string;
  a: string;
}

export default function Accordion({ items }: { items: QA[] }) {
  const [open, setOpen] = useState(-1);
  return (
    <div class="card divide-y divide-border">
      {items.map((it, i) => (
        <div>
          <button
            class="flex w-full items-center justify-between gap-4 px-5 py-3.5 text-left text-sm font-medium transition-colors hover:bg-muted/50"
            aria-expanded={open === i}
            onClick={() => setOpen(open === i ? -1 : i)}
          >
            <span>{it.q}</span>
            <Icon name="chevron-down" className={open === i ? "icon rotate-180 text-muted-foreground transition-transform" : "icon text-muted-foreground transition-transform"} />
          </button>
          {open === i && <p class="px-5 pb-4 text-sm leading-6 text-muted-foreground">{it.a}</p>}
        </div>
      ))}
    </div>
  );
}
