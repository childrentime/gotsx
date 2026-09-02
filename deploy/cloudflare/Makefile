# Build the shop demo as a Cloudflare Worker (Go → Wasm). Requires Go and Node (for wrangler).
.PHONY: build dev deploy clean

build:
	go run ../../cmd/gotsx build ../../shop
	mkdir -p public build && cp ../../shop/public/* public/
	go run github.com/syumai/workers/cmd/workers-assets-gen -mode=go
	GOOS=js GOARCH=wasm go build -trimpath -ldflags="-s -w" -o ./build/app.wasm .
	@ls -la build/app.wasm | awk '{printf "app.wasm: %.1f MB (", $$5/1e6}'; gzip -c build/app.wasm | wc -c | awk '{printf "%.2f MB gzipped)\n", $$1/1e6}'

dev: build
	npx wrangler dev

deploy: build
	npx wrangler deploy

clean:
	rm -rf build public
