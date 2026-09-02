# Conventions (gotsx {{version}})

## File kinds

| file | runs | may import | notes |
|---|---|---|---|
| `app/pages/**/*.server.tsx` | Go, per request | `host:*`, `./components/*.server`, islands (`*.client`), `gotsx` | `export default` a component taking `PageProps`; optional `export function meta(props?: PageProps): Meta` |
| `app/components/*.server.tsx` | Go | same as pages | any props; may take `children: Node` |
| `app/islands/*.client.tsx` | Go (first render) + browser (hydrated) | `gotsx`, other client modules, `import type` from `host:*`, **actions** from `host:*` | props must be JSON-serializable; no `children`; may use `useState`/`useEffect`, `async` handlers, `fetch`, DOM |
| `app/pages/**/_layout.server.tsx` | Go | same as pages | `LayoutProps` = `PageProps` + `meta` + `children`; nested directories may add their own; outer wraps inner |
| `app/pages/_404.server.tsx` | Go | same | → `gen.NotFound` (PageProps) |
| `app/pages/_error.server.tsx` | Go | same | → `gen.ErrorPage` (`ErrorProps` = PageProps + `message`) |
| `app/pages/**/_*.server.tsx` | ignored | | private helpers |

Routing: `pages/index.server.tsx` → `/`, `pages/about.server.tsx` → `/about`, `pages/posts/[id].server.tsx` →
`/posts/{id}` (`params.id`), `pages/docs/[...slug].server.tsx` → `/docs/{...slug}` (`params.slug = "a/b/c"`).
More specific routes win. Trailing slashes are normalized.

## PageProps

```ts
interface PageProps {
  params: Record<string, string>;   // route parameters
  query: Record<string, string>;    // ?a=b (first value)
  path: string;                     // request path (locale prefix stripped when i18n is on)
  locale: string;                   // "" when i18n is off
  cookies: Record<string, string>;
  session: Record<string, string>;  // signed session values (read-only here)
  flash: Flash[];                   // one-shot messages queued by actions/handlers: {kind, text}
  csrf: string;                     // token for classic <form method="post">: <input type="hidden" name="_csrf" value={csrf} />
}
```

Pages are plain functions: no data-fetching hooks, no loaders. Call Go directly:

```tsx
import type { PageProps, Meta } from "gotsx";
import { posts } from "host:data";

export function meta({ params }: PageProps): Meta {
  const p = posts.get(params.id);                    // (T, error) in Go: a wrapped ErrNotFound becomes a 404
  return { title: p.title, description: p.summary };
}

export default function Post({ params }: PageProps) {
  const p = posts.get(params.id);
  return <article><h1>{p.title}</h1><p>{p.body}</p></article>;
}
```

`Meta` fields: `title`, `description`, `canonical`, `image`, `noIndex` (all optional). Layouts render them:
`<title>{meta.title ? meta.title + " · Site" : "Site"}</title>`. `meta` runs before the page in the same
request; if both need the same record, keep that host method cheap (an in-memory lookup, or a cache in Go) —
there is no automatic per-request memoization.

Redirects and 404s from a page: `redirect("/login")`, `notFound()` (globals). Streaming: wrap slow parts in
`<Suspense fallback={<p>Loading…</p>}>…</Suspense>` (server only); each boundary renders in its own goroutine.

## Host modules (`host:*`)

The Go package `host` exports a `Registry`:

```go
var Registry = map[string]gotsx.HostModule{
    "data": {Value: Data, Go: "host.Data", Actions: []string{"Toggle", "Remove"}},
}
```

`gotsx build` reflects each module value: exported methods become functions (`todos.list()`), exported fields with
json tags become nested objects/modules, struct types become interfaces in `app/.gen/host.d.ts`. Read that file to
see exactly what is callable. Method names are lower-camel-cased (`Toggle` → `toggle`). A `(T, error)` method
"throws": in a page the error is a panic → 500, or 404 if it wraps `gotsx.ErrNotFound`.

Server components call host methods synchronously and directly (it is a Go call after compilation).

## Typed actions (islands → Go)

An **action** is a module-level Go method listed in `Actions`. Islands import it as a value and `await` it:

```tsx
import { toggle } from "host:data";        // Toggle(id string) (Todo, error)  →  toggle(id: string): Promise<Todo>
const t = await toggle(id);
```

