import type { PageProps, Meta } from "gotsx";
import { list } from "host:orders";
import Layout from "../../components/Layout.server";
import Icon from "../../ui/Icon";

export function meta(props: PageProps): Meta {
  const lc = props.locale !== "" ? props.locale : "en";
  return { title: t(lc, "orders.title"), noIndex: true };
}

export default function OrdersPage({ cookies, locale, path }: PageProps) {
  const lc = locale !== "" ? locale : "en";
  const sid = cookies.sid ?? "";
  const orders = list(sid);
  return (
    <Layout sid={sid} locale={lc} active="orders" wide path={path}>
      <h1 class="mb-5 text-xl font-semibold tracking-tight">{t(lc, "orders.title")}</h1>
      {orders.length === 0 ? (
        <div class="card flex flex-col items-center py-24 text-center">
          <span class="flex h-12 w-12 items-center justify-center rounded-full bg-muted text-muted-foreground"><Icon name="package" className="h-5 w-5" /></span>
          <p class="mt-4 font-medium">{t(lc, "orders.empty")}</p>
          <a href="/" class="btn btn-primary mt-6">{t(lc, "orders.first")}</a>
        </div>
      ) : (
        <div class="space-y-3">
          {orders.map((o) => (
            <a href={`/orders/${o.id}`} class="card block p-5 transition-colors hover:border-foreground/25">
              <div class="flex items-center justify-between">
                <span class="font-mono text-sm">{o.id}</span>
                <span class="badge badge-success">{t(lc, "status." + o.status)}</span>
              </div>
              <div class="mt-3 flex items-center gap-2">
                {o.items.slice(0, 8).map((it) => (
                  <span class="shot flex h-10 w-10 items-center justify-center rounded-md text-lg"><span class="emoji">{it.emoji}</span></span>
                ))}
                <span class="ml-auto text-xs text-muted-foreground">{o.createdFmt}</span>
              </div>
              <div class="mt-3 flex items-center justify-between border-t border-border pt-3 text-sm">
                <span class="text-muted-foreground">{plural(lc, "orders.count", o.count)}</span>
                <span class="text-muted-foreground">{t(lc, "orders.total")} <span class="text-base font-semibold text-foreground tabular-nums">{o.totalFmt}</span></span>
              </div>
            </a>
          ))}
        </div>
      )}
    </Layout>
  );
}
