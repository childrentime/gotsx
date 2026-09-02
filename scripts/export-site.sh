#!/usr/bin/env bash
# Static-export the docs site (site/) to dist/ for GitHub Pages — a thin wrapper around `gotsx export`.
# BASE is the project-page subpath (e.g. /gotsx; "" for a user/organization root site); SITE the public origin.
set -euo pipefail
cd "$(dirname "$0")/.."
BASE="${BASE:-/gotsx}"
SITE="${SITE:-https://childrentime.github.io}"
[ -x .tools/tailwindcss ] || go run ./cmd/gotsx tailwind
go run ./cmd/gotsx export site --out dist --base "$BASE" --site "$SITE" --port "${PORT:-4123}"
