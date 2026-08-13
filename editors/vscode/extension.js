const vscode = require("vscode");

async function api(path, method, body) {
  const cfg = vscode.workspace.getConfiguration("agent");
  const base = cfg.get("apiBase");
  const token = cfg.get("apiToken");
  const res = await fetch(base + path, {
    method,
    headers: { "Content-Type": "application/json", Authorization: "Bearer " + token },
    body: body ? JSON.stringify(body) : undefined,
  });
  if (!res.ok) throw new Error(path + " " + res.status);
  const ct = res.headers.get("content-type") || "";
  if (ct.includes("json")) return res.json();
  return res.text();
}

function activate(context) {
  let lastText = "";
  context.subscriptions.push(
    vscode.commands.registerCommand("agent.chat", async () => {
      const msg = await vscode.window.showInputBox({ prompt: "Agent task" });
      if (!msg) return;
      const created = await api("/v1/sessions", "POST", { message: msg });
      vscode.window.showInformationMessage("session " + created.session_id);
      const st = await api("/v1/sessions/" + created.session_id, "GET");
      lastText = st.final_text || "";
      const doc = await vscode.workspace.openTextDocument({ content: lastText || "(running — reopen status)", language: "markdown" });
      await vscode.window.showTextDocument(doc);
    }),
    vscode.commands.registerCommand("agent.openFile", async () => {
      const rel = await vscode.window.showInputBox({ prompt: "Workspace-relative path" });
      if (!rel || !vscode.workspace.workspaceFolders) return;
      const uri = vscode.Uri.joinPath(vscode.workspace.workspaceFolders[0].uri, rel);
      await vscode.window.showTextDocument(uri);
    }),
    vscode.commands.registerCommand("agent.applyLast", async () => {
      const editor = vscode.window.activeTextEditor;
      if (!editor || !lastText) {
        vscode.window.showWarningMessage("No last agent text");
        return;
      }
      await editor.edit((b) => b.insert(editor.selection.active, lastText));
    })
  );
}

function deactivate() {}
module.exports = { activate, deactivate };
