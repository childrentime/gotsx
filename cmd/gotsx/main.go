// gotsx: the dialect compiler, dev loop, project scaffolding and editor server.
//
//	gotsx new <dir>          scaffold a new app (own Go module, host module, pages, islands, tsconfig)
//	gotsx build [appdir]     compile app/ → gen/ (hostgen → tailwind → dialect → assets)
//	gotsx dev   [appdir]     build + go build + run; watches app/ host/ public/ and restarts (browser auto-reloads)
//	gotsx check [appdir]     type-check only, print diagnostics as file:line:col (--json for machines); exit 1 on errors
//	gotsx export [appdir]    static export: build, run, crawl every route, rewrite --base, copy assets into --out
//	gotsx docs [name]        print the version-matched docs (the same files gotsx build writes to app/.gen/docs/)
//	gotsx lsp                Language Server Protocol over stdio (diagnostics for editors)
//	gotsx tailwind           download the Tailwind standalone binary into .tools/ (no Node needed)
//	gotsx version            print the version
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime/debug"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/childrentime/gotsx/client"
	"github.com/childrentime/gotsx/compiler"
)

const usage = `gotsx — a TSX dialect that compiles to native Go

Usage:
  gotsx new <dir> [--module name] [--replace path] [--tailwind] [--db sqlite]
  gotsx build [appdir]
  gotsx dev   [appdir] [-addr :3000] [app flags...]
  gotsx check [appdir] [--json]
  gotsx export [appdir] [--out dist] [--base /subpath] [--site https://host] [--routes /,/docs] [--port 4123]
  gotsx docs [index|language|conventions|runtime|errors|agent-workflow]
  gotsx lsp
  gotsx tailwind [--dir .tools]
  gotsx version

appdir defaults to the current directory. It must contain app/ (pages, components, islands)
and a Go module (go.mod here or in a parent); gotsx.json is optional and only overrides
the inferred "module" / "hostPackage".`

// appConfig: gotsx.json (optional). Everything can be inferred from go.mod.
type appConfig struct {
	Module      string `json:"module"`      // Go import path of the app package (default: from go.mod)
	HostPackage string `json:"hostPackage"` // Go import path of the host package (default: <module>/host if host/ exists)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println(usage)
		os.Exit(2)
	}
	args := os.Args[2:]
	switch os.Args[1] {
	case "new":
		if err := newApp(args); err != nil {
			fail(err)
		}
	case "build":
		dir := appDirArg(args)
		cfg, err := readConfig(dir)
		if err != nil {
			fail(err)
		}
		if err := build(dir, cfg, false); err != nil {
			fail(err)
		}
	case "dev":
		dir := appDirArg(args)
		cfg, err := readConfig(dir)
		if err != nil {
			fail(err)
		}
		rest := args
		if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
			rest = args[1:]
		}
		dev(dir, cfg, rest)
	case "export":
		dir, o, err := parseExportArgs(args)
		if err != nil {
			fail(fmt.Errorf("%v\n%s", err, usage))
		}
		cfg, err := readConfig(dir)
		if err != nil {
			fail(err)
		}
		if err := export(dir, cfg, o); err != nil {
			fail(err)
		}
	case "docs":
		name := "index"
		if len(args) > 0 {
			name = args[0]
		}
		if err := printDoc(name); err != nil {
			fail(err)
		}
	case "check":
		dir := appDirArg(args)
		asJSON := false
		for _, a := range args {
			if a == "--json" {
				asJSON = true
			}
		}
		os.Exit(check(dir, asJSON))
	case "lsp":
		if err := runLSP(); err != nil {
			fail(err)
		}
	case "tailwind":
		dir := ".tools"
		for i, a := range args {
			if a == "--dir" && i+1 < len(args) {
				dir = args[i+1]
			}
		}
		if err := downloadTailwind(dir); err != nil {
			fail(err)
		}
	case "version", "--version", "-v":
		fmt.Println("gotsx", version())
	case "help", "--help", "-h":
		fmt.Println(usage)
	default:
		fmt.Fprintln(os.Stderr, "unknown command:", os.Args[1])
		fmt.Println(usage)
		os.Exit(2)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

