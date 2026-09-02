package main

// gotsx new <dir>: scaffold a self-contained app in its own Go module.
// The result builds with `gotsx dev` right away: a host module (Go), a page, a layout, a keyed-list island,
// an HTTP action, editor typings (tsconfig + app/.gen/*.d.ts) and a .gitignore.

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/childrentime/gotsx/design"
)

func newApp(args []string) error {
	var dir, module, replace string
	tailwind := false
	for i := 0; i < len(args); i++ {
		switch a := args[i]; {
		case a == "--module" && i+1 < len(args):
			module = args[i+1]
			i++
		case a == "--replace" && i+1 < len(args):
			replace = args[i+1]
			i++
		case a == "--tailwind":
			tailwind = true
		case strings.HasPrefix(a, "-"):
			return fmt.Errorf("unknown flag %s\n%s", a, usage)
		case dir == "":
			dir = a
		default:
			return fmt.Errorf("unexpected argument %s", a)
		}
	}
	if dir == "" {
		return fmt.Errorf("usage: gotsx new <dir> [--module name] [--replace path] [--tailwind]")
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	if entries, err := os.ReadDir(abs); err == nil && len(entries) > 0 {
		return fmt.Errorf("%s already exists and is not empty", abs)
	}
	if module == "" {
		module = sanitizeModule(filepath.Base(abs))
	}
	fw := frameworkModule()
	ver := version()
	if replace == "" && (ver == "(devel)" || strings.Contains(ver, "-")) {
		replace = findFrameworkCheckout(fw) // running from a checkout (go run ./cmd/gotsx): point the new module at it
	}
	if replace != "" {
		if replace, err = filepath.Abs(replace); err != nil {
			return err
		}
		ver = "v0.0.0"
	}
	goVer := strings.TrimPrefix(runtime.Version(), "go")
	if i := strings.LastIndexByte(goVer, '.'); i > 0 && strings.Count(goVer, ".") == 2 {
		goVer = goVer[:i] // 1.26.4 → 1.26
	}
	name := filepath.Base(abs)
	files := map[string]string{
		"go.mod":                          goModTmpl(module, fw, ver, goVer, replace),
		"main.go":                         strings.NewReplacer("{{module}}", module, "{{fw}}", fw).Replace(mainTmpl),
		"host/host.go":                    strings.ReplaceAll(hostTmpl, "{{fw}}", fw),
		"cmd/hostgen/main.go":             strings.NewReplacer("{{module}}", module, "{{fw}}", fw).Replace(hostgenTmpl),
		"app/pages/index.server.tsx":      indexTmpl,
		"app/pages/_layout.server.tsx":    strings.NewReplacer("{{name}}", name, "{{css}}", cssHref(tailwind)).Replace(layoutTmpl),
		"app/pages/_404.server.tsx":       notFoundTmpl,
		"app/islands/TodoList.client.tsx": todoListTmpl,
		"tsconfig.json":                   tsconfigTmpl,
		".gitignore":                      gitignoreTmpl,
		"README.md":                       strings.NewReplacer("{{name}}", name, "{{fw}}", fw).Replace(readmeTmpl),
	}
	files["public/robots.txt"] = "User-agent: *\nAllow: /\n" // public/ must hold at least one embeddable file (dotfiles don't count for go:embed)
	if tailwind {
		files["app/gotsx.css"] = design.TailwindCSS // the gotsx UI design system (tokens + component classes); edit freely
		files["app/tailwind.css"] = "@import \"./gotsx.css\";\n\n/* app-specific CSS below */\n"
	} else {
		files["public/app.css"] = design.PlainCSS + "\n/* app-specific CSS below */\n"
	}
	for rel, content := range files {
		p := filepath.Join(abs, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			return err
		}
	}
	fmt.Printf("gotsx: created %s (module %s)\n", abs, module)
	tidy := exec.Command("go", "mod", "tidy")
	tidy.Dir = abs
	if out, err := tidy.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "gotsx: go mod tidy failed (fix go.mod and rerun it):\n%s", out)
	}
	fmt.Printf("\nNext:\n  cd %s\n", dir)
	if tailwind {
		fmt.Println("  gotsx tailwind      # once: download the Tailwind standalone binary")
	}
	fmt.Println("  gotsx dev           # http://localhost:3000 — edit app/**/*.tsx and the browser reloads")
	return nil
}

