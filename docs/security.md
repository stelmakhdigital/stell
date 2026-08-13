# Security

## HMAC Brain ↔ Hands

`POST /v1/execute` requests are signed:

- `X-Agent-Timestamp` — unix seconds
- `X-Agent-Nonce` — unique id
- `X-Agent-Signature` — hex(HMAC-SHA256(key, `ts + "\n" + nonce + "\n" + sha256(body)`))

Skew window: 30s. Replay of a nonce is rejected. `/healthz` is unsigned.

Key: `hmac_key` in config or `STELL_HMAC_KEY`. Production runtime will not start without a key.

Keep clocks in sync (NTP), or valid requests will fall outside the window.

## Sandbox

Production: `--network=none`, `--user 65534:65534`, `--cap-drop ALL`, `--read-only` + tmpfs `/tmp`, `--pids-limit`, `no-new-privileges`.

Bash without Docker is rejected in Hands. Network ≠ `none` in production — startup is refused.

A non-root user may be unable to write to the workspace if the host uid differs — set `sandbox.user` for your environment.

## Manifests and allowlist

`configs/tools/*.yaml`. In production, a missing manifest → tool blocked in the model loop.

Empty `tool_allowlist` in dev = all tools. A non-empty allowlist hides the rest from the model (`Definitions` / `Execute`).

`blocked_in_model_loop` forbids a call from the main loop (sub-agents may still copy a tool via `Lookup`).

## Secrets

Do not commit keys to git. Redact in logs and DCL audit (`args_redacted` + `args_hash`). The output guardrail masks API keys / PEM.
