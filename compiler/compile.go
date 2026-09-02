package compiler

// 编译流程: 扫描 app/**/*.tsx → 解析 → 检查 → Go 后端(gen/*.go + routes) → JS 后端(gen/client/*.js)

import (
	"encoding/json"
	"fmt"
	"go/format"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Config struct {
	AppDir     string // .../example/app
	OutDir     string // .../example/gen
	Module     string // Go module path of the app, e.g. gotsx/example
	HostPkg    string // Go import path of host package
	RuntimePkg string // gotsx/runtime
	ClientFS   fs.FS  // runtime.js / loader.js / idiomorph.esm.js
}

type Report struct {
	Modules []string
	Routes  []string
	Islands []string
}

// Diagnostic: 一条带位置的编译错误(给 gotsx check / LSP 用)
type Diagnostic struct {
	File string `json:"file"`
	Line int    `json:"line"`
	Col  int    `json:"col"`
	Msg  string `json:"msg"`
}

func (d Diagnostic) String() string { return fmt.Sprintf("%s:%d:%d: %s", d.File, d.Line, d.Col, d.Msg) }

// ListSources 列出 app/ 下的全部 .tsx(跳过 .gen), 排好序
func ListSources(appDir string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(appDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && (d.Name() == ".gen" || d.Name() == "node_modules") {
			return filepath.SkipDir
		}
		if !d.IsDir() && strings.HasSuffix(p, ".tsx") {
			files = append(files, p)
		}
		return nil
	})
	sort.Strings(files)
	return files, err
}

// load: 扫描 + 解析 + 类型检查。overlay 里的内容优先于磁盘(编辑器未保存的缓冲区)。
// 返回 checker 和全部诊断; 有诊断时 checker 可能不完整, 不应继续生成。
func load(appDir string, overlay map[string]string) (*Checker, []Diagnostic, error) {
	files, err := ListSources(appDir)
	if err != nil {
		return nil, nil, err
	}
	for f := range overlay { // 编辑器里新建、还没落盘的文件
		if strings.HasSuffix(f, ".tsx") && strings.HasPrefix(f, appDir) && !containsStr(files, f) {
			files = append(files, f)
		}
	}
	sort.Strings(files)
	if len(files) == 0 {
		return nil, nil, fmt.Errorf("no .tsx files under %s", appDir)
	}
	c, err := NewChecker(ReadHostJSON(appDir))
	if err != nil {
		return nil, nil, err
	}
	c.HostDTS = HostDTSPath(appDir)
	var diags []Diagnostic
	for _, f := range files {
		var src string
		if s, ok := overlay[f]; ok {
			src = s
		} else {
			b, err := os.ReadFile(f)
			if err != nil {
				return nil, nil, err
			}
			src = string(b)
		}
		m, err := ParseModule(src, f)
		if err != nil {
			diags = append(diags, toDiag(err))
			continue
		}
		c.AddModule(m)
	}
	if len(diags) > 0 {
		return c, diags, nil
	}
	if err := c.CheckAll(); err != nil {
		for _, e := range c.Errors {
			diags = append(diags, toDiag(e))
		}
	}
	return c, diags, nil
}

func containsStr(xs []string, s string) bool {
	for _, x := range xs {
		if x == s {
			return true
		}
	}
	return false
}

func toDiag(err error) Diagnostic {
	switch e := err.(type) {
	case *parseError:
		return Diagnostic{File: e.Pos.File, Line: e.Pos.Line, Col: e.Pos.Col, Msg: e.Msg}
	case *CheckError:
		return Diagnostic{File: e.Pos.File, Line: e.Pos.Line, Col: e.Pos.Col, Msg: e.Msg}
	case *genError:
		return Diagnostic{File: e.Pos.File, Line: e.Pos.Line, Col: e.Pos.Col, Msg: e.Msg}
	}
	return Diagnostic{Msg: err.Error()}
}

