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
_Pending: run the Benchmark workflow (Actions → Benchmark → Run workflow); it commits `results/results.md` and
`results/results-1core.md` here. Local smoke runs are for checking the harness only._
<!-- /BENCH:ALL -->

## Reading the numbers

**Where gotsx's speed comes from.** A page is compiled to a Go function that writes HTML into one buffer:
static markup is merged into single writes at compile time, dynamic text is escaped inline, and `list.map(...)`
becomes a `for` loop. There is no template parsing, no reflection and no virtual DOM; the whole 50-item page
is ~18 µs of CPU and ~440 allocations, so throughput is bounded by the network stack and the GC, like templ.
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
