package compiler

import (
	"strings"
	"testing"
	"time"

	gotsx "github.com/childrentime/gotsx/runtime"
)

// 带 action 的宿主: todos 模块的 Toggle(req, id) / Add(req, title) / Count() 是 action, List() 不是
type Todo struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}
type Todos struct{}

func (Todos) List() []Todo                                   { return nil }
func (Todos) Toggle(req *gotsx.Req, id string) (Todo, error) { return Todo{}, nil }
func (Todos) Add(req *gotsx.Req, title string) error         { return nil }
func (Todos) Count() int                                     { return 0 }
func (Todos) Ping()                                          {}

var actionHostJSON = func() []byte {
	_, j := gotsx.GenerateHost(map[string]gotsx.HostModule{"todos": {Value: Todos{}, Go: "host.Todos", Actions: []string{"Toggle", "Add", "Count", "Ping"}}}, "host")
	return []byte(j)
}()

func actionChecker(t *testing.T, files map[string]string) (*Checker, error) {
	t.Helper()
	c, err := NewChecker(actionHostJSON)
	if err != nil {
		t.Fatalf("NewChecker: %v", err)
	}
	for f, src := range files {
		m, err := ParseModule(src, f)
		if err != nil {
			t.Fatalf("parse %s: %v", f, err)
		}
		c.AddModule(m)
	}
	return c, c.CheckAll()
}

func TestHostgenActions(t *testing.T) {
	dts, j := gotsx.GenerateHost(map[string]gotsx.HostModule{"todos": {Value: Todos{}, Go: "host.Todos", Actions: []string{"Toggle", "Add"}}}, "host")
	for _, want := range []string{`toggle(id: string): Promise<Todo>`, `add(title: string): Promise<void>`, `list(): Todo[]`, `"action": true`, `"req": true`} {
		if !strings.Contains(dts+j, want) {
			t.Errorf("host.d.ts / host.json should contain %q\n%s\n%s", want, dts, j)
		}
	}
	defer func() {
		if r := recover(); r == nil {
			t.Error("unknown action name should panic")
		}
	}()
	gotsx.GenerateHost(map[string]gotsx.HostModule{"todos": {Value: Todos{}, Go: "host.Todos", Actions: []string{"Nope"}}}, "host")
}

func TestActionClientCall(t *testing.T) {
	c, err := actionChecker(t, map[string]string{"islands/Row.client.tsx": `import { toggle, count, add } from "host:todos";
import type { Todo } from "host:todos";
import { useState } from "gotsx";
export default function Row({ id }: { id: string }) {
  const [done, setDone] = useState(false);
  const [n, setN] = useState(0);
  const go = async () => {
    const t: Todo = await toggle(id);
    setDone(t.done);
    setN(await count());
    await add("x");
  };
  return <button onClick={go} data-add={add}>{done ? "done" : "todo"} {n}</button>;
}`})
	if err != nil {
		t.Fatal(err)
	}
	m := c.Modules["islands/Row.client.tsx"]
	js, err := GenJS(c, m)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`G.act("todos", "toggle", [id])`, `G.act("todos", "count", [])`, `G.act("todos", "add", ["x"])`, `(...a) => G.act("todos", "add", a)`} {
		if !strings.Contains(js, want) {
			t.Errorf("JS should contain %q:\n%s", want, js)
		}
	}
}

func TestActionErrors(t *testing.T) {
	cases := []struct{ name, file, src, want string }{
		{"client imports non-action", "a.client.tsx", `import { list } from "host:todos";
export default function A() { return <b>{list().length}</b>; }`, "not an action"},
		{"server calls req action", "a.server.tsx", `import { toggle } from "host:todos";
export default function A() { return <b>{toggle("1").done}</b>; }`, "only be called from an island"},
		{"await result typed", "a.client.tsx", `import { toggle } from "host:todos";
export default function A() { const f = async () => { const t = await toggle("1"); return t.nope; }; return <b onClick={f}>x</b>; }`, "nope"},
	}
	for _, tc := range cases {
		_, err := actionChecker(t, map[string]string{tc.file: tc.src})
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: want error containing %q, got %v", tc.name, tc.want, err)
		}
	}
	// 服务端可以直接调用无 Req 的 action
	if _, err := actionChecker(t, map[string]string{"a.server.tsx": `import { count, list } from "host:todos";
export default function A() { return <b>{count() + list().length}</b>; }`}); err != nil {
		t.Errorf("server call of a plain action: %v", err)
	}
}

