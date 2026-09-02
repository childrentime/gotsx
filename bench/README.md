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
**GitHub-hosted runner (`ubuntu-latest`): INTEL(R) XEON(R) PLATINUM 8573C (4 vCPU) · go1.26.7 linux/amd64 · node v22.23.2 · bun 1.4.0 · 64 connections · 15 s per contender · 2026-09-02.** Raw data: [`results/`](results/). Runner hardware varies between runs (AMD EPYC or Intel Xeon); ratios are stable, absolute numbers move ±30%.

All cores (Go servers use all 4 vCPUs; Node and Bun are single-threaded):

| framework | req/s | p50 | p99 | peak RSS | cold start | build | artifact | HTML | JS (gz) | note |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---|
| gotsx | 19,620 | 1.96 ms | 14.26 ms | 21.2 MB | 20 ms | 3.04 s | 10.2 MB | 14.3 KB | 24.3 KB (9.2) | production mode: CSP nonce, security headers, request logging off |
| stdlib | 5,257 | 8.01 ms | 54.35 ms | 20.3 MB | 11 ms | 1.14 s | 12.2 MB | 14.4 KB | 0.0 KB (0.0) | net/http + html/template |
| gin | 4,981 | 8.60 ms | 61.88 ms | 27.2 MB | 11 ms | 16.24 s | 30.5 MB | 14.4 KB | 0.0 KB (0.0) | gin release mode + html/template |
| templ | 20,789 | 2.80 ms | 9.41 ms | 17.0 MB | 11 ms | 0.7 s | 9.4 MB | 13.9 KB | 0.0 KB (0.0) | a-h/templ generated Go |
| nextjs | 294 | 214.12 ms | 278.26 ms | 386.8 MB | 516 ms | 5.17 s | 500.7 MB | 50.8 KB | 442.8 KB (127.4) | App Router, force-dynamic, next start |
| astro | 1,675 | 37.30 ms | 57.36 ms | 215.0 MB | 170 ms | 1.48 s | 250.4 MB | 23.7 KB | 15.8 KB (6.5) | output: server, node adapter, preact island |
| hono | 3,918 | 15.92 ms | 31.08 ms | 76.0 MB | 27 ms | 0.0 s | 1.4 MB | 14.4 KB | 0.0 KB (0.0) | hono/jsx on bun, SSR only |

Go servers pinned to one core (`GOMAXPROCS=1`) — the per-core comparison (JS column not re-measured):

| framework | req/s | p50 | p99 | peak RSS | cold start | build | artifact | HTML | JS (gz) | note |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---|
| gotsx | 13,234 | 4.78 ms | 10.64 ms | 19.4 MB | 18 ms | 0.45 s | 10.2 MB | 14.3 KB | 0.0 KB (0.0) | production mode: CSP nonce, security headers, request logging off |
| stdlib | 2,360 | 28.16 ms | 46.95 ms | 18.2 MB | 11 ms | 0.35 s | 12.2 MB | 14.4 KB | 0.0 KB (0.0) | net/http + html/template |
| gin | 2,313 | 28.86 ms | 47.25 ms | 23.4 MB | 11 ms | 0.69 s | 30.5 MB | 14.4 KB | 0.0 KB (0.0) | gin release mode + html/template |
| templ | 9,183 | 6.95 ms | 14.68 ms | 14.8 MB | 11 ms | 0.33 s | 9.4 MB | 13.9 KB | 0.0 KB (0.0) | a-h/templ generated Go |
| nextjs | 313 | 200.60 ms | 250.45 ms | 139.9 MB | 18 ms | 1.86 s | 500.9 MB | 50.8 KB | 0.0 KB (0.0) | App Router, force-dynamic, next start |
| astro | 1,676 | 36.96 ms | 68.98 ms | 233.7 MB | 168 ms | 1.41 s | 250.4 MB | 23.7 KB | 0.0 KB (0.0) | output: server, node adapter, preact island |
| hono | 3,970 | 15.77 ms | 30.76 ms | 75.1 MB | 28 ms | 0.0 s | 1.4 MB | 14.4 KB | 0.0 KB (0.0) | hono/jsx on bun, SSR only |

**Takeaways**

- **Throughput / latency**: gotsx and templ, the two compiled-Go stacks, are within 6% of each other on 4 cores (19,620 vs 20,789 req/s); on one core gotsx is the fastest of the seven (13,234 req/s, p50 4.8 ms) because a page is straight-line writes with ~140 allocations. gotsx is 3.7× `html/template` (stdlib; Gin is the same template engine), 11.7× Astro and 66.8× Next.js, whose p50 is 214 ms.
- **Memory**: the Go binaries peak at 17–27 MB under load; Hono/Bun 76 MB, Astro 215 MB, Next.js 387 MB.
- **Cold start**: Go binaries answer within 20 ms of launch; Astro 170 ms, Next.js 516 ms.
- **Artifact**: 9–30 MB static binaries versus 250 MB (Astro) / 501 MB (Next.js) of `node_modules` + build output.
- **What the browser downloads**: ~14 KB of HTML everywhere except Next.js (51 KB with the RSC payload). First-load JS, gzipped: Next.js 127 KB (React runtime + page chunks), gotsx 9.2 KB (signals runtime + loader + the island; the morphing library for SPA navigation loads on hover or first navigation), Astro 6.5 KB (Preact + island); the template stacks ship no framework JS but also have no hydration story.
- **Build**: templ and plain Go build in under a second; gotsx adds `hostgen` (a `go run`) and the dialect compiler; Gin's 16 s is a cold module download on the runner; Next.js builds in 5 s.
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
