#!/usr/bin/env bash
set -euo pipefail
# Chaos: stop a runtime replica; Brain Failover should use a healthy one.
# Ephemeral only — do not run against shared prod.
NAME="${1:-stell-runtime-1}"
echo "killing ${NAME}"
docker kill "$NAME" || kubectl delete pod -l app=runtime --force --grace-period=0
echo "expected: gateway Execute retries; user sees tool error at worst, not a hang"
