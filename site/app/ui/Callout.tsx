import type { Node } from "gotsx";
import Icon from "./Icon";

function kindCls(kind: string): string {
  if (kind === "warn") return "border-warning/30 bg-warning/5";
  if (kind === "tip") return "border-success/30 bg-success/5";
  return "border-border bg-muted/40";
}

function iconCls(kind: string): string {
  if (kind === "warn") return "icon mt-0.5 text-warning";
  if (kind === "tip") return "icon mt-0.5 text-success";
  return "icon mt-0.5 text-muted-foreground";
}

function iconName(kind: string): string {
  if (kind === "warn") return "alert";
  if (kind === "tip") return "bulb";
  return "info";
}

export default function Callout({ kind = "info", title = "", children }: { kind?: string; title?: string; children?: Node }) {
  return (
    <div class={`my-4 flex gap-3 rounded-lg border px-4 py-3 text-sm leading-6 text-foreground ${kindCls(kind)}`}>
      <Icon name={iconName(kind)} className={iconCls(kind)} />
      <div class="min-w-0">
        {title !== "" && <div class="mb-0.5 font-medium">{title}</div>}
        {children}
      </div>
    </div>
  );
}
