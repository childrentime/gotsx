# gotsx — working notes for Claude (and humans)

gotsx is a full-stack framework that compiles a TSX dialect to native Go. Read `README.md` for what it is,
`STABILITY.md` for what is promised, `design/README.md` for how the demos must look.

## Language policy

- **English is the primary language.** Everything user-facing is English first: README, docs site (`/`),
  CLI output, compiler diagnostics, runtime logs, demo UI copy, demo seed data, commit messages, release notes,
  code comments in new code, identifiers.
- **Chinese is the secondary language**, delivered through the same i18n mechanisms as any other locale:
  `README.zh.md`, the `zh` locale of `site` and `shop` (`/zh/...`), `CHANGELOG.md` may stay bilingual.
- A demo must not show Chinese by default. If a demo has i18n, the default locale is `en`; if it has none, it
  is English only. Product names, user names, activity feeds and validation messages count as UI copy.
- Existing Chinese comments in older code are fine to leave; do not add new ones.

## Commands

```bash
make gen        # compile every demo app (gen/ is gitignored — always first)
make test       # gen + go test ./... (builds all four apps, runs the gotsx new → build → check e2e)
make test-fast  # compiler / runtime / CLI unit tests
make check      # gotsx check on every demo app (what the LSP runs)
make dev-shop   # dev server with live reload (dev-site / dev-admin / dev-example)
make lint fmt   # go vet / gofmt
go run ./cmd/gotsx tailwind   # once: Tailwind standalone binary into .tools/ (no Node)
python3 bench/run.py --duration 2   # benchmark harness smoke test only — published numbers come from the Benchmark workflow
python3 bench/update_docs.py        # after the workflow committed bench/results/: rewrite the README tables
```

A prebuilt CLI for parallel work: `go build -o .tools/gotsx-stable ./cmd/gotsx`, then `.tools/gotsx-stable build <app>`.

## Rules that keep the project coherent

- **One source, two backends, same semantics.** A language feature lands in the checker, the Go backend and the
  JS backend together, with a `compiler/lang_test.go` case per backend and a row in the syntax table
  (`site/app/content/site.server.tsx`). Anything observable (formatting, ordering, truthiness, absence) must be
  identical in Go and in the browser; the Go side decides, the JS side adapts.
- **Outside the subset is a positioned compile error**, never a silent miscompile.
- **Hydration markers must match**: Go `Dyn/Nodes/If/Marked` ↔ JS `text/each/cond` (`TestMarkerParity`).
- **Demos use only the design system** (`design/README.md`): semantic tokens and component classes, inline SVG
  icons, no palettes, no gradients, no emoji as UI chrome.
- **Generated code is internal**: `gen/`, `runtime` helpers called only by generated code, marker formats and
  `host.json` may change freely; `gotsx.Options`, the CLI, file conventions and the dialect are contracts.
- **Security defaults stay on** (CSP nonce, CSRF same-origin, escaping everywhere); a demo may not disable them.
- **Verify in a real browser** when touching the client runtime or a demo: Python Playwright is available
  (`from playwright.sync_api import sync_playwright`), use `bypass_csp=True` for Playwright's own helpers.

## Layout

`compiler/` (lexer, parser, check, gogen, jsgen, compile, query) · `runtime/` (node model, HTTP, builtins,
i18n, hostgen) · `client/` (signals runtime, loader) · `cmd/gotsx/` (CLI, LSP, scaffold) · `design/` (gotsx UI) ·
`editors/` (LSP clients) · `bench/` (reproducible benchmarks vs other frameworks) · `example/ site/ shop/ admin/`
(demo apps = integration tests; `example` has the language kitchen sink).
