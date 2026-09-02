# Benchmarks: gotsx vs. common full-stack frameworks

The same page — a product list of 50 items (title, brand, description, three tags, price) with a header, a
counter island and a footer, rendered from the same `data/products.json` — implemented in seven stacks and
measured the same way. Everything here is reproducible: `python3 bench/run.py` on any machine, or the
**Benchmark** GitHub Actions workflow (`workflow_dispatch`), whose numbers are the ones quoted below and in
the README. Numbers from a laptop are not published.

| stack | what it is | interactivity |
|---|---|---|
| **gotsx** | this framework, production mode (CSP nonce, security headers, gzip when asked) | a compiled island (signals) |
| **stdlib** | Go `net/http` + `html/template` | inline `<script>` |
| **gin** | Gin (release mode) + `html/template` | inline `<script>` |
| **templ** | Go `net/http` + [templ](https://templ.guide) generated code | inline `<script>` |
| **nextjs** | Next.js 16 App Router, `force-dynamic`, `next start` | React client component |
| **astro** | Astro 7, `output: "server"`, Node adapter | Preact island (`client:load`) |
| **hono** | Hono 4 JSX on Bun | inline `<script>` (no hydration story) |

## What is measured

- **req/s, p50, p99** — `bench/load` (a dependency-free Go load generator): 64 keep-alive connections, 2 s
  warm-up, then N seconds of continuous `GET /`, every response fully read and checked for 200.
- **peak RSS** — sampled every 200 ms during the load (process tree).
- **cold start** — milliseconds from process launch to the first 200.
- **build** — wall time of the production build; **artifact** — what you deploy (a binary, or `.next` /
  `dist` + `node_modules`).
- **HTML / JS** — what a headless Chromium downloads for the page: document bytes, and script bytes
  (uncompressed, plus gzip of the concatenated scripts). The counter is clicked to confirm it works.
- A second run pins the Go servers to **one core** (`GOMAXPROCS=1`); Node and Bun are single-threaded, so
  that table is the per-core comparison.

## Results

<!-- BENCH:ALL -->
**GitHub-hosted runner (`ubuntu-latest`: AMD EPYC 7763, 4 vCPU, 16 GB) · Go 1.26 · Node 22 · Bun 1.4 · 64 connections · 15 s per contender · 2026-09-02.** Raw data: [`results/`](results/).

All cores (Go servers use all 4 vCPUs; Node and Bun are single-threaded):

| framework | req/s | p50 | p99 | peak RSS | cold start | build | artifact | HTML | JS (gz) | note |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---|
| gotsx | 13,062 | 2.71 ms | 21.42 ms | 21.6 MB | 22 ms | 3.32 s | 10.2 MB | 14.3 KB | 72.4 KB (20.7) | production mode: CSP nonce, security headers, request logging off |
| stdlib | 3,837 | 9.69 ms | 80.02 ms | 20.2 MB | 11 ms | 0.94 s | 12.2 MB | 14.4 KB | 0.0 KB (0.0) | net/http + html/template |
| gin | 3,490 | 10.93 ms | 95.13 ms | 28.0 MB | 12 ms | 18.07 s | 30.5 MB | 14.4 KB | 0.0 KB (0.0) | gin release mode + html/template |
| templ | 13,498 | 3.98 ms | 16.77 ms | 18.9 MB | 11 ms | 0.75 s | 9.4 MB | 13.9 KB | 0.0 KB (0.0) | a-h/templ generated Go |
| nextjs | 253 | 251.72 ms | 297.16 ms | 377.4 MB | 696 ms | 5.85 s | 500.7 MB | 50.8 KB | 442.8 KB (127.6) | App Router, force-dynamic, next start |
| astro | 1,471 | 42.33 ms | 62.55 ms | 221.9 MB | 195 ms | 1.9 s | 250.4 MB | 23.7 KB | 15.8 KB (6.5) | output: server, node adapter, preact island |
| hono | 3,025 | 20.43 ms | 40.29 ms | 74.6 MB | 27 ms | 0.0 s | 1.4 MB | 14.4 KB | 0.0 KB (0.0) | hono/jsx on bun, SSR only |

Go servers pinned to one core (`GOMAXPROCS=1`) — the per-core comparison (JS column not re-measured):

| framework | req/s | p50 | p99 | peak RSS | cold start | build | artifact | HTML | JS (gz) | note |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---|
| gotsx | 7,267 | 8.81 ms | 18.44 ms | 19.3 MB | 24 ms | 0.56 s | 10.2 MB | 14.3 KB | 0.0 KB (0.0) | production mode: CSP nonce, security headers, request logging off |
| stdlib | 1,670 | 40.30 ms | 58.23 ms | 18.0 MB | 11 ms | 0.43 s | 12.2 MB | 14.4 KB | 0.0 KB (0.0) | net/http + html/template |
| gin | 1,612 | 42.90 ms | 65.49 ms | 23.5 MB | 11 ms | 0.83 s | 30.5 MB | 14.4 KB | 0.0 KB (0.0) | gin release mode + html/template |
| templ | 5,037 | 12.75 ms | 23.85 ms | 16.7 MB | 11 ms | 0.39 s | 9.4 MB | 13.9 KB | 0.0 KB (0.0) | a-h/templ generated Go |
| nextjs | 257 | 246.03 ms | 294.94 ms | 161.9 MB | 21 ms | 2.08 s | 500.9 MB | 50.8 KB | 0.0 KB (0.0) | App Router, force-dynamic, next start |
| astro | 1,437 | 42.77 ms | 83.16 ms | 215.5 MB | 206 ms | 1.91 s | 250.4 MB | 23.7 KB | 0.0 KB (0.0) | output: server, node adapter, preact island |
| hono | 3,086 | 20.26 ms | 39.90 ms | 74.6 MB | 28 ms | 0.0 s | 1.4 MB | 14.4 KB | 0.0 KB (0.0) | hono/jsx on bun, SSR only |

**Takeaways**

- **Throughput / latency**: gotsx and templ are the two compiled-Go stacks and land within 3% of each other at 4 cores (13.1k vs 13.5k req/s); on one core gotsx is the fastest of the seven (7.3k req/s, p50 8.8 ms) because its page is straight-line writes with ~140 allocations. Both are ~3.5× `html/template` (stdlib, Gin) and ~9× Astro; Next.js renders this page at 253 req/s with a p50 of 250 ms.
- **Memory**: the Go binaries peak at 19–28 MB under load; Hono/Bun at 75 MB, Astro at 220 MB, Next.js at 377 MB.
- **Cold start**: Go binaries answer in 11–24 ms from launch; Astro 200 ms, Next.js 700 ms (the `next start` server).
- **Artifact**: 9–12 MB static binaries versus 250 MB (Astro) / 500 MB (Next.js) of `node_modules` + build output.
- **What the browser downloads**: everyone ships ~14 KB of HTML except Next.js (51 KB, RSC payload). JS: Next.js 128 KB gz (React runtime + page chunks), gotsx 20.7 KB gz in this run — of which 12 KB is the morphing library for SPA navigation, now lazy-loaded on first navigation (initial load ≈ 10 KB gz: signals runtime + loader + the island); Astro 6.5 KB gz (Preact + island); the template stacks ship no framework JS but also have no hydration story.
- **Build**: `templ` and plain Go build in under a second; gotsx adds `hostgen` (a `go run`) and the dialect compiler; Gin's 18 s is a cold module download on the runner; Next.js builds in 6 s.
<!-- /BENCH:ALL -->

## Reading the numbers

**Where gotsx's speed comes from.** A page is compiled to a Go function that writes HTML into one buffer:
static markup is merged into single writes at compile time, dynamic text is escaped inline, and `list.map(...)`
becomes a `for` loop. There is no template parsing, no reflection and no virtual DOM; the whole 50-item page
is ~15 µs of CPU and ~140 allocations, so throughput is bounded by the network stack and the GC, like templ.
`html/template` (stdlib, gin) parses and escapes through reflection at request time, which is why it is 3× slower
on the same hardware. Next.js renders React on a single Node thread and ships the React runtime; Astro
renders once per request and ships only its islands.

**What the numbers do not say.** Throughput on a 50-item page is not the reason to pick a framework:

| | gotsx | templ + htmx | Next.js | Astro |
|---|---|---|---|---|
| Language | a static TSX dialect (compile error outside the subset) | Go + a template DSL | TypeScript, the whole language | TS + any UI framework |
| Interactivity | islands compiled to signals; SPA navigation with morphing | server round-trips (htmx) or hand-written JS | React everywhere, RSC, server actions | islands (React/Preact/Svelte/…) |
| Data access | direct Go calls (`host:*`), typed by reflection | direct Go | fetch / ORM in JS, RSC | fetch in the frontmatter |
| Streaming | `<Suspense>` resolved in goroutines | none built in | React streaming SSR | server islands |
| Ecosystem | none yet (component library in the dialect) | Go's | npm, the largest | npm + integrations |
| Deploy | one static binary (~10 MB) | one static binary | Node server or Vercel | Node server or adapters |
| Editor | LSP for the dialect + TS tooling | templ LSP | full TypeScript | Astro language server |

Pick **Next.js** when the team lives in React and needs the ecosystem; **Astro** for content sites with a
sprinkle of interactivity in any UI framework; **templ/htmx** when you want Go and are happy to express
interactivity as server round-trips; **gotsx** when you want Go's deploy story and performance *and* React's
component model in one source file — accepting that the language is a subset and the ecosystem is young.

## Running it yourself

```bash
python3 bench/run.py                 # all contenders, 10 s each; needs Go, Node 22 + npm, Bun, Python
python3 bench/run.py --only gotsx,templ --duration 5
python3 bench/run.py --single-core   # GOMAXPROCS=1 for the Go servers
```

The Node contenders install their pinned dependencies with `npm ci` / `bun install` in `bench/nextjs`,
`bench/astro` and `bench/hono` (the workflow does this). Playwright is optional (`pip install playwright &&
playwright install chromium`); without it the JS/HTML columns are skipped.

`go test -bench . ./bench/gotsx` runs the in-process benchmarks (`BenchmarkPage` for the full handler chain,
`BenchmarkRender` for the renderer alone) — the numbers behind "µs per page".
