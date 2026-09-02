import type { PageProps } from "gotsx";
import { cards, recent } from "host:stats";
import Shell from "../components/Shell.server";
import Icon from "../ui/Icon";

export default function Dashboard({ cookies }: PageProps) {
  const stat = cards();
  const acts = recent();
  const bars = [30, 45, 38, 52, 48, 61, 55, 70, 64, 78, 72, 88];
  return (
    <Shell title="仪表盘" active="home" name={cookies._name ?? ""} role={cookies._role ?? ""}>
      <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {stat.map((s) => (
          <div class="card p-5">
            <div class="flex items-center justify-between">
              <span class="text-sm text-muted-foreground">{s.label}</span>
              <span class={s.up ? "badge badge-success gap-1" : "badge badge-destructive gap-1"}><Icon name={s.up ? "up" : "down"} cls="h-3 w-3" />{s.delta}</span>
            </div>
            <div class="mt-2 text-2xl font-semibold tracking-tight">{s.value}</div>
          </div>
        ))}
      </div>
      <div class="mt-4 grid gap-4 lg:grid-cols-[1fr_320px]">
        <section class="card">
          <div class="card-header flex-row items-baseline justify-between">
            <h2 class="card-title">用户增长</h2>
            <span class="text-xs text-muted-foreground">近 12 周</span>
          </div>
          <div class="card-body">
            <div class="flex h-40 items-end gap-1.5">
              {bars.map((h, i) => (
                <div class={i === bars.length - 1 ? "flex-1 rounded-sm bg-primary" : "flex-1 rounded-sm bg-primary/20"} style={`height:${h}%`}></div>
              ))}
            </div>
            <div class="mt-2 flex justify-between text-[11px] text-muted-foreground"><span>12 周前</span><span>本周</span></div>
          </div>
        </section>
        <section class="card">
          <div class="card-header"><h2 class="card-title">最近动态</h2></div>
          <div class="card-body">
            <div class="divide-y divide-border">
              {acts.map((a) => (
                <div class="flex gap-3 py-3 text-sm first:pt-0 last:pb-0">
                  <span class="mt-1.5 h-1.5 w-1.5 shrink-0 rounded-full bg-primary/60"></span>
                  <div class="min-w-0 flex-1">
                    <div class="leading-5"><span class="font-medium">{a.who}</span> <span class="text-muted-foreground">{a.what}</span></div>
                    <div class="text-[11px] text-muted-foreground">{a.when}</div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </section>
      </div>
    </Shell>
  );
}
