#!/usr/bin/env bash
# 把官网(site, SSR)静态导出到 dist/, 供 GitHub Pages 托管。
# BASE 是项目页子路径(如 /gotsx); 用户/组织根页面用 BASE=""。
set -euo pipefail
cd "$(dirname "$0")/.."
BASE="${BASE:-/gotsx}"
PORT="${PORT:-4123}"
OUT="dist"
ROUTES=("/" "/docs" "/docs/language" "/docs/components" "/docs/architecture" \
        "/zh" "/zh/docs" "/zh/docs/language" "/zh/docs/components" "/zh/docs/architecture")

echo ">> build site"
[ -x .tools/tailwindcss ] || ./scripts/get-tailwind.sh
go run ./cmd/gotsx build site >/dev/null
go build -o site/.gotsx/export ./site

echo ">> run server :$PORT"
( cd site && ./.gotsx/export -addr ":$PORT" >/dev/null 2>&1 & echo $! > /tmp/export.pid )
trap 'kill $(cat /tmp/export.pid) 2>/dev/null || true' EXIT
for i in $(seq 1 60); do curl -sf -o /dev/null "http://localhost:$PORT/" && break; sleep 0.5; done

rm -rf "$OUT"; mkdir -p "$OUT/_gotsx" "$OUT/public"
cp site/gen/client/* "$OUT/_gotsx/"
cp -r site/public/* "$OUT/public/" 2>/dev/null || true
touch "$OUT/.nojekyll"   # 允许 _gotsx 目录(Jekyll 默认忽略下划线开头)

echo ">> crawl + rewrite (BASE=$BASE)"
for r in "${ROUTES[@]}"; do
  if [ "$r" = "/" ]; then dir="$OUT"; else dir="$OUT$r"; fi
  mkdir -p "$dir"
  curl -sf "http://localhost:$PORT$r" \
    | sed -E "s#(href|src)=\"/#\1=\"$BASE/#g" \
    | sed -E "s#\"/_gotsx/#\"$BASE/_gotsx/#g" \
    | sed -E "s#window.__GOTSX=\{#window.__GOTSX={\"base\":\"$BASE\",#" \
    > "$dir/index.html"
done
# 404 页(Pages 用 404.html)
cp "$OUT/index.html" "$OUT/404.html"
echo ">> done: $(find "$OUT" -name '*.html' | wc -l | tr -d ' ') 页, $(du -sh "$OUT" | cut -f1)"
