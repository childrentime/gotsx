# 更新日志

## 0.6.0 — 设计系统、流式 SSR、文件约定、编辑器跳转: 路线图清零

### 设计系统(demo 全部重做)
- 新增 `design/`: 对标 shadcn/ui neutral(zinc)主题的 **gotsx UI** —— 语义 token(`background/foreground/card/muted/border/primary/…`, 亮暗两套)、`@theme inline` 映射、组件类(`.btn-*` `.input` `.card` `.badge-*` `.table` `.nav-link` `.skeleton` …)。`gotsx.css` 是 Tailwind v4 层, `plain.css` 是同一套系统的手写 CSS; `design/README.md` 是规范(只用 token、一个中性强调色、发丝边框、8px 网格、SVG 图标不用 emoji 做 UI)。
- `site` / `shop` / `admin` / `example` 全部按规范重做: 去掉品牌色与渐变, 表格 / 表单 / 卡片 / 徽章 / 空状态统一, UI 图标换成内联 SVG, 亮暗模式由 token 驱动。
- `gotsx new` 生成的应用直接带设计系统(`--tailwind` 用 `app/gotsx.css`, 默认用 `public/app.css`)。

### 流式 SSR(Suspense)
- `import { Suspense } from "gotsx"`; `<Suspense fallback={…}>` 边界: fallback 随外壳先发, children 编成 thunk, 在外壳 flush 之后于**各自的 goroutine 里并发**求值, 完成一个就以 `<template data-gotsx-fill>` + 一行脚本追加一个(乱序到达也正确, 支持嵌套边界)。边界内的错误不炸整页(dev 显示错误, 生产留空并记日志)。SPA 跳转后的填充由 loader 应用; 岛在填充内容里照常挂载。仅服务端组件可用。
- `example` 首页的统计面板即演示(两个慢查询 600ms / 300ms 并发填充)。

### 文件约定: 嵌套布局 / 404 / 错误页
- `pages/**/_layout.server.tsx`(props `LayoutProps` = PageProps + children)自动包住其目录下的页面, 多层嵌套外层在外; `pages/_404.server.tsx` → `gen.NotFound`, `pages/_error.server.tsx`(`ErrorProps` = PageProps + message)→ `gen.ErrorPage`, 传给 `gotsx.Options` 即可; 其它 `_` 开头的文件不当路由。
- 页面组件的 Go 名按 app 相对路径加前缀(`pages_docs__layout_Docs`), 不同目录同名页面不再冲突。

### 编辑器: hover / 跳转定义
- `gotsx lsp` 新增 `textDocument/hover` 与 `textDocument/definition`: 变量 / 参数 / state / memo 给类型, 组件给 props 签名, 宿主方法给签名并**跳到 Go 源码的 file:line**(hostgen 反射得到), 宿主类型 / 字段跳到 `host.d.ts`, interface 跳到声明, 内建(useState / Suspense / redirect …)给文档; import 行上的名字也可用。`compiler.Load` / `HoverAt` / `DefinitionAt` 供工具调用。

### Record 缺席语义(路线图项)
- 两端一致: 读缺席的键得到值类型的零值(客户端编成 `m.k ?? ""`), `x === undefined` 对原始类型比较零值; **`Object.hasOwn(m, k)`** 判断键是否存在; **`delete m[k]` / `delete m.k`** 删除键。

### 更多内建
- **正则字面量** `/pattern/flags`(g i m s u), RE2 子集, 编译期校验(lookaround / 反向引用会报错); `re.test(s)`, `s.match(re)`(没匹配 → `[]`), `s.replace(re, "$1…")`, `s.replaceAll(re-g, …)`, `s.split(re)`, `s.search(re)`; 服务端缓存编译, 替换模板 `$& $1 $$` 两端一致。
- **`Date.now()` / `Date.parse(iso)` / `isoDate(ms)`** 两端一致(RFC3339 / `YYYY-MM-DD`; UTC 毫秒精度)。

