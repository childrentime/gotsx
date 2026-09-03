import type { LayoutProps } from "gotsx";
import { seed } from "gotsx";
import { url } from "host:site";
import { view } from "host:cart";
import { cart } from "../stores/cart";

// The document shell for every page. <head> is driven by the page's `export function meta(props)` (LayoutProps.meta):
// title, description, canonical, og:image and robots; components/Layout.server.tsx renders the store chrome in the body.
export default function Root({ path, locale, meta, cookies, children }: LayoutProps) {
  const lc = locale !== "" ? locale : "en";
  seed(cart, view(cookies.sid ?? ""));   // every island of this request (badge, cart page, checkout) renders this cart and hydrates from it
  const title = meta.title !== "" ? `${meta.title} · gomu` : "gomu";
  const desc = meta.description !== "" ? meta.description : t(lc, "meta.description");
  const canonical = meta.canonical !== "" ? meta.canonical : url(lpath(lc, path));
  const ogType = path.startsWith("/p/") ? "product" : "website";
  return (
    <html lang={lc}>
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1" />
        <meta name="theme-color" content="#ffffff" media="(prefers-color-scheme: light)" />
        <meta name="theme-color" content="#09090b" media="(prefers-color-scheme: dark)" />
        <link rel="manifest" href="/manifest.webmanifest" />
        <link rel="icon" href="/icon.svg" />
        <link rel="apple-touch-icon" href="/icon.svg" />
        <title>{title}</title>
        <meta name="description" content={desc} />
        {meta.noIndex && <meta name="robots" content="noindex" />}
        <link rel="canonical" href={canonical} />
        <meta property="og:site_name" content="gomu" />
        <meta property="og:title" content={meta.title !== "" ? meta.title : "gomu"} />
        <meta property="og:description" content={desc} />
        <meta property="og:type" content={ogType} />
        <meta property="og:url" content={canonical} />
        {meta.image !== "" && <meta property="og:image" content={meta.image} />}
        <meta name="twitter:card" content="summary_large_image" />
        <link rel="stylesheet" href="/public/tailwind.css" />
      </head>
      <body class="min-h-screen">
        <div id="gotsx-bar"></div>
        {children}
      </body>
    </html>
  );
}
