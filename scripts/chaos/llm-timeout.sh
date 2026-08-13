#!/usr/bin/env bash
set -euo pipefail
# Chaos: LLM timeout — Brain must return error, not hang.
# Uses a blackhole proxy if STELL_LLM_TIMEOUT_URL is set.
curl -sS -m 3 "${STELL_LLM_TIMEOUT_URL:-http://127.0.0.1:9}/v1/chat/completions" || true
echo "expected: llm client deadline → agent_error event"