func TestGenActions(t *testing.T) {
	c, err := actionChecker(t, map[string]string{})
	if err != nil {
		t.Fatal(err)
	}
	src, err := GenActions(c, "gen", "github.com/childrentime/gotsx/runtime", "example.com/app/host")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`{Module: "todos", Name: "add", Fn: func(req *gotsx.Req, args []json.RawMessage) (any, error) {`,
		"var a0 string\n\t\tif err := gotsx.Arg(args, 0, &a0); err != nil {",
		"return nil, host.Todos.Add(req, a0)",
		"return host.Todos.Toggle(req, a0)",
		"return host.Todos.Count(), nil",
		"host.Todos.Ping()\n\t\treturn nil, nil",
	} {
		if !strings.Contains(src, want) {
			t.Errorf("actions_gen should contain %q:\n%s", want, src)
		}
	}
	if strings.Contains(src, `Name: "list"`) {
		t.Error("non-action method must not be exposed")
	}
	// 没有 action 的宿主 → 空表, 仍然可编译
	c2, _ := NewChecker(testHostJSON)
	empty, err := GenActions(c2, "gen", "github.com/childrentime/gotsx/runtime", "x/host")
	if err != nil || !strings.Contains(empty, "var HostActions []gotsx.HostAction") {
		t.Errorf("empty registry: %v\n%s", err, empty)
	}
}

