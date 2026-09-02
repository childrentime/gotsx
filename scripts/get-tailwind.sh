#!/usr/bin/env bash
# 下载 Tailwind v4 standalone 二进制(无需 Node)到 .tools/tailwindcss
set -euo pipefail
cd "$(dirname "$0")/.."
os=$(uname -s | tr '[:upper:]' '[:lower:]')
arch=$(uname -m)
case "$os-$arch" in
  darwin-arm64)              f=tailwindcss-macos-arm64 ;;
  darwin-x86_64)             f=tailwindcss-macos-x64 ;;
  linux-x86_64)              f=tailwindcss-linux-x64 ;;
  linux-aarch64|linux-arm64) f=tailwindcss-linux-arm64 ;;
  *) echo "不支持的平台: $os-$arch (手动下载并设 GOTSX_TAILWIND)"; exit 1 ;;
esac
mkdir -p .tools
echo "下载 $f …"
curl -fsSL -o .tools/tailwindcss "https://github.com/tailwindlabs/tailwindcss/releases/latest/download/$f"
chmod +x .tools/tailwindcss
./.tools/tailwindcss --help >/dev/null 2>&1 && echo "Tailwind 就绪: .tools/tailwindcss"
