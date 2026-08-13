#!/usr/bin/env bash
set -euo pipefail
# Restore drill. RTO target: see docs/sla.md (< 15 min for session store).
ARCHIVE="${1:?usage: restore.sh <backup.tgz>}"
DST="${SESSION_STORE:-./data/sessions}"
mkdir -p "$DST"
tar -xzf "$ARCHIVE" -C "$(dirname "$DST")"
echo "restored into $DST — restart brain/gateway pods"
