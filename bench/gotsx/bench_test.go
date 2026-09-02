package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/childrentime/gotsx/bench/gotsx/gen"
	gotsx "github.com/childrentime/gotsx/runtime"
)

// BenchmarkPage drives the full handler chain (middleware + render + doc injection) in-process.
// go test -bench Page -benchmem -cpuprofile cpu.out ./gotsx && go tool pprof -top gotsx.test cpu.out
func BenchmarkPage(b *testing.B) {
	h := gotsx.Handler(gotsx.Options{Routes: gen.Routes, ClientDir: "gen/client", ClientFS: gen.ClientFS, PublicDir: "public", QuietLogs: true})
	req := httptest.NewRequest("GET", "/", nil)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			b.Fatal(rec.Code)
		}
	}
}

// BenchmarkRender isolates the renderer: Node tree construction + string output, no HTTP.
func BenchmarkRender(b *testing.B) {
	props := gotsx.PageProps{Params: map[string]string{}, Query: map[string]string{}, Path: "/"}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		gotsx.Render(gen.Routes[0].Render(props))
	}
}
