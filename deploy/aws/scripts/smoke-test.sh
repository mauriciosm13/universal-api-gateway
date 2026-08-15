#!/usr/bin/env bash
set -euo pipefail
source "$(dirname "$0")/common.sh"

ENV="staging"
while [[ $# -gt 0 ]]; do
  case "$1" in
    --env) ENV="$2"; shift 2 ;;
    *) die "unknown arg: $1" ;;
  esac
done

require_cmd curl
require_cmd jq

OUT="$(outputs_file "${ENV}")"
[[ -f "${OUT}" ]] || die "missing outputs file: ${OUT} — run deploy-lambda.sh first"

URL="$(jq -r '.functionUrl' "${OUT}")"
[[ "${URL}" != "null" && -n "${URL}" ]] || die "invalid function URL in ${OUT}"

log "Smoke test: ${URL}"

code="$(curl -s -o /dev/null -w '%{http_code}' "${URL}health/ready")"
if [[ "${code}" != "200" ]]; then
  die "/health/ready returned ${code}"
fi
log "OK /health/ready → ${code}"

code="$(curl -s -o /dev/null -w '%{http_code}' "${URL}health/live")"
if [[ "${code}" != "200" ]]; then
  die "/health/live returned ${code}"
fi
log "OK /health/live → ${code}"

code="$(curl -s -o /dev/null -w '%{http_code}' "${URL}get")"
if [[ "${code}" != "200" ]]; then
  die "proxy GET /get returned ${code}"
fi
log "OK proxy GET /get → ${code}"

code="$(curl -s -o /dev/null -w '%{http_code}' "${URL}api/smoke")"
case "${code}" in
  401)
    log "OK auth enforced on /api → 401"
    if [[ -n "${SMOKE_JWT:-}" ]]; then
      auth_code="$(curl -s -o /dev/null -w '%{http_code}' \
        -H "Authorization: Bearer ${SMOKE_JWT}" \
        "${URL}api/smoke")"
      if [[ "${auth_code}" != "200" ]]; then
        die "authenticated GET /api/smoke returned ${auth_code}"
      fi
      log "OK authenticated proxy /api/smoke → ${auth_code}"
    else
      log "SKIP authenticated proxy (set SMOKE_JWT to enable)"
    fi
    ;;
  200)
    log "OK proxy /api/smoke without auth (JWT not configured) → 200"
    ;;
  *)
    die "GET /api/smoke returned ${code}"
    ;;
esac

log "Smoke test passed"
