package main

import (
	"encoding/json"
	"fmt"
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

func TestExportHelpers(t *testing.T) {
	dir, o, err := parseExportArgs([]string{"site", "--base", "gotsx/", "--routes", "/, /docs", "--port", "5000"})
	if err != nil || !strings.HasSuffix(dir, "site") || o.base != "/gotsx" || len(o.routes) != 2 || o.port != 5000 || !strings.HasSuffix(o.out, "dist") {
		t.Fatalf("parseExportArgs: %s %+v %v", dir, o, err)
	}
	rw := rewriter("/gotsx", "http://127.0.0.1:4123", "https://u.github.io")
	in := `<a href="/docs">x</a><a href="//cdn/x">y</a><script src="/_gotsx/loader.js"></script><script>window.__GOTSX={"dev":false}</script><form action="/todos"><link hreflang="zh" href="http://127.0.0.1:4123/zh"><link rel="canonical" href="http://127.0.0.1:4123/">`
	got := string(rw([]byte(in)))
	for _, want := range []string{`href="/gotsx/docs"`, `href="//cdn/x"`, `src="/gotsx/_gotsx/loader.js"`, `window.__GOTSX={"base":"/gotsx","dev":false}`, `action="/gotsx/todos"`, `href="https://u.github.io/gotsx/zh"`, `href="https://u.github.io/gotsx/"`} {
		if !strings.Contains(got, want) {
			t.Errorf("rewrite: want %q in %s", want, got)
		}
	}
	plain := string(rewriter("", "http://127.0.0.1:4123", "")([]byte(in)))
	if !strings.Contains(plain, `href="/docs"`) || !strings.Contains(plain, `href="/zh"`) || strings.Contains(plain, "127.0.0.1") {
		t.Errorf("empty base: only the origin is rewritten: %s", plain)
	}
	ls := links([]byte(`<a href="/docs/">a</a><a href="/public/x.css">b</a><a href="/_gotsx/x">c</a><link rel="alternate" href="/zh/docs?x=1#f"><a href="/zh">z</a><link href="http://127.0.0.1:4123/zh/docs"><a href="https://example.com/x">ext</a><a href="http://127.0.0.1:4123">root</a>`), "http://127.0.0.1:4123")
	if strings.Join(ls, ",") != "/docs,/zh/docs,/zh,/zh/docs,/" {
		t.Errorf("links: %v", ls)
	}
}

func TestAgentFilesUpsert(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "app"), 0o755)
	// a directory without go.mod (a demo inside the repo): no AGENTS.md is created, but the docs are written
	ensureAgentFiles(dir)
	if _, err := os.Stat(filepath.Join(dir, "AGENTS.md")); err == nil {
		t.Error("must not create AGENTS.md outside a module root")
	}
	if _, err := os.Stat(filepath.Join(dir, "app", ".gen", "docs", "index.md")); err != nil {
		t.Error("docs should be written:", err)
	}
	// module root: AGENTS.md + CLAUDE.md are created
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module x\n"), 0o644)
	ensureAgentFiles(dir)
	raw, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if err != nil || !strings.Contains(string(raw), agentBlockBegin) || !strings.Contains(string(raw), "not React or Next.js") {
		t.Fatalf("AGENTS.md: %v\n%s", err, raw)
	}
	if c, _ := os.ReadFile(filepath.Join(dir, "CLAUDE.md")); string(c) != claudeFile {
		t.Errorf("CLAUDE.md = %q", c)
	}
	// user content is kept, the old block is replaced
	custom := "# my notes\nkeep me\n" + agentBlockBegin + "\nold rules\n" + agentBlockEnd + "\ntail\n"
	os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte(custom), 0o644)
	os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("mine\n"), 0o644)
	ensureAgentFiles(dir)
	raw, _ = os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	s := string(raw)
	if !strings.HasPrefix(s, "# my notes\nkeep me\n") || strings.Contains(s, "old rules") || !strings.HasSuffix(s, "\ntail\n") || strings.Count(s, agentBlockBegin) != 1 {
		t.Errorf("upsert kept/removed the wrong parts:\n%s", s)
	}
	if c, _ := os.ReadFile(filepath.Join(dir, "CLAUDE.md")); string(c) != "mine\n" {
		t.Error("existing CLAUDE.md must not be touched")
	}
	// an existing file without a block: the block is prepended
	os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("# existing\n"), 0o644)
	ensureAgentFiles(dir)
	raw, _ = os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if !strings.HasPrefix(string(raw), agentBlockBegin) || !strings.HasSuffix(string(raw), "# existing\n") {
		t.Errorf("prepend: %s", raw)
	}
	// idempotent: a second run leaves the content unchanged
	ensureAgentFiles(dir)
	if again, _ := os.ReadFile(filepath.Join(dir, "AGENTS.md")); string(again) != string(raw) {
		t.Error("not idempotent")
	}
}

