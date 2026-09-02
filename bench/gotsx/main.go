// gotsx: the same page as a gotsx app (production mode: CSP nonce, security headers, gzip when asked).
package main

import (
	"flag"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/childrentime/gotsx/bench/gotsx/gen"
	gotsx "github.com/childrentime/gotsx/runtime"
)

func main() {
	addr := flag.String("addr", ":3000", "")
	dev := flag.Bool("dev", false, "")
	flag.Parse()
	log.SetOutput(discard{}) // the other contenders do not log requests either
	if os.Getenv("PPROF") != "" {
		go http.ListenAndServe("127.0.0.1:6060", nil)
	}
	log.Fatal(gotsx.Serve(gotsx.Options{Addr: *addr, Dev: *dev, Routes: gen.Routes, ClientDir: gotsx.FindDir("gen/client"), ClientFS: gen.ClientFS, PublicDir: "public"}))
}

type discard struct{}

func (discard) Write(p []byte) (int, error) { return len(p), nil }
