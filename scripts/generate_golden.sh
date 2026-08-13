#!/usr/bin/env bash
# Regenerates synthetic golden cases (optional). Prefer editing cases by hand for quality.
set -euo pipefail
echo "Golden cases live in eval/golden/*.json — regenerate via project scripts if needed."
ls eval/golden | wc -l
