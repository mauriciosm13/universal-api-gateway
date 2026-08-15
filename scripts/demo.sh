#!/usr/bin/env bash
# E2E demo script — Week 8 (requires docker-compose.demo.yml + JWT from cmd/gen-jwt)
set -euo pipefail

BASE="${GATEWAY_URL:-http://localhost:8080}"
SECRET="${GATEWAY_JWT_HMAC_SECRET:-demo-secret-min-32-characters-long!!}"

echo "==> Health"
curl -sf "${BASE}/health/ready" | head -c 200
echo

echo "==> Route without auth (expect 401 when JWT configured)"
code=$(curl -s -o /dev/null -w '%{http_code}' "${BASE}/api/test" || true)
echo "HTTP ${code}"

if command -v go >/dev/null 2>&1; then
  TOKEN=$(go run ./cmd/gen-jwt --secret "${SECRET}" --sub demo-user 2>/dev/null || echo "")
  if [[ -n "${TOKEN}" ]]; then
    echo "==> Route with JWT"
    curl -sf -H "Authorization: Bearer ${TOKEN}" "${BASE}/api/test" || echo "(upstream/route may vary until Week 7 LB merged)"
  fi
fi

echo "==> Rate limit burst (expect some 429)"
for i in $(seq 1 12); do
  curl -s -o /dev/null -w '%{http_code} ' "${BASE}/health" || true
done
echo

echo "Demo complete. See docs/MVP_EXECUTION_PLAN.md for full checklist."
