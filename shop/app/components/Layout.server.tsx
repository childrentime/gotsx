import type { Node } from "gotsx";
import { categories } from "host:catalog";
import { view } from "host:cart";
import { url } from "host:site";
import CartBadge from "../islands/CartBadge.client";
import LocaleSwitch from "../islands/LocaleSwitch.client";
import Icon from "../ui/Icon";

export default function Layout({ title, sid, active = "", q = "", wide = false, desc = "", path = "", image = "", ld = "", ogType = "website", locale = "en", children }: { title: string; sid: string; active?: string; q?: string; wide?: boolean; desc?: string; path?: string; image?: string; ld?: string; ogType?: string; locale?: string; children?: Node }) {
  const lc = locale !== "" ? locale : "en";
  const d = desc !== "" ? desc : t(lc, "meta.description");
  const cart = view(sid);
  const cats = categories();
  const container = wide ? "mx-auto w-full max-w-[1200px] px-6" : "container-page";
  const other = lc === "en" ? "zh" : "en";
  const navCls = (on: boolean) => (on ? "nav-link-active shrink-0" : "nav-link shrink-0");
  return (
    <html lang={lc}>
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1" />
        <meta name="theme-color" content="#ffffff" />
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
        <div class="border-b border-border bg-muted/40 text-xs text-muted-foreground">
          <div class={`${container} flex h-8 items-center gap-4`}>
            <div class="flex flex-1 items-center gap-5 overflow-hidden whitespace-nowrap">
              <span class="flex items-center gap-1.5"><Icon name="truck" className="h-3.5 w-3.5" />{t(lc, "bar.freeship")}</span>
              <span class="hidden items-center gap-1.5 sm:flex"><Icon name="undo" className="h-3.5 w-3.5" />{t(lc, "bar.return")}</span>
              <span class="hidden items-center gap-1.5 md:flex"><Icon name="shield" className="h-3.5 w-3.5" />{t(lc, "bar.priceGuard")}</span>
              <span class="hidden items-center gap-1.5 md:flex"><Icon name="zap" className="h-3.5 w-3.5" />{t(lc, "bar.fastship")}</span>
            </div>
            <LocaleSwitch locale={lc} path={path} other={other} label={t(lc, "lang.other")} />
          </div>
        </div>
        <header class="page-header">
          <div class={`${container} flex h-14 items-center gap-4`}>
            <a href={lpath(lc, "/")} class="flex shrink-0 items-center gap-2">
              <span class="flex h-7 w-7 items-center justify-center rounded-md bg-primary font-mono text-sm font-semibold text-primary-foreground">g</span>
              <span class="text-base font-semibold tracking-tight">gomu</span>
            </a>
            <form method="get" action={lpath(lc, "/search")} class="flex min-w-0 flex-1 items-center gap-2">
              <div class="relative min-w-0 flex-1">
                <Icon name="search" className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
                <input name="q" value={q} placeholder={t(lc, "search.placeholder")} class="input pl-9" />
              </div>
              <button class="btn btn-primary hidden sm:inline-flex">{t(lc, "search.button")}</button>
            </form>
            <a href={lpath(lc, "/orders")} class={active === "orders" ? "hidden shrink-0 items-center gap-1.5 text-sm font-medium text-foreground sm:flex" : "hidden shrink-0 items-center gap-1.5 text-sm text-muted-foreground transition-colors hover:text-foreground sm:flex"}>
              <Icon name="package" /> {t(lc, "nav.orders")}
            </a>
            <CartBadge count={cart.count} label={t(lc, "nav.cart")} />
          </div>
          <div class={container}>
            <nav class="no-scrollbar -mx-1 flex items-center gap-0.5 overflow-x-auto pb-2 text-[13px]">
              <a href={lpath(lc, "/")} class={navCls(active === "home")}>{t(lc, "nav.home")}</a>
              {cats.map((c) => (
                <a href={lpath(lc, `/c/${c.key}`)} class={navCls(active === c.key)}>{c.label}</a>
              ))}
            </nav>
          </div>
        </header>
        <main class={`${container} fade-up py-8`}>{children}</main>
        <footer class="mt-24 border-t border-border">
          <div class={`${container} grid gap-8 py-12 sm:grid-cols-2 lg:grid-cols-4`}>
            <div>
              <div class="flex items-center gap-2">
                <span class="flex h-6 w-6 items-center justify-center rounded-md bg-primary font-mono text-xs font-semibold text-primary-foreground">g</span>
                <span class="font-semibold tracking-tight">gomu</span>
              </div>
              <p class="mt-3 max-w-xs text-[13px] leading-6 text-muted-foreground">{t(lc, "footer.tagline")}</p>
            </div>
            <div>
              <div class="mb-3 text-sm font-medium">{t(lc, "footer.guide")}</div>
              <ul class="space-y-2 text-[13px] text-muted-foreground"><li>{t(lc, "bar.freeship")}</li><li>{t(lc, "bar.return")}</li><li>{t(lc, "bar.fastship")}</li></ul>
            </div>
            <div>
              <div class="mb-3 text-sm font-medium">{t(lc, "footer.about")}</div>
              <ul class="space-y-2 text-[13px] text-muted-foreground"><li>gotsx</li><li>SSR · Go</li><li>islands · signals</li></ul>
            </div>
            <div>
              <div class="mb-3 text-sm font-medium">{t(lc, "footer.promise")}</div>
              <div class="flex gap-3 text-muted-foreground">
                <Icon name="shield" /><Icon name="undo" /><Icon name="truck" /><Icon name="zap" />
              </div>
            </div>
          </div>
          <div class="border-t border-border py-5 text-center text-xs text-muted-foreground">© 2026 gomu · built with gotsx</div>
        </footer>
      </body>
    </html>
  );
}