- Compiles to a same-origin `POST /_gotsx/act/data/toggle` with a JSON array of arguments; the server decodes
  by Go type, runs the method, returns the value. Non-2xx throws an `Error` with `e.status` and `e.fields`.
- Errors: `gotsx.Invalid(map[string]string{"title": "required"})` → 422 with `e.fields`; `gotsx.Fail("msg")` → 400;
  `gotsx.Unauthorized("sign in")` → 401; `gotsx.Forbidden("viewers cannot edit")` → 403; an error wrapping
  `gotsx.ErrNotFound` → 404; anything else → 500 (message only in dev). Authorization lives in the action itself:
  read `req.Session().Get("user")` and return `Unauthorized` / `Forbidden`.
- A method whose first parameter is `*gotsx.Req` gets the request injected: `req.Session()`, `req.Cookies`,
  `req.Locale`, `req.SetCookie(...)`. Such a method can only be called from islands.
- Actions without `*gotsx.Req` can also be called from server components like any host method.
- Arguments and results must be builtin types or `host` types (no `time.Time`; use strings).
- Method names become lower-camel-case identifiers in the dialect, so a Go method may not be named after a
  reserved word (`Delete` → `delete` clashes with the `delete m[k]` statement: call it `Remove`). Empty Go
  slices/maps reach islands as `[]` / `{}`, never `null`.
- Security is built in: POST only, same-origin check, the `X-Gotsx-Action` header, a 1 MB body limit, panics recovered.

Passing an action as a value (`onClick={remove}`) is allowed; it becomes a function that posts when called.
An action is a network request, so calling it in a component body (during render) is a compile error; call it
from a handler or an effect — `onClick={() => remove(id)}` fires and forgets, `await` inside an `async` handler
gives you the result.

## Classic form posts, sessions, flash, CSRF

For non-JS forms register a Go handler in `gotsx.Options.Actions` and use the session helpers:

```tsx
<form method="post" action="/todos">
  <input type="hidden" name="_csrf" value={csrf} />
  <input name="title" />
</form>
```

```go
"POST /todos": func(w http.ResponseWriter, r *http.Request) {
    if !gotsx.VerifyCSRF(r) { http.Error(w, "invalid CSRF token", 403); return }
    sess := gotsx.SessionOf(r)
    if _, err := host.Data.Todos.Add(r.FormValue("title")); err != nil { sess.Flash("error", err.Error()) } else { sess.Flash("ok", "Added") }
    sess.Save(w, r)                          // before writing the response
    http.Redirect(w, r, "/", http.StatusSeeOther)
},
```

The next page render receives `props.flash` (consumed once) and `props.session`. Render flashes with
`{flash.map((f) => <div class={"alert alert-" + f.kind}>{f.text}</div>)}`. Sessions are signed cookies
(`Options.SessionSecret`, HMAC-SHA256, HttpOnly, SameSite=Lax); values are strings; keep them small.
`props.csrf` is lazy: the token (and the session cookie) is created only when a page reads it, so pages that
don't use it stay cookie-free and cacheable. Read it in the page shell, not inside a `<Suspense>` boundary
(after the shell is flushed a cookie can no longer be set).

## Styling

Scaffolds ship the gotsx design system (`public/app.css`, or `app/gotsx.css` with `--tailwind`): tokens
(`--background`, `--foreground`, `--muted`, `--border`, `--primary`, `--destructive`, `--success`, `--warning`,
`--radius`) and classes `.btn .btn-primary .btn-outline .btn-ghost .btn-sm`, `.input`, `.card`, `.badge-*`,
`.alert .alert-ok .alert-error .alert-info .alert-warning`, `.table`, `.nav-link`, `.container-page`, `.stack`,
`.row`, `.muted`, `.empty`. Use only tokens for colors (dark mode follows `prefers-color-scheme`); icons are
inline SVG, not emoji. Attribute is `class`, not `className`.

## i18n (optional)

With `Options.I18n` set: `t("key")`, `tv("key", {n: 3})`, `plural("key", n)`, `lpath("/docs")` (locale-prefixed
link), `fmtNum / fmtCur / fmtDate`, and `props.locale`. Messages live in Go (`gotsx.I18n{Messages: …}`).
