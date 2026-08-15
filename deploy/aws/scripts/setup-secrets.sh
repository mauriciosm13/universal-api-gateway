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

require_aws

create_secret_if_missing() {
  local name="$1"
  local value="$2"
  if resource_exists aws secretsmanager describe-secret --secret-id "${name}" --region "${AWS_REGION}"; then
    log "Secret exists: ${name} (not overwritten)"
  else
    aws secretsmanager create-secret \
      --name "${name}" \
      --secret-string "${value}" \
      --region "${AWS_REGION}" \
      >/dev/null
    log "Created secret: ${name}"
  fi
}

JWT_SECRET="${UAG_JWT_SECRET:-change-me-min-32-chars-for-staging!!}"
API_KEYS="${UAG_API_KEYS:-demo-key:demo-user}"

create_secret_if_missing "uag/${ENV}/jwt-hmac-secret" "${JWT_SECRET}"
create_secret_if_missing "uag/${ENV}/api-keys" "${API_KEYS}"

log "Secrets ready under uag/${ENV}/"