// appDirArg: first non-flag argument, else "."
func appDirArg(args []string) string {
	dir := "."
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		dir = args[0]
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		fail(err)
	}
	return abs
}

// version: the module version this binary was built from ("(devel)" for go run inside the repo)
func version() string {
	if bi, ok := debug.ReadBuildInfo(); ok && bi.Main.Version != "" {
		return bi.Main.Version
	}
	return "(devel)"
}

// runtimeImport: the Go import path of the gotsx runtime, derived from this binary's own module path,
// so a fork keeps working and the generated code always imports the gotsx it was built with.
func runtimeImport() string {
	if bi, ok := debug.ReadBuildInfo(); ok && bi.Main.Path != "" {
		return bi.Main.Path + "/runtime"
	}
	return "github.com/childrentime/gotsx/runtime"
}

// frameworkModule: this binary's module path (e.g. github.com/childrentime/gotsx)
func frameworkModule() string {
	return strings.TrimSuffix(runtimeImport(), "/runtime")
}

// ---------- config ----------

// readConfig resolves the app's Go import path and host package.
// Order: gotsx.json values → inferred from the nearest go.mod → host/ directory presence.
func readConfig(dir string) (appConfig, error) {
	var cfg appConfig
	if b, err := os.ReadFile(filepath.Join(dir, "gotsx.json")); err == nil {
		if err := json.Unmarshal(b, &cfg); err != nil {
			return cfg, fmt.Errorf("gotsx.json: %w", err)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "app")); err != nil {
		return cfg, fmt.Errorf("%s: no app/ directory (run `gotsx new <dir>` to scaffold one)", dir)
	}
	if cfg.Module == "" {
		mod, modDir, err := findGoModule(dir)
		if err != nil {
			return cfg, err
		}
		rel, _ := filepath.Rel(modDir, dir)
		cfg.Module = mod
		if rel != "." && rel != "" {
			cfg.Module = mod + "/" + filepath.ToSlash(rel)
		}
	}
	if cfg.HostPackage == "" {
		if st, err := os.Stat(filepath.Join(dir, "host")); err == nil && st.IsDir() {
			cfg.HostPackage = cfg.Module + "/host"
		}
	}
	return cfg, nil
}

// findGoModule walks up from dir to the nearest go.mod and returns (module path, its directory)
func findGoModule(dir string) (string, string, error) {
	for d := dir; ; d = filepath.Dir(d) {
		b, err := os.ReadFile(filepath.Join(d, "go.mod"))
		if err == nil {
			for _, ln := range strings.Split(string(b), "\n") {
				ln = strings.TrimSpace(ln)
				if strings.HasPrefix(ln, "module ") {
					return strings.TrimSpace(strings.TrimPrefix(ln, "module ")), d, nil
				}
			}
			return "", "", fmt.Errorf("%s: no module line", filepath.Join(d, "go.mod"))
		}
		if filepath.Dir(d) == d {
			return "", "", fmt.Errorf("%s: no go.mod found here or in any parent (add one, or a gotsx.json with \"module\")", dir)
		}
	}
}

// ---------- build ----------

