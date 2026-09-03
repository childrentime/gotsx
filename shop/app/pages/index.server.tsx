import type { PageProps, Meta } from "gotsx";
import { flashCards, flashLeftMs, categories } from "host:catalog";
import { url, name as siteName } from "host:site";
import Layout from "../components/Layout.server";
import Icon from "../ui/Icon";
import Countdown from "../islands/Countdown.client";
import Feed from "../islands/Feed.client";

// Page meta → pages/_layout.server.tsx renders <title>, description, canonical and og:* from it
export function meta(props: PageProps): Meta {
  const lc = props.locale !== "" ? props.locale : "en";
  return { title: t(lc, "home.title") };
}

export default function Home({ cookies, locale, path }: PageProps) {
  const lc = locale !== "" ? locale : "en";
  const deals = flashCards().slice(0, 10);
  const cats = categories();
  const ld = JSON.stringify({
    "@context": "https://schema.org",
    "@type": "WebSite",
    name: siteName(),
    url: url("/"),
    potentialAction: { "@type": "SearchAction", target: url("/search?q={q}"), "query-input": "required name=q" },
  });
  const perks = [
    ["truck", t(lc, "perk.ship.t"), t(lc, "perk.ship.d")],
    ["zap", t(lc, "perk.fast.t"), t(lc, "perk.fast.d")],
    ["undo", t(lc, "perk.return.t"), t(lc, "perk.return.d")],
    ["shield", t(lc, "perk.price.t"), t(lc, "perk.price.d")],
  ];
  return (
    <Layout locale={lc} active="home" wide path={path}>
      {jsonLd(ld)}
      {/* Hero */}
      <section class="grid gap-4 lg:grid-cols-[1fr_360px]">
        <div class="card flex flex-col justify-center p-8 sm:p-10">
          <div>
            <span class="badge badge-secondary">{t(lc, "home.badge")}</span>
            <h1 class="mt-5 max-w-lg text-3xl font-semibold leading-tight tracking-tight sm:text-4xl">{t(lc, "home.heading1")}<br />{t(lc, "home.heading2")}</h1>
            <p class="mt-4 max-w-md text-sm leading-6 text-muted-foreground">{t(lc, "home.sub")}</p>
          </div>
          <div class="mt-8 flex flex-wrap items-center gap-3">
            <a href={`/c/${cats[0].key}`} class="btn btn-primary">{t(lc, "home.cta")} <Icon name="arrow-right" /></a>
            <a href="#deals" class="btn btn-outline">{t(lc, "home.deals")}</a>
            <span class="ml-1 hidden text-xs text-muted-foreground sm:inline">{t(lc, "home.coupons")}</span>
          </div>
        </div>
        <div class="card grid grid-cols-4 gap-1 p-3 lg:grid-cols-2">
          {cats.slice(0, 8).map((c) => (
            <a href={`/c/${c.key}`} class="flex flex-col items-center gap-2 rounded-md p-3 text-center transition-colors hover:bg-accent">
              <span class="flex h-11 w-11 items-center justify-center rounded-full bg-muted text-xl">{c.emoji}</span>
              <span class="text-xs font-medium">{c.label}</span>
            </a>
          ))}
        </div>
      </section>

      {/* Service perks */}
      <div class="mt-4 grid grid-cols-2 divide-border rounded-lg border border-border bg-card sm:grid-cols-4 sm:divide-x">
        {perks.map((f) => (
          <div class="flex items-center gap-3 px-4 py-3">
            <span class="flex h-8 w-8 shrink-0 items-center justify-center rounded-md bg-muted text-foreground"><Icon name={f[0]} /></span>
            <div class="min-w-0"><div class="text-[13px] font-medium">{f[1]}</div><div class="text-[11px] text-muted-foreground">{f[2]}</div></div>
          </div>
        ))}
      </div>

      {/* Flash sale */}
      <section id="deals" class="card mt-8 overflow-hidden">
        <div class="flex items-center gap-3 border-b border-border px-5 py-3">
          <Icon name="zap" />
          <span class="text-sm font-semibold">{t(lc, "flash.title")}</span>
          <span class="badge badge-secondary">{t(lc, "flash.badge")}</span>
          <div class="ml-auto flex items-center gap-2 text-xs text-muted-foreground">
            <span>{t(lc, "flash.ends")}</span><Countdown left0={flashLeftMs()} />
          </div>
        </div>
        <div class="no-scrollbar flex gap-4 overflow-x-auto p-5">
          {deals.map((p) => (
            <a href={`/p/${p.id}`} class="group w-32 shrink-0">
              <div class="shot flex aspect-square items-center justify-center rounded-md">
                <span class="emoji text-5xl">{p.emoji}</span>
              </div>
              <div class="mt-2 line-clamp-1 text-xs">{p.title}</div>
              <div class="mt-1 flex items-baseline gap-1.5">
                <span class="text-[14px] font-semibold tabular-nums">{p.priceFmt}</span>
                <span class="text-[11px] text-muted-foreground line-through tabular-nums">{p.origFmt}</span>
              </div>
              <div class="mt-2 h-1 overflow-hidden rounded-full bg-muted">
                <div class="h-full rounded-full bg-foreground" style={`width:${p.progress}%`}></div>
              </div>
              <div class="mt-1 text-[11px] text-muted-foreground tabular-nums">{tv(lc, "flash.claimed", { n: String(p.progress) })}</div>
            </a>
          ))}
        </div>
      </section>

      {/* For you (feed with skeleton loading) */}
      <section class="mt-10">
        <div class="mb-4 flex items-baseline gap-3">
          <h2 class="text-lg font-semibold tracking-tight">{t(lc, "feed.title")}</h2>
          <span class="text-xs text-muted-foreground">{t(lc, "feed.sub")}</span>
        </div>
        <Feed locale={lc} />
      </section>
    </Layout>
  );
}
