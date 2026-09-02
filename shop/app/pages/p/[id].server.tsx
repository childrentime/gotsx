import type { PageProps } from "gotsx";
import { get, catLabel, productReviews, flashLeftMs } from "host:catalog";
import { fmtPrice, fmtSold } from "host:intl";
import { url } from "host:site";
import Layout from "../../components/Layout.server";
import Stars from "../../ui/Stars";
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
  return (
    <Layout title={p.title} sid={sid} locale={locale} active={p.category} wide desc={meta} path={`/p/${p.id}`} image={url(`/img/p/${p.id}`)} ogType="product" ld={ld}>
      <nav class="mb-4 flex items-center gap-1.5 text-xs text-ink-400">
        <a href="/" class="transition hover:text-brand-600">首页</a><span>/</span>
        <a href={`/c/${p.category}`} class="transition hover:text-brand-600">{catLabel(p.category)}</a><span>/</span>
        <span class="line-clamp-1 text-ink-600">{p.title}</span>
      </nav>

      <div class="grid gap-6 rounded-xl2 border border-ink-100 bg-white p-5 md:grid-cols-[400px_1fr] sm:p-6">
        <ProductGallery id={p.id} hue={p.hue} count={p.gallery.length + 1} />
        <div>
          {p.flash && <span class="inline-block rounded-md bg-brand-500 px-2 py-0.5 text-xs font-bold text-white">⚡ 限时闪购</span>}
          <h1 class="mt-2 text-2xl font-extrabold leading-snug tracking-tight">{p.title}</h1>
          <div class="mt-2.5 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-ink-500">
            <span class="flex items-center gap-1"><Stars rating={p.rating} size="md" /><span class="font-semibold text-ink-800">{p.rating}</span></span>
            <span>{p.reviews} 条评价</span><span>·</span>
            <span>已售 {fmtSold(p.sold)}</span>
          </div>
          <div class="mt-4 rounded-xl2 bg-gradient-to-br from-brand-50 to-white p-5">
            <div class="flex flex-wrap items-baseline gap-2.5">
              <span class="text-sm font-bold text-brand-600">¥</span>
              <span class="-ml-1.5 text-4xl font-black tracking-tight text-brand-600">{fmtPrice(p.price).slice(1)}</span>
              <span class="text-sm text-ink-400 line-through">{fmtPrice(p.orig)}</span>
              {off > 0 && <span class="rounded-md bg-brand-500 px-2 py-0.5 text-xs font-bold text-white">立省 {off}%</span>}
              {p.flash && <span class="ml-auto flex items-center gap-1.5 text-xs text-ink-500">距结束 <Countdown left0={flashLeftMs()} /></span>}
            </div>
          </div>
          <AddToCart id={p.id} variants={p.variants} stock={p.stock} />
          <div class="mt-6 flex flex-wrap gap-x-5 gap-y-2 border-t border-ink-100 pt-4 text-xs text-ink-500">
            <span class="flex items-center gap-1">🚚 满 ¥69 免运费</span>
            <span class="flex items-center gap-1">↩️ 7 天无理由退换</span>
            <span class="flex items-center gap-1">🛡️ 正品保障 · 假一赔十</span>
            <span class="flex items-center gap-1">⚡ 24 小时内发货</span>
          </div>
        </div>
      </div>

      <div class="mt-5 grid gap-5 lg:grid-cols-[1fr_360px]">
        <section class="rounded-xl2 border border-ink-100 bg-white p-6">
          <h2 class="mb-3 flex items-center gap-2 text-base font-bold"><span class="h-4 w-1 rounded-full bg-brand-500"></span>商品详情</h2>
          <p class="text-sm leading-7 text-ink-600">{p.desc}</p>
          <div class="mt-4 grid grid-cols-2 gap-2.5 sm:grid-cols-3">
            {["材质精选", "工厂直发", "严格质检", "环保包装", "顺丰可选", "售后无忧"].map((t) => (
              <div class="flex items-center gap-2 rounded-lg bg-ink-50 px-3 py-2 text-[13px] text-ink-600"><span class="text-brand-500">✓</span>{t}</div>
            ))}
          </div>
          <p class="mt-4 text-xs leading-6 text-ink-400">本页面由 Go 在服务端渲染(数据接口含模拟延迟), 加购面板、图廊、相关推荐是编译成 signals 的岛。商品图为演示用 emoji 棚拍。</p>
        </section>
        <section class="h-fit rounded-xl2 border border-ink-100 bg-white p-6">
          <div class="mb-4 flex items-center justify-between">
            <h2 class="flex items-center gap-2 text-base font-bold"><span class="h-4 w-1 rounded-full bg-brand-500"></span>用户评价</h2>
            <span class="text-xs text-ink-400">{p.reviews} 条</span>
          </div>
          <div class="mb-4 flex items-center gap-4 rounded-xl bg-ink-50 p-4">
            <div class="text-center">
              <div class="text-3xl font-black text-brand-600">{p.rating}</div>
              <Stars rating={p.rating} />
            </div>
            <div class="flex-1 space-y-1">
              {[["5星", 78], ["4星", 16], ["3星", 6]].map((r) => (
                <div class="flex items-center gap-2 text-[11px] text-ink-400">
                  <span class="w-6">{r[0]}</span>
                  <div class="h-1.5 flex-1 overflow-hidden rounded-full bg-ink-200"><div class="h-full rounded-full bg-gold-400" style={`width:${r[1]}%`}></div></div>
                  <span class="w-8 text-right">{r[1]}%</span>
                </div>
              ))}
            </div>
          </div>
          <div class="space-y-4">
            {reviews.map((r) => (
              <div class="border-b border-ink-50 pb-3 last:border-0">
                <div class="flex items-center justify-between">
                  <span class="text-sm font-medium text-ink-700">{r.user}</span>
                  <Stars rating={r.stars} />
                </div>
                <p class="mt-1.5 text-[13px] leading-6 text-ink-600">{r.text}</p>
                <div class="mt-1 text-[11px] text-ink-300">{r.date}</div>
              </div>
            ))}
          </div>
        </section>
      </div>

      <section class="mt-8">
        <div class="mb-4 flex items-center gap-2">
          <span class="h-5 w-1 rounded-full bg-brand-500"></span>
          <h2 class="text-lg font-extrabold tracking-tight">相关推荐</h2>
        </div>
        <Related id={p.id} />
      </section>
    </Layout>
  );
}