// build: hostgen ∥ tailwind → dialect compiler. incremental (dev loop) skips hostgen while host/ is unchanged —
// it is a `go run`, the slowest step — so a TSX edit rebuilds in the time of the compiler alone.
func build(dir string, cfg appConfig, incremental bool) error {
	t := time.Now()
	var wg sync.WaitGroup
	var hostErr, twErr error
	var hostMs, twMs time.Duration
	hostgen := "skipped"
	if _, err := os.Stat(filepath.Join(dir, "cmd", "hostgen")); err == nil && (!incremental || hostStale(dir)) {
		hostgen = "ran"
		wg.Add(1)
		go func() {
			defer wg.Done()
			t0 := time.Now()
			cmd := exec.Command("go", "run", "./cmd/hostgen", filepath.Join("app", ".gen"))
			cmd.Dir = dir
			if out, err := cmd.CombinedOutput(); err != nil {
				hostErr = fmt.Errorf("hostgen failed: %v\n%s", err, out)
			}
			hostMs = time.Since(t0)
		}()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		t0 := time.Now()
		twErr = runTailwind(dir)
		twMs = time.Since(t0)
	}()
	wg.Wait()
	if hostErr != nil {
		return hostErr
	}
	if twErr != nil {
		return twErr
	}
	t1 := time.Now()
	rep, err := compiler.Build(compiler.Config{
		AppDir:     filepath.Join(dir, "app"),
		OutDir:     filepath.Join(dir, "gen"),
		Module:     cfg.Module,
		HostPkg:    cfg.HostPackage,
		RuntimePkg: runtimeImport(),
		ClientFS:   client.FS,
	})
	if err != nil {
		return err
	}
	ensureAgentFiles(dir) // AGENTS.md managed block + app/.gen/docs/ (version-matched docs for agents)
	steps := fmt.Sprintf("compile %s", time.Since(t1).Round(time.Millisecond))
	if hostgen == "ran" {
		steps = fmt.Sprintf("hostgen %s ∥ ", hostMs.Round(time.Millisecond)) + steps
	} else if incremental {
		steps = "hostgen skipped (host/ unchanged) · " + steps
	}
	if twMs > 0 && fileExists(filepath.Join(dir, "app", "tailwind.css")) {
		steps = fmt.Sprintf("tailwind %s ∥ ", twMs.Round(time.Millisecond)) + steps
	}
	fmt.Printf("gotsx: compiled %d modules → routes %v, islands %v (%s: %s)\n", len(rep.Modules), rep.Routes, rep.Islands, time.Since(t).Round(time.Millisecond), steps)
	return nil
}

// hostStale: host.json is older than anything under host/, cmd/hostgen/ or go.mod (or missing)
func hostStale(dir string) bool {
	st, err := os.Stat(filepath.Join(dir, "app", ".gen", "host.json"))
	if err != nil {
		return true
	}
	return latestMtime(filepath.Join(dir, "host"), filepath.Join(dir, "cmd", "hostgen"), filepath.Join(dir, "go.mod")).After(st.ModTime())
}

// check: analyze only; prints diagnostics. Returns the process exit code.
func check(dir string, asJSON bool) int {
	if _, err := os.Stat(filepath.Join(dir, "app")); err != nil {
		fmt.Fprintf(os.Stderr, "%s: no app/ directory\n", dir)
		return 2
	}
	diags, err := compiler.Analyze(filepath.Join(dir, "app"), nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	if asJSON {
		if diags == nil {
			diags = []compiler.Diagnostic{}
		}
		json.NewEncoder(os.Stdout).Encode(diags)
	} else {
		for _, d := range diags {
			rel, err := filepath.Rel(dir, d.File)
			if err != nil {
				rel = d.File
			}
			fmt.Printf("%s:%d:%d: %s\n", rel, d.Line, d.Col, d.Msg)
		}
		if len(diags) == 0 {
			fmt.Println("gotsx: ok")
		}
	}
	if len(diags) > 0 {
		return 1
	}
	return 0
}

// ---------- dev ----------

func dev(dir string, cfg appConfig, appArgs []string) {
	if st, alive := readDevState(dir); alive {
		fmt.Fprintf(os.Stderr, "gotsx dev is already running for this app: pid %d, %s (started %s).\nStop it first, or use it — .gotsx/dev.json is the source of truth.\n", st.PID, st.URL, st.Started)
		os.Exit(1)
	}
	port := portOf(appArgs)
	if err := writeDevState(dir, port); err != nil {
		fail(err)
	}
	var proc *exec.Cmd
	stop := func() {
		if proc != nil && proc.Process != nil {
			proc.Process.Kill()
			proc.Wait()
			proc = nil
		}
	}
	cleanup := func() {
		stop()
		os.Remove(devStatePath(dir))
	}
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sig
		cleanup()
		os.Exit(0)
	}()
	defer cleanup()
	gen := 0
	run := func() {
		// build first; only swap the process on success so the old server keeps serving during a compile error
		if err := build(dir, cfg, gen > 0); err != nil {
			fmt.Fprintln(os.Stderr, "\n"+err.Error()+"\n(the previous build keeps running; the browser shows the errors, so does .gotsx/diagnostics.json)")
			writeDiagnostics(dir, "gotsx: build failed", err)
			return
		}
		gen++
		bin := filepath.Join(dir, ".gotsx", fmt.Sprintf("app-%d%s", gen%2, exeSuffix())) // alternate names: never overwrite the running binary
		os.MkdirAll(filepath.Dir(bin), 0o755)
		t := time.Now()
		gb := exec.Command("go", "build", "-o", bin, ".")
		gb.Dir = dir
		if out, err := gb.CombinedOutput(); err != nil {
			fmt.Fprintf(os.Stderr, "\ngo build failed:\n%s\n(the previous build keeps running)\n", out)
			writeDiagnostics(dir, "go build failed", fmt.Errorf("%s", out))
			return
		}
		clearDiagnostics(dir)
		fmt.Printf("gotsx: go build %s\n", time.Since(t).Round(time.Millisecond))
		stop()
		proc = exec.Command(bin, append([]string{"-dev"}, appArgs...)...)
		proc.Dir = dir
		proc.Stdout, proc.Stderr = os.Stdout, os.Stderr
		if err := proc.Start(); err != nil {
			fmt.Fprintln(os.Stderr, "start failed:", err)
		} else if gen == 1 {
			fmt.Printf("gotsx: dev server → http://localhost:%d (state in .gotsx/dev.json)\n", port)
		}
	}
	run()
	watch := []string{filepath.Join(dir, "app"), filepath.Join(dir, "host"), filepath.Join(dir, "public"), filepath.Join(dir, "main.go")}
	last := latestMtime(watch...)
	for {
		time.Sleep(400 * time.Millisecond)
		now := latestMtime(watch...)
		if now.After(last) {
			last = now
			fmt.Println("gotsx: source changed, rebuilding…")
			run()
		}
	}
}

