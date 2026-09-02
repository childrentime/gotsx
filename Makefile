# gotsx —— TSX 方言 → Go 原生的全栈框架
APPS := example site shop

.PHONY: help gen build test test-fast lint dev-% tailwind clean fmt

help:
	@echo "gotsx make 目标:"
	@echo "  make tailwind   下载 Tailwind standalone 二进制到 .tools/"
	@echo "  make gen        编译所有示例应用的方言 → gen/ (含 hostgen + tailwind)"
	@echo "  make build      gen + go build ./..."
	@echo "  make test       gen + go test ./...   (gen 必须先于 test: gen/ 是 gitignore 的)"
	@echo "  make test-fast  只跑编译器/运行时单元测试(不构建应用)"
	@echo "  make lint       go vet"
	@echo "  make dev-shop   起 shop 开发服务器(dev-example / dev-site 同理)"
	@echo "  make clean      删除生成产物"

# gen 必须先于任何编译应用的命令: gen/ 是 gitignore 的, 干净检出里不存在
gen:
	@for a in $(APPS); do echo ">> gotsx build $$a"; go run ./cmd/gotsx build $$a || exit 1; done

build: gen
	go build ./...

test: gen
	go test ./...

test-fast:
	go test ./compiler/... ./runtime/...

lint:
	go vet ./compiler/... ./runtime/... ./cmd/... ./client/...

fmt:
	gofmt -w compiler runtime cmd client $(APPS)

dev-%:
	go run ./cmd/gotsx dev $* -addr :3000

clean:
	rm -rf $(addsuffix /gen,$(APPS)) $(addsuffix /.gotsx,$(APPS))
