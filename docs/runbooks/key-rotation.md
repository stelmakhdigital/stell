# Key rotation

## Hands HMAC (`STELL_HMAC_KEY` / `hmac_key`)

1. Generate a new key; dual-run is not supported — schedule a short drain.
2. Restart **runtime first** with the new key (old Brain will get 401).
3. Restart Brain/gateway with the same key.
4. Confirm `/v1/execute` 200 with a signed probe; old signatures must 401.

## Public API token (`STELL_API_TOKEN`)

1. Issue a new token; update clients (VS Code `stell.apiToken`, Web UI, JetBrains).
2. Restart gateway. Old Bearer tokens get 401.
3. HMAC and API token are **different** — do not reuse one for both.
