#!/usr/bin/env bash
set -euo pipefail
# Vegeta smoke against public API (informational). Requires vegeta.
BASE="${STELL_API:-http://127.0.0.1:8080}"
TOKEN="${STELL_API_TOKEN:-dev-token}"
echo "POST ${BASE}/v1/sessions" | vegeta attack -rate=5 -duration=10s \
  -header="Authorization: Bearer ${TOKEN}" \
  -header="Content-Type: application/json" \
  -body <(echo '{"message":"ping"}') | vegeta report
