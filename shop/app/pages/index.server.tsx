import type { PageProps } from "gotsx";
import { flashCards, flashLeftMs, categories } from "host:catalog";
import { url, name as siteName } from "host:site";
import Layout from "../components/Layout.server";
import Stars from "../ui/Stars";
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
  return (
    <Layout title="全球好物 · 工厂直发" sid={sid} locale={locale} active="home" wide path="/" ld={ld}>
      {/* Hero */}
      <section class="grid gap-4 lg:grid-cols-[1fr_320px]">
        <div class="relative overflow-hidden rounded-xl2 bg-gradient-to-br from-brand-500 via-brand-500 to-brand-700 p-8 text-white sm:p-10">
          <div class="absolute -right-8 -top-8 text-[200px] opacity-15 blur-[1px]">🛍️</div>
          <div class="relative">
            <span class="inline-block rounded-full bg-white/20 px-3 py-1 text-xs font-semibold backdrop-blur">🎉 新人首单立减 ¥15</span>
            <h1 class="mt-4 max-w-lg text-3xl font-black leading-tight tracking-tight sm:text-4xl">像逛工厂一样<br />把好物直接搬回家</h1>
            <p class="mt-3 max-w-md text-sm text-white/80">20 万+ 精选好物 · 全场满 ¥69 包邮 · 7 天无理由退换 · 90 天价保</p>
            <div class="mt-6 flex flex-wrap gap-2">
              <span class="rounded-lg bg-white/15 px-3 py-1.5 text-xs font-semibold backdrop-blur">满 ¥29 减 ¥5</span>
              <span class="rounded-lg bg-white/15 px-3 py-1.5 text-xs font-semibold backdrop-blur">满 ¥99 减 ¥20</span>
              <span class="rounded-lg bg-white/15 px-3 py-1.5 text-xs font-semibold backdrop-blur">品牌闪购 5 折起</span>
            </div>
          </div>
        </div>
        <div class="grid grid-cols-4 gap-2 rounded-xl2 border border-ink-100 bg-white p-4 lg:grid-cols-2">
          {cats.slice(0, 8).map((c) => (
            <a href={`/c/${c.key}`} class="flex flex-col items-center gap-1.5 rounded-xl p-2 text-center transition hover:bg-ink-50">
              <span class="flex h-12 w-12 items-center justify-center rounded-full bg-ink-50 text-2xl">{c.emoji}</span>
              <span class="text-xs font-medium text-ink-700">{c.label}</span>
            </a>
          ))}
        </div>
      </section>

      {/* 服务保障条 */}
      <div class="mt-4 grid grid-cols-2 gap-3 sm:grid-cols-4">
        {[["🚚", "满 ¥69 包邮", "全国大部分地区"], ["⚡", "24h 极速发货", "工厂直连"], ["↩️", "7 天退换", "无理由"], ["🛡️", "90 天价保", "买贵退差"]].map((f) => (
          <div class="flex items-center gap-3 rounded-xl border border-ink-100 bg-white px-4 py-3">
            <span class="text-2xl">{f[0]}</span>
            <div><div class="text-[13px] font-bold text-ink-800">{f[1]}</div><div class="text-[11px] text-ink-400">{f[2]}</div></div>
          </div>
        ))}
      </div>

      {/* 限时闪购 */}
      <section class="mt-6 overflow-hidden rounded-xl2 border border-brand-100 bg-white">
        <div class="flex items-center gap-3 bg-gradient-to-r from-brand-500 to-brand-400 px-5 py-3 text-white">
          <span class="text-lg font-black">⚡ 限时闪购</span>
          <span class="rounded-full bg-white/20 px-2 py-0.5 text-[11px] font-semibold backdrop-blur">5 折起</span>
          <div class="ml-auto flex items-center gap-2 rounded-full bg-white/15 px-3 py-1 backdrop-blur">
            <span class="text-xs">距结束</span><Countdown left0={flashLeftMs()} />
          </div>
        </div>
        <div class="no-scrollbar flex gap-3 overflow-x-auto p-4">
          {deals.map((p) => (
            <a href={`/p/${p.id}`} class="group w-32 shrink-0">
              <div class="shot flex aspect-square items-center justify-center rounded-xl border border-ink-100" style={`background:radial-gradient(120% 120% at 50% 18%, #fff, hsl(${p.hue} 46% 95%))`}>
                <span class="emoji text-5xl">{p.emoji}</span>
              </div>
              <div class="mt-2 line-clamp-1 text-xs text-ink-700">{p.title}</div>
              <div class="mt-0.5 flex items-baseline gap-1">
                <span class="text-[15px] font-extrabold text-brand-600">{p.priceFmt}</span>
                <span class="text-[10px] text-ink-300 line-through">{p.origFmt}</span>
              </div>
              <div class="mt-1 h-1.5 overflow-hidden rounded-full bg-brand-100">
                <div class="h-full rounded-full bg-gradient-to-r from-brand-500 to-brand-400" style={`width:${p.progress}%`}></div>
              </div>
              <div class="mt-0.5 text-[10px] text-brand-500">已抢 {p.progress}%</div>
            </a>
          ))}
        </div>
      </section>

      {/* 为你推荐(信息流, 骨架加载) */}
      <section class="mt-8">
        <div class="mb-4 flex items-center gap-2">
          <span class="h-5 w-1 rounded-full bg-brand-500"></span>
          <h2 class="text-lg font-extrabold tracking-tight">为你推荐</h2>
          <span class="text-xs text-ink-400">根据你的浏览猜你喜欢</span>
        </div>
        <Feed />
      </section>
    </Layout>
  );
}
