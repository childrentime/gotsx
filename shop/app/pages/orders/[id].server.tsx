import type { PageProps } from "gotsx";
import { get } from "host:orders";
import Layout from "../../components/Layout.server";

export default function OrderDetail({ params, cookies, locale }: PageProps) {
  const sid = cookies.sid ?? "";
  const o = get(sid, params.id);
  return (
    <Layout title={`订单 ${o.id}`} sid={sid} locale={locale} active="orders" wide>
      <div class="mb-5 flex items-center gap-4 rounded-xl2 border border-emerald-200 bg-gradient-to-r from-emerald-50 to-white px-6 py-5">
        <span class="flex h-12 w-12 items-center justify-center rounded-full bg-emerald-500 text-2xl text-white">✓</span>
        <div>
          <div class="text-lg font-black text-emerald-700">下单成功!</div>
          <div class="mt-0.5 text-xs text-emerald-600">订单号 {o.id} · {o.createdFmt} · {o.status}</div>
        </div>
        <a href="/orders" class="ml-auto hidden rounded-full border border-emerald-300 px-4 py-2 text-sm font-semibold text-emerald-700 transition hover:bg-emerald-50 sm:block">查看全部订单</a>
      </div>
      <div class="grid gap-5 lg:grid-cols-[1fr_340px]">
        <div class="overflow-hidden rounded-xl2 border border-ink-100 bg-white">
          <div class="border-b border-ink-50 px-5 py-3 text-sm font-bold">商品清单</div>
          <div class="divide-y divide-ink-50">
            {o.items.map((it) => (
              <div class="flex items-center gap-3.5 p-4 text-sm">
                <span class="flex h-14 w-14 shrink-0 items-center justify-center rounded-xl text-2xl" style={`background:radial-gradient(120% 120% at 50% 20%, #fff, hsl(${it.hue} 46% 95%))`}>{it.emoji}</span>
                <div class="min-w-0 flex-1">
                  <div class="line-clamp-1 text-ink-800">{it.title}</div>
                  {it.variant !== "" && <div class="mt-0.5 text-xs text-ink-400">{it.variant}</div>}
                </div>
                <span class="text-xs text-ink-400">×{it.qty}</span>
                <span class="w-20 text-right font-bold">{it.lineFmt}</span>
              </div>
            ))}
          </div>
        </div>
        <div class="h-fit space-y-4">
          <div class="rounded-xl2 border border-ink-100 bg-white p-5 text-sm">
            <h2 class="mb-3 flex items-center gap-2 text-base font-bold"><span>📍</span>收货信息</h2>
            <div class="space-y-1.5 text-ink-600">
              <div><span class="font-semibold text-ink-800">{o.name}</span> · {o.phone}</div>
              <div class="leading-6">{o.address}</div>
            </div>
          </div>
          <div class="rounded-xl2 border border-ink-100 bg-white p-5">
            <div class="flex items-baseline justify-between">
              <span class="text-sm text-ink-500">实付金额</span>
              <span class="text-2xl font-black text-brand-600">{o.totalFmt}</span>
            </div>
            <a href="/" class="mt-4 block rounded-full bg-brand-500 py-2.5 text-center text-sm font-bold text-white transition hover:bg-brand-600">继续购物</a>
          </div>
        </div>
      </div>
    </Layout>
  );
}
