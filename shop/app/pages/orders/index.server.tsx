import type { PageProps } from "gotsx";
import { list } from "host:orders";
import Layout from "../../components/Layout.server";
import Icon from "../../ui/Icon";

export default function OrdersPage({ cookies, locale }: PageProps) {
  const sid = cookies.sid ?? "";
  const orders = list(sid);
  return (
    <Layout title="我的订单" sid={sid} locale={locale} active="orders" wide>
      <h1 class="mb-5 text-xl font-semibold tracking-tight">我的订单</h1>
      {orders.length === 0 ? (
        <div class="card flex flex-col items-center py-24 text-center">
          <span class="flex h-12 w-12 items-center justify-center rounded-full bg-muted text-muted-foreground"><Icon name="package" className="h-5 w-5" /></span>
          <p class="mt-4 font-medium">还没有订单</p>
          <a href="/" class="btn btn-primary mt-6">去下第一单</a>
        </div>
      ) : (
        <div class="space-y-3">
          {orders.map((o) => (
            <a href={`/orders/${o.id}`} class="card block p-5 transition-colors hover:border-foreground/25">
              <div class="flex items-center justify-between">
                <span class="font-mono text-sm">{o.id}</span>
                <span class="badge badge-success">{o.status}</span>
              </div>
              <div class="mt-3 flex items-center gap-2">
                {o.items.slice(0, 8).map((it) => (
                  <span class="shot flex h-10 w-10 items-center justify-center rounded-md text-lg"><span class="emoji">{it.emoji}</span></span>
                ))}
                <span class="ml-auto text-xs text-muted-foreground">{o.createdFmt}</span>
              </div>
              <div class="mt-3 flex items-center justify-between border-t border-border pt-3 text-sm">
                <span class="text-muted-foreground">共 {o.count} 件</span>
                <span class="text-muted-foreground">合计 <span class="text-base font-semibold text-foreground tabular-nums">{o.totalFmt}</span></span>
              </div>
            </a>
          ))}
        </div>
      )}
    </Layout>
  );
}
