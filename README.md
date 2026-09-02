# gotsx

**借 React + TSX 的思想,编译到 Go 原生的全栈框架。** 一份 TSX 源码,两个编译器:服务端编成 Go 函数,客户端编成 signals。没有虚拟 DOM、没有 JS 引擎、没有 Node、没有 npm、没有 esbuild —— **工具链只有 Go**。

> **状态:v0.4 · C 端消费应用(含 i18n)· 仍在打磨。** 这是一个可用的端到端 PoC —— 编译器 + 两个后端 + 运行时 + dev 循环 + 三个真实应用(含一个完整电商)+ 测试套件。它足以证明"TSX 能编译到 Go 原生并撑起真实应用",但语言有已知长尾、未经安全审计、无 LSP。详见 [路线图](#路线图)。

```tsx
// app/pages/index.server.tsx —— 编译成 Go 函数, 永不进浏览器
import type { PageProps } from "gotsx";
import { models } from "host:data";              // Go 实现, 类型由反射生成, 调用零编组
import Layout from "../components/Layout.server";

export default function Home({ query }: PageProps) {
  const list = models.search(query.q ?? "");     // 同步: 并发由 goroutine 提供, 没有 async
  return <Layout title="模型">
    <ul>{list.map((m) => <li>{m.title} · ¥{m.price}</li>)}</ul>
  </Layout>;
}
```

```tsx
// app/islands/Counter.client.tsx —— 同一份源码: 服务端做 SSR, 客户端做交互
import { useState, useEffect } from "gotsx";
export default function Counter({ start }: { start: number }) {
  const [n, setN] = useState(start);
  const double = n * 2;                           // 依赖 n → 自动编译成 memo, 不需要 useMemo
  useEffect(() => { console.log(n); }, []);       // 空依赖 = 挂载跑一次
  return <button onClick={() => setN(n + 1)}>{n} ×2 = {double}{n > 4 && <b> 🔥</b>}</button>;
}
```

## 为什么

**核心洞察:SSR 是一次同步的单趟求值。** 没有重渲染、没有 effect、setter 永不被调用。所以服务端不需要 React 运行时,只需要组件的"渲染切片"—— 而它的语义小到可以直接编译成 Go。于是:一个列表页服务端渲染 **~30µs**(goja 跑 React+MUI 的版本要 ~50ms),客户端运行时 **~6KB**,`go build` 出单二进制,`delve`/`pprof`/`go test` 全能用。

## 快速开始

```bash
git clone <repo> gotsx && cd gotsx
./scripts/get-tailwind.sh          # 下载 Tailwind standalone 二进制到 .tools/(无需 Node)
make dev-shop                      # 起 Temu 风格电商示例 → http://localhost:3000
# 或:
make dev-site                      # 本框架官网(方言组件库 + 可搜索语法参考)
make dev-example                   # 分支集成 demo
```

构建 / 测试:

```bash
make gen      # 编译所有示例的方言 → gen/(gen 是 gitignore 的, 必须先跑)
make build    # gen + go build ./...
make test     # gen + go test ./...
```

> **注意**:`*/gen` 是 gitignore 的,干净检出里不存在。任何编译应用的命令之前必须先 `make gen`。CI 已按此顺序编排。

## 一份源码,两个编译器

```
.tsx ──▶ 前端: 自研 TSX 子集 parser + 类型检查 + 方言围栏 + server/client 边界
          ├─▶ Go 后端:  组件 → Go 函数, JSX → gotsx.El/Text/If/Nodes, hooks → 单趟语义,
          │             host:* → 直接 Go 调用; //line 指令让报错指回 .tsx
          └─▶ JS 后端:  组件 → 函数, useState → signal, 依赖 signal 的 const → memo,
                        JSX → el/t/text/cond/each; 走位 hydrate 不做 diff
Go 侧: 生成的 *_gen.go 与你的 main.go / host 包一起 go build = 单二进制
```

| 目录 | 内容 |
|---|---|
| `compiler/` | `lexer` / `parser` / `check` / `gogen` / `jsgen` / `compile` |
| `runtime/` | 节点模型、hydrate 标记、方言内建、HTTP、请求 Cookie、`Before` 钩子、宿主类型反射生成 |
| `client/` | signals、`el/t/text/cond/each`、走位 hydrate、岛加载器、SPA 跳转、进度条、跨岛 `emit`/`on` |
| `cmd/gotsx/` | `gotsx build` / `gotsx dev` |
| `example/` `site/` `shop/` | 示例应用,也是集成测试对象 |

## 语言:一门借 TSX 语法的静态子集

不是 TypeScript 的实现,而是**借 TSX 语法的静态语言**(AssemblyScript 的同类):类型系统限定在 Go 能表示的集合里,能推出静态类型且落在允许集合里就能编,否则是带 `文件:行:列` 的编译错误。

- **有**:函数组件 / props / 解构+默认值、`string`(rune)/`number`(float64)/`boolean`/数组/`Record`/对象、`map filter find some every forEach includes indexOf join slice concat sort reduce reverse flat at`、字符串方法(含 `padStart/padEnd`)、`Object.keys/values`、`Math.*`、模板字符串、`&& || ?? 三元`、`=== !==`、`if / for-of / try`、`useState/useMemo/useEffect(+[])`、JSX、模块级 `const`、`host:*`(服务端)、`fetch`/DOM/`await`(客户端)。
- **没有(报错,不是静默)**:`class`/`this`/原型、`any` 上取成员、`==`、`while`/`switch`、自定义泛型、`push/splice`、正则、`Date`(服务端走宿主)、给岛传 `children`。

完整语法表(可搜索)在官网 `site` 的 `/docs/language`。

## 特性

- **宿主模块**:`import { models } from "host:data"` 背后是 Go,编译后是直接调用、零编组;类型由 Go 反射生成 `host.d.ts`。`(T, error)` 的 error 变 panic 由请求层 recover(包了 `ErrNotFound` 的回 404)。
- **岛 + SPA 跳转**:页面零 JS,岛按需加载;跳转 = 拉 HTML → idiomorph morph,岛按 DOM 同一性存活,状态不丢;顶部进度条。
- **走位 hydrate**:服务端只在响应式的文本/条件/列表上留标记,客户端按同一编译器给出的结构顺序认领节点,复用现有 DOM,不做 diff。
- **会话 / 中间件**:请求 Cookie 进 `PageProps.Cookies`;`Options.Before` 钩子可种 cookie、做鉴权。
- **Source map**:`go build` 报错和 panic 栈指回 `.tsx` 行号。
- **Tailwind**:`class` 就是字符串,进程内跑 standalone CLI 扫描 `.tsx` 生成 CSS,无需 Node。

## 国际化(v0.4,可选)

开启 `Options.I18n` 即得完整 i18n:`t()`/`tv()`(插值)/`plural()`/`fmtNum`/`fmtCur`/`fmtDate`(服务端与客户端行为一致);**URL 前缀**(`/en/`,SEO 友好)或 **cookie/Accept-Language** 两种语言解析;自动 **hreflang**;loader **自动本地化内链**(站内导航保持同语言);`PageProps.Locale`。`shop` 已接入中/英 + 语言切换器,岛也能翻译。

## C 端能力(v0.3,面向真实商业消费应用)

- **SEO**:每页 title/description/**canonical**/**OpenGraph**/Twitter,商品页 **JSON-LD Product**(价格/库存/评分)+ 首页 WebSite SearchAction,**sitemap.xml** + **robots.txt**。新增 `jsonLd()` 安全内建。
- **真实图片**:服务端生成商品棚拍 SVG(`/img/p/{id}`),`Img` 组件懒加载 + 固定宽高防 CLS + alt;`og:image` 用商品图。
- **性能**:hover/touch **预取**,点击秒开(Core Web Vitals)。
- **可观测性**:客户端错误 + 页面浏览遥测(`sendBeacon` → `/_gotsx/client-log`,`Options.OnClientEvent` 接收)。
- **PWA**:manifest + 图标 + theme-color(可添加到主屏)。

## 生产就绪(v0.2)

- **HTTP 加固**:panic 恢复、安全头、**CSP+nonce**、gzip、**CSRF 同源校验**、内容哈希 immutable 缓存、请求 ID、访问日志、优雅关闭、`/healthz` `/readyz`、自定义 404/500、应用级中间件(鉴权)。
- **单二进制部署**:`go:embed` 客户端与静态资源,`go build` 出一个自包含二进制,`scp` 即跑。
- **客户端韧性**:响应式所有者/清理、按 DOM 范围重建块(嵌套表格刷新不残留)、岛错误边界(单岛失败不白屏)。
- 见 [`SECURITY.md`](SECURITY.md)。默认 `-dev=false` 为生产模式;`gotsx dev` 自动开发模式。

## 示例应用

| 应用 | 是什么 |
|---|---|
| `admin` | **企业后台**:登录 / 受保护路由 / 用户表格(搜索排序分页)/ CRUD + 服务端校验 / 模态框 / toast / 权限。用来把框架往企业级压 |
| `shop` | **Temu 风格全栈电商**:8 分类 / 192 商品 / 搜索排序分页 / 闪购倒计时 / 规格库存 / 购物车(金额服务端算)/ 心愿单 / 结算校验 / 订单 / 会话隔离 / 全接口模拟延迟 + 骨架屏 |
| `site` | 本框架官网:方言写的组件库 + 可搜索语法参考(高亮由 Go 写的 tokenizer 提供) |
| `example` | 分支集成管理 demo |

## 测试

```bash
make test        # 全部(含三应用集成构建)
make test-fast   # 只跑编译器 + 运行时单元测试
```

- `compiler/codegen_test.go` —— 方言片段 → 断言生成的 Go / JS 结构。
- `compiler/fence_test.go` —— 每种围栏违规 → 报错且带位置;合法程序不报错。
- `compiler/apps_test.go` —— 编译三个真实应用并 `go build` + `go vet`。
- `runtime/{builtins,render,security}_test.go` —— 内建正确性、hydrate 标记、XSS 转义(文本/属性/岛 props/全链路)。

## 安全

见 [`SECURITY.md`](SECURITY.md)。要点:方言无"注入原始 HTML"的口子,文本/属性/岛 props 全部经 `html.EscapeString`(有测试);`host:*` 只在服务端,客户端碰不到 Go;写操作在 Go。CSRF / 认证 / 业务校验由使用者负责。**未经独立安全审计。**

## 路线图

到"别人敢用"还缺:

- [ ] **编辑器 LSP**(方言专有规则错误目前只在构建时报)
- [ ] 语言长尾:`while`/`switch`、`push/splice`、自定义泛型、更多内建
- [ ] `each` 的 keyed diff(现在列表整块重建,已能正确增删但会重建 DOM)
- [ ] 流式 SSR、嵌套 layout 约定
- [ ] `Record` 的"键不存在 vs 空串"语义(Go map 无法区分)
- [ ] 增量编译、稳定性契约冻结、独立安全审计
- [ ] 跨平台打磨(Windows)

已做:两个后端、source map、测试套件、Tailwind、会话/中间件、跨岛事件、**生产 HTTP 加固(CSP/CSRF/gzip/缓存/健康检查/优雅关闭)**、**单二进制部署**、**响应式所有者与清理**、**岛错误边界**、**四个真实应用(含企业后台)**。

## 许可

[MIT](LICENSE)。gotsx 借鉴了 React/Solid/Svelte 的思想与 Astro 的岛模型;运行时是自己的。
