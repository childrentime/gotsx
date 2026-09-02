import type { PageProps, Meta } from "gotsx";
import { get } from "host:orders";
import Layout from "../../components/Layout.server";
import Icon from "../../ui/Icon";

export function meta(props: PageProps): Meta {
  const lc = props.locale !== "" ? props.locale : "en";
  return { title: tv(lc, "order.title", { id: props.params.id }), noIndex: true };
}

export default function OrderDetail({ params, cookies, locale, path, flash }: PageProps) {
  const lc = locale !== "" ? locale : "en";
  const sid = cookies.sid ?? "";
  const o = get(sid, params.id);
  return (
    <Layout sid={sid} locale={lc} active="orders" wide path={path}>
      {flash.map((f) => (
        <div class={"alert alert-" + f.kind + " mb-5"} role="status">{f.text}</div>
      ))}
      <div class="card mb-5 flex items-center gap-4 px-6 py-5">
        <span class="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-success/10 text-success"><Icon name="check" className="h-5 w-5" /></span>
        <div>
          <div class="text-base font-semibold">{t(lc, "order.success")}</div>
          <div class="mt-0.5 text-xs text-muted-foreground">{t(lc, "order.number")} <span class="font-mono">{o.id}</span> · {o.createdFmt} · {t(lc, "status." + o.status)}</div>
        </div>
        <a href="/orders" class="btn btn-outline ml-auto hidden sm:inline-flex">{t(lc, "order.viewAll")}</a>
      </div>
      <div class="grid gap-5 lg:grid-cols-[1fr_340px]">
        <div class="card overflow-hidden">
          <div class="border-b border-border px-5 py-3 text-sm font-medium">{t(lc, "order.items")}</div>
          <div class="divide-y divide-border">
            {o.items.map((it) => (
              <div class="flex items-center gap-4 p-4 text-sm">
                <span class="shot flex h-14 w-14 shrink-0 items-center justify-center rounded-md text-2xl"><span class="emoji">{it.emoji}</span></span>
                <div class="min-w-0 flex-1">
                  <div class="line-clamp-1">{it.title}</div>
                  {it.variant !== "" && <div class="mt-0.5 text-xs text-muted-foreground">{it.variant}</div>}
                </div>
                <span class="text-xs text-muted-foreground tabular-nums">×{it.qty}</span>
                <span class="w-20 text-right font-medium tabular-nums">{it.lineFmt}</span>
              </div>
            ))}
          </div>
        </div>
        <div class="h-fit space-y-4">
          <div class="card p-5 text-sm">
            <h2 class="mb-3 flex items-center gap-2 text-base font-semibold"><Icon name="map-pin" />{t(lc, "order.shipping")}</h2>
            <div class="space-y-1.5 text-muted-foreground">
              <div><span class="font-medium text-foreground">{o.name}</span> · {o.phone}</div>
              <div class="leading-6">{o.address}</div>
            </div>
          </div>
          <div class="card p-5">
            <div class="flex items-baseline justify-between">
              <span class="text-sm text-muted-foreground">{t(lc, "order.paid")}</span>
              <span class="text-2xl font-semibold tracking-tight tabular-nums">{o.totalFmt}</span>
            </div>
            <a href="/" class="btn btn-primary mt-4 w-full">{t(lc, "order.continue")}</a>
          </div>
        </div>
      </div>
    </Layout>
  );
}
