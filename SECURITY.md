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
- **Writes happen in Go**: the demo apps put every mutation (cart, orders, users) in Go `http.HandlerFunc`s (`Options.Actions`);
  the dialect reads, it does not write. Stock, prices and order consistency are Go's responsibility; the client cannot alter them.
- **HTTP layer defaults** (`runtime/server.go`, `runtime/server_test.go`): panic recovery without leaking stacks in production,
  `X-Content-Type-Options` / `X-Frame-Options` / `Referrer-Policy`, a **Content-Security-Policy with a per-response nonce**
  (only the framework's inline bootstrap script carries it), **CSRF same-origin checks on POST/PUT/PATCH/DELETE actions**
  (requests without an `Origin`/`Referer` are rejected), request-body limits on telemetry, `no-store` on pages.
- **Page control flow**: `redirect()` only accepts 3xx status codes (anything else falls back to 302); the URL is passed to
  `http.Redirect` as-is, so validate user-supplied redirect targets in your host code.
- **Dev mode**: `-dev` enables the `/_gotsx/dev` live-reload stream and detailed error pages. Never run production with `-dev`.

## What remains the application's responsibility

- **Authentication / authorization**: the framework provides the hooks (`Options.Before`, `Options.Middleware`, cookies in
  `PageProps.Cookies`), not a policy. The `admin` demo shows a session cookie + middleware pattern.
- **Input validation in host modules**: parameter *types* are enforced by the compiler; business validation (length, range,
  injection into SQL/shell) belongs to the host implementation.
- **Action deserialization**: the JSON body of an action is parsed by your handler; validate it.
- **Open redirects**: `redirect(query.next)` is only as safe as the check you do on `query.next`.

## Self-review (v0.6) and known gaps

An independent audit is still pending (it is the last open roadmap item). Until then, this is what has been
reviewed in-repo, so an auditor knows where to start:

| Area | What was checked | Status |
|---|---|---|
| Output escaping | text / attributes / island props / JSON-LD; Suspense fill content is rendered by the same node model (escaped) and inserted via `template.content`, never `innerHTML` of untrusted strings | tested |
| Inline scripts | the bootstrap script and every Suspense fill script carry the per-response CSP nonce; nothing else is inline | tested |
| CSRF | same-origin check on unsafe methods for `Options.Actions`; telemetry endpoint same-origin only, body capped at 16 KB | tested |
| Redirects | `redirect()` only accepts 3xx; the target is application data — validate `query.next`-style inputs in the page | documented |
| Dev-only surface | `/_gotsx/dev` (SSE) and detailed error pages exist only with `-dev`; production binaries never enable them | tested |
| Streaming | a boundary that panics is logged and rendered empty in production; a disconnected client drains remaining goroutines; the http.Server write timeout bounds a stuck boundary | tested |
| LSP | runs locally over stdio; reads only files under the app's `app/` directory plus `.gen/host.d.ts`; never executes app code | reviewed |
| Regex | patterns are compiled with Go's RE2 engine (linear time, no catastrophic backtracking) and validated at compile time | tested |
| Dependencies | none outside the Go standard library; `govulncheck ./...` on Go 1.26.4 reports only stdlib advisories that the framework's code paths do not call (GO-2026-6218 `net/url`, GO-2026-6090 `crypto/tls`, GO-2026-6089 `net/http`, GO-2026-5972 `encoding/asn1`) — upgrade the Go toolchain to clear them | checked |

Known gaps: no rate limiting or request-size limits on application actions (add middleware); `Options.SecurityHeaders`
can weaken the defaults if misused; host modules are trusted Go code.

## Reporting a vulnerability

Please do **not** open a public issue with exploit details. Email the maintainer (see the repository owner's GitHub
profile) or use GitHub's private vulnerability reporting on the repository. You will get an acknowledgement within a
week; fixes ship as a patch release with a changelog entry.
