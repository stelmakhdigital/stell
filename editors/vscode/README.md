# VS Code extension (MVP)

Settings: `stell.apiBase`, `stell.apiToken`.

Commands:

- **Stell: Send Chat** — `POST /v1/sessions`
- **Stell: Open File** — open a workspace path (Hands still execute tools)
- **Stell: Apply Last Suggestion** — insert last `final_text` at cursor

Requires `cmd/gateway` with `STELL_API_TOKEN`. Workspace trust: only open files inside the folder.
