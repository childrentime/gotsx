# Stability contract

gotsx is **0.x**: the ideas are settled, the surface is still being shaped by real apps. This document says
exactly what you can rely on, what may still move, and how changes are communicated — so that "0.x" is a
statement about polish, not about surprises.

## Versioning

- Semantic versioning. `0.MINOR.PATCH`: a **minor** release may contain breaking changes to *Experimental* items
  and, rarely, to *Stable* items — always listed under **Breaking** in `CHANGELOG.md` with a migration note.
  A **patch** release never breaks anything.
- Every release is a git tag (`v0.5.0`) on the module `github.com/childrentime/gotsx`; install a specific version
  with `go install github.com/childrentime/gotsx/cmd/gotsx@v0.5.0` and pin the same version in your `go.mod`.
- The CLI and the runtime are released together; generated code (`gen/`) is only guaranteed to build against the
  runtime version that generated it — `gotsx build` runs on every build, so this never matters in practice.
- Deprecations ship at least one minor release before removal, with a compile-time or startup warning naming
  the replacement.

## Tiers

### Stable — changes only with a documented migration

**The dialect.** Everything marked "supported" in the syntax table (`site` → `/docs/language`): the type
system (`string` `number` `boolean` arrays `Record` object types `interface` `type` literal unions `T | undefined`),
statements (`const let if for-of for while switch break continue try throw return`), expressions (operators incl.
`?? ++ %=`, template strings, arrows, optional chaining, `as`, `!`), JSX (elements, attributes, conditions, lists,
`key`, components, fragments), hooks (`useState useEffect useMemo`), builtins (`console JSON Math Object Date`,
array/string/number methods listed in the table, `String Number Boolean parseInt parseFloat isNaN
encodeURIComponent decodeURIComponent`), i18n builtins (`t tv plural fmtNum fmtCur fmtDate lpath`), page control
flow (`redirect notFound`), `jsonLd`, `isoDate`, regex literals (RE2 subset), the server/client fence and its
error semantics, the **assignability rules** (arguments, props, typed declarations, declared returns: any/unknown accepted, optional primitive ⊂ primitive, structural objects, `Node` accepts text), **typed actions** (value-importing an `Actions` member from `host:*` in an island, `await`
typed as `Promise<T>` from the Go signature, `Error.status` / `Error.fields` on failure), **page meta**
(`export function meta(props?: PageProps): Meta`, `Meta` / `Flash` importable from `"gotsx"`,
`LayoutProps.meta`) and the `PageProps` fields `session` / `flash` / `csrf`. Adding to the subset is never
breaking; removing from it always is.

**Semantic conventions.** Numbers are float64; strings are handled by rune on the server; an absent optional
primitive is its zero value (`""` `0` `false`) and `??` ≡ `||` for primitives; an absent optional object is
falsy and `=== undefined`; `Record` reads of a missing key yield the value type's zero value on both sides and
`Object.hasOwn` tests presence; `Object.keys` is sorted; `sort` copies; single-pass SSR (`useState` = initial
value, effects and setters do nothing on the server).

**File conventions.** `app/pages/**/*.server.tsx` → routes (`index`, `[param]`, `[...catchAll]`),
`_layout.server.tsx`, `_404.server.tsx`, `_error.server.tsx`; `app/components` `app/islands` `app/ui` are
conventions, not rules; `*.server.tsx` / `*.client.tsx` / no-suffix compile targets; `host/` + `cmd/hostgen`;
`gotsx.json` (optional) keys `module` and `hostPackage`; `gen/` layout (`Routes`, `NotFound`, `ErrorPage`,
`ClientFS`, `client/*.js`).

**Runtime API used from `main.go`.** `gotsx.Serve`, `gotsx.Handler` (the same handler chain without a server, for
your own `http.Server` / mux / tests), `gotsx.Options` and every exported field it has today,
`gotsx.Route`, `gotsx.PageProps` / `LayoutProps` / `ErrorProps` / `Meta` / `Flash`, `gotsx.HostModule`
(incl. `Actions`) / `GenerateHost`, `gotsx.HostAction` (as produced by `gen.HostActions`), `gotsx.Req`,
`gotsx.Session` and its methods, `gotsx.SessionOf`, `gotsx.VerifyCSRF`, `gotsx.Invalid` / `Fail` /
`Unauthorized` / `Forbidden` / `ValidationError`, `gotsx.Arg`, `gotsx.IsHTTPS`, `gotsx.ErrNotFound`, `gotsx.I18n`, `gotsx.ClientEvent`, `gotsx.FindDir`,
`gotsx.SameOrigin`, and the node constructors an app may use to build custom pages in Go
(`El A AB AN Text Frag Render`).

**CLI.** `gotsx new build dev check export docs lsp tailwind version`, their documented flags (incl. `new --db
sqlite`, `export --out --base --site --routes --port`), `gotsx check --json`'s `[{file,line,col,msg}]` shape,
and the exit codes (0 ok, 1 diagnostics, 2 usage/setup error).

**Dev state files and agent docs.** `.gotsx/dev.json` (`pid port url started`) and `.gotsx/diagnostics.json`
(`title`, `errors: [{file,line,col,msg}]`, `text`) written by `gotsx dev`; the `AGENTS.md` managed-block markers
`<!-- BEGIN:gotsx-agent-rules -->` / `<!-- END:gotsx-agent-rules -->` (content inside the block evolves; content
outside it is never touched); `app/.gen/docs/*.md` file names.

**HTTP surface.** `/_gotsx/*` is reserved; `/public/*` serves the app's public dir; `/healthz` `/readyz`;
security defaults (CSP nonce, CSRF same-origin check on unsafe methods, headers) are on unless disabled through
`Options`.

### Experimental — may change in a minor release (announced in the changelog)

- `<Suspense fallback>` streaming boundaries: the boundary semantics are settled, the wire format
  (`<gotsx-suspense>` + `<template data-gotsx-fill>`) is internal and may change.
- `gotsx lsp` beyond diagnostics (hover, definition).
- Client telemetry payload (`ClientEvent`) fields beyond `type` / `message` / `url`.
- The dev live-reload protocol (`/_gotsx/dev`, incl. the `diag` event) and the error overlay markup.
- The typed-action wire format (`POST /_gotsx/act/<module>/<name>`, JSON argument array, `{"ok","data"}` /
  `{"error","fields"}` bodies, the `X-Gotsx-Action` header): only the compiled client stub and `gen.HostActions`
  are expected to speak it.
- The session cookie encoding (`gotsx_session`, base64 JSON + HMAC): rotate `SessionSecret` to invalidate.
- `design/` (the demo design system) — it is a demo artifact, not a framework API.

### Internal — no guarantees

- Everything in `gotsx/runtime` that only generated code calls (`Map Filter Dyn Nodes Island Must Push Re …`),
  the hydration marker format, `runtime.js` / `loader.js` exports, `host.json`, the AST and checker packages.
  Import `github.com/childrentime/gotsx/compiler` at your own risk; `compiler.Analyze` and `compiler.Diagnostic`
  are the only functions intended for tools.

## What "1.0" will mean

Stable items frozen for the 1.x line, Suspense and the LSP promoted, an independent security review of the
runtime's HTTP layer and escaping, and incremental compilation in `gotsx dev` treated as a performance contract.
Until then, the changelog is the contract: every user-visible change is listed there with its tier.
