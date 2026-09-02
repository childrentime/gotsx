# gotsx

**English** · [中文](README.zh.md)

**A full-stack framework that borrows React + TSX ideas and compiles to native Go.** One TSX source, two compilers: server components become Go functions, client islands become signals. No virtual DOM, no JS engine, no Node, no npm, no esbuild — **the toolchain is just Go**.

> **Status: v0.4 · consumer-facing apps (with i18n) · still maturing.** A working end-to-end proof of concept — compiler + two backends + runtime + dev loop + four real apps (including a full e-commerce store) + a test suite. It shows that "TSX can compile to native Go and carry real apps," but the language has a known long tail, no independent security audit, and no LSP yet. See the [roadmap](#roadmap).

```tsx
// app/pages/index.server.tsx — compiles to a Go function, never ships to the browser
import type { PageProps } from "gotsx";
import { models } from "host:data";              // Go-backed; types are reflected, calls are zero-marshalling
import Layout from "../components/Layout.server";

export default function Home({ query }: PageProps) {
  const list = models.search(query.q ?? "");     // synchronous: concurrency comes from goroutines, no async
  return <Layout title="Products">
    <ul>{list.map((m) => <li>{m.title} · ${m.price}</li>)}</ul>
  </Layout>;
}
```

```tsx
// app/islands/Counter.client.tsx — one source: SSR on the server, interactivity on the client
import { useState, useEffect } from "gotsx";
export default function Counter({ start }: { start: number }) {
  const [n, setN] = useState(start);
  const double = n * 2;                           // depends on n → auto-compiled to a memo, no useMemo needed
  useEffect(() => { console.log(n); }, []);       // empty deps = run once on mount
  return <button onClick={() => setN(n + 1)}>{n} ×2 = {double}{n > 4 && <b> 🔥</b>}</button>;
}
```

## Why

**The core insight: SSR is a single, synchronous, one-pass evaluation.** No re-renders, no effects, setters are never called. So the server doesn't need a React runtime — only the component's "render slice," whose semantics are small enough to compile straight to Go. The result: a list page renders on the server in **~30µs** (a goja + React + MUI version takes ~50ms), the client runtime is **~6KB**, `go build` produces a single binary, and `delve` / `pprof` / `go test` all just work.

## Quick start

```bash
git clone https://github.com/childrentime/gotsx && cd gotsx
./scripts/get-tailwind.sh          # download the Tailwind standalone binary into .tools/ (no Node)
make dev-shop                      # run the Temu-style e-commerce demo → http://localhost:3000
# or:
make dev-site                      # this framework's docs site (dialect component library + searchable syntax reference)
make dev-example                   # branch-integration demo
```

Build / test:

```bash
make gen      # compile every demo's dialect → gen/ (gen is gitignored, so run this first)
make build    # gen + go build ./...
make test     # gen + go test ./...
```

> **Note:** `*/gen` is gitignored and does not exist in a clean checkout. Any command that compiles an app must run `make gen` first. CI is ordered this way.

## One source, two compilers

```
.tsx ──▶ frontend: hand-written TSX-subset parser + type checker + dialect fence + server/client boundary
          ├─▶ Go backend:  components → Go functions, JSX → gotsx.El/Text/If/Nodes, hooks → one-pass semantics,
          │                host:* → direct Go calls; //line directives point errors back to .tsx
          └─▶ JS backend:  components → functions, useState → signal, signal-dependent const → memo,
                           JSX → el/t/text/cond/each; resumable hydration, no diffing
Go side: the generated *_gen.go compiles together with your main.go / host package into one binary
```

| Directory | Contents |
|---|---|
| `compiler/` | `lexer` / `parser` / `check` / `gogen` / `jsgen` / `compile` |
| `runtime/` | node model, hydration markers, dialect builtins, HTTP (CSP/CSRF/gzip/caching/health/graceful shutdown), request cookies, `Before` hook, host-type reflection, i18n |
| `client/` | signals, `el/t/text/cond/each`, resumable hydration, island loader, SPA nav, progress bar, prefetch, cross-island `emit`/`on`, i18n |
| `cmd/gotsx/` | `gotsx build` / `gotsx dev` |
| `example/` `site/` `shop/` `admin/` | real apps, also used as integration tests |

## The language: a static subset that borrows TSX syntax

It's not an implementation of TypeScript, but **a static language that borrows TSX syntax** (a cousin of AssemblyScript): the type system is limited to what Go can represent. If every expression has a static type inside the allowed set, it compiles; otherwise you get a compile error with `file:line:col`.

- **Has**: function components / props / destructuring + defaults; `string` (rune) / `number` (float64) / `boolean` / arrays / `Record` / known-shape objects; `map filter find some every forEach includes indexOf join slice concat sort reduce reverse flat at`; string methods (incl. `padStart/padEnd`); `Object.keys/values`; `Math.*`; template strings; `&& || ?? ternary`; `=== !==`; `if / for-of / try`; `useState/useMemo/useEffect(+[])`; JSX; module-level `const`; `host:*` (server); `fetch`/DOM/`await` (client).
- **Does not have (a compile error, not silent)**: `class`/`this`/prototypes; member access on `any`; `==`; `while`/`switch`; custom generics; `push/splice`; regex; `Date` (server goes through a host module); passing `children` to an island.

The full, searchable syntax table lives in the `site` docs at `/docs/language`.

## Features

- **Host modules**: `import { models } from "host:data"` is backed by Go; after compilation it's a direct call with zero marshalling, and types are reflected into `host.d.ts`. A `(T, error)` method turns the error into a panic recovered by the request layer (an `ErrNotFound`-wrapped one becomes a 404).
- **Islands + SPA navigation**: pages ship zero JS; islands load on demand; navigation fetches HTML and morphs it (idiomorph), islands survive by DOM identity so state isn't lost; a top progress bar and hover prefetch make it feel instant.
- **Resumable hydration**: the server only marks the reactive text/conditions/lists; the client claims nodes in the same order the same compiler produced, reusing existing DOM without diffing.
- **Sessions / middleware**: request cookies flow into `PageProps.Cookies`; the `Options.Before` hook can set cookies or do auth; `Options.Middleware` is a standard middleware chain.
- **Source maps**: `go build` errors and panic stacks point back to `.tsx` line numbers.
- **Tailwind**: `class` is just a string; the standalone CLI runs in-process to scan `.tsx` and generate CSS — no Node.

## Production-ready (v0.2)

- **HTTP hardening**: panic recovery, security headers, **CSP + per-response nonce**, gzip, **CSRF same-origin check**, content-hash immutable caching, request IDs, access logging, graceful shutdown, `/healthz` `/readyz`, custom 404/500, application-level middleware (auth).
- **Single-binary deploy**: `go:embed` bundles the client and static assets, so `go build` produces one self-contained binary you can `scp` and run.
- **Client resilience**: reactive ownership/cleanup, range-based block rebuilds (nested tables don't leave stale rows on refresh), island error boundaries (one failing island doesn't blank the page).
- See [`SECURITY.md`](SECURITY.md). `-dev=false` is production mode by default; `gotsx dev` turns on dev mode.

## Consumer-facing capabilities (v0.3)

- **SEO**: per-page title/description/**canonical**/**OpenGraph**/Twitter, **JSON-LD Product** on product pages (price/stock/rating) + a WebSite SearchAction on the home page, **sitemap.xml** + **robots.txt**. New `jsonLd()` builtin.
- **Real images**: server-generated product "studio shots" as SVG (`/img/p/{id}`); the `Img` component does lazy loading + fixed width/height (CLS prevention) + alt text; `og:image` uses the product image.
- **Performance**: hover/touch **prefetch**, instant navigation on click (Core Web Vitals).
- **Observability**: client error + pageview telemetry (`sendBeacon` → `/_gotsx/client-log`, received by `Options.OnClientEvent`).
- **PWA**: manifest + icon + theme-color (installable to the home screen).

## Internationalization (v0.4, optional)

Turning on `Options.I18n` gives you complete i18n: `t()` / `tv()` (interpolation) / `plural()` / `fmtNum` / `fmtCur` / `fmtDate` behaving identically on the server and the client; two locale-resolution modes — **URL prefix** (`/en/`, SEO-friendly) or **cookie / Accept-Language**; automatic **hreflang**; the loader **auto-localizes internal links** (in-app navigation stays in the same language); and `PageProps.Locale`. The `shop` demo ships zh/en with a language switcher, and islands translate too.

## Example apps

| App | What it is |
|---|---|
| `admin` | **Back office**: login / protected routes / user table (search, sort, pagination) / CRUD + server-side validation / modals / toasts / roles. Built to push the framework toward enterprise use |
| `shop` | **Temu-style full-stack e-commerce**: 8 categories / 192 products / search-sort-paginate / flash-sale countdown / variants + stock / cart (totals computed server-side) / wishlist / checkout validation / orders / session isolation / simulated latency + skeletons |
| `site` | This framework's docs site: a component library written in the dialect + a searchable syntax reference (highlighting powered by a Go tokenizer) |
| `example` | Branch-integration management demo |

Live docs site (a static export of the `site` app): **https://childrentime.github.io/gotsx/**

## Testing

```bash
make test        # everything (incl. building all four apps)
make test-fast   # compiler + runtime unit tests only
```

- `compiler/codegen_test.go` — dialect snippets → assert the generated Go / JS structure.
- `compiler/fence_test.go` — every fence violation → an error with a position; valid programs don't error.
- `compiler/apps_test.go` — compile all four real apps and `go build` + `go vet`.
- `runtime/{builtins,render,security,server,i18n}_test.go` — builtin correctness, hydration markers, XSS escaping, HTTP middleware, i18n.

## Security

See [`SECURITY.md`](SECURITY.md). In short: the dialect has no "inject raw HTML" escape hatch, and text/attributes/island props all go through `html.EscapeString` (tested); `host:*` is server-only, the client can't touch Go; writes happen in Go. CSRF / auth / business validation are the app's responsibility. **Not independently security-audited.**

## Roadmap

Still missing before others should depend on it in production:

- [ ] **Editor LSP** (dialect-specific rule errors currently only surface at build time)
- [ ] Language long tail: `while`/`switch`, `push/splice`, custom generics, more builtins
- [ ] Keyed `each` diffing (lists currently rebuild wholesale — correct add/remove, but the DOM is rebuilt)
- [ ] Streaming SSR, nested-layout conventions
- [ ] `Record`'s "missing key vs empty string" semantics (a Go map can't distinguish them)
- [ ] Incremental compilation, a frozen stability contract, an independent security audit
- [ ] Cross-platform polish (Windows)

Done: both backends, source maps, a test suite, Tailwind, sessions/middleware, cross-island events, **production HTTP hardening (CSP/CSRF/gzip/caching/health/graceful shutdown)**, **single-binary deploy**, **reactive ownership and cleanup**, **island error boundaries**, **SEO / images / prefetch / telemetry / PWA**, **optional i18n**, and **four real apps (including a back office and an e-commerce store)**.

## License

[MIT](LICENSE). gotsx borrows ideas from React/Solid/Svelte and Astro's island model; the runtime is its own.
