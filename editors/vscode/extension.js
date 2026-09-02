// gotsx VS Code extension: start `gotsx lsp` for TSX documents. TypeScript's built-in service keeps
// handling completion/navigation; this adds the dialect's own diagnostics (fence, subset, types).
const vscode = require("vscode");
const { LanguageClient, TransportKind } = require("vscode-languageclient/node");

let client;

function activate(context) {
  const bin = vscode.workspace.getConfiguration("gotsx").get("path") || "gotsx";
  client = new LanguageClient(
    "gotsx",
    "gotsx",
    { command: bin, args: ["lsp"], transport: TransportKind.stdio },
    {
      documentSelector: [{ scheme: "file", language: "typescriptreact" }],
      synchronize: { fileEvents: vscode.workspace.createFileSystemWatcher("**/app/.gen/host.json") },
    }
  );
  client.start().catch((e) => {
    vscode.window.showErrorMessage(`gotsx: could not start "${bin} lsp" (${e.message}). Install it with: go install github.com/childrentime/gotsx/cmd/gotsx@latest`);
  });
  context.subscriptions.push({ dispose: () => client && client.stop() });
}

function deactivate() {
  return client ? client.stop() : undefined;
}

module.exports = { activate, deactivate };