### 增量 dev 循环
- `gotsx dev` 重建时 **`host/` 未变则跳过 hostgen**(`go run`, 最慢的一步), hostgen 与 Tailwind 并行; 输出各步耗时(`hostgen 786ms ∥ compile 13ms` → 只改 TSX 时 `compile 13ms`)。

### 稳定性契约与安全
- 新增 `STABILITY.md`: 三个层级(Stable / Experimental / Internal)、语义化版本、废弃策略、1.0 的含义。
- `SECURITY.md` 补充威胁模型与本轮自查清单(Suspense 填充脚本走 CSP nonce、LSP 路径处理、dev 端点仅 dev 模式、redirect 目标由应用校验), 附 `govulncheck` 结果; 独立审计仍是外部事项。

## 0.5.0 — 可安装、可编辑器集成、语言补齐长尾

目标从"证明可行"转向"能被别人依赖": 一条 `go install` 装好工具链, `gotsx new` 起一个独立模块的应用, 编辑器里实时看到方言的错误, 语言把最常见的长尾补齐, 列表有 keyed diff, 开发时浏览器自动刷新。

### 安装与脚手架
- **模块路径改为 `github.com/childrentime/gotsx`**(之前是 `gotsx`, 无法被 `go get`)。生成代码的运行时导入路径由 CLI 自己的模块路径推导, fork 也能用。
- **`gotsx new <dir>`**: 脚手架出独立 Go 模块 —— 宿主模块(Go)、页面、布局、keyed 列表岛、HTTP action、`tsconfig.json`、`.gitignore`、README; `--tailwind` 可选, `--replace <checkout>` 指向本地框架检出(从仓库 `go run` 时自动检测)。
- **`gotsx.json` 变为可选**: 应用的导入路径从最近的 `go.mod` 推断, `host/` 目录自动识别; 仍可用它覆盖。
- **`gotsx tailwind`**: 纯 Go 下载 Tailwind standalone(macOS / Linux / Windows), 替代 shell 脚本; `dev` 循环的路径处理适配 Windows。
- CLI 子命令: `new` / `build` / `dev` / `check` / `lsp` / `tailwind` / `version`; 输出改为英文。

### 编辑器
- **`gotsx check [dir] [--json]`**: 只检查不生成, 诊断 `file:line:col: message`, 有错 exit 1(CI 友好)。
- **`gotsx lsp`**: LSP over stdio —— 每次编辑对未保存的缓冲区做 解析 + 类型检查 + 两个后端 的内存编译, 发布诊断。`editors/` 提供 VS Code 扩展源码、Neovim / Helix / Zed 配置。
- **`app/.gen/gotsx.d.ts`**: 构建时生成 `"gotsx"` 模块与全局内建的类型声明, 加上各应用的 `tsconfig.json`, TypeScript 自带的补全/跳转不再报错。
- 编译器新增 `compiler.Analyze(appDir, overlay)`: 结构化诊断(文件 / 行 / 列 / 消息), 模块按文件名确定性遍历。

### 语言
- **控制流**: `while`、经典 `for (let i = 0; i < n; i++)`、`break` / `continue`、`switch` / `case` / `default`(空 case 合并, JS 贯穿语义翻译成 Go `fallthrough`, 支持 `switch (true)`)。
- **运算**: `++` / `--`(语句与表达式)、`%=`; 数组下标 / Record 键 / 字段的赋值(含宿主 int 字段的自动转换)。
- **原地修改的数组方法**: `push` / `pop` / `shift` / `unshift` / `splice`(Go 侧传地址, 语义与 JS 对齐; 直接改 `useState` 的数组是编译错误, 提示用不可变写法)。
- 更多内建: 数组 `findIndex` / `lastIndexOf`, 字符串 `trimStart` / `trimEnd` / `lastIndexOf` / `localeCompare`(两端都按码点比较)/ `at` / `toString`, 数字 `toString`, `slice()` 无参拷贝。
- **`interface A extends B`**(可多重, 子接口同名字段覆盖); 函数体内声明的 interface 也会生成 Go struct。
- **可选对象的语义补正**: `find()` 没找到 / 可选字段缺席 时, `if (x)`、`!x`、`x ?? y`、`x === undefined` 在 Go 侧按零值判断(之前结构体永远为真); 客户端的 `??` 对原始类型与 Go 同语义(与 `||` 等价)。
- `map` 回调继承期望的元素类型: `useState<Item[]>(xs.map((x) => ({ ... })))` 能对上。
- **页面级控制流**: `redirect(url, status?)` / `notFound()`(服务端页面; 中断渲染, 请求层回 3xx / 404 页)。
- **catch-all 路由**: `pages/docs/[...slug].server.tsx` → `/docs/{...slug}`, `params.slug = "a/b/c"`; 路由按具体程度排序(静态 > 参数 > catch-all)。

