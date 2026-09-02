import type { PageProps } from "gotsx";
import { view } from "host:cart";
import Layout from "../components/Layout.server";
import CheckoutForm from "../islands/CheckoutForm.client";

export default function Checkout({ cookies, locale }: PageProps) {
  const sid = cookies.sid ?? "";
  const cv = view(sid);
  return (
    <Layout title="确认订单" sid={sid} locale={locale} wide>
      <div class="mb-4 flex items-center gap-2 text-sm text-ink-400">
        <a href="/cart" class="transition hover:text-brand-600">购物车</a><span>›</span><span class="font-semibold text-ink-800">确认订单</span>
      </div>
      {cv.empty ? (
        <div class="rounded-xl2 border border-ink-100 bg-white py-20 text-center">
          <div class="text-6xl">🛒</div>
          <p class="mt-4 text-ink-500">购物车是空的</p>
          <a href="/" class="mt-5 inline-block rounded-full bg-brand-500 px-8 py-2.5 text-sm font-bold text-white transition hover:bg-brand-600">去逛逛</a>
        </div>
      ) : (
        <div class="grid gap-5 lg:grid-cols-[1fr_380px]">
          <div class="space-y-4">
            <div class="rounded-xl2 border border-ink-100 bg-white p-6">
              <h2 class="mb-4 flex items-center gap-2 text-base font-bold"><span class="text-lg">📍</span>收货信息</h2>
              <CheckoutForm totalFmt={cv.totalFmt} />
            </div>
          </div>
          <div class="h-fit rounded-xl2 border border-ink-100 bg-white p-6 lg:sticky lg:top-32">
            <h2 class="mb-4 text-base font-bold">商品清单 <span class="text-xs font-normal text-ink-400">({cv.count} 件)</span></h2>
            <div class="space-y-3.5">
              {cv.items.map((it) => (
                <div class="flex items-center gap-3 text-sm">
                  <span class="flex h-12 w-12 shrink-0 items-center justify-center rounded-lg text-2xl" style={`background:radial-gradient(120% 120% at 50% 20%, #fff, hsl(${it.hue} 46% 95%))`}>{it.emoji}</span>
                  <div class="min-w-0 flex-1">
                    <div class="line-clamp-1 text-ink-800">{it.title}</div>
                    {it.variant !== "" && <div class="text-xs text-ink-400">{it.variant} · ×{it.qty}</div>}
                  </div>
                  <span class="font-bold">{it.lineFmt}</span>
                </div>
              ))}
            </div>
            <div class="mt-5 space-y-2 border-t border-ink-100 pt-4 text-sm text-ink-600">
              <div class="flex justify-between"><span>商品小计</span><span>{cv.subtotalFmt}</span></div>
              <div class="flex justify-between"><span>运费</span><span class={cv.freeShip ? "font-semibold text-emerald-600" : ""}>{cv.shippingFmt}</span></div>
              <div class="flex items-baseline justify-between pt-2"><span class="text-sm text-ink-500">应付</span><span class="text-2xl font-black text-brand-600">{cv.totalFmt}</span></div>
            </div>
          </div>
        </div>
      )}
    </Layout>
  );
}
