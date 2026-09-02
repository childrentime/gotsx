import type { Node } from "gotsx";
import { categories } from "host:catalog";
import { view } from "host:cart";
import { url } from "host:site";
import CartBadge from "../islands/CartBadge.client";
import LocaleSwitch from "../islands/LocaleSwitch.client";

export default function Layout({ title, sid, active = "", q = "", wide = false, desc = "", path = "", image = "", ld = "", ogType = "website", locale = "zh", children }: { title: string; sid: string; active?: string; q?: string; wide?: boolean; desc?: string; path?: string; image?: string; ld?: string; ogType?: string; locale?: string; children?: Node }) {
  const lc = locale !== "" ? locale : "zh";
  const d = desc !== "" ? desc : "gomu — 全球好物, 工厂直发。20 万+ 精选好物, 满 ¥69 包邮, 7 天无理由退换。";
  const cart = view(sid);
  const cats = categories();
  const container = wide ? "max-w-[1240px]" : "max-w-6xl";
  const other = lc === "en" ? "zh" : "en";
  const navCls = (on: boolean) => on ? "shrink-0 rounded-full bg-ink-900 px-3.5 py-1.5 font-semibold text-white" : "shrink-0 rounded-full px-3.5 py-1.5 text-ink-600 transition hover:bg-ink-100";
  return (
    <html lang={lc}>
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1" />
        <meta name="theme-color" content="#f9491f" />
        <link rel="manifest" href="/manifest.webmanifest" />
        <link rel="icon" href="/icon.svg" />
        <link rel="apple-touch-icon" href="/icon.svg" />
        <title>{title} · gomu</title>
        <meta name="description" content={d} />
        <link rel="canonical" href={url(lpath(lc, path))} />
        <meta property="og:site_name" content="gomu" />
        <meta property="og:title" content={title} />
        <meta property="og:description" content={d} />
        <meta property="og:type" content={ogType} />
        <meta property="og:url" content={url(lpath(lc, path))} />
        {image !== "" && <meta property="og:image" content={image} />}
        <meta name="twitter:card" content="summary_large_image" />
        <link rel="stylesheet" href="/public/tailwind.css" />
        {ld !== "" && jsonLd(ld)}
      </head>
      <body class="min-h-screen">
        <div id="gotsx-bar"></div>
        <div class="bg-ink-900 text-[12px] text-ink-100">
          <div class={`mx-auto flex ${container} items-center gap-4 px-4 py-1.5`}>
            <div class="flex flex-1 items-center justify-center gap-6">
              <span>🚚 {t(lc, "bar.freeship")}</span><span class="hidden sm:inline">·</span>
              <span class="hidden sm:inline">↩️ {t(lc, "bar.return")}</span><span class="hidden md:inline">·</span>
              <span class="hidden md:inline">🛡️ {t(lc, "bar.priceGuard")}</span><span class="hidden md:inline">·</span>
              <span class="hidden md:inline">⚡ {t(lc, "bar.fastship")}</span>
            </div>
            <LocaleSwitch locale={lc} path={path} other={other} label={t(lc, "lang.other")} />
          </div>
        </div>
        <header class="sticky top-0 z-40 border-b border-ink-100 bg-white/85 backdrop-blur-md">
          <div class={`mx-auto flex ${container} items-center gap-5 px-4 py-3`}>
            <a href={lpath(lc, "/")} class="flex shrink-0 items-center gap-2">
              <span class="flex h-9 w-9 items-center justify-center rounded-xl bg-gradient-to-br from-brand-500 to-brand-600 text-lg font-black text-white shadow-sm">g</span>
              <span class="text-2xl font-black tracking-tight text-ink-900">gomu</span>
            </a>
            <form method="get" action={lpath(lc, "/search")} class="flex min-w-0 flex-1 items-center">
              <div class="flex h-11 min-w-0 flex-1 items-center rounded-full border-2 border-ink-900 bg-white pl-4 pr-1 transition focus-within:border-brand-500">
                <span class="mr-2 text-ink-300">🔍</span>
                <input name="q" value={q} placeholder={t(lc, "search.placeholder")} class="h-full min-w-0 flex-1 bg-transparent text-[15px] outline-none placeholder:text-ink-300" />
                <button class="h-8 shrink-0 rounded-full bg-brand-500 px-5 text-sm font-bold text-white transition hover:bg-brand-600">{t(lc, "search.button")}</button>
              </div>
            </form>
            <a href={lpath(lc, "/orders")} class={active === "orders" ? "hidden shrink-0 items-center gap-1.5 text-sm font-semibold text-brand-600 sm:flex" : "hidden shrink-0 items-center gap-1.5 text-sm text-ink-600 transition hover:text-brand-600 sm:flex"}>
              <span class="text-lg">📦</span> {t(lc, "nav.orders")}
            </a>
            <CartBadge count={cart.count} label={t(lc, "nav.cart")} />
          </div>
          <div class={`mx-auto ${container} px-4`}>
            <nav class="no-scrollbar flex items-center gap-1 overflow-x-auto pb-2.5 text-[13px]">
              <a href={lpath(lc, "/")} class={navCls(active === "home")}>🏠 {t(lc, "nav.home")}</a>
              {cats.map((c) => (
                <a href={lpath(lc, `/c/${c.key}`)} class={navCls(active === c.key)}>{c.emoji} {c.label}</a>
              ))}
            </nav>
          </div>
        </header>
        <main class={`mx-auto ${container} px-4 py-6`}>{children}</main>
        <footer class="mt-20 border-t border-ink-100 bg-white">
          <div class={`mx-auto ${container} grid gap-8 px-4 py-12 sm:grid-cols-2 lg:grid-cols-4`}>
            <div>
              <div class="flex items-center gap-2">
                <span class="flex h-8 w-8 items-center justify-center rounded-lg bg-brand-500 text-base font-black text-white">g</span>
                <span class="text-xl font-black">gomu</span>
              </div>
              <p class="mt-3 text-[13px] leading-6 text-ink-400">{t(lc, "footer.tagline")}</p>
            </div>
            <div>
              <div class="mb-3 text-sm font-bold text-ink-800">{t(lc, "footer.guide")}</div>
              <ul class="space-y-2 text-[13px] text-ink-500"><li>·</li><li>·</li><li>·</li></ul>
            </div>
            <div>
              <div class="mb-3 text-sm font-bold text-ink-800">{t(lc, "footer.about")}</div>
              <ul class="space-y-2 text-[13px] text-ink-500"><li>·</li><li>·</li><li>·</li></ul>
            </div>
            <div>
              <div class="mb-3 text-sm font-bold text-ink-800">{t(lc, "footer.promise")}</div>
              <div class="flex flex-wrap gap-2 text-2xl">
                <span>🔒</span><span>🛡️</span><span>💸</span><span>🚚</span>
              </div>
            </div>
          </div>
          <div class="border-t border-ink-100 py-5 text-center text-[12px] text-ink-400">
            © 2026 gomu · gotsx
          </div>
        </footer>
      </body>
    </html>
  );
}
