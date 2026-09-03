package compiler

import (
	"strings"
	"testing"
	"testing/fstest"

	gotsx "github.com/childrentime/gotsx/runtime"
)

// 测试用的宿主模块: host:data 暴露 models(store) 和 Model 类型
type Model struct {
	ID    string   `json:"id"`
	Title string   `json:"title"`
	Price int      `json:"price"`
	Tags  []string `json:"tags"`
}
type Store struct{}

func (Store) List() []Model                { return nil }
func (Store) Get(id string) (Model, error) { return Model{}, nil }
func (Store) Search(q string) []Model      { return nil }

type Data struct {
	Models Store `json:"models"`
}

var testHostJSON = func() []byte {
	_, j := gotsx.GenerateHost(map[string]gotsx.HostModule{"data": {Value: Data{}, Go: "host.Data"}}, "host")
	return []byte(j)
}()

// compileOne 跑完整前端 + 两个后端, 返回生成的 Go / JS。stripLine 去掉 //line 指令(路径随机)
func compileOne(t *testing.T, file, src string) (goSrc, jsSrc string) {
	t.Helper()
	c, err := NewChecker(testHostJSON)
	if err != nil {
		t.Fatalf("NewChecker: %v", err)
	}
	m, err := ParseModule(src, file)
	if err != nil {
		t.Fatalf("parse %s: %v", file, err)
	}
	c.AddModule(m)
	if err := c.CheckAll(); err != nil {
		t.Fatalf("check %s: %v", file, err)
	}
	gs, err := GenGo(c, m, "gen", "github.com/childrentime/gotsx/runtime", "github.com/childrentime/gotsx/example/host")
	if err != nil {
		t.Fatalf("GenGo %s: %v", file, err)
	}
	goSrc = stripLineDirectives(gs)
	if m.Kind != "server" {
		js, err := GenJS(c, m)
		if err != nil {
			t.Fatalf("GenJS %s: %v", file, err)
		}
		jsSrc = js
	}
	return goSrc, jsSrc
}

func stripLineDirectives(s string) string {
	var b strings.Builder
	for _, ln := range strings.Split(s, "\n") {
		if strings.HasPrefix(strings.TrimSpace(ln), "//line ") {
			continue
		}
		b.WriteString(ln)
		b.WriteByte('\n')
	}
	return b.String()
}

// compileErr 期望编译失败, 返回错误信息
func compileErr(file, src string) (string, bool) {
	c, err := NewChecker(testHostJSON)
	if err != nil {
		return err.Error(), true
	}
	m, err := ParseModule(src, file)
	if err != nil {
		return err.Error(), true
	}
	c.AddModule(m)
	if err := c.CheckAll(); err != nil {
		return err.Error(), true
	}
	if _, err := GenGo(c, m, "gen", "github.com/childrentime/gotsx/runtime", "github.com/childrentime/gotsx/example/host"); err != nil {
		return err.Error(), true
	}
	if m.Kind != "server" {
		if _, err := GenJS(c, m); err != nil {
			return err.Error(), true
		}
	}
	return "", false
}

// fakeClientFS: Build 需要客户端运行时文件; 测试里用占位内容
func fakeClientFS() fstest.MapFS {
	return fstest.MapFS{
		"runtime.js":       {Data: []byte("// runtime")},
		"loader.js":        {Data: []byte("// loader")},
		"idiomorph.esm.js": {Data: []byte("// idiomorph")},
	}
}

// compileMany: several modules at once (a store module, the island reading it, the page seeding it). Returns the
// generated Go / JS keyed by file name; JS is empty for server modules.
func compileMany(t *testing.T, files map[string]string) (goSrc, jsSrc map[string]string) {
	t.Helper()
	c, err := NewChecker(testHostJSON)
	if err != nil {
		t.Fatalf("NewChecker: %v", err)
	}
	var mods []*Module
	for f, src := range files {
		m, err := ParseModule(src, f)
		if err != nil {
			t.Fatalf("parse %s: %v", f, err)
		}
		c.AddModule(m)
		mods = append(mods, m)
	}
	if err := c.CheckAll(); err != nil {
		t.Fatalf("check: %v", err)
	}
	goSrc, jsSrc = map[string]string{}, map[string]string{}
	for _, m := range mods {
		gs, err := GenGo(c, m, "gen", "github.com/childrentime/gotsx/runtime", "github.com/childrentime/gotsx/example/host")
		if err != nil {
			t.Fatalf("GenGo %s: %v", m.File, err)
		}
		goSrc[m.File] = stripLineDirectives(gs)
		if m.Kind != "server" {
			js, err := GenJS(c, m)
			if err != nil {
				t.Fatalf("GenJS %s: %v", m.File, err)
			}
			jsSrc[m.File] = js
		}
	}
	return goSrc, jsSrc
}

// compileManyErr: like compileMany but expects a failure somewhere (check or either backend); returns the message
func compileManyErr(files map[string]string) (string, bool) {
	c, err := NewChecker(testHostJSON)
	if err != nil {
		return err.Error(), true
	}
	var mods []*Module
	for f, src := range files {
		m, err := ParseModule(src, f)
		if err != nil {
			return err.Error(), true
		}
		c.AddModule(m)
		mods = append(mods, m)
	}
	if err := c.CheckAll(); err != nil {
		return err.Error(), true
	}
	for _, m := range mods {
		if _, err := GenGo(c, m, "gen", "github.com/childrentime/gotsx/runtime", "github.com/childrentime/gotsx/example/host"); err != nil {
			return err.Error(), true
		}
		if m.Kind != "server" {
			if _, err := GenJS(c, m); err != nil {
				return err.Error(), true
			}
		}
	}
	return "", false
}