// Analyze: 只检查不落盘 —— 解析、类型检查、两个后端都跑一遍(生成到内存), 收集全部带位置的错误。
// gotsx check 和 LSP 用它; 返回的诊断按文件、行排序。
func Analyze(appDir string, overlay map[string]string) ([]Diagnostic, error) {
	c, diags, err := load(appDir, overlay)
	if err != nil {
		return nil, err
	}
	if len(diags) == 0 {
		files := make([]string, 0, len(c.Modules))
		for f := range c.Modules {
			files = append(files, f)
		}
		sort.Strings(files)
		pagesDir := filepath.Join(appDir, "pages")
		for _, f := range files {
			m := c.Modules[f]
			if _, err := GenGo(c, m, "gen", "rt", "host"); err != nil {
				diags = append(diags, toDiag(err))
				continue
			}
			if m.Kind != "server" {
				if _, err := GenJS(c, m); err != nil {
					diags = append(diags, toDiag(err))
					continue
				}
			}
			if strings.HasPrefix(f, pagesDir+string(filepath.Separator)) && m.Kind == "server" {
				if msg := pageProblem(c, m, pageKind(f)); msg != "" {
					diags = append(diags, Diagnostic{File: f, Line: 1, Col: 1, Msg: msg})
				}
			}
		}
	}
	sort.Slice(diags, func(i, j int) bool {
		if diags[i].File != diags[j].File {
			return diags[i].File < diags[j].File
		}
		if diags[i].Line != diags[j].Line {
			return diags[i].Line < diags[j].Line
		}
		return diags[i].Col < diags[j].Col
	})
	return diags, nil
}

// pageKind: pages/ 下文件的角色 —— page | layout (_layout) | notfound (_404) | error (_error) | ignored (其它 _ 开头)
func pageKind(f string) string {
	base := strings.TrimSuffix(filepath.Base(f), ".server.tsx")
	switch {
	case base == "_layout":
		return "layout"
	case base == "_404":
		return "notfound"
	case base == "_error":
		return "error"
	case strings.HasPrefix(base, "_"):
		return "ignored"
	}
	return "page"
}

// pageProblem: pages/ 下的服务端模块必须 export default 一个组件, props 按角色固定
func pageProblem(c *Checker, m *Module, kind string) string {
	if kind == "ignored" {
		return ""
	}
	if m.Default == nil || m.Default.Kind != SComp {
		return "a page must export default a component"
	}
	switch kind {
	case "layout":
		if m.Default.Comp.Props != c.layoutProps {
			return "a _layout component's props must be LayoutProps (PageProps + children)"
		}
	case "error":
		if m.Default.Comp.Props != c.errorProps {
			return "an _error component's props must be ErrorProps (PageProps + message)"
		}
	default:
		if m.Default.Comp.Props != c.pageProps {
			return "a page component's props must be PageProps"
		}
	}
	if sym := m.Scope.syms["meta"]; sym != nil && m.Exports["meta"] == nil && kind != "layout" && sym.Kind == SFunc {
		return "function meta is not exported: write `export function meta(props: PageProps): Meta` for the layout to see it"
	}
	if ms, ok := m.Exports["meta"]; ok && kind != "layout" {
		ft, _ := ms.Type.(*FnT)
		if ms.Kind != SFunc || ft == nil || ft.Ret != c.metaT || len(ft.Params) > 1 || (len(ft.Params) == 1 && ft.Params[0].Type != c.pageProps) {
			return "export function meta must have the signature (props: PageProps) => Meta (import Meta from \"gotsx\")"
		}
	}
	return ""
}

// metaExpr: the Go call of the page's meta export (zero value when absent)
func metaExpr(m *Module) string {
	ms, ok := m.Exports["meta"]
	if !ok {
		return "gotsx.Meta{}"
	}
	if ft, _ := ms.Type.(*FnT); ft != nil && len(ft.Params) == 1 {
		return ms.Go + "(p)"
	}
	return ms.Go + "()"
}

