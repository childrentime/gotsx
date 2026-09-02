package gotsx

import (
	"strings"
	"testing"
)

func TestRenderElement(t *testing.T) {
	n := El("a", []Attr{A("href", "/x"), A("class", "c")}, Text("hi"))
	got := Render(n)
	if got != `<a href="/x" class="c">hi</a>` {
		t.Errorf("got %q", got)
	}
}

func TestVoidElement(t *testing.T) {
	if Render(El("input", []Attr{A("name", "q")})) != `<input name="q">` {
		t.Error("void 元素不应有闭合标签")
	}
	if Render(El("br", nil)) != `<br>` {
		t.Error("br")
	}
}

func TestBoolAttr(t *testing.T) {
	if Render(El("input", []Attr{AB("disabled", true)})) != `<input disabled>` {
		t.Error("bool true")
	}
	if Render(El("input", []Attr{AB("disabled", false)})) != `<input>` {
		t.Error("bool false 应省略")
	}
}

func TestFragAndNodes(t *testing.T) {
	got := Render(Frag(Text("a"), Text("b")))
	if got != "ab" {
		t.Errorf("frag: %q", got)
	}
	got = Render(NodesPlain([]Node{Text("x"), Text("y")}))
	if got != "xy" {
		t.Errorf("nodes: %q", got)
	}
}

// 岛内部动态部分带 hydrate 走位标记
func TestIslandMarkers(t *testing.T) {
	inner := Frag(Dyn("5"), If(true, func() Node { return Text("A") }))
	got := Render(Island("Counter", map[string]any{"start": 0}, inner))
	for _, want := range []string{`name="Counter"`, `props="{&#34;start&#34;:0}"`, "<!--$-->5<!--/-->", "<!--[-->A<!--]-->"} {
		if !strings.Contains(got, want) {
			t.Errorf("岛标记缺少 %q\n%s", want, got)
		}
	}
}

// 岛外(纯服务端)不带标记
func TestNoMarkersOutsideIsland(t *testing.T) {
	got := Render(Frag(Dyn("5"), If(true, func() Node { return Text("A") })))
	if strings.Contains(got, "<!--") {
		t.Errorf("岛外不应有标记: %q", got)
	}
	if got != "5A" {
		t.Errorf("got %q", got)
	}
}