### 客户端
- **keyed 列表**: `xs.map((x) => <li key={x.id}>…</li>)` 编译成带 key 函数的 `each`, 变化时按 key 复用 / 移动 / 销毁 DOM(最少 `insertBefore`, 每项独立所有者), 输入框状态、焦点、行内 effect 在增删排序后都保留; 无 key 的列表行为不变。
- **dev 自动刷新**: 服务端 dev 模式提供 `/_gotsx/dev` SSE(带本进程 bootID), loader 检测到重启后整页刷新; 编译失败时旧进程继续跑, 不刷新。
- 页面没有 `<head>` 时, 运行时脚本注入到 `</body>` 前(之前静默不注入)。

### 示例与测试
- `example` 新增 `/kitchen`(语言"厨房水槽": 所有新语法在一页, 也是集成测试)、`/docs/[...slug]`(catch-all)、`Todos` keyed 列表岛。
- 新增 `compiler/lang_test.go`(两端 codegen + 围栏 + Analyze + keyed 标记一致性)、`runtime/lang_test.go`(数组方法 / 零值语义 / 路由 / redirect / dev SSE)、`cmd/gotsx/cli_test.go`(`new → build → go build → check` 端到端)。
- 浏览器实测(Playwright): keyed 列表重排后 DOM 同一性与输入状态保留、SPA 跳转岛状态存活、redirect / 404 / catch-all、编辑源码后自动刷新。
- CI 加入 gofmt 检查与 `make check`。

## 0.4.0 — 可选的国际化(i18n)完整支持

框架级 i18n,可选开启(`Options.I18n`),在电商 `shop` 上验证(中/英)。

### 能力
- **翻译**:`t(locale, key)`;**插值** `tv(locale, key, {name})`;**复数** `plural(locale, key, n)`(CLDR-lite:英语 one/other,中文单形式,文案用 `one|other` 分隔,`{n}` 占位)。
- **本地化格式化**:`fmtNum`(千分位)、`fmtCur`(按语言货币符号,不换算汇率)、`fmtDate`(zh `年月日` / en `Mon D, Y`)。
- **服务端 + 客户端一致**:同名内建,服务端 → `gotsx.Tr/...`,客户端 → `G.tr/...`,行为逐一对应;活动语言的目录由服务端注入到 `window.__GOTSX.i18n`,岛也能翻译。
- **语言解析**:两种模式——**URL 前缀**(`/en/...`,SEO 友好,默认语言不加前缀)或 **cookie `lang` + Accept-Language**。解析后 `PageProps.Locale`,路由用去前缀的路径匹配。
- **SEO**:自动注入 **hreflang** 备用链接(每语言一条 + x-default);`canonical` / `og:url` 走 `lpath`;`<html lang>` 由应用设置。
- **本地化导航**:前缀模式下,客户端 loader **自动给内链补当前语言前缀**(站内点击保持同语言),无需给每个链接手写前缀;`lpath(locale, path)` 供 SSR 链接用。
- **语言切换**:示例 `LocaleSwitch` 岛整页导航到目标语言前缀。