// wrapLayouts wraps inner in the _layout.server.tsx files along the page's directory chain (from pages/ down to its directory) — outer layouts outside.
// Returns the function body (with return): when layouts exist the page meta is computed once and every layout receives the same value
func wrapLayouts(c *Checker, pagesDir, file string, m *Module, inner string) string {
	dir := filepath.Dir(file)
	var chain []string
	for d := dir; ; d = filepath.Dir(d) {
		if lm, ok := c.Modules[filepath.Join(d, "_layout.server.tsx")]; ok && lm.Default != nil {
			chain = append([]string{lm.Default.Go}, chain...)
		}
		if d == pagesDir || filepath.Dir(d) == d {
			break
		}
	}
	if len(chain) == 0 {
		return "return " + inner
	}
	for i := len(chain) - 1; i >= 0; i-- {
		inner = chain[i] + "(gotsx.LayoutProps{PageProps: p, Meta: _m, Children: " + inner + "})"
	}
	return "_m := " + metaExpr(m) + "\n" + "return " + inner
}

func Build(cfg Config) (*Report, error) {
	c, diags, err := load(cfg.AppDir, nil)
	if err != nil {
		return nil, err
	}
	if len(diags) > 0 {
		var b strings.Builder
		for _, d := range diags {
			b.WriteString("  " + d.String() + "\n")
		}
		return nil, fmt.Errorf("compile failed:\n%s", b.String())
	}
	files := make([]string, 0, len(c.Modules))
	for f := range c.Modules {
		files = append(files, f)
	}
	sort.Strings(files)

	rep := &Report{}
	os.MkdirAll(cfg.OutDir, 0o755)
	clientDir := filepath.Join(cfg.OutDir, "client")
	os.MkdirAll(clientDir, 0o755)
	// 清掉旧产物
	old, _ := filepath.Glob(filepath.Join(cfg.OutDir, "*_gen.go"))
	for _, f := range old {
		os.Remove(f)
	}
	oldJS, _ := filepath.Glob(filepath.Join(clientDir, "*.js"))
	for _, f := range oldJS {
		os.Remove(f)
	}

	var genErrs []string
	pagesDir := filepath.Join(cfg.AppDir, "pages")
	var routes []string
	notFound, errorPage := "nil", "nil" // pages/_404.server.tsx / pages/_error.server.tsx
	for _, f := range files {
		m := c.Modules[f]
		rel, _ := filepath.Rel(cfg.AppDir, f)
		rep.Modules = append(rep.Modules, rel)
		src, err := GenGo(c, m, "gen", cfg.RuntimePkg, cfg.HostPkg)
		if err != nil {
			genErrs = append(genErrs, "  "+err.Error())
			continue
		}
		formatted, ferr := format.Source([]byte(src))
		if ferr != nil {
			genErrs = append(genErrs, fmt.Sprintf("  %s: generated Go does not format: %v\n%s", rel, ferr, numbered(src)))
			continue
		}
		name := strings.TrimSuffix(rel, ".tsx")
		name = strings.NewReplacer("/", "_", "[", "_", "]", "_", ".", "_", "-", "_").Replace(name)
		if err := os.WriteFile(filepath.Join(cfg.OutDir, name+"_gen.go"), formatted, 0o644); err != nil {
			return nil, err
		}
		if m.Kind != "server" {
			js, err := GenJS(c, m)
			if err != nil {
				genErrs = append(genErrs, "  "+err.Error())
				continue
			}
			if err := os.WriteFile(filepath.Join(clientDir, m.Name+".js"), []byte(js), 0o644); err != nil {
				return nil, err
			}
			if m.Kind == "client" {
				rep.Islands = append(rep.Islands, m.Name)
			}
		}
		if strings.HasPrefix(f, pagesDir+string(filepath.Separator)) && m.Kind == "server" {
			kind := pageKind(f)
			if msg := pageProblem(c, m, kind); msg != "" {
				genErrs = append(genErrs, fmt.Sprintf("  %s: %s", rel, msg))
				continue
			}
			switch kind {
			case "page":
				pattern, segs := routeOf(pagesDir, f)
				render := wrapLayouts(c, pagesDir, f, m, m.Default.Go+"(p)")
				routes = append(routes, fmt.Sprintf("\t{Pattern: %q, Segs: %s, Render: func(p gotsx.PageProps) gotsx.Node {\n\t\t%s\n\t}},", pattern, goStrSlice(segs), strings.ReplaceAll(render, "\n", "\n\t\t")))
				rep.Routes = append(rep.Routes, pattern)
			case "notfound":
				notFound = "func(p gotsx.PageProps) gotsx.Node {\n\t" + strings.ReplaceAll(wrapLayouts(c, pagesDir, f, m, m.Default.Go+"(p)"), "\n", "\n\t") + "\n}"
			case "error":
				errorPage = "func(p gotsx.PageProps, err error) gotsx.Node {\n\t" + strings.ReplaceAll(wrapLayouts(c, pagesDir, f, m, m.Default.Go+"(gotsx.ErrorProps{PageProps: p, Message: err.Error()})"), "\n", "\n\t") + "\n}"
			}
		}
	}
	if len(genErrs) > 0 {
		return nil, fmt.Errorf("codegen failed:\n%s", strings.Join(genErrs, "\n"))
	}
	rt := fmt.Sprintf("// Code generated by gotsx. DO NOT EDIT.\n\npackage gen\n\nimport gotsx %q\n\nvar Routes = []gotsx.Route{\n%s\n}\n\n"+
		"// NotFound / ErrorPage come from pages/_404.server.tsx and pages/_error.server.tsx (nil when absent); pass them to gotsx.Options.\n"+
		"var NotFound func(gotsx.PageProps) gotsx.Node = %s\n\nvar ErrorPage func(gotsx.PageProps, error) gotsx.Node = %s\n",
		cfg.RuntimePkg, strings.Join(routes, "\n"), notFound, errorPage)
	if fs, err := format.Source([]byte(rt)); err == nil { // gofmt-clean route table (layout wrapping is a multi-line function body)
		rt = string(fs)
	}
	if err := os.WriteFile(filepath.Join(cfg.OutDir, "routes_gen.go"), []byte(rt), 0o644); err != nil {
		return nil, err
	}
	ij, _ := json.Marshal(rep.Islands)
	if err := os.WriteFile(filepath.Join(clientDir, "islands.json"), ij, 0o644); err != nil {
		return nil, err
	}
	acts, err := GenActions(c, "gen", cfg.RuntimePkg, cfg.HostPkg)
	if err != nil {
		return nil, err
	}
	if fs, err := format.Source([]byte(acts)); err == nil {
		acts = string(fs)
	}
	if err := os.WriteFile(filepath.Join(cfg.OutDir, "actions_gen.go"), []byte(acts), 0o644); err != nil {
		return nil, err
	}
	// 内嵌客户端资源, 支持单二进制部署
	assets := "// Code generated by gotsx. DO NOT EDIT.\n\npackage gen\n\n" +
		"import (\n\t\"embed\"\n\t\"io/fs\"\n)\n\n" +
		"//go:embed client\n" +
		"var clientEmbed embed.FS\n\n" +
		"// ClientFS holds the embedded client assets (single-binary deploys)\n" +
		"var ClientFS, _ = fs.Sub(clientEmbed, \"client\")\n"
	if err := os.WriteFile(filepath.Join(cfg.OutDir, "assets_gen.go"), []byte(assets), 0o644); err != nil {
		return nil, err
	}
	// 客户端运行时
	for _, n := range []string{"runtime.js", "loader.js", "idiomorph.esm.js"} {
		b, err := fs.ReadFile(cfg.ClientFS, n)
		if err != nil {
			return nil, fmt.Errorf("read embedded %s: %w", n, err)
		}
		if err := os.WriteFile(filepath.Join(clientDir, n), b, 0o644); err != nil {
			return nil, err
		}
	}
	// 编辑器用的类型声明: "gotsx" 模块 + 全局内建(让 VS Code 的 TS 服务不再报 找不到模块)
	genDir := filepath.Join(cfg.AppDir, ".gen")
	os.MkdirAll(genDir, 0o755)
	if err := os.WriteFile(filepath.Join(genDir, "gotsx.d.ts"), []byte(GotsxDTS), 0o644); err != nil {
		return nil, err
	}
	return rep, nil
}

