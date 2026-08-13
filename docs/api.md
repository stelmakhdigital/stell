# Public API

Token auth (`Authorization: Bearer` or `X-API-Token`). **Not** Hands HMAC.

- OpenAPI: `api/openapi.yaml`
- Proto: `api/proto/agent.proto`
- Events: `pkg/protocol`

```bash
export AGENT_API_TOKEN=dev-token
go run ./cmd/gateway -addr 127.0.0.1:8080
curl -H "Authorization: Bearer $AGENT_API_TOKEN" \
  -d '{"message":"hello"}' http://127.0.0.1:8080/v1/sessions
```

| Method | Path | Action |
|---|---|---|
| POST | `/v1/sessions` | start run |
| GET | `/v1/sessions/{id}` | status |
| GET | `/v1/sessions/{id}/events` | SSE |
| POST | `/v1/sessions/{id}/cancel` | cancel loop |
| POST | `/v1/sessions/{id}/hitl` | `{decision: approve\|deny}` |

Web UI: `web/index.html`. VS Code: `editors/vscode`. JetBrains: `editors/jetbrains`.
