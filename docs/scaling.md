# Scaling

Brain replicas are **stateless**. Session JSON lives on a shared volume (`agent.session_store`, default unset = memory).

Hands replicas sit behind nginx / k8s Service. Brain uses `runtime_urls` + HMAC. Unhealthy `/healthz` replicas are skipped (`internal/runtimeclient.Failover`). Optional sticky by workspace hash.

```text
Brain pods  → FileStore /data/sessions
            → HMAC → LB → runtime pods
```

## Compose (dev)

Fill images or run binaries on the host:

```bash
STELL_HMAC_KEY=dev-hmac go run ./cmd/runtime -addr 127.0.0.1:8081
STELL_API_TOKEN=dev-token STELL_HMAC_KEY=dev-hmac go run ./cmd/gateway -runtime http -runtime-url http://127.0.0.1:8081
```

Two runtimes: `-runtime-url http://127.0.0.1:8081,http://127.0.0.1:8082`.

## k8s

`deploy/k8s/stell.yaml` — 2 runtime + 2 brain, shared PVC for sessions, secrets `hmac` / `api`.

Do not disable HMAC in production. Workspace data must be a volume both runtime replicas can see, or use sticky sessions.
