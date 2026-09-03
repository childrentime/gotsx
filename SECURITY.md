# Security

gotsx is pre-1.0 and has **not** been independently audited. Do not put it in front of untrusted input in
production unless you have reviewed it yourself.

## Defenses in place (all covered by tests)

- **XSS / HTML injection**: the dialect has no `dangerouslySetInnerHTML` and no other way to emit raw HTML.
  - Text nodes (`{expr}`) go through `gotsx.Text` / `gotsx.Dyn` → `html.EscapeString`.
  - Element attributes go through `gotsx.A` → `html.EscapeString`, so attribute break-out is impossible.
  - Island props are serialized to JSON and then HTML-escaped into the attribute.
  - `jsonLd()` only accepts a JSON string (normally `JSON.stringify(...)`, whose output already escapes `<>&`) and additionally neutralizes `</`.
  - See `runtime/security_test.go`: text, attributes, island props and full-chain XSS payloads are all asserted escaped.
- **Server / client boundary**: `host:*` (Go capabilities) can only be imported by server components; client code may only `import type`.
  Client code cannot reach Go, the database or secrets. Violations are compile errors (`compiler/fence_test.go`).
- **Writes happen in Go**: every mutation is a Go method. Islands reach them only through **typed actions**
  (`Registry[...].Actions`, `POST /_gotsx/act/<module>/<name>`): the compiler generates the route and the client stub,
  the runtime requires `POST`, a same-origin `Origin`/`Referer` **and** the `X-Gotsx-Action` header (HTML forms cannot
  set it, so a cross-site `<form>` never reaches an action even with `DisableCSRF`), caps the body at 1 MB, decodes the
  JSON arguments by Go type, recovers panics, and maps errors to 400/401/403/404/422/500 with the message hidden in
  production. Classic form posts stay Go `http.HandlerFunc`s (`Options.Actions`) behind the same-origin check plus
  `gotsx.VerifyCSRF` (a per-session token). Stock, prices and order consistency are Go's responsibility.
- **Sessions**: an HMAC-SHA256-signed cookie (`gotsx_session`, HttpOnly, SameSite=Lax, Secure over TLS or behind a
  proxy that says so) carrying string values, one-shot flash messages and the CSRF token. Verification is constant-time;
  a tampered or foreign cookie is an empty session. The key comes from `Options.SessionSecret`; when it is empty a random
  key is generated per process (sessions do not survive a restart and are not shared between replicas) — logged at
  startup when the app registers typed actions or runs in dev, and again the first time a session cookie is actually issued.
  Session data is signed, not encrypted: never store secrets in it.
