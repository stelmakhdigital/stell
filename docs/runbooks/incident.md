# Incident response

1. Check `/healthz` on gateway and runtime; Prometheus `:9090/metrics`.
2. If Hands down: Failover should skip unhealthy replicas. If all down, users see tool errors — not silent hangs.
3. If LLM down: `agent_error` + cancel in-flight sessions.
4. Preserve `eval/results/audit.jsonl` and session store before restart.
5. After fix: run `go test ./...` and a golden subset (`go run ./cmd/eval -fixed-answer ...`).
6. Rotate keys if HMAC/API token may have leaked (`docs/runbooks/key-rotation.md`).
