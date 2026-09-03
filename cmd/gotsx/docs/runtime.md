# The Go side (gotsx {{version}})

Import path: `github.com/childrentime/gotsx/runtime` (alias `gotsx`). Generated code lives in `gen/`.

## Serving

```go
log.Fatal(gotsx.Serve(gotsx.Options{
    Addr: ":3000", Dev: *dev,
    Routes: gen.Routes, NotFound: gen.NotFound, ErrorPage: gen.ErrorPage,
    ClientDir: gotsx.FindDir("gen/client"), ClientFS: gen.ClientFS,   // island JS: disk in dev, embedded in prod
    PublicDir: gotsx.FindDir("public"), PublicFS: public,             // /public/*
    HostActions: gen.HostActions,                                     // typed actions
    SessionSecret: os.Getenv("SESSION_SECRET"),                       // empty → random per start
    Actions: map[string]http.HandlerFunc{"POST /todos": …},           // classic handlers (same-origin checked)
    Middleware: []func(http.Handler) http.Handler{auth, metrics},     // optional, outermost first
    I18n: &gotsx.I18n{…}, QuietLogs: false,
}))
```

`gotsx.Handler(opt)` returns an `http.Handler` instead of listening (tests, Wasm/Cloudflare, custom servers).
Built in: request IDs, access log, panic recovery (500 / `_error` page), security headers and a CSP nonce for the
runtime scripts, gzip on request, `/healthz`, `/readyz`, graceful shutdown, streaming Suspense, dev SSE reload.

## Host modules

```go
type TodoStore struct{ mu sync.Mutex; items []Todo }
func (s *TodoStore) List() []Todo                  // todos.list(): Todo[]
func (s *TodoStore) Get(id string) (Todo, error)   // todos.get(id): Todo   (throws; ErrNotFound → 404)
type DataModule struct{ Todos *TodoStore `json:"todos"` }
func (d *DataModule) Toggle(id string) (Todo, error)                    // action candidate
func (d *DataModule) Rename(req *gotsx.Req, id, title string) error     // action with the request injected
var Data = &DataModule{…}
var Registry = map[string]gotsx.HostModule{"data": {Value: Data, Go: "host.Data", Actions: []string{"Toggle", "Rename"}}}
```

Type mapping (Go → dialect): `string`→`string`; `bool`→`boolean`; all ints/floats→`number`; `[]T`→`T[]`;
`map[string]T`→`Record<string, T>` (maps with other key types are exposed the same way; JSON turns the keys into strings); struct with json tags→interface (field names from the tags; `omitempty`
does not make a field optional); `*T`→`T` (a nil pointer reaches the dialect as the zero value / `null`);
`time.Time`→`string` (RFC 3339; `isoDate()` formats) — but not as an action parameter, where only builtin and
`host` types are decoded; `(T, error)`→ throws. Rejected by hostgen with a message: channels, funcs, non-empty
interfaces, complex numbers; `any` becomes `unknown` (JSON passthrough).

Host methods run concurrently (one goroutine per request, more with Suspense): guard shared state with a mutex or
use a database. Methods should be fast; there is no per-request caching.

Regenerate after changing `host/`: `gotsx build` (or `gotsx dev` does it on save) runs `cmd/hostgen`, which
rewrites `app/.gen/host.d.ts` and `host.json`. The compiler then knows the new signatures.

## Stores (generated, nothing to configure)

`export const cart = createStore<T>(init)` compiles to a `gotsx.NewStore(name, init)` package var in `gen/`; `seed(cart, value)`
in a page or layout body compiles to `gotsx.Seed(props, store, value)`, which records the value for the current request
(islands render with it through the render context, Suspense goroutines included) and writes it into `<head>` as
`<script type="application/json" data-gotsx-stores>` — a data block, no CSP nonce needed, `<` escaped, nil slices as `[]`.
The Go side never mutates a store, so the vars are safe to share across requests.

## Request helpers

- `gotsx.Req` (first param of an action): `W`, `R`, `Cookies`, `Locale`, `Session()`, `SetCookie(c)`.
- `gotsx.Session`: `Get/Set/Delete/Clear`, `Flash(kind, text)`, `Values()`, `CSRF()`, `Save(w, r)` (actions save automatically).
- `gotsx.SessionOf(r)` / `gotsx.VerifyCSRF(r)` in classic handlers.
- Errors: `gotsx.Invalid(fields)` (422), `gotsx.Fail(msg)` (400), `gotsx.Unauthorized(msg)` (401), `gotsx.Forbidden(msg)` (403), `gotsx.ErrNotFound` (wrap with `%w` → 404). Redirects from a page are the dialect's `redirect(url)`; from a Go handler use `http.Redirect`.
- `PageProps` reaches pages with `Session`, `Flash`, `CSRF` filled in; `LayoutProps.Meta` carries the page meta.

## Deploy

`gotsx build && CGO_ENABLED=0 go build -o app .` produces one static binary (client JS and `public/` embedded).
Set `SESSION_SECRET`; run behind any reverse proxy or directly. `gotsx export` produces a static site for pages
that need no server state. A Cloudflare Workers target exists via Go→Wasm (`gotsx.Handler` + syumai/workers).
