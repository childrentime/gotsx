import type { PageProps } from "gotsx";
import DocsLayout from "../../components/DocsLayout.server";
import Section from "../../components/Section.server";
import Card from "../../ui/Card";
import Callout from "../../ui/Callout";
import CodeBlock from "../../ui/CodeBlock.server";
import { loc } from "../../ui/i18n";
import { sampleHTML, sampleIslandJS, sampleIslandGo } from "../../content/site.server";

interface Step {
  n: string;
  title: string;
  desc: string;
}

function pipeline(locale: string): Step[] {
  return [
    { n: "1", title: loc(locale, "Parse", "解析"), desc: loc(locale, "A hand-written parser for the TSX subset (lexer + recursive descent + JSX mode), about 900 lines of Go.", "自己写的 TSX 子集 parser(词法 + 递归下降 + JSX 模式), 约 900 行 Go。") },
    { n: "2", title: loc(locale, "Type-check", "类型检查"), desc: loc(locale, "Local inference + boundary annotations: props, host signatures, hooks, the builtin method table. Stepping outside the subset errors.", "局部推断 + 边界标注: props、宿主签名、hooks、内建方法表。出子集就报错。") },
    { n: "3", title: loc(locale, "Reactivity analysis", "响应性判定"), desc: loc(locale, "Which expressions read a signal/memo. This single analysis drives both backends, so the hydration markers are guaranteed to line up.", "哪个表达式读到了 signal/memo。这个判定同时驱动两个后端, 所以 hydrate 标记一定对得上。") },
    { n: "4", title: loc(locale, "Go backend", "Go 后端"), desc: loc(locale, "Component → Go function, JSX → gotsx.El/Text/If/Nodes, hooks → single-pass semantics, host:* → direct call. gofmt validates the syntax.", "组件 → Go 函数, JSX → gotsx.El/Text/If/Nodes, hooks → 单趟语义, host:* → 直接调用。gofmt 校验语法。") },
    { n: "5", title: loc(locale, "JS backend", "JS 后端"), desc: loc(locale, "Component → function, useState → G.signal, a signal-dependent const → G.memo, JSX → G.el/t/text/cond/each.", "组件 → 函数, useState → G.signal, 依赖 signal 的 const → G.memo, JSX → G.el/t/text/cond/each。") },
    { n: "6", title: "go build", desc: loc(locale, "The generated Go compiles together with your main.go and host package into a single binary; the client JS is served directly as ES modules.", "生成的 Go 和你的 main.go、host 包一起编成一个二进制; 客户端 JS 直接作为 ES module 服务。") },
  ];
}

