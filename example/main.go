// 示例应用: 路由来自编译器生成的 gen 包, 宿主模块在 host 包, 动作是普通 Go handler。
package main

import (
	"embed"
	"encoding/json"
	"flag"
	"io/fs"
	"log"
	"net/http"

	"github.com/childrentime/gotsx/example/gen"
	"github.com/childrentime/gotsx/example/host"
	gotsx "github.com/childrentime/gotsx/runtime"
)

//go:embed public
var publicEmbed embed.FS

func main() {
	addr := flag.String("addr", ":3000", "监听地址")
	dev := flag.Bool("dev", false, "开发模式")
	flag.Parse()
	err := gotsx.Serve(gotsx.Options{
		Addr:      *addr,
		Dev:       *dev,
		Routes:    gen.Routes,
		ClientDir: gotsx.FindDir("gen/client"),
		ClientFS:  gen.ClientFS,
		PublicFS:  mustSub(publicEmbed, "public"),
		PublicDir: gotsx.FindDir("public"),
		NotFound:  gen.NotFound,  // pages/_404.server.tsx
		ErrorPage: gen.ErrorPage, // pages/_error.server.tsx
		Actions: map[string]http.HandlerFunc{
			"POST /actions/like": func(w http.ResponseWriter, r *http.Request) {
				n, err := host.Data.Models.Like(r.URL.Query().Get("id"))
				w.Header().Set("Content-Type", "application/json")
				if err != nil {
					w.WriteHeader(404)
					json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
					return
				}
				json.NewEncoder(w).Encode(map[string]int{"likes": n})
			},
		},
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
