import type { PageProps } from "gotsx";
import { list } from "host:orders";
import Layout from "../../components/Layout.server";

export default function OrdersPage({ cookies, locale }: PageProps) {
  const sid = cookies.sid ?? "";
  const orders = list(sid);
  return (
    <Layout title="我的订单" sid={sid} locale={locale} active="orders" wide>
      <h1 class="mb-4 text-xl font-extrabold tracking-tight">我的订单</h1>
      {orders.length === 0 ? (
        <div class="rounded-xl2 border border-ink-100 bg-white py-24 text-center">
          <div class="text-6xl">📦</div>
          <p class="mt-4 text-ink-500">还没有订单</p>
          <a href="/" class="mt-5 inline-block rounded-full bg-brand-500 px-8 py-2.5 text-sm font-bold text-white transition hover:bg-brand-600">去下第一单</a>
        </div>
      ) : (
        <div class="space-y-4">
          {orders.map((o) => (
            <a href={`/orders/${o.id}`} class="block rounded-xl2 border border-ink-100 bg-white p-5 transition hover:border-brand-200 hover:shadow-card">
              <div class="flex items-center justify-between">
                <span class="font-mono text-sm font-bold text-ink-700">{o.id}</span>
                <span class="rounded-full bg-emerald-50 px-2.5 py-1 text-xs font-semibold text-emerald-600">{o.status}</span>
              </div>
              <div class="mt-3 flex items-center gap-2">
                {o.items.slice(0, 8).map((it) => (
                  <span class="flex h-11 w-11 items-center justify-center rounded-lg text-xl" style={`background:radial-gradient(120% 120% at 50% 20%, #fff, hsl(${it.hue} 46% 95%))`}>{it.emoji}</span>
                ))}
                <span class="ml-auto text-xs text-ink-400">{o.createdFmt}</span>
              </div>
              <div class="mt-3 flex items-center justify-between border-t border-ink-50 pt-3 text-sm">
                <span class="text-ink-400">共 {o.count} 件</span>
                <span>合计 <span class="text-lg font-black text-brand-600">{o.totalFmt}</span></span>
              </div>
            </a>
          ))}
        </div>
      )}
    </Layout>
  );
}
