# Errors and fixes (gotsx {{version}})

Compile errors are `file:line:col: message`. `gotsx check --json` returns them as `[{file,line,col,msg}]`.
`gotsx dev` also writes them to `.gotsx/diagnostics.json` and shows an overlay in the browser.

| message contains | meaning | fix |
|---|---|---|
| ``is not an action, so client code may only `import type` it`` | an island imports a host value that is not in `Actions` | add the method name to `host.Registry[...].Actions` (module-level method), or move the call to a server component and pass the data as a prop |
| `takes *gotsx.Req, so it can only be called from an island` | a server component calls an action that needs the request | call it from an island, or drop the `*gotsx.Req` parameter |
| `is not an action; only actions can be called from client code` | island calls a plain host method | same as above |
| `no host module named "host:x"` | Registry key missing or `gotsx build` not run after editing `host/` | check `Registry`, rebuild |
| `host:x does not export "y"` | method/field not exported or not in `app/.gen/host.d.ts` | export it (capitalized, json tag for fields), rebuild |
| `a page must export default a component` | page file has no default export or exports a lowercase function | `export default function Page(props: PageProps)` |
| `a page component's props must be PageProps` / `LayoutProps` / `ErrorProps` | wrong props type on a page/layout/error file | use the framework type from `"gotsx"` |
| `export function meta must have the signature (props: PageProps) => Meta` | wrong meta signature | `export function meta(props: PageProps): Meta` (or no params) |
| `cannot find module` | wrong relative path, or a server component imported from an island | check the path; islands may only import client modules and types |
| `missing prop` / `has no prop` | component call does not match its props type | fix the JSX attributes |
| `await can only be used in client code` / `… only … server components (*.server.tsx)` | server code uses browser-only constructs, or an island uses `redirect()` / `notFound()` / `Suspense` | move the code to the other side: islands for `await`/DOM, server components for page control flow |
| `actions return a Promise` … `not during render` | an action called in a component body (render is synchronous) | call it from a handler or effect: `onClick={() => toggle(id)}` (fire-and-forget) or `onClick={async () => { const t = await toggle(id); … }}` |
| `actions may only take builtin types and host types` (from `gotsx build`) | an action parameter has a type from another package (`time.Time`, …) | take a string / number and convert in Go |
| `unsupported` / `not supported` (syntax) | construct outside the subset | see the "no" rows in `language.md` for rewrites |
| `Go backend: object spread is not supported` | `{...a}` in server code | build the object explicitly |
| `cannot determine the type of the object literal` | an untyped `{}` literal | annotate: `const x: T = {...}` or add a return type |
| `hostgen failed` … `has Go type … which the dialect cannot represent` / `*gotsx.Req must be the first` / `lists action … but has no such method` | `cmd/hostgen` rejected a Registry entry | the message names the method and type; use strings/numbers/slices/maps/structs, put `*gotsx.Req` first, fix the Actions list |
| `go build failed` | the generated Go or your Go code does not compile | read the Go error; usually `host/` or `main.go` |

## Runtime

| symptom | cause | fix |
|---|---|---|
| action returns 403 | cross-origin request or missing `X-Gotsx-Action` header (calls from outside the compiled island) | use the imported action; for external clients use a classic handler |
| action returns 404 `{"error":"unknown action"}` | method not in `Actions`, or `HostActions: gen.HostActions` missing in `main.go` | add both |
| action returns 422 with `fields` | `gotsx.Invalid` from the method | show `e.fields` next to the inputs |
| action returns 401 / 403 | `gotsx.Unauthorized` / `gotsx.Forbidden` from the method (session check) | sign in / hide the control for that role |
| `delete only works on a Record key: delete m[key] / delete m.key` at a call like `delete(id)` | a Go method named `Delete` becomes the reserved word `delete` | rename the method (`Remove`) |
| form post returns 403 `invalid CSRF token` | no `_csrf` hidden input, or the session cookie was not sent | add `<input type="hidden" name="_csrf" value={csrf} />`; same site only |
| page shows an old value after an action | islands hold their own state; the server rendered the previous one | update the signal from the action's return value, or navigate/reload |
| `[browser] error: …` in the dev terminal | a JS error in an island (forwarded by the dev server) | fix the island; stack is in the browser console |
| 500 with `_error` page | a host method returned an error / panicked | wrap not-found errors with `gotsx.ErrNotFound` for 404s; check the log line with the request id |
| styles missing after `gotsx new --tailwind` | Tailwind binary not downloaded | `gotsx tailwind` once |
| `gotsx dev is already running for this app` | another dev server owns `.gotsx/dev.json` | use it (url in the file) or stop it; a stale file from a dead process is ignored |