### 编译器
- 新增内建 `t` / `tv` / `plural` / `fmtNum` / `fmtCur` / `fmtDate` / `lpath`(显式 locale,类型检查,与 gotsx 无 Context 的设计一致)。
- `PageProps.locale` 字段。

### 测试
- 运行时 i18n 单测(查表/回退/插值/复数/格式化/语言解析/前缀);编译器 i18n 内建映射测试;shop 浏览器验证语言切换 + 岛翻译 + 前缀粘性导航。

## 0.3.0 — C 端能力(面向真实商业消费应用)

在电商示例 `shop` 上补齐并验证了消费端(C 端)应用的定义性能力。

### SEO(可被搜索引擎发现 —— C 端的生命线)
- 每页 `<title>` / `description` / **canonical**(绝对 URL)/ 完整 **OpenGraph** + Twitter Card。
- **JSON-LD 结构化数据**:商品页 `Product`(offers 价格/币种/库存 + aggregateRating),首页 `WebSite`(SearchAction)。新增安全内建 `jsonLd(JSON.stringify(...))`(Go `json.Marshal` 已转义 `<>&`,再防 `</script>` 逃逸)。
- **sitemap.xml**(首页 + 分类 + 全部商品)、**robots.txt**(含 sitemap 链接 + 禁私有路径)。
- 修复:`JSON.stringify` 对方言对象在服务端也用原始键(如 `@context`);任意 JSON 键(`@` / 数字开头 / 连字符)映射成合法 Go 字段。

### 真实图片(商业视觉 + 性能)
- 服务端生成**商品棚拍 SVG**(`/img/p/{id}`:径向渐变背景 + 落地投影 + 大图,immutable 缓存)。
- `Img` 组件:**懒加载**(`loading=lazy`)、异步解码、**固定宽高防 CLS**、**alt 文本**(SEO / 无障碍)。卡片与图廊全部换成真实 `<img>`;`og:image` 用商品图。

### 性能(Core Web Vitals)
- **hover / touch 预取**:悬停 60ms 预拉目标页 HTML,点击命中缓存**秒开**(0 额外请求),有上限防内存膨胀。

### 可观测性
- **客户端遥测**:JS 错误 / 未处理 rejection / 页面浏览 通过 `sendBeacon` 上报到 `/_gotsx/client-log`(仅同源,有上限)。服务端 `Options.OnClientEvent` 接收。

### PWA
- `manifest.webmanifest` + 品牌图标 + `theme-color` + apple-touch-icon(可"添加到主屏")。

## 0.2.0 — 企业加固

真实用框架搭了一个后台管理系统(`admin`: 认证 / 受保护路由 / 用户表格 CRUD / 服务端校验 / 模态框 / toast),反复驱动,逐一修掉暴露的问题。

### 生产 HTTP 层(`runtime/server.go`,均有测试)
- 中间件链:请求 ID、访问日志、**panic 恢复**(500 页,prod 不泄露堆栈)、安全响应头、gzip 压缩。
- **CSP + 每响应 nonce**(内联脚本带 nonce);`X-Frame-Options` / `X-Content-Type-Options` / `Referrer-Policy`。
- **CSRF**:对写操作(POST/PUT/PATCH/DELETE)做同源校验。
- 静态资源**内容哈希 + immutable 长缓存**;页面 `no-store`。
- 自定义 404 / 500 页(用方言写);`/healthz`、`/readyz`;优雅关闭;应用级中间件(鉴权重定向)。

### 单二进制部署
- 编译器生成 `gen/assets_gen.go`(`go:embed` 客户端资源),应用内嵌 `public/`;`go build` 出一个自包含二进制,`scp` 到任意目录即可运行(已验证从 `/tmp` 独立启动全功能)。

### 客户端运行时
- **响应式所有者/清理**:嵌套 effect 随其所属块一起销毁,不再更新已卸载的 DOM。
- **块按 DOM 范围重建**:`each`/`cond` 重建时删除 start/end 之间的一切,修掉深层嵌套(表格行)刷新后**翻倍/残留**的 bug。
- **岛错误边界**:单个岛失败不白屏、不影响其它岛;hydrate 不匹配回退客户端渲染。
- `useEffect(fn, [])` = 挂载跑一次(onMount)。

