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
	var dir, module, replace, db string
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
		case a == "--db" && i+1 < len(args):
			db = args[i+1]
			i++
			if db != "sqlite" && db != "memory" {
				return fmt.Errorf("--db must be sqlite or memory (default)")
			}
		case strings.HasPrefix(a, "-"):
			return fmt.Errorf("unknown flag %s\n%s", a, usage)
		case dir == "":
			dir = a
		default:
			return fmt.Errorf("unexpected argument %s", a)
		}
	}
	if dir == "" {
		return fmt.Errorf("usage: gotsx new <dir> [--module name] [--replace path] [--tailwind] [--db sqlite]")
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
		"host/host.go":                    strings.ReplaceAll(hostFor(db), "{{fw}}", fw),
		"cmd/hostgen/main.go":             strings.NewReplacer("{{module}}", module, "{{fw}}", fw).Replace(hostgenTmpl),
		"app/pages/index.server.tsx":      indexTmpl,
		"app/pages/_layout.server.tsx":    strings.NewReplacer("{{name}}", name, "{{css}}", cssHref(tailwind)).Replace(layoutTmpl),
		"app/pages/_404.server.tsx":       notFoundTmpl,
		"app/islands/TodoList.client.tsx": todoListTmpl,
		"tsconfig.json":                   tsconfigTmpl,
		".gitignore":                      gitignoreTmpl,
		"README.md":                       strings.NewReplacer("{{name}}", name, "{{fw}}", fw).Replace(readmeTmpl),
		"AGENTS.md":                       agentsFile(name),
		"CLAUDE.md":                       claudeFile,
	}
	files["public/robots.txt"] = "User-agent: *\nAllow: /\n" // public/ must hold at least one embeddable file (dotfiles don't count for go:embed)
	if db == "sqlite" {
		files[".gitignore"] += "\n# SQLite database (DATABASE_PATH, default ./todos.db)\n*.db\n*.db-wal\n*.db-shm\n"
	}
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
	if db == "sqlite" {
		fmt.Println("  (data lives in ./todos.db — set DATABASE_PATH to move it)")
	}
	return nil
}

// hostFor: the host template — in-memory (default) or SQLite (--db sqlite, pure-Go modernc.org/sqlite, no cgo)
func hostFor(db string) string {
	if db == "sqlite" {
		return hostSqliteTmpl
	}
	return hostTmpl
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
// data comes from the host package, islands call host methods through typed actions (gen.HostActions),
// and classic form posts are plain Go handlers (Actions).
package main

import (
	"embed"
	"flag"
	"io/fs"
	"log"
	"net/http"
	"os"

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

		// Typed actions: host methods listed in host.Registry[...].Actions. Islands import and await them;
		// the compiler generates the JSON decoding, the route and the same-origin/CSRF checks.
		HostActions: gen.HostActions,
		// Signed session cookie (PageProps.session / flash / csrf; req.Session() in actions).
		// Empty → a random key per start, so sessions do not survive restarts: set it in production.
		SessionSecret: os.Getenv("SESSION_SECRET"),

		Actions: map[string]http.HandlerFunc{
			// A classic form post, no JavaScript involved: verify the CSRF token from the form,
			// write, queue a flash message for the next page, redirect (POST → redirect → GET).
			"POST /todos": func(w http.ResponseWriter, r *http.Request) {
				if !gotsx.VerifyCSRF(r) {
					http.Error(w, "invalid CSRF token", http.StatusForbidden)
					return
				}
				sess := gotsx.SessionOf(r)
				if _, err := host.Data.Todos.Add(r.FormValue("title")); err != nil {
					sess.Flash("error", err.Error())
				} else {
					sess.Flash("ok", "Added “"+r.FormValue("title")+"”")
				}
				sess.Save(w, r)
				http.Redirect(w, r, "/", http.StatusSeeOther)
			},
		},
	}))
}
`

const hostTmpl = `// Package host is the Go side of ` + "`import { todos } from \"host:data\"`" + `.
// Everything the dialect can see is exactly what this package exposes; server-side calls compile to
// direct Go calls, and methods listed in Registry[...].Actions become typed actions islands can await.
// Replace the in-memory store with a database — the pages don't change.
package host

