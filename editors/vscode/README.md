# VS Code extension (MVP)

Settings: `agent.apiBase`, `agent.apiToken`.

Commands:

- **Agent: Send Chat** — `POST /v1/sessions`
- **Agent: Open File** — open a workspace path (Hands still execute tools)
- **Agent: Apply Last Suggestion** — insert last `final_text` at cursor

Requires `cmd/gateway` with `AGENT_API_TOKEN`. Workspace trust: only open files inside the folder.
