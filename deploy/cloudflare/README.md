# Shop demo on Cloudflare Workers

The shop compiled to Wasm and served by a Worker — no container, no server to keep running.

```bash
cd deploy/cloudflare
npx wrangler login          # once
make dev                    # http://localhost:8787 (local workerd)
make deploy                 # publishes gotsx-shop.<your-subdomain>.workers.dev
```

How it works: `shop/server.Options()` is the same configuration `shop/main.go` uses; here it is handed to
`gotsx.Handler` and to [`syumai/workers`](https://github.com/syumai/workers), which adapts Cloudflare's
`fetch` event to `net/http`. `make build` compiles the dialect, copies `shop/public`, generates the JS shim
(`build/worker.mjs` + `wasm_exec.js`) and builds `build/app.wasm` (~8 MB, ~2.2 MB compressed; the free plan
allows 3 MB, paid 10 MB).

Differences from the native binary: Workers have no persistent memory between requests on different isolates, so
the demo's in-memory cart/orders survive only within one isolate; the `<Suspense>` streaming falls back to a
buffered response (the adapter has no `http.Flusher`); there is no `/healthz` TCP server (Workers do not need one).
For a persistent store, back the host module with D1 or KV — `syumai/workers` exposes both.

Continuous deploy: `.github/workflows/deploy-cloudflare.yml` builds and publishes on every push that touches the
shop; add the `CLOUDFLARE_API_TOKEN` and `CLOUDFLARE_ACCOUNT_ID` secrets and set the repository variable
`CLOUDFLARE_DEPLOY=true`.

## Sessions on Workers

The signed session cookie (flash messages after checkout, `props.session`) needs a fixed key shared by every isolate.
Set it once as a Worker secret; without it each isolate signs with its own random key:

```bash
npx wrangler secret put SESSION_SECRET
```
