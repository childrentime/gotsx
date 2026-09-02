import type { PageProps } from "gotsx";
import { get, catLabel, productReviews, flashLeftMs } from "host:catalog";
import { fmtPrice, fmtSold } from "host:intl";
import { url } from "host:site";
import Layout from "../../components/Layout.server";
import Stars from "../../ui/Stars";
import Icon from "../../ui/Icon";
import AddToCart from "../../islands/AddToCart.client";
import Countdown from "../../islands/Countdown.client";
import ProductGallery from "../../islands/ProductGallery.client";
import Related from "../../islands/Related.client";

export default function ProductPage({ params, cookies, locale }: PageProps) {
  const sid = cookies.sid ?? "";
  const p = get(params.id);
  const reviews = productReviews(p.id);
  const off = Math.round((1 - p.price / p.orig) * 100);
  const ld = JSON.stringify({
    "@context": "https://schema.org",
    "@type": "Product",
    name: p.title,
    description: p.desc,
    sku: p.id,
    category: catLabel(p.category),
    offers: {
      "@type": "Offer",
      price: (p.price / 100).toFixed(2),
      priceCurrency: "CNY",
      availability: p.stock > 0 ? "https://schema.org/InStock" : "https://schema.org/OutOfStock",
      url: url(`/p/${p.id}`),
    },
    aggregateRating: { "@type": "AggregateRating", ratingValue: p.rating, reviewCount: p.reviews },
  });
  const meta = `${p.desc} 现价 ${fmtPrice(p.price)}, 已售 ${fmtSold(p.sold)}, ${p.reviews} 条好评, 满 ¥69 包邮。`;
  const perks = [["truck", "满 ¥69 免运费"], ["undo", "7 天无理由退换"], ["shield", "正品保障 · 假一赔十"], ["zap", "24 小时内发货"]];
  const dist = [["5 星", 78], ["4 星", 16], ["3 星", 6]];
  return (
    <Layout title={p.title} sid={sid} locale={locale} active={p.category} wide desc={meta} path={`/p/${p.id}`} image={url(`/img/p/${p.id}`)} ogType="product" ld={ld}>
      <nav class="mb-5 flex items-center gap-1.5 text-xs text-muted-foreground" aria-label="breadcrumb">
        <a href="/" class="transition-colors hover:text-foreground">首页</a><Icon name="chevron-right" className="h-3 w-3" />
        <a href={`/c/${p.category}`} class="transition-colors hover:text-foreground">{catLabel(p.category)}</a><Icon name="chevron-right" className="h-3 w-3" />
        <span class="line-clamp-1 text-foreground">{p.title}</span>
      </nav>

      <div class="card grid gap-8 p-6 md:grid-cols-[400px_1fr]">
        <ProductGallery id={p.id} hue={p.hue} count={p.gallery.length + 1} />
        <div>
          <div class="flex flex-wrap items-center gap-2">
            {p.flash && <span class="badge badge-primary">闪购</span>}
            <span class="badge badge-outline">{catLabel(p.category)}</span>
            {p.flash && <span class="ml-auto flex items-center gap-2 text-xs text-muted-foreground">距结束 <Countdown left0={flashLeftMs()} /></span>}
          </div>
          <h1 class="mt-3 text-2xl font-semibold leading-snug tracking-tight">{p.title}</h1>
          <div class="mt-2 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-muted-foreground">
            <span class="flex items-center gap-1.5"><Stars rating={p.rating} size="md" /><span class="font-medium text-foreground">{p.rating}</span></span>
            <span>{p.reviews} 条评价</span><span>·</span>
            <span>已售 {fmtSold(p.sold)}</span>
          </div>
          <div class="mt-5 flex flex-wrap items-baseline gap-3 border-y border-border py-5">
            <span class="text-3xl font-semibold tracking-tight tabular-nums">{fmtPrice(p.price)}</span>
            <span class="text-sm text-muted-foreground line-through tabular-nums">{fmtPrice(p.orig)}</span>
            {off > 0 && <span class="badge badge-secondary tabular-nums">立省 {off}%</span>}
          </div>
          <AddToCart id={p.id} variants={p.variants} stock={p.stock} />
          <div class="mt-6 flex flex-wrap gap-x-5 gap-y-2 text-xs text-muted-foreground">
            {perks.map((f) => <span class="flex items-center gap-1.5"><Icon name={f[0]} className="h-3.5 w-3.5" />{f[1]}</span>)}
          </div>
        </div>
      </div>

      <div class="mt-5 grid gap-5 lg:grid-cols-[1fr_360px]">
        <section class="card h-fit p-6">
          <h2 class="mb-3 text-base font-semibold">商品详情</h2>
          <p class="text-sm leading-7 text-muted-foreground">{p.desc}</p>
          <div class="mt-5 grid grid-cols-2 gap-2 sm:grid-cols-3">
            {["材质精选", "工厂直发", "严格质检", "环保包装", "顺丰可选", "售后无忧"].map((s) => (
              <div class="flex items-center gap-2 rounded-md border border-border px-3 py-2 text-[13px]"><Icon name="check" className="h-3.5 w-3.5 text-success" />{s}</div>
            ))}
          </div>
          <p class="mt-5 text-xs leading-6 text-muted-foreground">本页面由 Go 在服务端渲染(数据接口含模拟延迟), 加购面板、图廊、相关推荐是编译成 signals 的岛。商品图为演示用 emoji 棚拍。</p>
        </section>
        <section class="card h-fit p-6">
          <div class="mb-4 flex items-baseline justify-between">
            <h2 class="text-base font-semibold">用户评价</h2>
            <span class="text-xs text-muted-foreground">{p.reviews} 条</span>
          </div>
          <div class="mb-5 flex items-center gap-5 rounded-md bg-muted p-4">
            <div class="text-center">
              <div class="text-3xl font-semibold tracking-tight tabular-nums">{p.rating}</div>
              <Stars rating={p.rating} />
            </div>
            <div class="flex-1 space-y-1.5">
              {dist.map((r) => (
                <div class="flex items-center gap-2 text-[11px] text-muted-foreground">
                  <span class="w-7">{r[0]}</span>
                  <div class="h-1 flex-1 overflow-hidden rounded-full bg-border"><div class="h-full rounded-full bg-foreground" style={`width:${r[1]}%`}></div></div>
                  <span class="w-8 text-right tabular-nums">{r[1]}%</span>
                </div>
              ))}
            </div>
          </div>
          <div class="divide-y divide-border">
            {reviews.map((r) => (
              <div class="py-3 first:pt-0 last:pb-0">
                <div class="flex items-center justify-between">
                  <span class="text-sm font-medium">{r.user}</span>
                  <Stars rating={r.stars} />
                </div>
                <p class="mt-1.5 text-[13px] leading-6 text-muted-foreground">{r.text}</p>
                <div class="mt-1 text-[11px] text-muted-foreground">{r.date}</div>
              </div>
            ))}
          </div>
        </section>
      </div>

      <section class="mt-10">
        <h2 class="mb-4 text-lg font-semibold tracking-tight">相关推荐</h2>
        <Related id={p.id} />
      </section>
    </Layout>
  );
}
