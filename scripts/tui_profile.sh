#!/usr/bin/env bash
# CPU profile of the TUI (localhost pprof). Requires a running TUI with STELL_TUI_PPROF=1.
set -euo pipefail
ADDR="${STELL_TUI_PPROF_ADDR:-127.0.0.1:6060}"
OUT="${1:-/tmp/tui-cpu.pprof}"
SECONDS_N="${2:-10}"

echo "Recording ${SECONDS_N}s CPU profile from http://${ADDR}/debug/pprof/profile"
echo "Start the TUI in another terminal:"
echo "  STELL_TUI_PPROF=1 go run ./tui -ui-only"
curl -fsS "http://${ADDR}/debug/pprof/profile?seconds=${SECONDS_N}" -o "${OUT}"
echo "Wrote ${OUT}"
go tool pprof -top "${OUT}"
