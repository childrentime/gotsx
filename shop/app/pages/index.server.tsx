import type { PageProps } from "gotsx";
import { flashCards, flashLeftMs, categories } from "host:catalog";
import { url, name as siteName } from "host:site";
import Layout from "../components/Layout.server";
import Icon from "../ui/Icon";
import Countdown from "../islands/Countdown.client";
import Feed from "../islands/Feed.client";

export default function Home({ cookies, locale }: PageProps) {
  const sid = cookies.sid ?? "";
  const deals = flashCards().slice(0, 10);
  const cats = categories();
  const ld = JSON.stringify({
    "@context": "https://schema.org",
    "@type": "WebSite",
    name: siteName(),
    url: url("/"),
    potentialAction: { "@type": "SearchAction", target: url("/search?q={q}"), "query-input": "required name=q" },
  });
  const perks = [["truck", "满 ¥69 包邮", "全国大部分地区"], ["zap", "24h 极速发货", "工厂直连"], ["undo", "7 天退换", "无理由"], ["shield", "90 天价保", "买贵退差"]];
  return (
    <Layout title="全球好物 · 工厂直发" sid={sid} locale={locale} active="home" wide path="/" ld={ld}>
      {/* Hero */}
      <section class="grid gap-4 lg:grid-cols-[1fr_360px]">
        <div class="card flex flex-col justify-center p-8 sm:p-10">
          <div>
            <span class="badge badge-secondary">新人首单立减 ¥15</span>
            <h1 class="mt-5 max-w-lg text-3xl font-semibold leading-tight tracking-tight sm:text-4xl">像逛工厂一样,<br />把好物直接搬回家</h1>
            <p class="mt-4 max-w-md text-sm leading-6 text-muted-foreground">20 万+ 精选好物 · 全场满 ¥69 包邮 · 7 天无理由退换 · 90 天价保</p>
          </div>
          <div class="mt-8 flex flex-wrap items-center gap-3">
            <a href={`/c/${cats[0].key}`} class="btn btn-primary">开始逛 <Icon name="arrow-right" /></a>
            <a href="#deals" class="btn btn-outline">今日闪购</a>
            <span class="ml-1 hidden text-xs text-muted-foreground sm:inline">满 ¥29 减 ¥5 · 满 ¥99 减 ¥20</span>
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

      {/* 服务保障条 */}
      <div class="mt-4 grid grid-cols-2 divide-border rounded-lg border border-border bg-card sm:grid-cols-4 sm:divide-x">
        {perks.map((f) => (
          <div class="flex items-center gap-3 px-4 py-3">
            <span class="flex h-8 w-8 shrink-0 items-center justify-center rounded-md bg-muted text-foreground"><Icon name={f[0]} /></span>
            <div class="min-w-0"><div class="text-[13px] font-medium">{f[1]}</div><div class="text-[11px] text-muted-foreground">{f[2]}</div></div>
          </div>
        ))}
      </div>

      {/* 限时闪购 */}
      <section id="deals" class="card mt-8 overflow-hidden">
        <div class="flex items-center gap-3 border-b border-border px-5 py-3">
          <Icon name="zap" />
          <span class="text-sm font-semibold">限时闪购</span>
          <span class="badge badge-secondary">5 折起</span>
          <div class="ml-auto flex items-center gap-2 text-xs text-muted-foreground">
            <span>距结束</span><Countdown left0={flashLeftMs()} />
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
              <div class="mt-1 text-[11px] text-muted-foreground tabular-nums">已抢 {p.progress}%</div>
            </a>
          ))}
        </div>
      </section>

      {/* 为你推荐(信息流, 骨架加载) */}
      <section class="mt-10">
        <div class="mb-4 flex items-baseline gap-3">
          <h2 class="text-lg font-semibold tracking-tight">为你推荐</h2>
          <span class="text-xs text-muted-foreground">根据你的浏览猜你喜欢</span>
        </div>
        <Feed />
      </section>
    </Layout>
  );
}
