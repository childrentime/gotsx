package main

import (
	"os"
	"path/filepath"

	"github.com/childrentime/gotsx/bench/gotsx/host"
	gotsx "github.com/childrentime/gotsx/runtime"
)

func main() {
	dir := filepath.Join("app", ".gen")
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}
	dts, js := gotsx.GenerateHost(host.Registry, "host")
	os.MkdirAll(dir, 0o755)
	os.WriteFile(filepath.Join(dir, "host.d.ts"), []byte(dts), 0o644)
	os.WriteFile(filepath.Join(dir, "host.json"), []byte(js), 0o644)
}
