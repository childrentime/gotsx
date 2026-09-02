package gotsx

import (
	"strings"
	"testing"
)

// 文本节点必须 HTML 转义(防 XSS)
func TestTextEscaping(t *testing.T) {
	got := Render(Text(`</script><img src=x onerror=alert(1)>`))
	if strings.Contains(got, "<img") || strings.Contains(got, "</script>") {
		t.Errorf("文本未转义: %q", got)
	}
	if !strings.Contains(got, "&lt;img") {
		t.Errorf("期望转义后的 &lt;img: %q", got)
	}
}

func TestDynEscaping(t *testing.T) {
	got := Render(Island("X", nil, Dyn(`"><script>alert(1)</script>`)))
	if strings.Contains(got, "<script>alert") {
		t.Errorf("Dyn 未转义: %q", got)
	}
}

// 属性值必须转义, 防止属性逃逸
func TestAttrEscaping(t *testing.T) {
	got := Render(El("a", []Attr{A("title", `" onmouseover="alert(1)`)}))
	if strings.Contains(got, `onmouseover="alert`) {
		t.Errorf("属性未转义, 可逃逸: %q", got)
	}
}

// 岛 props(进 HTML 属性的 JSON)必须转义
func TestIslandPropsEscaping(t *testing.T) {
	got := Render(Island("X", map[string]any{"title": `"><script>alert(1)</script>`}, nil))
	if strings.Contains(got, "<script>alert") {
		t.Errorf("岛 props 未转义: %q", got)
	}
}

// 全链路: 一个带 XSS 载荷的"商品标题"经过 El+Text 仍然安全
func TestPipelineXSS(t *testing.T) {
	title := `<script>steal()</script>`
	page := El("div", []Attr{A("class", "card")}, El("h1", nil, Text(title)))
	got := Render(page)
	if strings.Contains(got, "<script>steal") {
		t.Errorf("XSS 穿透渲染管线: %q", got)
	}
}