- **HTTP layer defaults** (`runtime/server.go`, `runtime/server_test.go`): panic recovery without leaking stacks in production,
  `X-Content-Type-Options` / `X-Frame-Options` / `Referrer-Policy`, a **Content-Security-Policy with a per-response nonce**
  (only the framework's inline bootstrap script carries it), **CSRF same-origin checks on POST/PUT/PATCH/DELETE actions**
  (requests without an `Origin`/`Referer` are rejected), request-body limits on telemetry, `no-store` on pages.
- **Page control flow**: `redirect()` only accepts 3xx status codes (anything else falls back to 302); the URL is passed to
  `http.Redirect` as-is, so validate user-supplied redirect targets in your host code.
- **Dev mode**: `-dev` enables the `/_gotsx/dev` live-reload / diagnostics stream, detailed error pages, the in-page
  error overlay and the browser-error log endpoint (`/_gotsx/client-log`, same-origin, 16 KB, newlines stripped before
  logging). Never run production with `-dev`.

## What remains the application's responsibility

- **Authentication / authorization**: the framework provides the hooks (`Options.Before`, `Options.Middleware`, cookies in
  `PageProps.Cookies`), not a policy. The `admin` demo shows a session cookie + middleware pattern.
- **Input validation in host modules**: parameter *types* are enforced by the compiler; business validation (length, range,
  injection into SQL/shell) belongs to the host implementation.
- **Action arguments**: typed actions decode arguments by Go type, so a string parameter is always a string — but
  business validation (length, range, ownership, injection into SQL/shell) belongs to the method; return
  `gotsx.Invalid` / `Fail` / `Unauthorized` / `Forbidden` so the client sees a status, not a stack.
- **Authorization inside actions**: an action is reachable by anyone who can load the page; check
  `req.Session()` yourself (the `admin` demo's `requireUser` / `requireEditor`).
- **Open redirects**: `redirect(query.next)` is only as safe as the check you do on `query.next`.

## Self-review (v0.7) and known gaps

An independent audit is still pending (it is the last open roadmap item). Until then, this is what has been
reviewed in-repo, so an auditor knows where to start:

| Area | What was checked | Status |
|---|---|---|
| Output escaping | text / attributes / island props / JSON-LD; Suspense fill content is rendered by the same node model (escaped) and inserted via `template.content`, never `innerHTML` of untrusted strings | tested |
| Inline scripts | the bootstrap script and every Suspense fill script carry the per-response CSP nonce; nothing else is inline | tested |
| CSRF | same-origin check on unsafe methods for `Options.Actions`; typed actions additionally require the `X-Gotsx-Action` header (unconditionally) and a 1 MB body cap; classic forms verify a per-session token (`VerifyCSRF`); telemetry endpoint same-origin only, body capped at 16 KB | tested |
| Sessions | HMAC-SHA256 signed cookie, constant-time compare, HttpOnly + SameSite=Lax + Secure (TLS / `X-Forwarded-Proto` / `Forwarded` / `CF-Visitor`), random per-process key when unset (logged), flash consumed exactly once, lazily created so anonymous pages set no cookie | tested |
| Typed actions | JSON decoded by Go type (`gotsx.Arg`), panics recovered, error detail only in dev, unknown action → 404, duplicate registrations logged, `*gotsx.Req` injection only for module-level methods | tested |
| Island props | serialized with nil slices/maps as `[]`/`{}`; the reflection walk is cycle-safe and depth-capped; marshaler types pass through unchanged; still HTML-escaped into the attribute | tested |
| Static export | `gotsx export` refuses to empty anything but an empty directory or a previous export, never the app dir, its ancestors, `/` or `$HOME`; crawled links never leave the local origin | tested |
| Redirects | `redirect()` only accepts 3xx; the target is application data — validate `query.next`-style inputs in the page | documented |
| Dev-only surface | `/_gotsx/dev` (SSE) and detailed error pages exist only with `-dev`; production binaries never enable them | tested |
| Streaming | a boundary that panics is logged and rendered empty in production; a disconnected client drains remaining goroutines; the http.Server write timeout bounds a stuck boundary | tested |
| LSP | runs locally over stdio; reads only files under the app's `app/` directory plus `.gen/host.d.ts`; never executes app code | reviewed |
| Regex | patterns are compiled with Go's RE2 engine (linear time, no catastrophic backtracking) and validated at compile time | tested |
| Dependencies | none outside the Go standard library; `govulncheck ./...` on Go 1.26.4 reports only stdlib advisories that the framework's code paths do not call (GO-2026-6218 `net/url`, GO-2026-6090 `crypto/tls`, GO-2026-6089 `net/http`, GO-2026-5972 `encoding/asn1`) — upgrade the Go toolchain to clear them | checked |

Known gaps: no rate limiting (add middleware) and no request-size limit on classic `Options.Actions` handlers (typed
actions are capped at 1 MB); `Options.SecurityHeaders` can weaken the defaults if misused; host modules are trusted Go
code; the session cookie is signed but not encrypted; the checker does not yet enforce every assignment (the Go
compiler does, at build time).

## Reporting a vulnerability

Please do **not** open a public issue with exploit details. Email the maintainer (see the repository owner's GitHub
profile) or use GitHub's private vulnerability reporting on the repository. You will get an acknowledgement within a
week; fixes ship as a patch release with a changelog entry.
