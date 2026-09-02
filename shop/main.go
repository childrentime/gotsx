// gomu: a Temu-style full-stack e-commerce demo. Pages are Go compiled from the dialect; writes are plain Go actions;
// the server definition lives in package server so the Cloudflare Worker (deploy/cloudflare) can reuse it.
package main

import (
	"embed"
	"flag"
	"io/fs"
	"log"

	gotsx "github.com/childrentime/gotsx/runtime"
	"github.com/childrentime/gotsx/shop/server"
)

//go:embed public
var publicEmbed embed.FS

func main() {
	addr := flag.String("addr", ":3000", "listen address")
	dev := flag.Bool("dev", false, "development mode (error pages with stacks, assets no-cache, live reload)")
	flag.Parse()
	public, _ := fs.Sub(publicEmbed, "public")
	opt := server.Options(*dev, public)
	opt.Addr = *addr
	opt.ClientDir = gotsx.FindDir("gen/client")
	opt.PublicDir = gotsx.FindDir("public")
	log.Fatal(gotsx.Serve(opt))
}
