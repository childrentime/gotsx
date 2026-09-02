package main

// gotsx export: static export — build, start a local server, crawl the routes, rewrite the subpath, copy assets. For GitHub Pages or any static host.

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

type exportOpts struct {
	out    string
	base   string
	site   string // the site's public origin (e.g. https://user.github.io): absolute links (hreflang / canonical) are rewritten to site+base; empty → base-relative paths
	routes []string
	port   int
}

func parseExportArgs(args []string) (string, exportOpts, error) {
	o := exportOpts{out: "dist", port: 4123}
	dir := "."
	for i := 0; i < len(args); i++ {
		a := args[i]
		next := func() (string, error) {
			if i+1 >= len(args) {
				return "", fmt.Errorf("%s needs a value", a)
			}
			i++
			return args[i], nil
		}
		var err error
		switch {
		case a == "--out":
			o.out, err = next()
		case a == "--base":
			o.base, err = next()
		case a == "--site":
			o.site, err = next()
			o.site = strings.TrimSuffix(o.site, "/")
		case a == "--routes":
			var v string
			v, err = next()
			for _, r := range strings.Split(v, ",") {
				if r = strings.TrimSpace(r); r != "" {
					o.routes = append(o.routes, r)
				}
			}
		case a == "--port":
			var v string
			v, err = next()
			fmt.Sscanf(v, "%d", &o.port)
		case strings.HasPrefix(a, "-"):
			return "", o, fmt.Errorf("unknown flag %s", a)
		default:
			dir = a
		}
		if err != nil {
			return "", o, err
		}
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", o, err
	}
	o.base = strings.TrimSuffix(o.base, "/")
	if o.base != "" && !strings.HasPrefix(o.base, "/") {
		o.base = "/" + o.base
	}
	if !filepath.IsAbs(o.out) {
		o.out = filepath.Join(abs, o.out)
	}
	return abs, o, nil
}