// TypeScript-side CSS link: plain stylesheet or the Tailwind output
func cssHref(tailwind bool) string {
	if tailwind {
		return "/public/tailwind.css"
	}
	return "/public/app.css"
}

var modBad = regexp.MustCompile(`[^a-z0-9._/-]+`)

func sanitizeModule(name string) string {
	s := modBad.ReplaceAllString(strings.ToLower(name), "-")
	s = strings.Trim(s, "-.")
	if s == "" {
		return "app"
	}
	return s
}

// findFrameworkCheckout: the nearest ancestor of the working directory whose go.mod declares the framework module
func findFrameworkCheckout(fw string) string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	for d := wd; ; d = filepath.Dir(d) {
		if b, err := os.ReadFile(filepath.Join(d, "go.mod")); err == nil && strings.Contains(string(b), "module "+fw+"\n") {
			return d
		}
		if filepath.Dir(d) == d {
			return ""
		}
	}
}

func goModTmpl(module, fw, ver, goVer, replace string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "module %s\n\ngo %s\n\nrequire %s %s\n", module, goVer, fw, ver)
	if replace != "" {
		fmt.Fprintf(&b, "\nreplace %s => %s\n", fw, filepath.ToSlash(replace))
	}
	return b.String()
}

const mainTmpl = `// {{module}}: a gotsx app. Pages are Go functions compiled from app/**/*.tsx (see gen/),
// data comes from the host package, and writes are plain Go HTTP handlers (Actions).
package main

import (
	"embed"
	"encoding/json"
	"flag"
	"io/fs"
	"log"
	"net/http"

	"{{module}}/gen"
	"{{module}}/host"
	gotsx "{{fw}}/runtime"
)

//go:embed public
var publicEmbed embed.FS

func main() {
	addr := flag.String("addr", ":3000", "listen address")
	dev := flag.Bool("dev", false, "development mode (gotsx dev passes this)")
	flag.Parse()
	public, _ := fs.Sub(publicEmbed, "public")
	log.Fatal(gotsx.Serve(gotsx.Options{
		Addr:      *addr,
		Dev:       *dev,
		Routes:    gen.Routes,
		ClientDir: gotsx.FindDir("gen/client"), // dev: served from disk
		ClientFS:  gen.ClientFS,                // prod: embedded → one self-contained binary
		PublicDir: gotsx.FindDir("public"),
		PublicFS:  public,
		NotFound:  gen.NotFound,  // pages/_404.server.tsx (nil if absent → built-in 404)
		ErrorPage: gen.ErrorPage, // pages/_error.server.tsx
		Actions: map[string]http.HandlerFunc{
			// Writes stay in Go. Islands reach them over HTTP; CSRF same-origin checks are on by default.
			"POST /actions/toggle": func(w http.ResponseWriter, r *http.Request) {
				var body struct {
					ID string ` + "`json:\"id\"`" + `
				}
				json.NewDecoder(r.Body).Decode(&body)
				todo, err := host.Data.Todos.Toggle(body.ID)
				w.Header().Set("Content-Type", "application/json")
				if err != nil {
					w.WriteHeader(http.StatusNotFound)
					json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
					return
				}
				json.NewEncoder(w).Encode(map[string]any{"todo": todo})
			},
		},
	}))
}
`

