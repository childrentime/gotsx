# Contributing to gotsx

## Environment

- Go (version in `go.mod`). **No Node required.**
- The Tailwind standalone binary for the demo apps: `go run ./cmd/gotsx tailwind` (or set `GOTSX_TAILWIND` to an existing one).

## Everyday commands

```bash
make gen        # compile every demo app's dialect → gen/ (gen is gitignored, so run this first)
make build      # gen + go build ./...
make test       # gen + go test ./... (builds all four apps, runs the gotsx new → build → check e2e)
make test-fast  # compiler / runtime / CLI unit tests only
make check      # gotsx check on every demo app (what the LSP runs)
make dev-shop   # run the shop dev server (dev-example / dev-site / dev-admin likewise)
make lint       # go vet
make fmt        # gofmt
```

> **Important:** `gen/` is gitignored and does not exist in a clean checkout. Any command that compiles an app
> (`go build ./...`, `go test ./...`) must be preceded by `make gen`, otherwise the `*/gen` packages are missing.
> CI is ordered this way (and also checks `gofmt` and `make check`).

## Repository layout

| Directory | Contents |
|---|---|
| `compiler/` | the dialect compiler: `lexer` / `parser` / `check` (types) / `gogen` (Go backend) / `jsgen` (JS backend) / `compile` (pipeline, `Analyze` for check/LSP) |
| `runtime/` | what generated Go depends on: node model, hydration markers, dialect builtins, HTTP, host-type reflection, i18n, redirect/notFound |
| `client/` | the browser runtime: signals, `el/t/text/cond/each` (keyed), resumable hydration; island loader + morph navigation + dev live reload |
| `cmd/gotsx/` | the CLI: `new` (scaffold) / `build` / `dev` (hostgen → tailwind → compile → go build → run → watch) / `check` / `lsp` / `tailwind` |
| `editors/` | LSP setup: VS Code extension source, Neovim / Helix / Zed snippets |
| `design/` | gotsx UI, the design system every demo and `gotsx new` use (`gotsx.css` Tailwind layer, `plain.css` twin, `README.md` rules). Demo UI changes must stay inside its tokens and component classes |
| `example/` `site/` `shop/` `admin/` | demo apps, also the integration-test targets; `example/app/pages/kitchen.server.tsx` exercises the whole language |

## Test conventions

- **`compiler/codegen_test.go`, `compiler/lang_test.go`**: dialect snippet → assert the generated Go / JS contains the expected structure. `TestMarkerParity` / `TestKeyedMarkerParity` guard the invariant that both backends emit the same hydration markers.
- **`compiler/fence_test.go`, `TestLongTailFence`**: every fence violation → an error carrying `file:line:col`.
- **`compiler/apps_test.go`**: compile the four real apps and `go build` + `go vet` (the regression net; skipped with `-short`).
- **`runtime/*_test.go`**: builtin correctness, rendering / hydration markers, XSS escaping, HTTP middleware, routing, page control flow.
- **`cmd/gotsx/cli_test.go`**: `gotsx new` → `build` → `go build` → `check` in a temp module (skipped with `-short`).

**The rule for adding a language feature**: it lands in the checker **and** the Go backend **and** the JS backend
**and** the runtime (if needed), with a line in `lang_test.go` for each backend; the two backends must agree
semantically (number formatting, strings by rune, sorted map keys, zero-value-as-undefined). Outside the subset,
report an error with a position — never miscompile silently. If the feature is user-visible, add it to
`example/app/pages/kitchen.server.tsx` (so it is compiled and `go build`-checked by `make test`) and to the
syntax table in `site/app/content/site.server.tsx`.

## Design principles

- **Go is the single source of truth**: routing, data, permissions and host capabilities live in Go; what the dialect can do is exactly what Go exposes.
- **The subset is defined by the type system**: if a static type can be inferred and it falls inside the allowed set, it compiles; otherwise it is a positioned compile error.
- **SSR is a single pass**: `useState` = the initial value, `useEffect` = nothing, setters = no-ops; this is what makes TSX compilable to Go.
- **One source, two backends, same semantics**: anything observable (formatting, ordering, truthiness) must behave the same in Go and in the browser; when in doubt, the Go side wins and the JS side adapts (e.g. `??` ≡ `||` for primitives, `localeCompare` by code point).
