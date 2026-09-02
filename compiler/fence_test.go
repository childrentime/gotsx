package compiler

import (
	"regexp"
	"strings"
	"testing"
)

var posRe = regexp.MustCompile(`\.tsx:\d+:\d+`)

// 方言围栏: 每种违规都必须报错, 且带 文件:行:列 位置
func TestFence(t *testing.T) {
	cases := []struct {
		name string
		file string
		src  string
		want string // 错误信息应包含
	}{
		{"客户端import宿主值", "c.client.tsx",
			`import { models } from "host:data";
export default function C() { return <b>{models.list().length}</b>; }`, "只能在服务端"},
		{"客户端import服务端组件", "c.client.tsx",
			`import X from "./page.server";
export default function C() { return <X />; }`, "找不到模块"},
		{"缺少必填prop", "c.server.tsx",
			`function Child({ n }: { n: number }) { return <b>{n}</b>; }
export default function C() { return <Child />; }`, "缺少 prop"},
		{"不存在的prop", "c.server.tsx",
			`function Child({ n }: { n: number }) { return <b>{n}</b>; }
export default function C() { return <Child n={1} extra="x" />; }`, "没有 prop"},
		{"用了双等号", "c.server.tsx",
			`export default function C() { const a = 1; return <b>{a == 1 ? "y" : "n"}</b>; }`, "=== / !=="},
		{"class语法", "c.server.tsx",
			`class Foo {}
export default function C() { return <b>1</b>; }`, "不支持 class"},
		{"未定义标识符", "c.server.tsx",
			`export default function C() { return <b>{ghost}</b>; }`, "未定义的标识符"},
		{"空数组无类型", "c.server.tsx",
			`export default function C() { const xs = []; return <b>{xs.length}</b>; }`, "空数组"},
		{"数字上调字符串方法", "c.server.tsx",
			`export default function C() { const n = 1; return <b>{n.toUpperCase()}</b>; }`, "数字方法"},
		{"岛props含Node", "c.client.tsx",
			`import type { Node } from "gotsx";
export default function C({ children }: { children?: Node }) { return <div>{children}</div>; }`, "JSON 序列化"},
		{"属性展开", "c.server.tsx",
			`export default function C() { const p = { a: 1 }; return <div {...p} />; }`, "属性展开"},
		{"未知宿主模块", "c.server.tsx",
			`import { x } from "host:nope";
export default function C() { return <b>{x}</b>; }`, "宿主模块"},
		{"while循环", "c.server.tsx",
			`export default function C() { while (true) {} return <b>1</b>; }`, "while"},
		{"new表达式", "c.server.tsx",
			`export default function C() { const d = new Thing(); return <b>1</b>; }`, "不支持 new"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			msg, failed := compileErr(tc.file, tc.src)
			if !failed {
				t.Fatalf("期望编译失败, 但通过了")
			}
			if !strings.Contains(msg, tc.want) {
				t.Errorf("错误信息应含 %q, 实际: %s", tc.want, msg)
			}
			if !posRe.MatchString(msg) {
				t.Errorf("错误信息应带 文件:行:列 位置, 实际: %s", msg)
			}
		})
	}
}

// 合法子集代码不应报错
func TestValidPrograms(t *testing.T) {
	cases := []string{
		`import { useState } from "gotsx";
export default function C() {
  const [n, setN] = useState(0);
  const double = n * 2;
  return <button onClick={() => setN(n + 1)}>{n} {double}{n > 4 && <b>!</b>}</button>;
}`,
		`import type { PageProps } from "gotsx";
import { models } from "host:data";
export default function C({ query }: PageProps) {
  const list = models.search(query.q ?? "");
  return <ul>{list.map((m) => <li>{m.title} {m.price}</li>)}</ul>;
}`,
	}
	for i, src := range cases {
		file := "c.client.tsx"
		if strings.Contains(src, "host:") {
			file = "c.server.tsx"
		}
		if msg, failed := compileErr(file, src); failed {
			t.Errorf("用例 %d 本应通过, 却报错: %s", i, msg)
		}
	}
}
