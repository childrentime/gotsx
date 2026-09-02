package compiler

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// 集成测试: 编译三个真实应用并 go build + go vet, 作为回归网。慢, -short 跳过。
func TestAppsBuild(t *testing.T) {
	if testing.Short() {
		t.Skip("-short: 跳过应用集成构建")
	}
	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	for _, app := range []string{"example", "site", "shop", "admin"} {
		app := app
		t.Run(app, func(t *testing.T) {
			run := func(name string, args ...string) {
				cmd := exec.Command(name, args...)
				cmd.Dir = root
				if out, err := cmd.CombinedOutput(); err != nil {
					t.Fatalf("%s %v 失败: %v\n%s", name, args, err, out)
				}
			}
			// hostgen(如有) → gotsx build → go build → go vet
			if _, err := os.Stat(filepath.Join(root, app, "cmd", "hostgen")); err == nil {
				cmd := exec.Command("go", "run", "./cmd/hostgen", "app/.gen")
				cmd.Dir = filepath.Join(root, app)
				if out, err := cmd.CombinedOutput(); err != nil {
					t.Fatalf("hostgen 失败: %v\n%s", err, out)
				}
			}
			run("go", "run", "./cmd/gotsx", "build", app)
			run("go", "build", "-o", os.DevNull, "./"+app+"/")
			run("go", "vet", "./"+app+"/gen/")
		})
	}
}
