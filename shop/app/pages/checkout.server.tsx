import type { PageProps } from "gotsx";
import { view } from "host:cart";
import Layout from "../components/Layout.server";
import Icon from "../ui/Icon";
import CheckoutForm from "../islands/CheckoutForm.client";

export default function Checkout({ cookies, locale }: PageProps) {
  const sid = cookies.sid ?? "";
  const cv = view(sid);
  return (
    <Layout title="确认订单" sid={sid} locale={locale} wide>
      <nav class="mb-5 flex items-center gap-1.5 text-xs text-muted-foreground" aria-label="breadcrumb">
        <a href="/cart" class="transition-colors hover:text-foreground">购物车</a><Icon name="chevron-right" className="h-3 w-3" /><span class="text-foreground">确认订单</span>
      </nav>
      {cv.empty ? (
        <div class="card flex flex-col items-center py-24 text-center">
          <span class="flex h-12 w-12 items-center justify-center rounded-full bg-muted text-muted-foreground"><Icon name="cart" className="h-5 w-5" /></span>
          <p class="mt-4 font-medium">购物车是空的</p>
          <a href="/" class="btn btn-primary mt-6">去逛逛</a>
        </div>
      ) : (
        <div class="grid gap-5 lg:grid-cols-[1fr_380px]">
          <div class="space-y-4">
            <div class="card p-6">
              <h2 class="mb-5 flex items-center gap-2 text-base font-semibold"><Icon name="map-pin" />收货信息</h2>
              <CheckoutForm totalFmt={cv.totalFmt} />
            </div>
          </div>
          <div class="card h-fit p-6 lg:sticky lg:top-32">
            <h2 class="mb-4 text-base font-semibold">商品清单 <span class="text-xs font-normal text-muted-foreground">({cv.count} 件)</span></h2>
            <div class="space-y-3.5">
              {cv.items.map((it) => (
                <div class="flex items-center gap-3 text-sm">
                  <span class="shot flex h-12 w-12 shrink-0 items-center justify-center rounded-md text-2xl"><span class="emoji">{it.emoji}</span></span>
                  <div class="min-w-0 flex-1">
                    <div class="line-clamp-1">{it.title}</div>
                    {it.variant !== "" && <div class="text-xs text-muted-foreground">{it.variant} · ×{it.qty}</div>}
                  </div>
                  <span class="font-medium tabular-nums">{it.lineFmt}</span>
                </div>
              ))}
            </div>
            <div class="mt-5 space-y-2 border-t border-border pt-4 text-sm text-muted-foreground">
              <div class="flex justify-between"><span>商品小计</span><span class="tabular-nums">{cv.subtotalFmt}</span></div>
              <div class="flex justify-between"><span>运费</span><span class={cv.freeShip ? "text-success" : "tabular-nums"}>{cv.shippingFmt}</span></div>
              <div class="flex items-baseline justify-between pt-2"><span>应付</span><span class="text-2xl font-semibold tracking-tight text-foreground tabular-nums">{cv.totalFmt}</span></div>
            </div>
          </div>
        </div>
      )}
    </Layout>
  );
}
