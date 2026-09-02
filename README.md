# gotsx

[![CI](https://github.com/childrentime/gotsx/actions/workflows/ci.yml/badge.svg)](https://github.com/childrentime/gotsx/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/childrentime/gotsx.svg)](https://pkg.go.dev/github.com/childrentime/gotsx/runtime)
[![Go Report Card](https://goreportcard.com/badge/github.com/childrentime/gotsx)](https://goreportcard.com/report/github.com/childrentime/gotsx)
[![Docs](https://img.shields.io/badge/docs-childrentime.github.io%2Fgotsx-111?logo=readthedocs&logoColor=white)](https://childrentime.github.io/gotsx/)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

**English** · [中文](README.zh.md)

**A full-stack framework that borrows React + TSX ideas and compiles to native Go.** One TSX source, two compilers: server components become Go functions, client islands become signals. No virtual DOM, no JS engine, no Node, no npm, no esbuild — **the toolchain is just Go**.

> **Status: v0.6 · the roadmap is done, the surface is documented in [`STABILITY.md`](STABILITY.md) · still pre-1.0.** A working end-to-end framework — compiler + two backends + runtime + streaming SSR + dev loop with live reload + scaffolding + editor server (diagnostics, hover, go-to-definition) + a shared design system + four real apps (including a full e-commerce store) + a test suite. The language is a deliberate static subset (see below); what is still missing before 1.0 is an independent security audit. See the [roadmap](#roadmap).

```tsx
// app/pages/index.server.tsx — compiles to a Go function, never ships to the browser
import type { PageProps } from "gotsx";
import { models } from "host:data";              // Go-backed; types are reflected, calls are zero-marshalling
import Layout from "../components/Layout.server";

export default function Home({ query }: PageProps) {
  if (query.legacy !== "") return redirect("/");  // page-level control flow: redirect() / notFound()
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

<p align="center">
  <img src="docs/screenshots/shop.png" alt="The shop demo: a Temu-style store rendered by Go" width="820">
  <br><sub>The <code>shop</code> demo — 192 products, cart, checkout, i18n — every page is a Go function compiled from TSX; the interactive parts are islands.</sub>
</p>

<table>
  <tr>
    <td><img src="docs/screenshots/admin.png" alt="The admin demo: back office with a data table" width="400"></td>
    <td><img src="docs/screenshots/site-dark.png" alt="The docs site in dark mode" width="400"></td>
  </tr>
  <tr>
    <td align="center"><sub><code>admin</code>: auth, CRUD table, modals, toasts</sub></td>
    <td align="center"><sub><code>site</code>: the docs, built with gotsx, dark mode via design tokens</sub></td>
  </tr>
</table>

## Why

**The core insight: SSR is a single, synchronous, one-pass evaluation.** No re-renders, no effects, setters are never called. So the server doesn't need a React runtime — only the component's "render slice," whose semantics are small enough to compile straight to Go. The result: a list page renders on the server in **~30µs** (a goja + React + MUI version takes ~50ms), the client runtime is **~6KB gzipped** (plus a ~4KB loader), `go build` produces a single binary, and `delve` / `pprof` / `go test` all just work.

## Quick start

```bash
go install github.com/childrentime/gotsx/cmd/gotsx@latest
gotsx new hello && cd hello   # scaffold: own Go module, host module, a page, a keyed-list island, an action, tsconfig
gotsx dev                     # http://localhost:3000 — edit app/**/*.tsx, the browser reloads by itself
go build -o hello . && ./hello -addr :8080        # one self-contained binary for production
```

To hack on the framework itself:

```bash
git clone https://github.com/childrentime/gotsx && cd gotsx
go run ./cmd/gotsx tailwind   # once: download the Tailwind standalone binary into .tools/ (no Node)
make dev-shop                 # the Temu-style e-commerce demo → http://localhost:3000
make dev-site                 # this framework's docs site  ·  make dev-example  ·  make dev-admin
make test                     # gen + go test ./... (also builds all four demo apps)
```

> `*/gen` is gitignored, so any command that compiles a demo app must run `make gen` first. CI is ordered this way.

## The CLI

| Command | What it does |
|---|---|
| `gotsx new <dir>` | Scaffold an app in its own module (`--module`, `--tailwind`, `--replace <checkout>` for local framework development) |
| `gotsx build [dir]` | hostgen → Tailwind → dialect → `gen/` (Go + client JS + embedded assets + editor typings) |
| `gotsx dev [dir]` | build + `go build` + run, watch `app/ host/ public/ main.go`, restart on change; the browser **live-reloads** (a compile error keeps the old build serving); rebuilds are **incremental** (hostgen only when `host/` changed, Tailwind in parallel: a TSX edit recompiles in ~15ms) |
| `gotsx check [dir]` | Type-check only; diagnostics as `file:line:col: message` (`--json` for tools); exit 1 on errors |
| `gotsx lsp` | Language Server over stdio: diagnostics as you type, **hover** (types, prop signatures, host method signatures) and **go-to-definition** (into the Go source of a host method) — see [`editors/`](editors) for VS Code, Neovim, Helix, Zed |
| `gotsx tailwind` | Download the Tailwind v4 standalone CLI for this OS/arch (macOS, Linux, Windows) |

`gotsx.json` is optional: the app's import path is inferred from `go.mod`, and `host/` is picked up automatically.

## One source, two compilers

```
.tsx ──▶ frontend: hand-written TSX-subset parser + type checker + dialect fence + server/client boundary
          ├─▶ Go backend:  components → Go functions, JSX → gotsx.El/Text/If/Nodes, hooks → one-pass semantics,
          │                host:* → direct Go calls; //line directives point errors back to .tsx
          └─▶ JS backend:  components → functions, useState → signal, signal-dependent const → memo,
                           JSX → el/t/text/cond/each (keyed when you write key={…}); resumable hydration, no diffing
Go side: the generated *_gen.go compiles together with your main.go / host package into one binary
```

| Directory | Contents |
|---|---|
| `compiler/` | `lexer` / `parser` / `check` / `gogen` / `jsgen` / `compile` (+ `Analyze` for check/LSP) |
| `runtime/` | node model, hydration markers, dialect builtins, HTTP (CSP/CSRF/gzip/caching/health/graceful shutdown), request cookies, `Before` hook, host-type reflection, i18n, redirect/notFound |
| `client/` | signals, `el/t/text/cond/each` (keyed reuse), resumable hydration, island loader, SPA nav, progress bar, prefetch, cross-island `emit`/`on`, i18n, dev live reload |
| `cmd/gotsx/` | `new` / `build` / `dev` / `check` / `lsp` / `tailwind` |
| `editors/` | LSP setup for VS Code (extension source), Neovim, Helix, Zed |
| `design/` | **gotsx UI**: the shared design system (shadcn-style neutral tokens + component classes) — a Tailwind v4 layer and a plain-CSS twin, used by every demo and by `gotsx new` |
| `example/` `site/` `shop/` `admin/` | real apps, also used as integration tests (`example` has a language kitchen-sink page, a streaming `Suspense` panel and the `_layout` / `_404` / `_error` conventions) |

## The language: a static subset that borrows TSX syntax

It's not an implementation of TypeScript, but **a static language that borrows TSX syntax** (a cousin of AssemblyScript): the type system is limited to what Go can represent. If every expression has a static type inside the allowed set, it compiles; otherwise you get a compile error with `file:line:col` — at build time, from `gotsx check`, or live in your editor.

- **Has**: function components / props / destructuring + defaults; `string` (rune) / `number` (float64) / `boolean` / arrays / `Record` / known-shape objects / `interface … extends`; `if` / `for-of` / `for (;;)` / `while` / `switch` / `break` / `continue` / `try`; `++ -- += -= *= /= %=`; `&& || ?? ternary`; `=== !==`; template strings; **regex literals** (RE2 subset, checked at compile time: `re.test`, `s.match/replace/replaceAll/split/search`); array `map filter find findIndex some every forEach includes indexOf lastIndexOf join slice concat sort reduce reverse flat at` and in-place `push pop shift unshift splice`; string methods (incl. `padStart/padEnd/trimStart/trimEnd/localeCompare/at`); `Object.keys/values/hasOwn`, `delete m[k]`; `Math.*`; **`Date.now/Date.parse/isoDate`**; `useState/useMemo/useEffect(+[])`; JSX with `key`; **`<Suspense fallback>`** (server); module-level `const`; `host:*` (server); `redirect()/notFound()` (server pages); `fetch`/DOM/`await` (client).
- **Does not have (a compile error, not silent)**: `class`/`this`/prototypes/`new`; member access on `any`; `==`; `do-while`, `for-in`; custom generics; regex lookaround/backreferences; passing `children` to an island; mutating a `useState` array in place (use `setXs([...xs, x])`).
- **Semantic conventions** (identical on both sides): an optional primitive is its zero value (`""`/`0`/`false`), so `??` and `||` are the same operator; an optional object (e.g. a `find()` miss) is falsy and `=== undefined`; a `Record` read of an absent key is the zero value and `Object.hasOwn` tests presence; objects are values on the server. The full list is the **Stable** tier in [`STABILITY.md`](STABILITY.md).

The full, searchable syntax table lives in the `site` docs at `/docs/language`.

## Features

- **Host modules**: `import { models } from "host:data"` is backed by Go; after compilation it's a direct call with zero marshalling, and types are reflected into `host.d.ts`. A `(T, error)` method turns the error into a panic recovered by the request layer (an `ErrNotFound`-wrapped one becomes a 404).
- **File routing**: `pages/p/[id].server.tsx` → `/p/{id}`, `pages/docs/[...slug].server.tsx` → catch-all; more specific routes win. **Nested layouts**: `pages/**/_layout.server.tsx` (`LayoutProps` = `PageProps` + `children`) wrap the pages below them; `_404` / `_error` become `gen.NotFound` / `gen.ErrorPage`. `redirect(url, status?)` and `notFound()` abort a render.
- **Streaming SSR**: `<Suspense fallback={…}>` ships the fallback with the shell and renders its children **in their own goroutine** after the shell is flushed; several boundaries resolve concurrently and stream in as they finish (out of order, nested, with error isolation). No async in the language: the boundary is where the slow host call goes.
- **Islands + SPA navigation**: pages ship zero JS; islands load on demand; navigation fetches HTML and morphs it (idiomorph), islands survive by DOM identity so state isn't lost; a top progress bar and hover prefetch make it feel instant.
- **Keyed lists**: `xs.map((x) => <li key={x.id}>…</li>)` reuses, moves and disposes DOM per key — inputs, focus and per-row effects survive reorders; lists without `key` rebuild as before.
- **Resumable hydration**: the server only marks the reactive text/conditions/lists; the client claims nodes in the same order the same compiler produced, reusing existing DOM without diffing.
- **Sessions / middleware**: request cookies flow into `PageProps.Cookies`; the `Options.Before` hook can set cookies or do auth; `Options.Middleware` is a standard middleware chain.
- **Source maps**: `go build` errors and panic stacks point back to `.tsx` line numbers.
- **Editor support**: `tsconfig.json` + generated `app/.gen/gotsx.d.ts` / `host.d.ts` make TypeScript tooling happy; `gotsx lsp` adds the dialect's diagnostics, hover and go-to-definition (a host method jumps into its Go source).
- **Design system**: [`design/`](design) ships gotsx UI — shadcn-style neutral tokens, dark mode, and component classes — as a Tailwind layer and a plain-CSS twin; every demo and every scaffolded app uses it, so a new app looks finished on day one.
- **Tailwind**: `class` is just a string; the standalone CLI runs in-process to scan `.tsx` and generate CSS — no Node.

## Production-ready

- **HTTP hardening**: panic recovery, security headers, **CSP + per-response nonce**, gzip, **CSRF same-origin check**, content-hash immutable caching, request IDs, access logging, graceful shutdown, `/healthz` `/readyz`, custom 404/500, application-level middleware (auth).
- **Single-binary deploy**: `go:embed` bundles the client and static assets, so `go build` produces one self-contained binary you can `scp` and run.
- **Client resilience**: reactive ownership/cleanup, range-based block rebuilds, island error boundaries (one failing island doesn't blank the page).
- **Consumer-facing**: per-page SEO (canonical / OpenGraph / Twitter / JSON-LD / sitemap / robots), lazy images, hover prefetch, client telemetry, PWA manifest.
- **Internationalization** (optional): `t()` / `tv()` / `plural()` / `fmtNum` / `fmtCur` / `fmtDate` identical on server and client; URL-prefix or cookie/Accept-Language locale resolution; automatic `hreflang`; localized internal links.
- See [`SECURITY.md`](SECURITY.md). `-dev=false` is production mode by default; `gotsx dev` turns on dev mode.

## Example apps

| App | What it is |
|---|---|
| `admin` | **Back office**: login / protected routes / user table (search, sort, pagination) / CRUD + server-side validation / modals / toasts / roles |
| `shop` | **Temu-style full-stack e-commerce**: 8 categories / 192 products / search-sort-paginate / flash-sale countdown / variants + stock / cart / wishlist / checkout validation / orders / sessions / zh + en |
| `site` | This framework's docs site: a component library written in the dialect + a searchable syntax reference |
| `example` | Small demo + the **language kitchen sink** (`/kitchen`: loops, switch, mutation, redirect, catch-all `/docs/a/b`, keyed list) |

Live docs site (a static export of the `site` app): **https://childrentime.github.io/gotsx/**

## Testing

```bash
make test        # everything (incl. building all four apps and the gotsx new → build → check end-to-end)
make test-fast   # compiler + runtime + CLI unit tests only
make check       # gotsx check on every demo app
```

- `compiler/codegen_test.go`, `compiler/lang_test.go` — dialect snippets → assert the generated Go / JS structure (incl. marker parity between the two backends).
- `compiler/fence_test.go` — every fence violation → an error with a position; valid programs don't error.
- `compiler/apps_test.go` — compile all four real apps and `go build` + `go vet`.
- `runtime/*_test.go` — builtin correctness, hydration markers, XSS escaping, HTTP middleware, routing, redirect/notFound, dev live reload, i18n.
- `cmd/gotsx/cli_test.go` — `gotsx new` → `build` → `go build` → `check` in a temp module.

## Security

See [`SECURITY.md`](SECURITY.md). In short: the dialect has no "inject raw HTML" escape hatch, and text/attributes/island props all go through `html.EscapeString` (tested); `host:*` is server-only, the client can't touch Go; writes happen in Go. CSRF is checked by default; auth and business validation are the app's responsibility. **Not independently security-audited.**

## Roadmap

- [ ] An **independent security audit** (the self-review, threat model and `govulncheck` results are in [`SECURITY.md`](SECURITY.md))
- [ ] **1.0**: freeze the Stable tier of [`STABILITY.md`](STABILITY.md), promote `Suspense` and the LSP out of Experimental

Done: both backends, source maps, a test suite, Tailwind, sessions/middleware, cross-island events, production HTTP hardening, single-binary deploy, reactive ownership and cleanup, island error boundaries, SEO / images / prefetch / telemetry / PWA, optional i18n, four real apps, installable module + `gotsx new`, the language long tail (loops, switch, in-place array methods, `interface extends`, regex, `Date`), keyed list diffing, redirect / notFound / catch-all routes, `gotsx check` + LSP (diagnostics, hover, definition) + editor typings, dev live reload with incremental rebuilds, Windows-friendly toolchain, **streaming SSR with `Suspense`**, **nested layouts / `_404` / `_error`**, **`Record` absence semantics**, **a stability contract**, **a shared design system across all demos**.

## License

[MIT](LICENSE). gotsx borrows ideas from React/Solid/Svelte and Astro's island model; the runtime is its own.
