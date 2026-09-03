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
**GitHub-hosted runner (`ubuntu-latest`): AMD EPYC 7763 64-Core Processor (4 vCPU) · go1.26.7 linux/amd64 · node v22.23.2 · bun 1.4.0 · 64 connections · 15 s per contender · 2026-09-03.** Raw data: [`results/`](results/). Runner hardware varies between runs (AMD EPYC or Intel Xeon); ratios are stable, absolute numbers move ±30%.

All cores (Go servers use all 4 vCPUs; Node and Bun are single-threaded):

| framework | req/s | p50 | p99 | peak RSS | cold start | build | artifact | HTML | JS (gz) | note |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---|
| gotsx | 12,712 | 2.79 ms | 22.41 ms | 20.4 MB | 22 ms | 3.3 s | 10.3 MB | 14.3 KB | 27.8 KB (10.6) | production mode: CSP nonce, security headers, request logging off |
| stdlib | 3,656 | 9.83 ms | 85.14 ms | 20.3 MB | 11 ms | 0.91 s | 12.2 MB | 14.4 KB | 0.0 KB (0.0) | net/http + html/template |
| gin | 3,372 | 11.13 ms | 104.47 ms | 26.0 MB | 12 ms | 18.35 s | 30.5 MB | 14.4 KB | 0.0 KB (0.0) | gin release mode + html/template |
| templ | 13,013 | 4.10 ms | 17.71 ms | 17.0 MB | 11 ms | 0.78 s | 9.4 MB | 13.9 KB | 0.0 KB (0.0) | a-h/templ generated Go |
| nextjs | 242 | 262.20 ms | 306.16 ms | 386.5 MB | 718 ms | 6.0 s | 500.7 MB | 50.8 KB | 442.8 KB (127.4) | App Router, force-dynamic, next start |
| astro | 1,441 | 43.19 ms | 62.60 ms | 227.5 MB | 206 ms | 1.96 s | 250.4 MB | 23.7 KB | 15.8 KB (6.5) | output: server, node adapter, preact island |
| hono | 2,915 | 21.19 ms | 41.77 ms | 77.9 MB | 40 ms | 0.0 s | 1.4 MB | 14.4 KB | 0.0 KB (0.0) | hono/jsx on bun, SSR only |

Go servers pinned to one core (`GOMAXPROCS=1`) — the per-core comparison (JS column not re-measured):

| framework | req/s | p50 | p99 | peak RSS | cold start | build | artifact | HTML | JS (gz) | note |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---|
| gotsx | 6,939 | 9.23 ms | 19.15 ms | 17.6 MB | 21 ms | 0.55 s | 10.3 MB | 14.3 KB | 0.0 KB (0.0) | production mode: CSP nonce, security headers, request logging off |
| stdlib | 1,635 | 41.23 ms | 59.15 ms | 18.0 MB | 11 ms | 0.42 s | 12.2 MB | 14.4 KB | 0.0 KB (0.0) | net/http + html/template |
| gin | 1,583 | 43.87 ms | 66.51 ms | 23.5 MB | 11 ms | 0.81 s | 30.5 MB | 14.4 KB | 0.0 KB (0.0) | gin release mode + html/template |
| templ | 4,899 | 13.10 ms | 24.32 ms | 14.7 MB | 11 ms | 0.39 s | 9.4 MB | 13.9 KB | 0.0 KB (0.0) | a-h/templ generated Go |
| nextjs | 262 | 244.32 ms | 268.57 ms | 157.0 MB | 16 ms | 2.16 s | 500.9 MB | 50.8 KB | 0.0 KB (0.0) | App Router, force-dynamic, next start |
| astro | 1,454 | 42.51 ms | 64.64 ms | 224.2 MB | 206 ms | 1.87 s | 250.4 MB | 23.7 KB | 0.0 KB (0.0) | output: server, node adapter, preact island |
| hono | 2,938 | 21.33 ms | 42.14 ms | 74.7 MB | 28 ms | 0.0 s | 1.4 MB | 14.4 KB | 0.0 KB (0.0) | hono/jsx on bun, SSR only |

**Takeaways**

- **Throughput / latency**: gotsx and templ, the two compiled-Go stacks, are within 2% of each other on 4 cores (12,712 vs 13,013 req/s); on one core gotsx is the fastest of the seven (6,939 req/s, p50 9.2 ms) because a page is straight-line writes with ~140 allocations. gotsx is 3.5× `html/template` (stdlib; Gin is the same template engine), 8.8× Astro and 52.6× Next.js, whose p50 is 262 ms.
- **Memory**: the Go binaries peak at 17–26 MB under load; Hono/Bun 78 MB, Astro 228 MB, Next.js 386 MB.
- **Cold start**: Go binaries answer within 22 ms of launch; Astro 206 ms, Next.js 718 ms.
- **Artifact**: 9–30 MB static binaries versus 250 MB (Astro) / 501 MB (Next.js) of `node_modules` + build output.
- **What the browser downloads**: ~14 KB of HTML everywhere except Next.js (51 KB with the RSC payload). First-load JS, gzipped: Next.js 127 KB (React runtime + page chunks), gotsx 10.6 KB (signals runtime + loader + the island; the morphing library for SPA navigation loads on hover or first navigation), Astro 6.5 KB (Preact + island); the template stacks ship no framework JS but also have no hydration story.
- **Build**: templ and plain Go build in under a second; gotsx adds `hostgen` (a `go run`) and the dialect compiler; Gin's 18 s is a cold module download on the runner; Next.js builds in 6 s.
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

After the workflow has committed fresh `results/`, `python3 bench/update_docs.py` rewrites the tables above and the
summary in both READMEs from `results/*.json` (machine and date included), so the published numbers always come from
one identified run.