const hostTmpl = `// Package host is the Go side of ` + "`import { todos } from \"host:data\"`" + `.
// Everything the dialect can see is exactly what this package exposes; calls compile to direct Go calls.
// Replace the in-memory store with a database — the pages don't change.
package host

import (
	"fmt"
	"sync"

	gotsx "{{fw}}/runtime"
)

type Todo struct {
	ID    string ` + "`json:\"id\"`" + `
	Title string ` + "`json:\"title\"`" + `
	Done  bool   ` + "`json:\"done\"`" + `
}

type TodoStore struct {
	mu    sync.Mutex
	items []Todo
}

// List returns a copy (host methods run concurrently: one goroutine per request).
func (s *TodoStore) List() []Todo {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]Todo(nil), s.items...)
}

// Get: a (T, error) method — the error becomes a panic recovered by the request layer;
// wrapping gotsx.ErrNotFound turns it into a 404.
func (s *TodoStore) Get(id string) (Todo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, t := range s.items {
		if t.ID == id {
			return t, nil
		}
	}
	return Todo{}, fmt.Errorf("%w: todo %q", gotsx.ErrNotFound, id)
}

func (s *TodoStore) Toggle(id string) (Todo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.items {
		if s.items[i].ID == id {
			s.items[i].Done = !s.items[i].Done
			return s.items[i], nil
		}
	}
	return Todo{}, fmt.Errorf("%w: todo %q", gotsx.ErrNotFound, id)
}

type DataModule struct {
	Todos *TodoStore ` + "`json:\"todos\"`" + `
}

var Data = &DataModule{Todos: &TodoStore{items: []Todo{
	{ID: "1", Title: "Read the gotsx README", Done: true},
	{ID: "2", Title: "Edit app/pages/index.server.tsx", Done: false},
	{ID: "3", Title: "Add a host method and call it from a page", Done: false},
}}}

// Registry: module name → value + the Go expression generated code uses to reach it.
// cmd/hostgen reflects this into app/.gen/host.d.ts (editor) and host.json (compiler).
var Registry = map[string]gotsx.HostModule{
	"data": {Value: Data, Go: "host.Data"},
}
`

const hostgenTmpl = `// hostgen: reflect host.Registry → app/.gen/host.d.ts + host.json. gotsx build runs it automatically.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"{{module}}/host"
	gotsx "{{fw}}/runtime"
)

func main() {
	dir := filepath.Join("app", ".gen")
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}
	dts, js := gotsx.GenerateHost(host.Registry, "host")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for name, content := range map[string]string{"host.d.ts": dts, "host.json": js} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}
`

const indexTmpl = `import type { PageProps } from "gotsx";
import { todos } from "host:data";               // Go-backed: a direct call after compilation
import TodoList from "../islands/TodoList.client";

// A page: compiled to a Go function, never shipped to the browser. pages/_layout.server.tsx wraps it.
export default function Home({ query }: PageProps) {
  const list = todos.list();                     // synchronous — concurrency comes from goroutines
  const done = list.filter((x) => x.done).length;
  const q = query.q ?? "";
  const shown = q === "" ? list : list.filter((x) => x.title.toLowerCase().includes(q.toLowerCase()));
  return (
    <div class="stack">
      <div>
        <h1>Todos</h1>
        <p class="muted">{done} of {list.length} done · rendered by Go in microseconds</p>
      </div>
      <form method="get" action="/" class="row">
        <input class="input" style="flex:1" name="q" value={q} placeholder="Filter…" />
        <button class="btn btn-primary">Filter</button>
      </form>
      {shown.length === 0 && <p class="empty">Nothing matches “{q}”.</p>}
      <TodoList initial={shown} />
    </div>
  );
}
`

const notFoundTmpl = `import type { PageProps } from "gotsx";

// pages/_404.server.tsx → gen.NotFound, wrapped by _layout like any page.
export default function NotFound({ path }: PageProps) {
  return (
    <div class="empty">
      <h1>404</h1>
      <p class="muted">There is nothing at {path}.</p>
      <a href="/" class="btn btn-outline">Home</a>
    </div>
  );
}
`

