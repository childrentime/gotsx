// hostgen: 反射 host.Registry → app/.gen/host.d.ts + host.json
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gotsx/example/host"
	gotsx "gotsx/runtime"
)

func main() {
	dir := "app/.gen"
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}
	dts, js := gotsx.GenerateHost(host.Registry, "host")
	os.MkdirAll(dir, 0o755)
	must(os.WriteFile(filepath.Join(dir, "host.d.ts"), []byte(dts), 0o644))
	must(os.WriteFile(filepath.Join(dir, "host.json"), []byte(js), 0o644))
	fmt.Println("hostgen: 写入", filepath.Join(dir, "host.d.ts"), "和 host.json")
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
