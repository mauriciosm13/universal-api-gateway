#!/usr/bin/env bash
# E2E demo — health, auth, rate limit, round-robin load balancing.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT}"

BASE="${GATEWAY_URL:-http://localhost:8080}"
SECRET="${GATEWAY_JWT_HMAC_SECRET:-demo-secret-min-32-characters-long!!}"
FAILURES=0

fail() {
  echo "FAIL: $*" >&2
  FAILURES=$((FAILURES + 1))
}

pass() {
  echo "PASS: $*"
}

wait_for_gateway() {
  echo "==> Waiting for gateway"
  for _ in $(seq 1 60); do
    if curl -sf "${BASE}/health/ready" >/dev/null 2>&1; then
      pass "gateway ready"
      return 0
    fi
    sleep 1
  done
  fail "gateway not ready after 60s"
  return 1
}

http_code() {
  curl -s -o /dev/null -w '%{http_code}' "$@"
}

assert_status() {
  local label=$1 expected=$2
  shift 2
  local code
  code="$(http_code "$@")"
  if [[ "${code}" == "${expected}" ]]; then
    pass "${label} → HTTP ${code}"
  else
    fail "${label} → HTTP ${code}, want ${expected}"
  fi
}

require_go() {
  if ! command -v go >/dev/null 2>&1; then
    fail "go required for JWT generation (cmd/gen-jwt)"
    return 1
  fi
}

gen_jwt() {
  GATEWAY_JWT_HMAC_SECRET="${SECRET}" go run ./cmd/gen-jwt --sub demo-user
}

wait_for_gateway

echo "==> 1. Health"
assert_status "GET /health/ready" "200" "${BASE}/health/ready"

echo "==> 2. Route without auth (expect 401)"
assert_status "GET /api/echo no auth" "401" "${BASE}/api/echo"

echo "==> 3. Route with valid JWT (expect 200 + upstream body)"
require_go || true
TOKEN=""
if command -v go >/dev/null 2>&1; then
  TOKEN="$(gen_jwt)"
fi

if [[ -z "${TOKEN}" ]]; then
  fail "could not generate JWT"
else
  body_file="$(mktemp)"
  code="$(curl -s -o "${body_file}" -w '%{http_code}' \
    -H "Authorization: Bearer ${TOKEN}" \
    "${BASE}/api/echo")"
  body="$(cat "${body_file}")"
  rm -f "${body_file}"

  if [[ "${code}" == "200" && "${body}" == *"upstream"* ]]; then
    pass "GET /api/echo with JWT → HTTP 200, body contains upstream"
  else
    fail "GET /api/echo with JWT → HTTP ${code}, body=${body}"
  fi
fi

echo "==> 4. Round robin (expect both upstream-a and upstream-b)"
if [[ -n "${TOKEN}" ]]; then
  seen_a=0
  seen_b=0
  for _ in $(seq 1 8); do
    headers="$(curl -s -D - -o /dev/null \
      -H "Authorization: Bearer ${TOKEN}" \
      "${BASE}/api/lb-check")"
    if echo "${headers}" | grep -qi 'x-upstream-id: upstream-a'; then
      seen_a=1
    fi
    if echo "${headers}" | grep -qi 'x-upstream-id: upstream-b'; then
      seen_b=1
    fi
  done
  if [[ "${seen_a}" -eq 1 && "${seen_b}" -eq 1 ]]; then
    pass "round robin saw upstream-a and upstream-b"
  else
    fail "round robin missing upstream (a=${seen_a} b=${seen_b})"
  fi
else
  fail "round robin skipped (no JWT)"
fi

echo "==> 5. Rate limit burst (expect some 429)"
if [[ -n "${TOKEN}" ]]; then
  got_429=0
  for i in $(seq 1 15); do
    code="$(http_code -H "Authorization: Bearer ${TOKEN}" "${BASE}/api/burst-${i}")"
    if [[ "${code}" == "429" ]]; then
      got_429=1
    fi
  done
  if [[ "${got_429}" -eq 1 ]]; then
    pass "rate limit burst produced 429"
  else
    fail "rate limit burst: no 429 in 15 requests"
  fi
else
  fail "rate limit test skipped (no JWT)"
fi

echo
if [[ "${FAILURES}" -gt 0 ]]; then
  echo "Demo FAILED (${FAILURES} assertion(s))" >&2
  exit 1
fi

echo "Demo PASSED — all 5 scenarios OK"