const layoutTmpl = `import type { LayoutProps } from "gotsx";

// pages/_layout.server.tsx wraps every page under pages/ (nested directories can add their own _layout).
// The whole document is rendered by the dialect; props are PageProps + children.
export default function Layout({ path, children }: LayoutProps) {
  return (
    <html lang="en">
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1" />
        <title>{{name}}</title>
        <link rel="stylesheet" href="{{css}}" />
      </head>
      <body>
        <div id="gotsx-bar"></div>
        <header class="page-header">
          <div class="container-page flex h-14 items-center justify-between">
            <a href="/" class="brand"><span class="mark">g</span>{{name}}</a>
            <nav class="nav flex items-center gap-1">
              <a href="/" class={path === "/" ? "nav-link nav-link-active" : "nav-link"}>Todos</a>
            </nav>
          </div>
        </header>
        <main class="container-page fade-up py-8">{children}</main>
        <footer class="muted py-8 text-center">Built with gotsx — TSX compiled to native Go.</footer>
      </body>
    </html>
  );
}
`

const todoListTmpl = `import { useState } from "gotsx";
import type { Todo } from "host:data";           // types only: the client never touches Go

// An island: SSR'd by Go from the same source, then hydrated as signals in the browser (~6KB runtime).
export default function TodoList({ initial }: { initial: Todo[] }) {
  const [items, setItems] = useState(initial);
  const [busy, setBusy] = useState("");
  const remaining = items.filter((x) => !x.done).length;   // reads a signal → automatically a memo

  const toggle = async (id: string) => {
    setBusy(id);
    const r = await fetch("/actions/toggle", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ id }),
    });
    const d = await r.json();
    if (d.todo) setItems(items.map((x) => (x.id === id ? (d.todo as Todo) : x)));
    setBusy("");
  };

  return (
    <div class="stack">
      <ul class="todos">
        {items.map((x) => (
          <li key={x.id} class={x.done ? "done" : ""}>
            <button class="btn btn-outline btn-sm" disabled={busy === x.id} onClick={() => toggle(x.id)}>
              {x.done ? "Done" : "Todo"}
            </button>
            <span>{x.title}</span>
          </li>
        ))}
      </ul>
      <p class="muted">{remaining} remaining</p>
    </div>
  );
}
`

const tsconfigTmpl = `{
  "compilerOptions": {
    "target": "ES2022",
    "lib": ["ES2022", "DOM", "DOM.Iterable"],
    "module": "ESNext",
    "moduleResolution": "Bundler",
    "jsx": "preserve",
    "strict": true,
    "noEmit": true,
    "skipLibCheck": true,
    "types": []
  },
  "include": ["app/**/*.tsx", "app/.gen/*.d.ts"]
}
`

const gitignoreTmpl = `# generated by gotsx build (regenerate with gotsx build)
gen/
.gotsx/
app/.gen/
public/tailwind.css
# local tools
.tools/
`

const readmeTmpl = `# {{name}}

A [gotsx]({{fw}}) app: TSX compiled to native Go. No Node, no npm — the toolchain is just Go.

` + "```" + `bash
gotsx dev            # build + run on http://localhost:3000, rebuild + reload on change
gotsx build          # compile app/ → gen/
go build -o {{name}} . && ./{{name}} -addr :8080   # one self-contained binary
gotsx check          # type-check only (also what the editor LSP runs)
` + "```" + `

- ` + "`app/pages/*.server.tsx`" + ` → routes (` + "`[id]`" + ` params, ` + "`[...slug]`" + ` catch-all); ` + "`_layout`" + ` / ` + "`_404`" + ` / ` + "`_error`" + ` are conventions. Compiled to Go, never reach the browser.
- ` + "`app/islands/*.client.tsx`" + ` → islands: server-rendered by Go, interactive in the browser as signals.
- ` + "`host/`" + ` → Go code the dialect can import as ` + "`host:data`" + `. Add a method, run ` + "`gotsx build`" + `, use it.
- ` + "`main.go`" + ` → the server: routes from ` + "`gen/`" + `, HTTP actions for writes, production hardening by default.
`