// export function meta: 页面级元数据 → 路由表里算一次, 每层布局都拿到
func TestPageMeta(t *testing.T) {
	c, err := actionChecker(t, map[string]string{
		"/x/app/pages/_layout.server.tsx": `import type { LayoutProps } from "gotsx";
export default function Root({ meta, children }: LayoutProps) { return <html><head><title>{meta.title ? meta.title + " · Site" : "Site"}</title>{meta.description && <meta name="description" content={meta.description} />}</head><body>{children}</body></html>; }`,
		"/x/app/pages/index.server.tsx": `import type { PageProps, Meta } from "gotsx";
export function meta(): Meta { return { title: "Home" }; }
export default function Home(p: PageProps) { return <h1>hi</h1>; }`,
		"/x/app/pages/about.server.tsx": `import type { PageProps, Meta } from "gotsx";
export function meta({ params }: PageProps): Meta { return { title: "About " + params.x, noIndex: true }; }
export default function About(p: PageProps) { return <h1>about</h1>; }`,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := metaExpr(c.Modules["/x/app/pages/index.server.tsx"]); got != "pages_index_meta()" {
		t.Errorf("zero-param meta: %s", got)
	}
	if got := metaExpr(c.Modules["/x/app/pages/about.server.tsx"]); got != "pages_about_meta(p)" {
		t.Errorf("one-param meta: %s", got)
	}
	if got := metaExpr(c.Modules["/x/app/pages/_layout.server.tsx"]); got != "gotsx.Meta{}" {
		t.Errorf("no meta: %s", got)
	}
	gs, err := GenGo(c, c.Modules["/x/app/pages/index.server.tsx"], "gen", "github.com/childrentime/gotsx/runtime", "x/host")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(gs, "func pages_index_meta() gotsx.Meta") || !strings.Contains(gs, `gotsx.Meta{Title: "Home"}`) {
		t.Errorf("meta codegen:\n%s", gs)
	}
	for _, tc := range []struct{ src, want string }{
		{`import type { PageProps, Meta } from "gotsx";
export function meta(): string { return "x"; }
export default function P(p: PageProps) { return <b/>; }`, "signature"},
		{`import type { PageProps, Meta } from "gotsx";
export function meta(a: string): Meta { return { title: a }; }
export default function P(p: PageProps) { return <b/>; }`, "signature"},
	} {
		c2, err := actionChecker(t, map[string]string{"/x/app/pages/p.server.tsx": tc.src})
		if err != nil {
			t.Fatal(err)
		}
		if msg := pageProblem(c2, c2.Modules["/x/app/pages/p.server.tsx"], "page"); !strings.Contains(msg, tc.want) {
			t.Errorf("bad meta should be rejected, got %q", msg)
		}
	}
}

// 普通函数的解构参数(meta({ params }: PageProps)) 在 Go 后端要展开
func TestDestructuredFnParams(t *testing.T) {
	c, err := actionChecker(t, map[string]string{"/x/app/pages/p.server.tsx": `import type { PageProps, Meta } from "gotsx";
export function meta({ params, query }: PageProps): Meta { return { title: params.id + query.q }; }
function pick({ a, b = 2 }: { a: number; b?: number }): number { return a + b; }
export default function P(p: PageProps) { const f = ({ x }: { x: string }) => x; return <b>{pick({ a: 1 })}{f({ x: "y" })}</b>; }`})
	if err != nil {
		t.Fatal(err)
	}
	gs, err := GenGo(c, c.Modules["/x/app/pages/p.server.tsx"], "gen", "github.com/childrentime/gotsx/runtime", "x/host")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"func pages_p_meta(p0 gotsx.PageProps) gotsx.Meta {", "params := p0.Params", "query := p0.Query", "a := p0.A", "b := p0.B"} {
		if !strings.Contains(gs, want) {
			t.Errorf("want %q in:\n%s", want, gs)
		}
	}
	if strings.Contains(gs, "\thost ") {
		t.Errorf("no host import needed when only gotsx.Meta is used:\n%s", gs)
	}
}

func TestActionReviewFixes(t *testing.T) {
	// an action called during render (outside an async function) is a positioned error
	_, err := actionChecker(t, map[string]string{"a.client.tsx": `import { count } from "host:todos";
export default function A() { const n = count(); return <b>{n}</b>; }`})
	if err == nil || !strings.Contains(err.Error(), "not during render") {
		t.Errorf("sync action call: %v", err)
	}
	// fire-and-forget from a plain arrow handler, and a non-async callback inside an async function, are fine
	if _, err := actionChecker(t, map[string]string{"a.client.tsx": `import { toggle } from "host:todos";
export default function A({ ids }: { ids: string[] }) { const all = async () => { ids.forEach((i) => toggle(i)); }; return <b onClick={() => toggle("1")} onInput={all}>x</b>; }`}); err != nil {
		t.Errorf("fire-and-forget action calls: %v", err)
	}
	// inside an effect / async arrow it is fine
	if _, err := actionChecker(t, map[string]string{"a.client.tsx": `import { count } from "host:todos";
import { useState, useEffect } from "gotsx";
export default function A() { const [n, setN] = useState(0); useEffect(async () => { setN(await count()); }, []); return <b onClick={async () => setN(await count())}>{n}</b>; }`}); err != nil {
		t.Errorf("async action call: %v", err)
	}
	// parameter names from Go reach the checker's messages (and the hover signature, see TestHoverAndDefinition)
	_, err = actionChecker(t, map[string]string{"a.client.tsx": `import { toggle } from "host:todos";
export default function A() { const f = async () => { await toggle(); }; return <b onClick={f}/>; }`})
	if err == nil || !strings.Contains(err.Error(), "missing argument id") {
		t.Errorf("a missing argument should be named after the Go parameter: %v", err)
	}
	// a non-exported meta in a page is diagnosed
	c2, err := actionChecker(t, map[string]string{"/x/app/pages/p.server.tsx": `import type { PageProps, Meta } from "gotsx";
function meta(): Meta { return { title: "x" }; }
export default function P(p: PageProps) { return <b>{meta().title}</b>; }`})
	if err != nil {
		t.Fatal(err)
	}
	if msg := pageProblem(c2, c2.Modules["/x/app/pages/p.server.tsx"], "page"); !strings.Contains(msg, "not exported") {
		t.Errorf("non-exported meta: %q", msg)
	}
	// GenActions rejects parameters from foreign packages
	c3, _ := NewChecker(foreignHostJSON)
	if _, err := GenActions(c3, "gen", "github.com/childrentime/gotsx/runtime", "x/host"); err == nil || !strings.Contains(err.Error(), "package \"time\"") {
		t.Errorf("foreign package parameter: %v", err)
	}
}

type Clock struct{}

func (Clock) At(when time.Time) string { return "" }

var foreignHostJSON = func() []byte {
	_, j := gotsx.GenerateHost(map[string]gotsx.HostModule{"clock": {Value: Clock{}, Go: "host.Clock", Actions: []string{"At"}}}, "host")
	return []byte(j)
}()
