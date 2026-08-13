# SLO / SLA (starting point — iterate with real traffic)

| Indicator | Target | Notes |
|---|---|---|
| Gateway availability | 99.5% monthly | `/healthz` |
| Session create p95 | < 5s (excl. LLM) | k6 `scripts/load/k6.js` |
| Tool execute p95 (Hands) | < 30s | sandbox timeout |
| Error rate (5xx gateway) | < 1% | |
| Eval golden aggregate | ≥ 0.85 | nightly informational; release gate |
| RPO (sessions/audit) | last backup | `scripts/dr/backup.sh` |
| RTO (session restore) | < 15 min | `scripts/dr/restore.sh` drill |

Load tests are informational in CI; block a release if p95 or error rate exceeds the table on a dedicated env.
