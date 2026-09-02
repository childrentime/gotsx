# 更新日志

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
