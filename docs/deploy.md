# Deploying a gotsx app

A gotsx app is one static binary: the client runtime, the compiled islands and `public/` are embedded with
`go:embed`, so a deploy is "copy the binary somewhere and run it". Nothing else is needed at runtime.

## Bare binary

```bash
gotsx build            # app/ → gen/  (also runs Tailwind if app/tailwind.css exists)
CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o myapp .
scp myapp server:/srv/myapp && ssh server /srv/myapp -addr :3000
```

Run it behind any TLS terminator (Caddy, nginx, a cloud load balancer). The server sets security headers, a CSP with
per-response nonces, gzip, immutable caching for hashed assets, `/healthz` and `/readyz`, and shuts down gracefully
on `SIGTERM`. If the proxy terminates TLS, forward `X-Forwarded-Proto: https` so canonical / hreflang URLs are right.

## Docker

The repository `Dockerfile` builds any demo (or an app laid out like one) into a distroless image (~15 MB):

```bash
docker build --build-arg APP=shop -t gotsx-shop .
docker run --rm -p 3000:3000 gotsx-shop
```

For your own app, copy the `Dockerfile` and change `APP` to `.`.

## Cloudflare Workers (Go → Wasm, no container)

`deploy/cloudflare/` runs the shop demo as a Worker: `gotsx.Handler(server.Options(...))` handed to
[`syumai/workers`](https://github.com/syumai/workers) and compiled with `GOOS=js GOARCH=wasm` (~8 MB Wasm,
~2.3 MB compressed — under the free plan's 3 MB).

```bash
cd deploy/cloudflare
npx wrangler login      # once
make dev                # local workerd on :8787
make deploy             # → gotsx-shop.<subdomain>.workers.dev
```

Or from GitHub: add the `CLOUDFLARE_API_TOKEN` / `CLOUDFLARE_ACCOUNT_ID` secrets and run the
"Deploy shop to Cloudflare Workers" workflow. Caveats: request memory is per isolate (use D1/KV for real state),
`<Suspense>` streaming degrades to a buffered response, and the size limit rules out very large apps unless you use
TinyGo. For your own app, copy the directory and point `main.go` at your app's options.

## Fly.io

```bash
fly launch --copy-config --no-deploy   # picks up fly.toml
fly deploy
```

`fly.toml` runs one `shared-cpu-1x` machine with `auto_stop_machines`; the demo idles at ~10 MB RSS and cold-starts
in well under a second, so scale-to-zero is fine.

## Render

`render.yaml` is a Blueprint for the same Docker image on the free plan. Import the repository in the Render
dashboard and it deploys the `shop` demo.

## Static export (GitHub Pages, Netlify, S3, any file host)

For pages that need no per-request server state, `gotsx export` turns the app into a directory of HTML:

```bash
gotsx export . --out dist --base /my-repo --site https://user.github.io   # project page under a subpath
gotsx export . --out dist                                                   # root site
```

It builds, starts the binary on a local port, crawls every same-origin link from `/` (locale alternates from
`hreflang` included; or pass `--routes /,/docs,...`), rewrites root-relative URLs and absolute local URLs
(`hreflang`, `canonical`) to `--site` + `--base`, copies `gen/client` → `_gotsx/` and `public/`, and writes
`404.html` + `.nojekyll`. Islands keep working (their JS is static); typed actions and forms need a server.
This repository's docs site is deployed exactly this way (`.github/workflows/pages.yml`).

## Configuration

- `SESSION_SECRET` — the HMAC key for the signed session cookie (`Options.SessionSecret`). Required in production for `props.session` / flash / CSRF tokens to survive restarts and to be shared between replicas.

- `-addr :3000` — listen address (flag, the demos also honour it).
- `-dev` — development mode only (detailed errors, live-reload endpoint). Never in production.
- Environment is read by your `main.go`/host code as usual; gotsx itself needs no environment variables.
