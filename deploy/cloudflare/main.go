// Cloudflare Worker entry for the shop demo: the same gotsx.Options as shop/main.go, served through
// github.com/syumai/workers (Go compiled to Wasm). Build with `make build`, run locally with `npx wrangler dev`,
// publish with `npx wrangler deploy`. The Wasm is ~8 MB (≈2.2 MB compressed), within the free plan's 3 MB limit.
package main

import (
	"embed"
	"io/fs"

	gotsx "github.com/childrentime/gotsx/runtime"
	"github.com/childrentime/gotsx/shop/server"
	"github.com/syumai/workers"
	"github.com/syumai/workers/cloudflare"
)

// public/ is copied from ../../shop/public by `make build` (go:embed cannot reach outside the package directory).
//
//go:embed public
var publicEmbed embed.FS

func main() {
	public, _ := fs.Sub(publicEmbed, "public")
	opt := server.Options(false, public)
	// Workers have no process environment: the session key comes from a Worker secret
	// (`npx wrangler secret put SESSION_SECRET`). Without it every isolate signs with its own random key
	// and sessions / flash messages do not survive across isolates.
	opt.SessionSecret = cloudflare.Getenv("SESSION_SECRET")
	workers.Serve(gotsx.Handler(opt))
}