export default function Architecture({ locale, path }: PageProps) {
  const lc = locale !== "" ? locale : "en";
  return (
    <DocsLayout title={loc(lc, "Architecture & internals", "架构与原理")} active="architecture" locale={lc} path={path}>
      <p class="text-[15px] leading-7">
        {loc(lc, "The core insight is a single sentence: ", "核心洞察只有一句: ")}<strong>{loc(lc, "SSR is a single synchronous one-pass evaluation.", "SSR 是一次同步的单趟求值。")}</strong>{loc(lc, " No re-render, no effects, setters are never called. So the server needs no React runtime, only the component's \"render slice\" — and its semantics are small enough to compile to Go.", " 没有重渲染、没有 effect、setter 永远不会被调用。所以服务端不需要 React 运行时, 只需要组件的\"渲染切片\"——而它的语义小到可以编译成 Go。")}
      </p>

      <Section title={loc(lc, "The compile pipeline", "编译流水线")}>
        <div class="grid gap-3 sm:grid-cols-2">
          {pipeline(lc).map((s) => (
            <Card title={`${s.n}. ${s.title}`}>{s.desc}</Card>
          ))}
        </div>
      </Section>

      <Section title={loc(lc, "What hooks are in Go", "hooks 在 Go 里是什么")} lead={loc(lc, "The Go output of the same Counter.client.tsx", "同一份 Counter.client.tsx 的 Go 输出")}>
        <CodeBlock code={sampleIslandGo} lang="go" title="gen/islands_Counter_client_gen.go" />
        <ul class="list-disc space-y-1 pl-5">
          <li><code class="font-mono text-sm">useState(start)</code>{loc(lc, " → ", " → ")}<code class="font-mono text-sm">var n float64 = props.Start</code>{loc(lc, "; the setter is an empty function.", "; setter 是空函数。")}</li>
          <li><code class="font-mono text-sm">useEffect</code>{loc(lc, ", event handlers, async functions → not generated.", "、事件处理器、async 函数 → 不生成。")}</li>
          <li>{loc(lc, "The reactive ", "响应式的 ")}<code class="font-mono text-sm">{"{n}"}</code>{loc(lc, " → ", " → ")}<code class="font-mono text-sm">gotsx.Dyn</code>{loc(lc, " (with a marker); static text → ", "(带标记); 静态文本 → ")}<code class="font-mono text-sm">gotsx.Text</code>{loc(lc, ".", "。")}</li>
          <li>{loc(lc, "The island shell ", "岛的外壳 ")}<code class="font-mono text-sm">gotsx.Island(name, props, inner)</code>{loc(lc, " serializes props into an attribute and turns on marker mode.", " 把 props 序列化进属性, 并打开标记模式。")}</li>
        </ul>
      </Section>

      <Section title={loc(lc, "Resumable hydration", "走位 hydrate")} lead={loc(lc, "The HTML the server emits (excerpt)", "服务端输出的 HTML(节选)")}>
        <CodeBlock code={sampleHTML} lang="html" title={loc(lc, "the markers inside an island", "岛内部的标记")} />
        <p>{loc(lc, "The client receives JS of the same structure:", "客户端拿到同一份结构的 JS:")}</p>
        <CodeBlock code={sampleIslandJS} lang="js" title="gen/client/Counter.js" />
        <p>
          <code class="font-mono text-sm">G.el</code>{loc(lc, " claims the next element, ", " 认领下一个元素, ")}<code class="font-mono text-sm">G.t</code>{loc(lc, " claims the next text node, ", " 认领下一个文本节点, ")}<code class="font-mono text-sm">G.text</code>{loc(lc, " claims the text between ", " 认领 ")}<code class="font-mono text-sm">{"<!--$-->…<!--/-->"}</code>{loc(lc, " and binds an effect, ", " 之间的文本并绑 effect, ")}<code class="font-mono text-sm">G.cond / G.each</code>{loc(lc, " claim the ", " 认领 ")}<code class="font-mono text-sm">{"<!--[-->…<!--]-->"}</code>{loc(lc, " block. Children are deferred with thunks until after the parent is claimed, so the order is source order. No diff, no rebuild, the existing DOM is reused as-is.", " 区块。子节点用 thunk 延迟到父元素认领之后再执行, 顺序就是源码顺序。没有 diff, 没有重建, 已有 DOM 原样复用。")}
        </p>
        <Callout kind="info">{loc(lc, "Why it can be this simple: both sides' structure comes from the same compiler and the same reactivity analysis. React needs a diff because it doesn't know what the server actually rendered.", "为什么能这么简单: 两端的结构来自同一个编译器、同一个响应性判定。React 需要 diff 是因为它不知道服务端到底渲染了什么。")}</Callout>
      </Section>

      <Section title={loc(lc, "Host modules & the fence", "宿主模块与围栏")}>
        <ul class="list-disc space-y-2 pl-5">
          <li>{loc(lc, "A host module is a Go value; ", "宿主模块是 Go 值; ")}<code class="font-mono text-sm">hostgen</code>{loc(lc, " reflects it into ", " 反射它生成 ")}<code class="font-mono text-sm">host.d.ts</code>{loc(lc, " (for the editor) and ", "(给编辑器)和 ")}<code class="font-mono text-sm">host.json</code>{loc(lc, " (for the compiler, with Go names and numeric types).", "(给编译器, 含 Go 名和数值类型)。")}</li>
          <li>{loc(lc, "What the dialect can do is exactly what Go exposes. Go is the single source of truth: routing, data, permissions, and actions all live in Go.", "方言能做的严格等于 Go 暴露的。Go 是唯一的真相源: 路由、数据、权限、动作都在 Go。")}</li>
          <li>{loc(lc, "The fence is at the type-check stage: a client importing host, a client importing a server component, await on the server, an unknown prop, member access on any — all compile errors.", "围栏在类型检查阶段: 客户端 import host、客户端 import 服务端组件、服务端出现 await、未知 prop、any 上取成员——全是编译错误。")}</li>
        </ul>
      </Section>

      <Section title={loc(lc, "SPA navigation", "SPA 跳转")}>
        <p>{loc(lc, "Click a link → fetch the new page's HTML → idiomorph morphs the body over → islands survive by DOM identity (same name and props stay put, changed props rebuild, gone ones unmount). Form GET is the same; back/forward go through popstate; concurrent navigations honor only the last; a non-HTML response falls back to a full page load.", "点击链接 → fetch 新页 HTML → idiomorph 把 body morph 过去 → 岛按 DOM 同一性存活(同名同 props 不动, props 变了重建, 消失了卸载)。表单 GET 同理; 后退/前进走 popstate; 并发导航只认最后一次; 非 HTML 响应退回整页加载。")}</p>
      </Section>

      <Section title={loc(lc, "Compared with the goja route", "和 goja 路线的对比")} lead={loc(lc, "Same machine, same kind of page", "同一台机器, 同一类页面")}>
        <div class="overflow-x-auto rounded-lg border border-border">
          <table class="table">
            <thead><tr><th>{loc(lc, "Item", "项")}</th><th>goja + React + MUI</th><th>gotsx</th></tr></thead>
            <tbody>
              <tr><td class="text-muted-foreground">{loc(lc, "list page render", "列表页渲染")}</td><td class="text-muted-foreground">45–60 ms</td><td class="font-medium">~30 µs</td></tr>
              <tr><td class="text-muted-foreground">{loc(lc, "throughput", "吞吐")}</td><td class="text-muted-foreground">{loc(lc, "~31 req/s (4 VMs)", "~31 req/s(4 VM)")}</td><td class="font-medium">~28k req/s</td></tr>
              <tr><td class="text-muted-foreground">{loc(lc, "client runtime", "客户端运行时")}</td><td class="text-muted-foreground">react-dom 62 KB gz</td><td class="font-medium">6 KB</td></tr>
              <tr><td class="text-muted-foreground">{loc(lc, "debugging", "调试")}</td><td class="text-muted-foreground">{loc(lc, "no breakpoints", "没有断点")}</td><td class="font-medium">delve / pprof / go test</td></tr>
              <tr><td class="text-muted-foreground">{loc(lc, "npm ecosystem", "npm 生态")}</td><td class="text-muted-foreground">{loc(lc, "pure-JS packages work", "纯 JS 包可用")}</td><td class="text-muted-foreground">{loc(lc, "unavailable; libraries are written in the dialect", "不可用, 库用方言写")}</td></tr>
            </tbody>
          </table>
        </div>
      </Section>
    </DocsLayout>
  );
}
