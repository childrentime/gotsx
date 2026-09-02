# 参与 gotsx

## 环境

- Go(版本见 `go.mod`)。**不需要 Node。**
- Tailwind standalone 二进制:`./scripts/get-tailwind.sh`(或设 `GOTSX_TAILWIND` 指向已有的)。

## 常用命令

```bash
make gen        # 编译所有示例应用的方言 → gen/(gen 是 gitignore 的, 必须先跑)
make build      # gen + go build ./...
make test       # gen + go test ./...
make test-fast  # 只跑编译器/运行时单元测试
make dev-shop   # 起 shop 开发服务器
make lint       # go vet
```

> **重要**:`gen/` 目录是 gitignore 的,干净检出里不存在。任何编译应用的命令(`go build ./...` / `go test ./...`)之前必须先 `make gen`,否则找不到 `*/gen` 包。CI 已按此顺序编排。

## 仓库结构

| 目录 | 内容 |
|---|---|
| `compiler/` | 方言编译器:`lexer` / `parser` / `check`(类型) / `gogen`(Go 后端) / `jsgen`(JS 后端) / `compile`(流程) |
| `runtime/` | 生成的 Go 依赖的运行时:节点模型、hydrate 标记、方言内建、HTTP、宿主类型反射生成 |
| `client/` | 浏览器运行时:signals、`el/t/text/cond/each`、走位 hydrate;岛加载器 + morph 跳转 |
| `cmd/gotsx/` | CLI:`gotsx build` / `gotsx dev`(hostgen → tailwind → 编译 → go build → 运行 → 监视重来) |
| `example/` `site/` `shop/` | 示例应用,也是集成测试对象 |

## 测试约定

- **`compiler/codegen_test.go`**:方言片段 → 断言生成的 Go / JS 含预期结构。
- **`compiler/fence_test.go`**:每种围栏违规 → 报错且带 `文件:行:列`。
- **`compiler/apps_test.go`**:编译三个真实应用并 `go build` + `go vet`(集成回归网,`-short` 跳过)。
- **`runtime/*_test.go`**:内建函数正确性、渲染 / hydrate 标记、XSS 转义。

**加语言特性的规矩**:一个特性要同时落地 checker + Go 后端 + JS 后端 + 运行时(如需),并补 `codegen_test.go` 一行;两个后端语义必须一致(数字格式化、字符串按 rune、map 键排序等)。出子集要报错,不要静默编错。

## 设计原则

- **Go 是唯一真相源**:路由、数据、权限、宿主能力在 Go;方言能做的严格等于 Go 暴露的。
- **子集由类型系统定义**:能推出静态类型且落在允许集合里就能编,否则是带位置的编译错误。
- **SSR 是单趟求值**:`useState`=初值、`useEffect`=空、setter=空函数;这是 TSX 能编成 Go 的前提。
