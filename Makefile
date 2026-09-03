# gotsx — a TSX dialect that compiles to native Go
APPS := example site shop admin

.PHONY: help gen build test test-fast test-browser check lint fmt tailwind dev-% clean

help:
	@echo "gotsx make targets:"
	@echo "  make tailwind   download the Tailwind standalone binary into .tools/ (no Node)"
	@echo "  make gen        compile every demo app's dialect → gen/ (hostgen + tailwind + compiler)"
	@echo "  make build      gen + go build ./..."
	@echo "  make test       gen + go test ./...   (gen must run first: gen/ is gitignored)"
	@echo "  make test-fast  compiler / runtime / cli unit tests only (no app builds)"
	@echo "  make test-browser  Playwright suites: demos + a fresh scaffold + dev overlay (needs a browser)"
	@echo "  make check      type-check every demo app (what the editor LSP runs)"
	@echo "  make lint       go vet"
	@echo "  make dev-shop   run the shop dev server (dev-example / dev-site / dev-admin likewise)"
	@echo "  make clean      remove generated output"

# gen must precede anything that compiles an app: gen/ is gitignored and absent in a clean checkout
gen:
	@for a in $(APPS); do echo ">> gotsx build $$a"; go run ./cmd/gotsx build $$a || exit 1; done

build: gen
	go build ./...

test: gen
	go test ./...

test-browser: ## real-browser suites (Python Playwright): demos + a fresh scaffold
	python3 tests/browser/run.py

test-fast:
	go test -short ./compiler/... ./runtime/... ./cmd/...

check:
	@for a in $(APPS); do echo ">> gotsx check $$a"; go run ./cmd/gotsx check $$a || exit 1; done

lint:
	go vet ./compiler/... ./runtime/... ./cmd/... ./client/...

fmt:
	gofmt -w compiler runtime cmd client $(APPS)

tailwind:
	go run ./cmd/gotsx tailwind

dev-%:
	go run ./cmd/gotsx dev $* -addr :3000

clean:
	rm -rf $(addsuffix /gen,$(APPS)) $(addsuffix /.gotsx,$(APPS))
