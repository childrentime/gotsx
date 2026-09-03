# gotsx {{version}} — docs for coding agents

gotsx is a full-stack web framework whose source language is a **static TSX dialect compiled to Go**.
Pages and components become Go functions that write HTML; islands become a small signals runtime in the
browser; data comes from ordinary Go methods. It is *not* React, Next.js, Solid or Astro, even though the
syntax is familiar. These files describe exactly what this version accepts; prefer them over memory.

Reading order: `language.md` (syntax table) → `conventions.md` (files, pages, islands, actions, sessions) →
`runtime.md` (the Go side) → `errors.md` (messages and fixes) → `agent-workflow.md` (edit/check/verify loop).

## Project layout

```
main.go                  gotsx.Serve(gotsx.Options{...}) — routes, actions, session secret, static files
host/                    Go package "host": data, services, and the Registry the dialect can see
cmd/hostgen/main.go      reflects host.Registry → app/.gen/host.d.ts + host.json (run by gotsx build)
app/pages/**.server.tsx  file routing: index, [id], [...slug], _layout, _404, _error
app/components/*.server.tsx  server components (Go only)
app/islands/*.client.tsx     islands: rendered by Go, then hydrated in the browser
app/stores/*.client.tsx      stores: state shared by islands (createStore), seeded by pages (seed)
app/.gen/                GENERATED: host.d.ts (what Go exposes), host.json, docs/ (these files)
gen/                     GENERATED: pages_gen.go…, routes_gen.go, actions_gen.go, client/*.js, assets
public/                  static files served at /public/* (embedded into the binary)
.gotsx/                  dev state: dev.json (running server), diagnostics.json (last failed build), binaries
```

## Commands

| command | what it does |
|---|---|
| `gotsx dev [dir]` | build → `go build` → run; watches `app/ host/ public/ main.go`; browser reloads; errors overlay in the browser and land in `.gotsx/diagnostics.json` |
| `gotsx check [dir] [--json]` | type-check only (fast); JSON diagnostics `[{file,line,col,msg}]`; exit 1 on errors |
| `gotsx build [dir]` | hostgen ∥ tailwind → compile → `gen/`; also refreshes `app/.gen/docs/` and the AGENTS.md block |
| `gotsx export [dir] --out dist --base /sub` | static export (crawl every route) for GitHub Pages etc. |
| `gotsx new <dir> [--tailwind] [--db sqlite]` | scaffold an app |
| `gotsx docs [name]` | print one of these documents |
| `go build .` | after `gotsx build`: one self-contained binary |

## The mental model in five lines

1. A `.server.tsx` file runs in Go on every request. It may call Go directly (`import { todos } from "host:data"`), synchronously. No `async`, `fetch`, DOM, timers, `Promise`.
2. A `.client.tsx` file is an island: Go renders it once, the browser hydrates it as signals. Its props must be JSON-serializable. It talks to Go only through **typed actions** (`await toggle(id)`). State shared by several islands lives in a **store** (`createStore` in a client module) that the page seeds per request (`seed(store, value)`).
3. Types are bounded by Go: `number` is float64, absent primitives are zero values (`""`, `0`, `false`), `Record<string, T>` is a Go map, arrays are slices, `T | undefined` is a pointer-ish optional. No `any`, no classes, no generics of your own, no runtime type tests.
4. Everything is compiled: a construct outside the subset is a compile error with `file:line:col`, never a runtime surprise.
5. `gen/` and `app/.gen/` are outputs. Change `app/`, `host/`, `main.go`, `public/` and rebuild.
