# 安全说明

gotsx 是实验性项目(v0.1),尚未经过独立安全审计。请勿用于处理不可信输入的生产系统,除非你自己完成了审计。

## 已建立的防线(有测试覆盖)

- **XSS / HTML 注入**:方言里没有 `dangerouslySetInnerHTML`,也没有任何"注入原始 HTML"的口子。
  - 文本节点(`{expr}`)经 `gotsx.Text` / `gotsx.Dyn` → `html.EscapeString`。
  - 元素属性经 `gotsx.A` → `html.EscapeString`,防止属性逃逸。
  - 岛的 props 序列化成 JSON 后进 HTML 属性,同样经 `html.EscapeString`。
  - 见 `runtime/security_test.go`:文本、属性、岛 props、全链路 XSS 载荷都断言被转义。
- **服务端 / 客户端边界**:`host:*`(Go 能力)只能在服务端组件里 `import`;客户端只能 `import type`。
  客户端代码碰不到 Go、数据库、密钥。违规是编译错误(见 `compiler/fence_test.go`)。
- **写操作在 Go**:示例应用(shop)把加购/下单/改库存全部放在 Go 的 `http.HandlerFunc`,
  方言只读不写。库存、价格、订单一致性由 Go 保证,客户端改不了。

## 已知需要使用者负责的部分

- **CSRF**:框架不内置 CSRF 防护。写操作(actions)应自行加同源校验 / token。
- **认证 / 授权**:框架不提供。会话、权限判定在宿主模块和 action 里自己实现(shop 示例用 sid cookie 演示会话)。
- **宿主模块的输入校验**:宿主方法的参数来自方言,类型受编译器约束,但业务级校验(长度、范围、注入)由宿主实现负责。
- **动作(action)反序列化**:action 的 JSON body 由使用者的 handler 解析,自行校验。

## 报告漏洞

这是 PoC 仓库,尚无正式披露流程。发现问题请开 issue(去掉可利用的细节)或私下联系维护者。
