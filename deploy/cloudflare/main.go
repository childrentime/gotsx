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
)

// public/ is copied from ../../shop/public by `make build` (go:embed cannot reach outside the package directory).
//
//go:embed public
var publicEmbed embed.FS

func main() {
	public, _ := fs.Sub(publicEmbed, "public")
	workers.Serve(gotsx.Handler(server.Options(false, public)))
}
