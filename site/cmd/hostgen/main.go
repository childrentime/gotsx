package main

import (
	"fmt"
	"os"
	"path/filepath"

	gotsx "github.com/childrentime/gotsx/runtime"
	"github.com/childrentime/gotsx/site/host"
)

func main() {
	dir := "app/.gen"
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}
	dts, js := gotsx.GenerateHost(host.Registry, "host")
	os.MkdirAll(dir, 0o755)
	if err := os.WriteFile(filepath.Join(dir, "host.d.ts"), []byte(dts), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := os.WriteFile(filepath.Join(dir, "host.json"), []byte(js), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
