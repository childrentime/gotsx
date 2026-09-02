import type { Node } from "gotsx";
import Icon from "./Icon";

export default function Card({ title = "", icon = "", className = "", children }: { title?: string; icon?: string; className?: string; children?: Node }) {
  return (
    <div class={`card p-5 ${className}`}>
      {icon !== "" && (
        <div class="mb-3 inline-flex h-8 w-8 items-center justify-center rounded-md border border-border bg-muted/50 text-foreground">
          <Icon name={icon} />
        </div>
      )}
      {title !== "" && <h3 class="mb-1.5 text-[15px] font-semibold tracking-tight">{title}</h3>}
      <div class="text-sm leading-6 text-muted-foreground">{children}</div>
    </div>
  );
}
