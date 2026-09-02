# gomu — a full-stack e-commerce demo (gotsx)

A complete store written in the gotsx dialect with a Go host, used to push the framework toward "enterprise-grade".
English is the default locale; Chinese is available under `/zh/...`.

```bash
make dev-shop          # from the repo root → http://localhost:3000
SHOP_NOLAG=1 …         # disable the simulated backend latency
```

## What it covers

| Scenario | How |
|---|---|
| Home / 8 categories / search / sort / pagination | file routing + SSR; 192 products; one host method `catalog.listCards(cat, q, sort, page)` does filtering, sorting and paging |
| Product page with variants and reviews | server-rendered; reviews are generated deterministically per product id in Go |
| Flash sale with countdown | the server passes the remaining milliseconds, the `Countdown` island ticks with `setInterval` |
| Add to cart | `AddToCart` island: variant selection check (client) + stock check (**server wins**) + quantity stepper |
| Live cart badge | cross-island event `emit("cart:changed")` → the header `CartBadge` island listens with `on(...)` |
| Cart edits | every amount (subtotal / shipping / free-shipping gap / total) is **computed in Go**; the island only displays and sends commands |
| Wishlist | `WishButton` island; persisted per session on the server, so a reload SSRs the filled heart |
| Checkout | form validation in a Go action (name / international phone / address), field-level errors in the island; success navigates (SPA) to the order page |
| Orders list / detail | session-scoped: one order list per `sid` |
| Sessions | the `Before` hook sets the `sid` cookie, which reaches pages through `PageProps.Cookies`; cart, wishlist and orders are keyed by it |
| i18n | `Options.I18n` in URL-prefix mode: `/` is English (default), `/zh/...` is Chinese; server pages use `t / tv / plural`, islands receive the locale as a prop and use the same builtins (the active catalog is injected client-side) |
| SEO / PWA | per-page title / description / canonical / OpenGraph, JSON-LD (`WebSite` + `Product`), `sitemap.xml`, `robots.txt`, manifest, server-generated SVG product images |

## Layout

- `app/pages/` — routes: `index`, `c/[cat]`, `p/[id]`, `search`, `cart`, `checkout`, `orders/index`, `orders/[id]`
- `app/components/` — `Layout` (document, header, nav, footer), `Listing` (grid + sort + pager)
- `app/islands/` — `AddToCart`, `CartBadge`, `CartPage`, `CheckoutForm`, `Countdown`, `Feed`, `LocaleSwitch`, `ProductGallery`, `Related`, `WishButton`
- `app/ui/` — shared pieces compiled to both sides (`CardShell`, `Stars`, `Icon`, …)
- `host/host.go` — host modules `catalog / cart / wish / orders / intl / site` (in memory behind mutexes; a database would be an implementation detail the dialect never sees)
- `main.go` — `gotsx.Serve`: routes from `gen/`, the i18n catalog (`messages`, keys grouped by area), and the write actions

## Architecture notes

- **Writes live in Go**: add-to-cart, quantity changes, checkout and wishlist are plain `http.HandlerFunc`s (`Actions` in `main.go`); the dialect reads, it never writes. Stock, prices and order consistency are guaranteed by Go — the client cannot alter them.
- **Money is only computed on the server**: islands never do price arithmetic, they echo the `*Fmt` strings Go returns.
- **Host modules are the domain layer**: swapping the in-memory stores for Redis / a database touches only `host/`.
- **One full page load**: the whole site navigates as a SPA (fetch HTML → morph); islands survive by DOM identity, so the header cart badge keeps its state across pages.
