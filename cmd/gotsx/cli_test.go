package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// 端到端: gotsx new → gotsx build → go build → gotsx check(含 --json)。慢, -short 跳过。
func TestNewBuildCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("-short")
	}
	root := t.TempDir()
	dir := filepath.Join(root, "my-app")
	gotsx := func(args ...string) (string, error) {
		cmd := exec.Command("go", append([]string{"run", "."}, args...)...)
		out, err := cmd.CombinedOutput()
		return string(out), err
	}
	if out, err := gotsx("new", dir); err != nil {
		t.Fatalf("gotsx new: %v\n%s", err, out)
	}
	for _, f := range []string{"go.mod", "main.go", "host/host.go", "cmd/hostgen/main.go", "app/pages/index.server.tsx", "app/islands/TodoList.client.tsx", "tsconfig.json"} {
		if _, err := os.Stat(filepath.Join(dir, f)); err != nil {
			t.Errorf("脚手架缺 %s", f)
		}
	}
	gm, _ := os.ReadFile(filepath.Join(dir, "go.mod"))
	if !strings.Contains(string(gm), "module my-app") || !strings.Contains(string(gm), "replace github.com/childrentime/gotsx =>") {
		t.Errorf("从检出运行时 go.mod 应带 replace:\n%s", gm)
	}
	if out, err := gotsx("build", dir); err != nil {
		t.Fatalf("gotsx build: %v\n%s", err, out)
	}
	for _, f := range []string{"gen/routes_gen.go", "gen/client/TodoList.js", "app/.gen/host.d.ts", "app/.gen/gotsx.d.ts"} {
		if _, err := os.Stat(filepath.Join(dir, f)); err != nil {
			t.Errorf("构建产物缺 %s", f)
		}
	}
	gb := exec.Command("go", "build", "-o", os.DevNull, ".")
	gb.Dir = dir
	if out, err := gb.CombinedOutput(); err != nil {
		t.Fatalf("go build: %v\n%s", err, out)
	}
	if out, err := gotsx("check", dir); err != nil || !strings.Contains(out, "gotsx: ok") {
		t.Fatalf("gotsx check: %v\n%s", err, out)
	}
	// 引入一个错误: check 报位置, --json 给机器
	page := filepath.Join(dir, "app", "pages", "index.server.tsx")
	src, _ := os.ReadFile(page)
	os.WriteFile(page, []byte(strings.Replace(string(src), "todos.list()", "todos.nope()", 1)), 0o644)
	out, err := gotsx("check", dir)
	if err == nil || !strings.Contains(out, "index.server.tsx:") || !strings.Contains(out, "nope") {
		t.Fatalf("check 应失败并指出位置: %v\n%s", err, out)
	}
	jc := exec.Command("go", "run", ".", "check", dir, "--json")
	stdout, _ := jc.Output() // stdout only: go run 会把 "exit status 1" 写到 stderr
	var diags []map[string]any
	if jerr := json.Unmarshal(stdout, &diags); jerr != nil || len(diags) != 1 || diags[0]["line"].(float64) < 1 {
		t.Fatalf("check --json 应输出诊断数组: %v\n%s", jerr, stdout)
	}
}

func TestReadConfigInfersFromGoMod(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/x\n\ngo 1.26\n"), 0o644)
	app := filepath.Join(root, "apps", "web")
	os.MkdirAll(filepath.Join(app, "app"), 0o755)
	os.MkdirAll(filepath.Join(app, "host"), 0o755)
	cfg, err := readConfig(app)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Module != "example.com/x/apps/web" || cfg.HostPackage != "example.com/x/apps/web/host" {
		t.Errorf("推断错误: %+v", cfg)
	}
	// gotsx.json 覆盖
	os.WriteFile(filepath.Join(app, "gotsx.json"), []byte(`{"hostPackage":"example.com/x/shared/host"}`), 0o644)
	cfg, _ = readConfig(app)
	if cfg.Module != "example.com/x/apps/web" || cfg.HostPackage != "example.com/x/shared/host" {
		t.Errorf("gotsx.json 应覆盖: %+v", cfg)
	}
}

func TestLSPHelpers(t *testing.T) {
	p := uriToPath("file:///tmp/a%20b/app/pages/x.server.tsx")
	if !strings.HasSuffix(p, filepath.Join("a b", "app", "pages", "x.server.tsx")) {
		t.Errorf("uriToPath: %q", p)
	}
	if appDirOf(p) != filepath.Join(string(filepath.Separator)+"tmp", "a b", "app") {
		t.Errorf("appDirOf: %q", appDirOf(p))
	}
	if appDirOf("/nowhere/x.tsx") != "" {
		t.Error("app 外的文件应返回空")
	}
	if u := pathToURI("/tmp/a b/x.tsx"); u != "file:///tmp/a%20b/x.tsx" {
		t.Errorf("pathToURI: %q", u)
	}
}
