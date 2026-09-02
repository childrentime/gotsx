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

## Fly.io (used for the hosted shop demo)

```bash
fly launch --copy-config --no-deploy   # picks up fly.toml
fly deploy
```

`fly.toml` runs one `shared-cpu-1x` machine with `auto_stop_machines`; the demo idles at ~10 MB RSS and cold-starts
in well under a second, so scale-to-zero is fine.

## Render

`render.yaml` is a Blueprint for the same Docker image on the free plan. Import the repository in the Render
dashboard and it deploys the `shop` demo.

## Configuration

- `-addr :3000` — listen address (flag, the demos also honour it).
- `-dev` — development mode only (detailed errors, live-reload endpoint). Never in production.
- Environment is read by your `main.go`/host code as usual; gotsx itself needs no environment variables.
