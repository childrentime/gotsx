// 示例应用: 路由来自编译器生成的 gen 包, 宿主模块在 host 包, 岛通过类型化 action(gen.HostActions)回到 Go。
package main

import (
	"embed"
	"flag"
	"io/fs"
	"log"
	"os"

	"github.com/childrentime/gotsx/example/gen"
	gotsx "github.com/childrentime/gotsx/runtime"
)

//go:embed public
var publicEmbed embed.FS

func main() {
	addr := flag.String("addr", ":3000", "监听地址")
	dev := flag.Bool("dev", false, "开发模式")
	flag.Parse()
	err := gotsx.Serve(gotsx.Options{
		Addr:          *addr,
		Dev:           *dev,
		Routes:        gen.Routes,
		ClientDir:     gotsx.FindDir("gen/client"),
		ClientFS:      gen.ClientFS,
		PublicFS:      mustSub(publicEmbed, "public"),
		PublicDir:     gotsx.FindDir("public"),
		NotFound:      gen.NotFound,                // pages/_404.server.tsx
		ErrorPage:     gen.ErrorPage,               // pages/_error.server.tsx
		HostActions:   gen.HostActions,             // typed actions: DataModule.Like → islands call like(id)
		SessionSecret: os.Getenv("SESSION_SECRET"), // empty → random per start (sessions do not survive restarts)
	})
	log.Fatal(err)
}

func mustSub(f embed.FS, dir string) fs.FS {
	sub, err := fs.Sub(f, dir)
	if err != nil {
		panic(err)
	}
	return sub
}
