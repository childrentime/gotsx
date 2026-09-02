import type { PageProps } from "gotsx";
import { cards, recent } from "host:stats";
import Shell from "../components/Shell.server";

export default function Dashboard({ cookies }: PageProps) {
  const stat = cards();
  const acts = recent();
  return (
    <Shell title="仪表盘" active="home" name={cookies._name ?? ""} role={cookies._role ?? ""}>
      <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {stat.map((s) => (
          <div class="rounded-xl border border-slate-200 bg-white p-5">
            <div class="flex items-center justify-between">
              <span class="text-2xl">{s.icon}</span>
              <span class={s.up ? "rounded-full bg-emerald-50 px-2 py-0.5 text-xs font-semibold text-emerald-600" : "rounded-full bg-rose-50 px-2 py-0.5 text-xs font-semibold text-rose-600"}>{s.delta}</span>
            </div>
            <div class="mt-3 text-2xl font-black">{s.value}</div>
            <div class="text-xs text-slate-400">{s.label}</div>
          </div>
        ))}
      </div>
      <div class="mt-5 grid gap-5 lg:grid-cols-[1fr_320px]">
        <section class="rounded-xl border border-slate-200 bg-white p-5">
          <div class="mb-4 flex items-center justify-between">
            <h2 class="font-bold">用户增长</h2>
            <span class="text-xs text-slate-400">近 12 周</span>
          </div>
          <div class="flex h-40 items-end gap-1.5">
            {[30, 45, 38, 52, 48, 61, 55, 70, 64, 78, 72, 88].map((h) => (
              <div class="flex-1 rounded-t bg-gradient-to-t from-brand-500 to-brand-300" style={`height:${h}%`}></div>
            ))}
          </div>
        </section>
        <section class="rounded-xl border border-slate-200 bg-white p-5">
          <h2 class="mb-4 font-bold">最近动态</h2>
          <div class="space-y-4">
            {acts.map((a) => (
              <div class="flex gap-3 text-sm">
                <span class="text-lg">{a.icon}</span>
                <div class="min-w-0 flex-1">
                  <div class="text-slate-700"><b>{a.who}</b> {a.what}</div>
                  <div class="text-[11px] text-slate-400">{a.when}</div>
                </div>
              </div>
            ))}
          </div>
        </section>
      </div>
    </Shell>
  );
}
