const path = require("path");
const fs = require("fs");
const vscode = require("vscode");
const lc = require("vscode-languageclient");

let client = null;

function findServerBinary() {
  const extDir = path.dirname(__dirname);
  const resolved = fs.realpathSync(extDir);
  const projectRoot = path.resolve(resolved, "..", "..");
  const candidates = [
    path.join(projectRoot, "bin", "gsl-lsp"),
    "/tmp/gsl-lsp",
    path.join(projectRoot, "gsl-lsp"),
  ];
  for (const p of candidates) {
    if (fs.existsSync(p)) {
      console.log("GSL LSP: using binary at", p);
      return p;
    }
  }
  return candidates[0];
}

function createClient() {
  const serverOpts = {
    command: findServerBinary(),
    args: [],
    options: { stdio: "pipe" },
  };

  const clientOpts = {
    documentSelector: [
      { scheme: "file", language: "gsl" },
      { scheme: "file", language: "gql" },
    ],
    synchronize: {
      configurationSection: "gsl",
    },
    diagnosticCollectionName: "gsl",
  };

  const c = new lc.LanguageClient("gsl-lsp", "GSL Language Server", serverOpts, clientOpts);
  c.onDidChangeState((e) => {
    console.log("GSL LSP: state changed to", e.newState);
    if (e.newState === lc.State.Starting) {
      vscode.window.setStatusBarMessage("Starting GSL language server...");
    } else if (e.newState === lc.State.Running) {
      vscode.window.setStatusBarMessage("GSL language server ready", 3000);
    } else if (e.newState === lc.State.Stopped) {
      vscode.window.showWarningMessage("GSL language server stopped");
    }
  });
  return c;
}

function activate(context) {
  console.log("GSL LSP: activating");

  client = createClient();
  const disposable = client.start();
  context.subscriptions.push(disposable);

  context.subscriptions.push(
    vscode.commands.registerCommand("gsl.restartLSP", async () => {
      console.log("GSL LSP: restarting");
      if (client) {
        await client.stop();
      }
      client = createClient();
      client.start();
      vscode.window.showInformationMessage("GSL language server restarted");
    })
  );
}

function deactivate() {
  if (client) {
    return client.stop();
  }
}

module.exports = { activate, deactivate };
