// gotsx 官网: 用 gotsx 自己搭的。
package main

import (
	"embed"
	"flag"
	"io/fs"
	"log"

	gotsx "gotsx/runtime"
	"gotsx/site/gen"
)

//go:embed public
var publicEmbed embed.FS

func main() {
	addr := flag.String("addr", ":3000", "监听地址")
	dev := flag.Bool("dev", false, "开发模式")
	flag.Parse()
	log.Fatal(gotsx.Serve(gotsx.Options{
		Addr:      *addr,
		Dev:       *dev,
		Routes:    gen.Routes,
		ClientDir: gotsx.FindDir("gen/client"),
		ClientFS:  gen.ClientFS,
		PublicFS:  mustSub(publicEmbed, "public"),
		PublicDir: gotsx.FindDir("public"),
	}))
}

func mustSub(f embed.FS, dir string) fs.FS {
	sub, err := fs.Sub(f, dir)
	if err != nil {
		panic(err)
	}
	return sub
}
