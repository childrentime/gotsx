# Editor support

gotsx ships a Language Server (`gotsx lsp`, LSP over stdio) that reports the **dialect's own rules** as you type:
fence violations (client code importing `host:*`), unsupported syntax, missing props, `==`, type errors, page
constraints — everything `gotsx check` reports, with `file:line:col`, re-run on every keystroke against the
unsaved buffer. It also answers **hover** (types of variables, state and memos, component prop signatures, host
method signatures, builtin docs) and **go-to-definition** — including jumping from `models.search(...)` straight
into the Go method in your `host/` package, and from a host type to its line in `app/.gen/host.d.ts`.

TypeScript's own tooling keeps doing completion, hover and navigation: every app has a `tsconfig.json` and
`gotsx build` writes `app/.gen/gotsx.d.ts` (the `"gotsx"` module + global builtins) and `app/.gen/host.d.ts`
(your Go host modules, reflected). So the two layers together give you a normal TSX editing experience.

Install the CLI once:

```bash
go install github.com/childrentime/gotsx/cmd/gotsx@latest
```

## VS Code

The extension source lives in [`vscode/`](vscode). It is a ~40-line client that launches `gotsx lsp` for
`.tsx` files in a workspace that contains a gotsx app. To run it from source:

```bash
cd editors/vscode
npm install            # only vscode-languageclient
code --extensionDevelopmentPath=$PWD
```

or package it with `npx @vscode/vsce package` and install the `.vsix`.

## Neovim (0.10+)

```lua
vim.api.nvim_create_autocmd("FileType", {
  pattern = "typescriptreact",
  callback = function(ev)
    local root = vim.fs.root(ev.buf, { "go.mod", "gotsx.json" })
    if not root or vim.fn.isdirectory(root .. "/app") == 0 then return end
    vim.lsp.start({ name = "gotsx", cmd = { "gotsx", "lsp" }, root_dir = root })
  end,
})
```

## Helix

```toml
# ~/.config/helix/languages.toml
[language-server.gotsx]
command = "gotsx"
args = ["lsp"]

[[language]]
name = "tsx"
language-servers = ["typescript-language-server", "gotsx"]
```

## Zed

```json
// settings.json
{ "lsp": { "gotsx": { "binary": { "path": "gotsx", "arguments": ["lsp"] } } },
  "languages": { "TSX": { "language_servers": ["gotsx", "..."] } } }
```

## Anything else

Any LSP client works: run `gotsx lsp`, speak JSON-RPC over stdio, open `.tsx` documents. The server only needs
`textDocument/didOpen|didChange|didSave|didClose` (full sync) and publishes `textDocument/publishDiagnostics`.
From the command line, `gotsx check [appdir]` prints the same diagnostics (`--json` for tooling).