func export(dir string, cfg appConfig, o exportOpts) error {
	if err := build(dir, cfg, false); err != nil {
		return err
	}
	bin := filepath.Join(dir, ".gotsx", "export-bin"+exeSuffix())
	os.MkdirAll(filepath.Dir(bin), 0o755)
	gb := exec.Command("go", "build", "-o", bin, ".")
	gb.Dir = dir
	if out, err := gb.CombinedOutput(); err != nil {
		return fmt.Errorf("go build failed:\n%s", out)
	}
	proc := exec.Command(bin, "-addr", fmt.Sprintf("127.0.0.1:%d", o.port))
	proc.Dir = dir
	proc.Stderr = io.Discard
	if err := proc.Start(); err != nil {
		return err
	}
	defer func() { proc.Process.Kill(); proc.Wait() }()
	origin := fmt.Sprintf("http://127.0.0.1:%d", o.port)
	client := &http.Client{Timeout: 20 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	var ready bool
	for i := 0; i < 100; i++ {
		if r, err := client.Get(origin + "/healthz"); err == nil {
			r.Body.Close()
			ready = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if !ready {
		return fmt.Errorf("the app did not start on %s", origin)
	}

	if err := prepareOut(dir, o.out); err != nil {
		return err
	}
	for _, d := range []string{filepath.Join(o.out, "_gotsx"), filepath.Join(o.out, "public")} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return err
		}
	}
	if err := copyTree(filepath.Join(dir, "gen", "client"), filepath.Join(o.out, "_gotsx")); err != nil {
		return err
	}
	copyTree(filepath.Join(dir, "public"), filepath.Join(o.out, "public")) // optional
	os.WriteFile(filepath.Join(o.out, ".nojekyll"), nil, 0o644)            // Pages: keep the _gotsx directory

	// routes: --routes if given; otherwise crawl same-origin links from "/" (<a href> and <link rel=alternate href>)
	seen := map[string]bool{}
	queue := append([]string{}, o.routes...)
	crawl := len(o.routes) == 0
	if crawl {
		queue = []string{"/"}
	}
	rewrite := rewriter(o.base, origin, o.site)
	var written []string
	fetch := func(route string) ([]byte, int, error) {
		r, err := client.Get(origin + route)
		if err != nil {
			return nil, 0, err
		}
		defer r.Body.Close()
		b, err := io.ReadAll(r.Body)
		return b, r.StatusCode, err
	}
	for len(queue) > 0 {
		route := queue[0]
		queue = queue[1:]
		if seen[route] || len(seen) > 2000 {
			continue
		}
		seen[route] = true
		body, status, err := fetch(route)
		if err != nil {
			return fmt.Errorf("%s: %v", route, err)
		}
		if status != 200 {
			fmt.Fprintf(os.Stderr, "gotsx export: %s → %d (skipped)\n", route, status)
			continue
		}
		outDir := filepath.Join(o.out, filepath.FromSlash(strings.TrimPrefix(route, "/")))
		if err := os.MkdirAll(outDir, 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(outDir, "index.html"), rewrite(body), 0o644); err != nil {
			return err
		}
		written = append(written, route)
		if crawl {
			for _, l := range links(body, origin) {
				if !seen[l] {
					queue = append(queue, l)
				}
			}
		}
	}
	// 404.html: the app's own 404 page (Pages serves it)
	if body, status, err := fetch("/__gotsx_export_404__"); err == nil && status == 404 && len(body) > 0 {
		os.WriteFile(filepath.Join(o.out, "404.html"), rewrite(body), 0o644)
	} else if idx, err := os.ReadFile(filepath.Join(o.out, "index.html")); err == nil {
		os.WriteFile(filepath.Join(o.out, "404.html"), idx, 0o644)
	}
	sort.Strings(written)
	fmt.Printf("gotsx: exported %d pages to %s (base %q)\n", len(written), o.out, o.base)
	return nil
}

// assetExt: link targets that are files, not pages
var assetExt = map[string]bool{".css": true, ".js": true, ".mjs": true, ".map": true, ".json": true, ".xml": true, ".txt": true, ".pdf": true, ".zip": true, ".wasm": true, ".webmanifest": true,
	".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".svg": true, ".webp": true, ".avif": true, ".ico": true, ".woff": true, ".woff2": true, ".ttf": true, ".otf": true, ".mp4": true, ".webm": true, ".mp3": true}

var (
	hrefRe   = regexp.MustCompile(`(href|src|action)="/([^/"])`)
	assetRe  = regexp.MustCompile(`"/_gotsx/`)
	bootRe   = regexp.MustCompile(`window\.__GOTSX=\{`)
	linkRe   = regexp.MustCompile(`href="([^"#?]*)`)
	skipLink = regexp.MustCompile(`^/(_gotsx/|public/|healthz|readyz|actions/)`)
)

// rewriter moves root-relative paths under the subpath (a GitHub project page lives under /repo/) and rewrites the local
// origin (http://127.0.0.1:port in hreflang / canonical) to site+base (base-relative when --site is not given)
func rewriter(base, origin, site string) func([]byte) []byte {
	return func(b []byte) []byte {
		b = bytes.ReplaceAll(b, []byte(origin+"/"), []byte(site+base+"/"))
		b = bytes.ReplaceAll(b, []byte(`"`+origin+`"`), []byte(`"`+site+base+`/"`))
		if base == "" {
			return b
		}
		b = hrefRe.ReplaceAll(b, []byte(`${1}="`+base+`/${2}`))
		b = assetRe.ReplaceAll(b, []byte(`"`+base+`/_gotsx/`))
		b = bootRe.ReplaceAll(b, []byte(`window.__GOTSX={"base":"`+base+`",`))
		return b
	}
}

// prepareOut: refuse to wipe anything that is not an empty directory or a previous export (marked by .nojekyll + _gotsx/),
// and never the app directory, one of its ancestors, the filesystem root or the home directory
func prepareOut(appDir, out string) error {
	out = filepath.Clean(out)
	if real, err := filepath.EvalSymlinks(out); err == nil { // macOS: /var → /private/var
		out = real
	}
	if real, err := filepath.EvalSymlinks(appDir); err == nil {
		appDir = real
	}
	home, _ := os.UserHomeDir()
	if out == filepath.Clean(appDir) || out == filepath.VolumeName(out)+string(filepath.Separator) || (home != "" && out == filepath.Clean(home)) || strings.HasPrefix(filepath.Clean(appDir), out+string(filepath.Separator)) {
		return fmt.Errorf("refusing to export into %s: --out must be a separate directory (it is emptied first)", out)
	}
	entries, err := os.ReadDir(out)
	if err != nil {
		if os.IsNotExist(err) {
			return os.MkdirAll(out, 0o755)
		}
		return err
	}
	if len(entries) == 0 {
		return nil
	}
	if _, err := os.Stat(filepath.Join(out, "_gotsx")); err != nil || !fileExists(filepath.Join(out, ".nojekyll")) {
		return fmt.Errorf("refusing to empty %s: it is not empty and not a previous gotsx export (delete it yourself, or pick another --out)", out)
	}
	if err := os.RemoveAll(out); err != nil {
		return err
	}
	return os.MkdirAll(out, 0o755)
}

// links: same-origin paths in a page (including hreflang written as absolute origin URLs; excluding assets, actions and query strings); trailing slash removed
func links(body []byte, origin string) []string {
	var out []string
	for _, m := range linkRe.FindAllSubmatch(body, -1) {
		l := string(m[1])
		if origin != "" && strings.HasPrefix(l, origin) {
			l = strings.TrimPrefix(l, origin)
			if l == "" {
				l = "/"
			}
		}
		if !strings.HasPrefix(l, "/") || strings.HasPrefix(l, "//") || skipLink.MatchString(l) || assetExt[strings.ToLower(path.Ext(path.Base(l)))] {
			continue // assets (…/app.css, …/logo.svg) are copied from public/, not crawled; /blog/v1.2 is a page
		}
		if len(l) > 1 {
			l = strings.TrimSuffix(l, "/")
		}
		out = append(out, l)
	}
	return out
}

func copyTree(src, dst string) error {
	return filepath.WalkDir(src, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, p)
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		return os.WriteFile(target, b, 0o644)
	})
}