import (
	"fmt"
	"strconv"
	"strings"
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
	next  int
}

// List returns a copy (host methods run concurrently: one goroutine per request).
func (s *TodoStore) List() []Todo {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]Todo(nil), s.items...)
}

// Get: a (T, error) method — in a page the error becomes a panic recovered by the request layer;
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

// Add validates its input: gotsx.Invalid becomes a 422 with per-field messages for actions,
// and a readable error for form handlers.
func (s *TodoStore) Add(title string) (Todo, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return Todo{}, gotsx.Invalid(map[string]string{"title": "Title is required"})
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.next++
	t := Todo{ID: strconv.Itoa(s.next), Title: title}
	s.items = append(s.items, t)
	return t, nil
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

func (s *TodoStore) Remove(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.items {
		if s.items[i].ID == id {
			s.items = append(s.items[:i], s.items[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("%w: todo %q", gotsx.ErrNotFound, id)
}

type DataModule struct {
	Todos *TodoStore ` + "`json:\"todos\"`" + `
}

// Actions live on the module (they may take a *gotsx.Req first parameter for the session/cookies).
func (d *DataModule) Toggle(id string) (Todo, error) { return d.Todos.Toggle(id) }
func (d *DataModule) Remove(id string) error         { return d.Todos.Remove(id) }

var Data = &DataModule{Todos: &TodoStore{next: 3, items: []Todo{
	{ID: "1", Title: "Read the gotsx README", Done: true},
	{ID: "2", Title: "Edit app/pages/index.server.tsx", Done: false},
	{ID: "3", Title: "Add a host method and call it from a page", Done: false},
}}}

// Registry: module name → value + the Go expression generated code uses to reach it.
// Actions lists the methods islands may call (POST /_gotsx/act/data/<name>).
// cmd/hostgen reflects this into app/.gen/host.d.ts (editor) and host.json (compiler).
var Registry = map[string]gotsx.HostModule{
	"data": {Value: Data, Go: "host.Data", Actions: []string{"Toggle", "Remove"}},
}
`

const hostSqliteTmpl = `// Package host is the Go side of ` + "`import { todos } from \"host:data\"`" + `, backed by SQLite
// (modernc.org/sqlite: pure Go, no cgo, one file). Everything the dialect can see is exactly what this
// package exposes; server-side calls compile to direct Go calls, and methods listed in
// Registry[...].Actions become typed actions islands can await.
package host

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	_ "modernc.org/sqlite"

	gotsx "{{fw}}/runtime"
)

type Todo struct {
	ID    string ` + "`json:\"id\"`" + `
	Title string ` + "`json:\"title\"`" + `
	Done  bool   ` + "`json:\"done\"`" + `
}

type TodoStore struct{ db *sql.DB }

// Methods return (T, error): in a page a failure becomes a 500 (or 404 when the error wraps gotsx.ErrNotFound);
// in an action it becomes the matching HTTP status. Host methods run concurrently; database/sql pools connections.
func (s *TodoStore) List() ([]Todo, error) {
	rows, err := s.db.Query(` + "`SELECT id, title, done FROM todos ORDER BY id`" + `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Todo
	for rows.Next() {
		var t Todo
		var id int64
		if err := rows.Scan(&id, &t.Title, &t.Done); err != nil {
			return nil, err
		}
		t.ID = strconv.FormatInt(id, 10)
		out = append(out, t)
	}
	return out, rows.Err()
}

func (s *TodoStore) Get(id string) (Todo, error) {
	var t Todo
	err := s.db.QueryRow(` + "`SELECT title, done FROM todos WHERE id = ?`" + `, id).Scan(&t.Title, &t.Done)
	if errors.Is(err, sql.ErrNoRows) {
		return Todo{}, fmt.Errorf("%w: todo %q", gotsx.ErrNotFound, id)
	}
	t.ID = id
	return t, err
}

// Add validates its input: gotsx.Invalid becomes a 422 with per-field messages for actions,
// and a readable error for form handlers.
func (s *TodoStore) Add(title string) (Todo, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return Todo{}, gotsx.Invalid(map[string]string{"title": "Title is required"})
	}
	res, err := s.db.Exec(` + "`INSERT INTO todos (title, done) VALUES (?, 0)`" + `, title)
	if err != nil {
		return Todo{}, err
	}
	id, _ := res.LastInsertId()
	return Todo{ID: strconv.FormatInt(id, 10), Title: title}, nil
}

func (s *TodoStore) Toggle(id string) (Todo, error) {
	res, err := s.db.Exec(` + "`UPDATE todos SET done = NOT done WHERE id = ?`" + `, id)
	if err != nil {
		return Todo{}, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return Todo{}, fmt.Errorf("%w: todo %q", gotsx.ErrNotFound, id)
	}
	return s.Get(id)
}

func (s *TodoStore) Remove(id string) error {
	res, err := s.db.Exec(` + "`DELETE FROM todos WHERE id = ?`" + `, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("%w: todo %q", gotsx.ErrNotFound, id)
	}
	return nil
}

type DataModule struct {
	Todos *TodoStore ` + "`json:\"todos\"`" + `
}

// Actions live on the module (they may take a *gotsx.Req first parameter for the session/cookies).
func (d *DataModule) Toggle(id string) (Todo, error) { return d.Todos.Toggle(id) }
func (d *DataModule) Remove(id string) error         { return d.Todos.Remove(id) }

// open: DATABASE_PATH (default ./todos.db); WAL + busy timeout so concurrent requests don't fail with "database is locked"
func open() *sql.DB {
	path := os.Getenv("DATABASE_PATH")
	if path == "" {
		path = "todos.db"
	}
	db, err := sql.Open("sqlite", "file:"+path+"?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)")
	if err != nil {
		panic(err)
	}
	if _, err := db.Exec(` + "`CREATE TABLE IF NOT EXISTS todos (id INTEGER PRIMARY KEY AUTOINCREMENT, title TEXT NOT NULL, done INTEGER NOT NULL DEFAULT 0)`" + `); err != nil {
		panic(err)
	}
	var n int
	if db.QueryRow(` + "`SELECT count(*) FROM todos`" + `).Scan(&n); n == 0 {
		db.Exec(` + "`INSERT INTO todos (title, done) VALUES ('Read the gotsx README', 1), ('Edit app/pages/index.server.tsx', 0), ('Add a host method and call it from a page', 0)`" + `)
	}
	return db
}

var Data = &DataModule{Todos: &TodoStore{db: open()}}

// Registry: module name → value + the Go expression generated code uses to reach it.
// Actions lists the methods islands may call (POST /_gotsx/act/data/<name>).
// cmd/hostgen reflects this into app/.gen/host.d.ts (editor) and host.json (compiler).
var Registry = map[string]gotsx.HostModule{
	"data": {Value: Data, Go: "host.Data", Actions: []string{"Toggle", "Remove"}},
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

const indexTmpl = `import type { PageProps, Meta } from "gotsx";
import { todos } from "host:data";               // Go-backed: a direct call after compilation
import TodoList from "../islands/TodoList.client";

// Page metadata: the layout renders it (props.meta) into <title> and <meta name="description">.
export function meta(): Meta {
  return { title: "Todos", description: "A gotsx starter: pages compiled to Go, islands compiled to signals." };
}

// A page: compiled to a Go function, never shipped to the browser. pages/_layout.server.tsx wraps it.
export default function Home({ query, flash, csrf }: PageProps) {
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
      {flash.map((f) => (
        <div class={"alert alert-" + f.kind} role="status">{f.text}</div>
      ))}
      <form method="post" action="/todos" class="row">
        <input type="hidden" name="_csrf" value={csrf} />
        <input class="input" style="flex:1" name="title" placeholder="New todo…" aria-label="New todo" autocomplete="off" />
        <button class="btn btn-primary">Add</button>
      </form>
      <form method="get" action="/" class="row">
        <input class="input" style="flex:1" name="q" value={q} placeholder="Filter…" aria-label="Filter todos" />
        <button class="btn btn-outline">Filter</button>
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
// The whole document is rendered by the dialect; props are PageProps + the page's meta + children.
export default function Layout({ path, meta, children }: LayoutProps) {
  return (
    <html lang="en">
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1" />
        <title>{meta.title ? meta.title + " · {{name}}" : "{{name}}"}</title>
        {meta.description && <meta name="description" content={meta.description} />}
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
import { toggle, remove } from "host:data";     // typed actions: Go methods called over a same-origin POST
import type { Todo } from "host:data";

// An island: SSR'd by Go from the same source, then hydrated as signals in the browser (~6KB runtime).
export default function TodoList({ initial }: { initial: Todo[] }) {
  const [items, setItems] = useState(initial);
  const [busy, setBusy] = useState("");
  const [error, setError] = useState("");
  const remaining = items.filter((x) => !x.done).length;   // reads a signal → automatically a memo

  const onToggle = async (id: string) => {
    setBusy(id);
    setError("");
    try {
      const t = await toggle(id);                            // Promise<Todo>, typed from Go
      setItems(items.map((x) => (x.id === id ? t : x)));
    } catch (e) {
      setError(e.message);                                   // 404 / 422 / 500 from the action
    } finally {
      setBusy("");
    }
  };
  const onRemove = async (id: string) => {
    try {
      await remove(id);
      setItems(items.filter((x) => x.id !== id));
    } catch (e) {
      setError(e.message);
    }
  };

  return (
    <div class="stack">
      {error !== "" && <div class="alert alert-error" role="status">{error}</div>}
      <ul class="todos">
        {items.map((x) => (
          <li key={x.id} class={x.done ? "done" : ""}>
            <button class="btn btn-outline btn-sm" disabled={busy === x.id} onClick={() => onToggle(x.id)}>
              {x.done ? "Done" : "Todo"}
            </button>
            <span style="flex:1">{x.title}</span>
            <button class="btn btn-ghost btn-sm" aria-label="Remove" onClick={() => onRemove(x.id)}>×</button>
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

- ` + "`app/pages/*.server.tsx`" + ` → routes (` + "`[id]`" + ` params, ` + "`[...slug]`" + ` catch-all); ` + "`_layout`" + ` / ` + "`_404`" + ` / ` + "`_error`" + ` are conventions; ` + "`export function meta`" + ` feeds the layout's ` + "`<title>`" + `. Compiled to Go, never reach the browser.
- ` + "`app/islands/*.client.tsx`" + ` → islands: server-rendered by Go, interactive in the browser as signals. They call Go through **typed actions**: methods listed in ` + "`host.Registry[...].Actions`" + `, imported from ` + "`host:data`" + ` and awaited.
- ` + "`host/`" + ` → Go code the dialect can import as ` + "`host:data`" + `. Add a method, run ` + "`gotsx build`" + `, use it.
- ` + "`main.go`" + ` → the server: routes and ` + "`HostActions`" + ` from ` + "`gen/`" + `, a CSRF-protected form handler with a flash message, signed sessions (` + "`SESSION_SECRET`" + `), production hardening by default.
- ` + "`AGENTS.md`" + ` / ` + "`CLAUDE.md`" + ` → instructions for coding agents; ` + "`app/.gen/docs/`" + ` holds the docs for the gotsx version in use (` + "`gotsx docs`" + ` prints them).
- Data lives in memory (or SQLite with ` + "`gotsx new --db sqlite`" + `: ` + "`DATABASE_PATH`" + `, default ` + "`./todos.db`" + `); swap the store in ` + "`host/`" + ` — the pages don't change.
`