func TestDevStateHelpers(t *testing.T) {
	if p := portOf(nil); p != 3000 {
		t.Errorf("default port %d", p)
	}
	if p := portOf([]string{"-addr", ":8080"}); p != 8080 {
		t.Errorf("-addr: %d", p)
	}
	if p := portOf([]string{"--addr=127.0.0.1:9000"}); p != 9000 {
		t.Errorf("--addr=: %d", p)
	}
	dir := t.TempDir()
	if err := writeDevState(dir, 3456); err != nil {
		t.Fatal(err)
	}
	st, alive := readDevState(dir)
	if !alive || st.Port != 3456 || st.URL != "http://localhost:3456" || st.PID != os.Getpid() {
		t.Errorf("dev state: %+v alive=%v", st, alive)
	}
	os.WriteFile(devStatePath(dir), []byte(`{"pid": 999999999, "port": 1}`), 0o644)
	if _, alive := readDevState(dir); alive {
		t.Error("dead pid must not count as running")
	}
	ds := parseGoErrors(dir, "# x\nhost/host.go:12:5: undefined: foo\nsome other line\n")
	if len(ds) != 1 || ds[0].Line != 12 || ds[0].Col != 5 || ds[0].Msg != "undefined: foo" || !strings.HasSuffix(ds[0].File, filepath.Join("host", "host.go")) {
		t.Errorf("parseGoErrors: %+v", ds)
	}
	writeDiagnostics(dir, "go build failed", fmt.Errorf("host/host.go:12:5: undefined: foo"))
	raw, err := os.ReadFile(diagPath(dir))
	if err != nil || !strings.Contains(string(raw), `"undefined: foo"`) || !strings.Contains(string(raw), `"title": "go build failed"`) {
		t.Errorf("diagnostics.json: %v %s", err, raw)
	}
	clearDiagnostics(dir)
	if _, err := os.Stat(diagPath(dir)); err == nil {
		t.Error("clearDiagnostics should remove the file")
	}
}

func TestExportSafety(t *testing.T) {
	app := t.TempDir()
	os.WriteFile(filepath.Join(app, "main.go"), []byte("package main\n"), 0o644)
	for _, out := range []string{app, filepath.Dir(app)} {
		if err := prepareOut(app, out); err == nil {
			t.Errorf("prepareOut must refuse %s", out)
		}
	}
	other := filepath.Join(t.TempDir(), "notes")
	os.MkdirAll(other, 0o755)
	os.WriteFile(filepath.Join(other, "keep.txt"), []byte("x"), 0o644)
	if err := prepareOut(app, other); err == nil {
		t.Error("a non-empty directory that is not an export must be refused")
	}
	if _, err := os.Stat(filepath.Join(other, "keep.txt")); err != nil {
		t.Error("refusal must not delete anything")
	}
	prev := filepath.Join(t.TempDir(), "dist")
	os.MkdirAll(filepath.Join(prev, "_gotsx"), 0o755)
	os.WriteFile(filepath.Join(prev, ".nojekyll"), nil, 0o644)
	os.WriteFile(filepath.Join(prev, "index.html"), []byte("old"), 0o644)
	if err := prepareOut(app, prev); err != nil {
		t.Fatalf("a previous export should be replaceable: %v", err)
	}
	if _, err := os.Stat(filepath.Join(prev, "index.html")); err == nil {
		t.Error("previous export should have been emptied")
	}
	fresh := filepath.Join(t.TempDir(), "new", "dist")
	if err := prepareOut(app, fresh); err != nil {
		t.Errorf("missing --out should be created: %v", err)
	} else if st, serr := os.Stat(fresh); serr != nil || !st.IsDir() {
		t.Error("missing --out should exist as a directory")
	}
	ls := links([]byte(`<a href="/blog/v1.2">a</a><a href="/users/jane.doe">b</a><a href="/public/x.css">c</a><a href="/img/logo.svg">d</a>`), "")
	if strings.Join(ls, ",") != "/blog/v1.2,/users/jane.doe" {
		t.Errorf("links with dots in segments: %v", ls)
	}
	ds := parseGoErrors(`C:\`, `C:\app\host\host.go:12:5: undefined: foo`)
	if len(ds) != 1 || ds[0].Line != 12 || !strings.HasSuffix(ds[0].File, "host.go") {
		t.Errorf("windows path: %+v", ds)
	}
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module x\n"), 0o644)
	os.MkdirAll(filepath.Join(dir, "app"), 0o755)
	os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("# x\n"+agentBlockEnd+"\nstuff\n"+agentBlockBegin+"\n"), 0o644)
	ensureAgentFiles(dir) // malformed block (END before BEGIN): the file must still end up with exactly one valid block
	raw, _ := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if strings.Count(string(raw), agentBlockBegin) < 1 || strings.Index(string(raw), agentBlockBegin) > strings.Index(string(raw), agentBlockEnd) {
		t.Errorf("malformed block not repaired:\n%s", raw)
	}
	if err := printDoc("nope"); err == nil || !strings.Contains(err.Error(), "available") {
		t.Errorf("unknown doc: %v", err)
	}
}
