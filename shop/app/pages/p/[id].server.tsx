import type { PageProps, Meta } from "gotsx";
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

// Meta runs before the page: title / description / og:image for the product (get() 404s for an unknown id,
// so the not-found page is served before the page body is even attempted)
export function meta(props: PageProps): Meta {
  const lc = props.locale !== "" ? props.locale : "en";
  const p = get(props.params.id);
  return {
    title: p.title,
    description: tv(lc, "product.meta", { desc: p.desc, price: fmtPrice(p.price), sold: fmtSold(p.sold), reviews: String(p.reviews) }),
    image: url(`/img/p/${p.id}`),
  };
}

export default function ProductPage({ params, cookies, locale, path }: PageProps) {
  const lc = locale !== "" ? locale : "en";
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
      priceCurrency: "USD",
      availability: p.stock > 0 ? "https://schema.org/InStock" : "https://schema.org/OutOfStock",
      url: url(`/p/${p.id}`),
    },
    aggregateRating: { "@type": "AggregateRating", ratingValue: p.rating, reviewCount: p.reviews },
  });
  const perks = [["truck", t(lc, "pperk.ship")], ["undo", t(lc, "pperk.return")], ["shield", t(lc, "pperk.auth")], ["zap", t(lc, "pperk.fast")]];
  const dist = [{ label: tv(lc, "product.stars", { n: "5" }), pct: 78 }, { label: tv(lc, "product.stars", { n: "4" }), pct: 16 }, { label: tv(lc, "product.stars", { n: "3" }), pct: 6 }];
  const features = [t(lc, "feature.1"), t(lc, "feature.2"), t(lc, "feature.3"), t(lc, "feature.4"), t(lc, "feature.5"), t(lc, "feature.6")];
  return (
    <Layout sid={sid} locale={lc} active={p.category} wide path={path}>
      {jsonLd(ld)}
      <nav class="mb-5 flex items-center gap-1.5 text-xs text-muted-foreground" aria-label="breadcrumb">
        <a href="/" class="transition-colors hover:text-foreground">{t(lc, "nav.home")}</a><Icon name="chevron-right" className="h-3 w-3" />
        <a href={`/c/${p.category}`} class="transition-colors hover:text-foreground">{catLabel(p.category)}</a><Icon name="chevron-right" className="h-3 w-3" />
        <span class="line-clamp-1 text-foreground">{p.title}</span>
      </nav>

      <div class="card grid gap-8 p-6 md:grid-cols-[400px_1fr]">
        <ProductGallery id={p.id} hue={p.hue} count={p.gallery.length + 1} locale={lc} />
        <div>
          <div class="flex flex-wrap items-center gap-2">
            {p.flash && <span class="badge badge-primary">{t(lc, "product.flash")}</span>}
            <span class="badge badge-outline">{catLabel(p.category)}</span>
            {p.flash && <span class="ml-auto flex items-center gap-2 text-xs text-muted-foreground">{t(lc, "flash.ends")} <Countdown left0={flashLeftMs()} /></span>}
          </div>
          <h1 class="mt-3 text-2xl font-semibold leading-snug tracking-tight">{p.title}</h1>
          <div class="mt-2 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-muted-foreground">
            <span class="flex items-center gap-1.5"><Stars rating={p.rating} size="md" /><span class="font-medium text-foreground">{p.rating}</span></span>
            <span>{tv(lc, "product.reviews", { n: String(p.reviews) })}</span><span>·</span>
            <span>{tv(lc, "product.sold", { n: fmtSold(p.sold) })}</span>
          </div>
          <div class="mt-5 flex flex-wrap items-baseline gap-3 border-y border-border py-5">
            <span class="text-3xl font-semibold tracking-tight tabular-nums">{fmtPrice(p.price)}</span>
            <span class="text-sm text-muted-foreground line-through tabular-nums">{fmtPrice(p.orig)}</span>
            {off > 0 && <span class="badge badge-secondary tabular-nums">{tv(lc, "product.save", { n: String(off) })}</span>}
          </div>
          <AddToCart id={p.id} variants={p.variants} stock={p.stock} locale={lc} />
          <div class="mt-6 flex flex-wrap gap-x-5 gap-y-2 text-xs text-muted-foreground">
            {perks.map((f) => <span class="flex items-center gap-1.5"><Icon name={f[0]} className="h-3.5 w-3.5" />{f[1]}</span>)}
          </div>
        </div>
      </div>

      <div class="mt-5 grid gap-5 lg:grid-cols-[1fr_360px]">
        <section class="card h-fit p-6">
          <h2 class="mb-3 text-base font-semibold">{t(lc, "product.details")}</h2>
          <p class="text-sm leading-7 text-muted-foreground">{p.desc}</p>
          <div class="mt-5 grid grid-cols-2 gap-2 sm:grid-cols-3">
            {features.map((s) => (
              <div class="flex items-center gap-2 rounded-md border border-border px-3 py-2 text-[13px]"><Icon name="check" className="h-3.5 w-3.5 text-success" />{s}</div>
            ))}
          </div>
          <p class="mt-5 text-xs leading-6 text-muted-foreground">{t(lc, "product.note")}</p>
        </section>
        <section class="card h-fit p-6">
          <div class="mb-4 flex items-baseline justify-between">
            <h2 class="text-base font-semibold">{t(lc, "product.reviewsTitle")}</h2>
            <span class="text-xs text-muted-foreground">{tv(lc, "product.reviewsCount", { n: String(p.reviews) })}</span>
          </div>
          <div class="mb-5 flex items-center gap-5 rounded-md bg-muted p-4">
            <div class="text-center">
              <div class="text-3xl font-semibold tracking-tight tabular-nums">{p.rating}</div>
              <Stars rating={p.rating} />
            </div>
            <div class="flex-1 space-y-1.5">
              {dist.map((r) => (
                <div class="flex items-center gap-2 text-[11px] text-muted-foreground">
                  <span class="w-12">{r.label}</span>
                  <div class="h-1 flex-1 overflow-hidden rounded-full bg-border"><div class="h-full rounded-full bg-foreground" style={`width:${r.pct}%`}></div></div>
                  <span class="w-8 text-right tabular-nums">{r.pct}%</span>
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
        <h2 class="mb-4 text-lg font-semibold tracking-tight">{t(lc, "product.related")}</h2>
        <Related id={p.id} locale={lc} />
      </section>
    </Layout>
  );
}