// GotsxDTS 是 app/.gen/gotsx.d.ts 的内容: 方言对编辑器暴露的类型面。编译器自己不读它, 只给 TS 语言服务看。
const GotsxDTS = `// Generated by gotsx. DO NOT EDIT. Editor-only typings for the gotsx dialect.

declare module "gotsx" {
  /** A rendered node (JSX). */
  export type Node = JSX.Element;
  /** A one-shot message queued by an action (req.Session().Flash) and delivered to the next page. */
  export interface Flash { kind: string; text: string; }
  /** Props every page component receives. */
  export interface PageProps {
    params: Record<string, string>;
    query: Record<string, string>;
    path: string;
    locale: string;
    cookies: Record<string, string>;
    /** Session values (read-only here; write them in actions via req.Session()). */
    session: Record<string, string>;
    /** Flash messages consumed by this render. */
    flash: Flash[];
    /** CSRF token for classic <form method="post"> handlers (name="_csrf"). */
    csrf: string;
  }
  /** Props of pages/**\/_layout.server.tsx: the page props plus the wrapped content. */
  /** Page metadata: a page exports "function meta(props: PageProps): Meta"; layouts read it as props.meta. */
  export interface Meta { title?: string; description?: string; canonical?: string; image?: string; noIndex?: boolean; }
  export interface LayoutProps extends PageProps { meta: Meta; children: Node; }
  /** Props of pages/_error.server.tsx. */
  export interface ErrorProps extends PageProps { message: string; }
  /** Streaming boundary (server components only): the fallback ships with the shell, children render afterwards in their own goroutine. */
  export function Suspense(props: { fallback: Node; children?: Node }): Node;
  /** Server: the initial value. Client: a signal. */
  export function useState<T>(init: T): [T, (v: T | ((prev: T) => T)) => void];
  /** Client only. deps [] = run once on mount; otherwise dependencies are tracked automatically. */
  export function useEffect(fn: () => void, deps?: unknown[]): void;
  export function useMemo<T>(fn: () => T): T;
  /** Cross-island event bus (client only). */
  export function emit(name: string, detail?: unknown): void;
  export function on(name: string, fn: (detail: any) => void): void;
}

declare global {
  namespace JSX {
    interface Element {}
    interface ElementChildrenAttribute { children: {} }
    interface IntrinsicElements { [tag: string]: any }
  }
  /** i18n (server + client). */
  function t(locale: string, key: string): string;
  function tv(locale: string, key: string, vars: Record<string, string>): string;
  function plural(locale: string, key: string, n: number): string;
  function fmtNum(locale: string, n: number): string;
  function fmtCur(locale: string, cents: number): string;
  function fmtDate(locale: string, iso: string): string;
  function lpath(locale: string, path: string): string;
  /** Safe JSON-LD <script> (pass JSON.stringify(...)). */
  function jsonLd(json: string): JSX.Element;
  /** Server only: abort rendering and respond with a redirect (default 302). */
  function redirect(url: string, status?: number): JSX.Element;
  /** Server only: abort rendering and respond with the 404 page. */
  function notFound(): JSX.Element;
}

export {};
`

// routeOf: pages/a/[id].server.tsx → /a/{id}; pages/docs/[...slug].server.tsx → /docs/{...slug}(catch-all)
func routeOf(pagesDir, file string) (string, []string) {
	r, _ := filepath.Rel(pagesDir, file)
	r = strings.TrimSuffix(r, ".server.tsx")
	r = strings.TrimSuffix(r, "index")
	r = strings.Trim(filepath.ToSlash(r), "/")
	var segs []string
	if r != "" {
		for _, s := range strings.Split(r, "/") {
			if strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") {
				s = "{" + s[1:len(s)-1] + "}" // [id] → {id}, [...slug] → {...slug}
			}
			segs = append(segs, s)
		}
	}
	return "/" + strings.Join(segs, "/"), segs
}

func goStrSlice(ss []string) string {
	if len(ss) == 0 {
		return "nil"
	}
	var q []string
	for _, s := range ss {
		q = append(q, fmt.Sprintf("%q", s))
	}
	return "[]string{" + strings.Join(q, ", ") + "}"
}

func numbered(src string) string {
	var b strings.Builder
	for i, ln := range strings.Split(src, "\n") {
		fmt.Fprintf(&b, "%4d  %s\n", i+1, ln)
	}
	return b.String()
}
