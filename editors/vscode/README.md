# gotsx for VS Code

Dialect diagnostics for [gotsx](https://github.com/childrentime/gotsx) apps, powered by `gotsx lsp`.

1. `go install github.com/childrentime/gotsx/cmd/gotsx@latest` (or set `gotsx.path`).
2. Open a folder containing a gotsx app (`app/pages/*.server.tsx`).
3. Errors the Go compiler would report at build time now appear as you type.

Build from source: `npm install && npx @vscode/vsce package`.
