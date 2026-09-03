import type { PageProps, Meta } from "gotsx";
import DocsLayout from "../../components/DocsLayout.server";
import Section from "../../components/Section.server";
import Callout from "../../ui/Callout";
import CodeBlock from "../../ui/CodeBlock.server";
import { loc } from "../../ui/i18n";
import { sampleAction, sampleActionIsland, sampleActionErrors, sampleActionCatch, sampleForm, sampleFormHandler, sampleMeta } from "../../content/site.server";

export function meta({ locale }: PageProps): Meta {
  const lc = locale !== "" ? locale : "en";
  return { title: loc(lc, "Actions & sessions", "Action 与会话"), description: loc(lc, "How islands call Go: typed actions, error mapping, signed sessions, flash messages, CSRF tokens and page meta.", "岛怎么调用 Go: 类型化 action、错误映射、签名会话、flash 消息、CSRF token 与页面 meta。") };
}

export default function Actions({ locale, path }: PageProps) {
  const lc = locale !== "" ? locale : "en";
  return (
    <DocsLayout title={loc(lc, "Actions & sessions", "Action 与会话")} active="actions" locale={lc} path={path}>
      <p class="text-[15px] leading-7">
        {loc(lc, "A server component calls Go directly. An island runs in the browser, so its way back to Go is HTTP — but you never write the HTTP: list a Go method in ", "服务端组件直接调用 Go。岛跑在浏览器里, 回到 Go 只能走 HTTP —— 但 HTTP 不用你写: 把 Go 方法列进 ")}<code class="font-mono text-sm">Actions</code>{loc(lc, ", import it in the island, await it. The compiler generates the route, the JSON decoding, the same-origin and header checks, the error mapping and the client stub, and the return type comes from the Go signature.", ", 在岛里 import 然后 await。路由、JSON 解码、同源与标头校验、错误映射和客户端桩由编译器生成, 返回类型来自 Go 签名。")}
      </p>

      <Section title={loc(lc, "1. Declare an action", "1. 声明一个 action")} lead={loc(lc, "Module-level methods only; a *gotsx.Req first parameter gets the request injected", "只能是模块级方法; 第一个参数写 *gotsx.Req 就能拿到请求")}>
        <CodeBlock code={sampleAction} lang="go" title="host/host.go" />
        <p class="mt-3">{loc(lc, "hostgen reflects the method into ", "hostgen 把方法反射进 ")}<code class="font-mono text-sm">app/.gen/host.d.ts</code>{loc(lc, " as ", " 变成 ")}<code class="font-mono text-sm">like(id: string): Promise&lt;number&gt;</code>{loc(lc, " — parameter names come from the Go source. Arguments must be builtin or host types (results may be anything hostgen can reflect).", " —— 参数名来自 Go 源码。参数必须是内建类型或 host 类型(返回值只要 hostgen 能反射即可)。")}</p>
      </Section>

      <Section title={loc(lc, "2. Call it from an island", "2. 在岛里调用")} lead={loc(lc, "A value import from host:* is allowed for actions; everything else stays import type", "对 action 允许从 host:* 做值导入; 其它成员仍只能 import type")}>
        <CodeBlock code={sampleActionIsland} lang="tsx" title="app/islands/LikeButton.client.tsx" />
        <Callout kind="info" title={loc(lc, "Render is synchronous", "渲染是同步的")}>
          {loc(lc, "Calling an action in a component body is a compile error. Call it from a handler or an effect: ", "在组件体里调用 action 是编译错误。放进事件处理器或 effect 里: ")}<code class="font-mono text-sm">onClick={"{() => toggle(id)}"}</code>{loc(lc, " fires and forgets; ", " 发出就不管; ")}<code class="font-mono text-sm">await</code>{loc(lc, " inside an async handler gives you the result.", " 在 async 处理器里能拿到结果。")}
        </Callout>
      </Section>

      <Section title={loc(lc, "3. Errors become statuses", "3. 错误变成状态码")} lead={loc(lc, "The island's catch sees e.status, e.fields and e.message", "岛里的 catch 拿到 e.status、e.fields、e.message")}>
        <CodeBlock code={sampleActionErrors} lang="go" title="host/host.go" />
        <CodeBlock code={sampleActionCatch} lang="tsx" title="app/islands/RenameForm.client.tsx" />
        <table class="table mt-4">
          <thead><tr><th>Go</th><th>HTTP</th><th>{loc(lc, "island", "岛")}</th></tr></thead>
          <tbody>
            <tr><td><code class="font-mono text-sm">gotsx.Invalid(fields)</code></td><td>422</td><td><code class="font-mono text-sm">e.fields</code></td></tr>
            <tr><td><code class="font-mono text-sm">gotsx.Fail(msg)</code></td><td>400</td><td><code class="font-mono text-sm">e.message</code></td></tr>
            <tr><td><code class="font-mono text-sm">gotsx.Unauthorized(msg)</code> / <code class="font-mono text-sm">Forbidden(msg)</code></td><td>401 / 403</td><td><code class="font-mono text-sm">e.status</code></td></tr>
            <tr><td><code class="font-mono text-sm">fmt.Errorf("%w", gotsx.ErrNotFound)</code></td><td>404</td><td></td></tr>
            <tr><td>{loc(lc, "any other error or panic", "其它错误或 panic")}</td><td>500</td><td>{loc(lc, "message only in dev", "消息只在 dev 显示")}</td></tr>
          </tbody>
        </table>
      </Section>

      <Section title={loc(lc, "4. Sessions, flash messages, CSRF", "4. 会话、flash 消息、CSRF")} lead={loc(lc, "A signed cookie; pages read it, actions and handlers write it", "签名 cookie; 页面读, action 和 handler 写")}>
        <p>{loc(lc, "Pages receive ", "页面拿到 ")}<code class="font-mono text-sm">props.session</code>{loc(lc, " (read-only string values), ", "(只读键值)、")}<code class="font-mono text-sm">props.flash</code>{loc(lc, " (one-shot messages, consumed by the render) and ", "(一次性消息, 渲染一次即消费)和 ")}<code class="font-mono text-sm">props.csrf</code>{loc(lc, " (a token for classic forms, created lazily so pages that don't use it set no cookie). Actions write through ", "(经典表单用的 token, 惰性生成, 不用它的页面不会种 cookie)。action 通过 ")}<code class="font-mono text-sm">req.Session().Set / Flash / Clear</code>{loc(lc, "; Go handlers through ", " 写; Go handler 通过 ")}<code class="font-mono text-sm">gotsx.SessionOf(r)</code>{loc(lc, " and ", " 和 ")}<code class="font-mono text-sm">sess.Save(w, r)</code>{loc(lc, ". Set ", "。生产环境设置 ")}<code class="font-mono text-sm">SESSION_SECRET</code>{loc(lc, " in production; without it every start signs with a fresh random key.", "; 不设的话每次启动都用新的随机密钥。")}</p>
        <CodeBlock code={sampleForm} lang="tsx" title="app/pages/todos.server.tsx" />
        <CodeBlock code={sampleFormHandler} lang="go" title="main.go" />
      </Section>

      <Section title={loc(lc, "5. Page meta", "5. 页面 meta")} lead={loc(lc, "export function meta next to the page; the layout renders it", "页面旁边 export function meta, 由布局渲染")}>
        <CodeBlock code={sampleMeta} lang="tsx" title="app/pages" />
        <p class="mt-3">{loc(lc, "Fields: title, description, canonical, image, noIndex — all optional. meta runs before the page in the same request, so keep the host call it makes cheap.", "字段: title、description、canonical、image、noIndex, 全部可选。meta 在同一请求里先于页面执行, 它调用的宿主方法要便宜。")}</p>
      </Section>
    </DocsLayout>
  );
}
