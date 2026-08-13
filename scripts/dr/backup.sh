#!/usr/bin/env bash
set -euo pipefail
SRC="${SESSION_STORE:-./data/sessions}"
DST="${BACKUP_DIR:-./backups}/sessions-$(date -u +%Y%m%dT%H%M%SZ).tgz"
AUDIT="${AUDIT_PATH:-./eval/results/audit.jsonl}"
mkdir -p "$(dirname "$DST")"
tar -czf "$DST" -C "$(dirname "$SRC")" "$(basename "$SRC")" 2>/dev/null || true
if [[ -f "$AUDIT" ]]; then
  cp "$AUDIT" "${DST%.tgz}-audit.jsonl"
fi
echo "backup $DST"