func exeSuffix() string {
	if isWindows() {
		return ".exe"
	}
	return ""
}

func latestMtime(paths ...string) time.Time {
	var latest time.Time
	genDir := string(filepath.Separator) + ".gen"
	for _, d := range paths {
		filepath.WalkDir(d, func(p string, e os.DirEntry, err error) error {
			if err != nil || strings.Contains(p, genDir) {
				return nil
			}
			if info, err := e.Info(); err == nil && info.ModTime().After(latest) {
				latest = info.ModTime()
			}
			return nil
		})
	}
	return latest
}

// ---------- tailwind ----------

// runTailwind: if app/tailwind.css exists, run the standalone CLI (no Node) to scan app/**/*.tsx into public/tailwind.css.
// Binary lookup: $GOTSX_TAILWIND → <dir or any parent>/.tools/tailwindcss[.exe] → tailwindcss on PATH
func runTailwind(dir string) error {
	in := filepath.Join(dir, "app", "tailwind.css")
	if _, err := os.Stat(in); err != nil {
		return nil
	}
	bin := findTailwind(dir)
	if bin == "" {
		fmt.Fprintln(os.Stderr, "gotsx: app/tailwind.css found but no tailwindcss binary (run `gotsx tailwind`, or set GOTSX_TAILWIND); skipping")
		return nil
	}
	t := time.Now()
	os.MkdirAll(filepath.Join(dir, "public"), 0o755)
	cmd := exec.Command(bin, "-i", in, "-o", filepath.Join(dir, "public", "tailwind.css"), "--minify")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("tailwind failed: %v\n%s", err, out)
	}
	_ = t
	return nil
}

func findTailwind(dir string) string {
	if bin := os.Getenv("GOTSX_TAILWIND"); bin != "" {
		return bin
	}
	for d := dir; ; d = filepath.Dir(d) {
		for _, name := range []string{"tailwindcss", "tailwindcss.exe"} {
			if p := filepath.Join(d, ".tools", name); fileExists(p) {
				return p
			}
		}
		if filepath.Dir(d) == d {
			break
		}
	}
	if p, err := exec.LookPath("tailwindcss"); err == nil {
		return p
	}
	return ""
}

func fileExists(p string) bool {
	st, err := os.Stat(p)
	return err == nil && !st.IsDir()
}