### 编译器 bug(真实使用中发现,均补回归测试)
- 任意 JSON 键(`_`、数字开头、连字符)映射成合法 Go 字段名。
- **嵌套响应式条件**内层也要保持响应式;**组件根返回条件**(模态框 `open ? A : B`)要响应式。
- 关键不变量:客户端与服务端的 hydrate 标记必须逐一对应(`TestMarkerParity` 护栏)。

### 语言
- 新增 `sort/reduce/reverse/flat/at`、`Object.keys/values`、`padStart/padEnd`、`Math.pow/sign/trunc`。

### 测试
- 服务端中间件测试(安全头 / CSP nonce / gzip / CSRF / 404 / panic 恢复 / 健康检查)。
- 三→四个真实应用集成构建回归(`example` / `site` / `shop` / `admin`)。
- 标记一致性护栏 + 怪异 JSON 键 + 嵌套条件回归。

## 0.1.0 — 首个实验性发布

第一个可用的端到端版本。**实验性,勿用于生产。**

### 编译器
- 自研 TSX 子集 parser + 类型检查器(不依赖 typescript-go)。
- 两个后端:Go(SSR,JSX → 字符串直写,hooks → 单趟求值,`host:*` → 直接 Go 调用)、
  JS(signals + 精确 DOM 绑定,`useState`→signal、依赖 signal 的 const 自动 memo)。
- **Source map**:生成的 Go 带 `//line` 指令,`go build` 报错和运行时 panic 指回 `.tsx` 行号。
- `useEffect(fn, [])` = 挂载跑一次(onMount)语义,避免同步读 signal 造成的追踪循环。

### 语言(子集)
- 组件、props、解构 + 默认值、`map/filter/find/some/every/forEach/includes/indexOf/join/slice/concat`、
  `sort/reduce/reverse/flat/at`、字符串方法(含 `padStart/padEnd`)、`Object.keys/values`、
  `Math.*`、模板字符串、`&&`/`||`/`??`/三元、`===`/`!==`、`if`/`for-of`/`try`、
  `useState`/`useMemo`/`useEffect`、JSX(元素/属性/条件/列表/组件/fragment)、模块级 `const`。
- 明确不支持的语法会报带位置的编译错误(而非静默行为差异)。

### 运行时
- Go:节点模型、hydrate 走位标记(`<!--$-->` / `<!--[-->`)、方言内建、HTTP 服务、
  请求 Cookie → `PageProps.Cookies`、`Before` 中间件、宿主类型反射生成。
- 客户端(~6KB):signals、`el/t/text/cond/each`、走位 hydrate、岛加载器、
  SPA 跳转(fetch HTML → idiomorph)、顶部进度条、跨岛事件 `emit`/`on`。

### 工具链
- `gotsx build` / `gotsx dev`,进程内 Tailwind(standalone 二进制,无 Node)。
- 无 JS 引擎、无 Node、无 npm、无 esbuild —— 工具链只有 Go。

### 示例应用
- `example` —— 分支集成 demo;`site` —— 本框架官网(方言组件库 + 语法参考);
  `shop` —— Temu 风格全栈电商(会话 / 购物车 / 订单 / 模拟延迟 / 骨架屏)。

### 测试
- 编译器 codegen / fence 单元测试、三应用集成构建、运行时内建 / 渲染 / XSS 安全测试。

### 已知限制
- 语言长尾未覆盖(`while`/`switch`/`enum`、自定义泛型、`push/splice`、正则、`Date`(服务端走宿主))。
- `Record<string,string>` 在 Go 里无法区分"键不存在"和"空字符串"(见 `docs`)。
- 无 LSP(方言专有规则错误只在构建时报);无增量编译;无流式 SSR;`each` 无 keyed diff(列表变化整块重建)。
- 未经独立安全审计。
