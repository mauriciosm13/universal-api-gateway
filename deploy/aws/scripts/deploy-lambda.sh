#!/usr/bin/env bash
set -euo pipefail
source "$(dirname "$0")/common.sh"

ENV="staging"
TAG="latest"
MEMORY=512
TIMEOUT=30

while [[ $# -gt 0 ]]; do
  case "$1" in
    --env) ENV="$2"; shift 2 ;;
    --tag) TAG="$2"; shift 2 ;;
    --memory) MEMORY="$2"; shift 2 ;;
    --timeout) TIMEOUT="$2"; shift 2 ;;
    *) die "unknown arg: $1" ;;
  esac
done

require_aws
require_cmd jq

FUNC_NAME="$(lambda_name "${ENV}")"
ROLE_ARN="arn:aws:iam::$(account_id):role/$(lambda_role_name "${ENV}")"
IMAGE_URI="$(ecr_image_uri "${TAG}")"
CONFIG_FILE="$(env_config_file "${ENV}")"

[[ -f "${CONFIG_FILE}" ]] || die "missing config: ${CONFIG_FILE}"

log "Deploying ${FUNC_NAME} image=${IMAGE_URI} config=${CONFIG_FILE}"

merge_secrets() {
  local base
  base="$(cat "${CONFIG_FILE}")"
  if resource_exists aws secretsmanager describe-secret \
      --secret-id "uag/${ENV}/jwt-hmac-secret" --region "${AWS_REGION}"; then
    local jwt
    jwt="$(aws secretsmanager get-secret-value \
      --secret-id "uag/${ENV}/jwt-hmac-secret" \
      --region "${AWS_REGION}" \
      --query SecretString --output text)"
    base="$(echo "${base}" | jq --arg v "${jwt}" '. + {GATEWAY_JWT_HMAC_SECRET: $v}')"
  fi
  if resource_exists aws secretsmanager describe-secret \
      --secret-id "uag/${ENV}/api-keys" --region "${AWS_REGION}"; then
    local keys
    keys="$(aws secretsmanager get-secret-value \
      --secret-id "uag/${ENV}/api-keys" \
      --region "${AWS_REGION}" \
      --query SecretString --output text)"
    base="$(echo "${base}" | jq --arg v "${keys}" '. + {GATEWAY_API_KEYS: $v}')"
  fi
  echo "${base}"
}

ENV_VARS="$(merge_secrets | jq -c '{Variables: .}')"

if resource_exists aws lambda get-function --function-name "${FUNC_NAME}" --region "${AWS_REGION}"; then
  log "Updating existing function"
  aws lambda update-function-code \
    --function-name "${FUNC_NAME}" \
    --image-uri "${IMAGE_URI}" \
    --region "${AWS_REGION}" \
    >/dev/null

  aws lambda wait function-updated-v2 --function-name "${FUNC_NAME}" --region "${AWS_REGION}"

  aws lambda update-function-configuration \
    --function-name "${FUNC_NAME}" \
    --timeout "${TIMEOUT}" \
    --memory-size "${MEMORY}" \
    --environment "${ENV_VARS}" \
    --region "${AWS_REGION}" \
    >/dev/null
else
  log "Creating function"
  aws lambda create-function \
    --function-name "${FUNC_NAME}" \
    --package-type Image \
    --code "ImageUri=${IMAGE_URI}" \
    --role "${ROLE_ARN}" \
    --timeout "${TIMEOUT}" \
    --memory-size "${MEMORY}" \
    --architectures x86_64 \
    --environment "${ENV_VARS}" \
    --region "${AWS_REGION}" \
    >/dev/null

  aws lambda wait function-active-v2 --function-name "${FUNC_NAME}" --region "${AWS_REGION}"
fi

# Function URL (idempotent)
if resource_exists aws lambda get-function-url-config --function-name "${FUNC_NAME}" --region "${AWS_REGION}"; then
  log "Function URL already exists"
else
  aws lambda create-function-url-config \
    --function-name "${FUNC_NAME}" \
    --auth-type NONE \
    --region "${AWS_REGION}" \
    >/dev/null
  log "Created Function URL"
fi

# Allow public invoke via URL
aws lambda add-permission \
  --function-name "${FUNC_NAME}" \
  --statement-id "FunctionURLAllowPublicAccess" \
  --action lambda:InvokeFunctionUrl \
  --principal "*" \
  --function-url-auth-type NONE \
  --region "${AWS_REGION}" \
  2>/dev/null || true

FUNCTION_URL="$(aws lambda get-function-url-config \
  --function-name "${FUNC_NAME}" \
  --region "${AWS_REGION}" \
  --query 'FunctionUrl' --output text)"

OUT="$(outputs_file "${ENV}")"
jq -n \
  --arg env "${ENV}" \
  --arg function "${FUNC_NAME}" \
  --arg url "${FUNCTION_URL}" \
  --arg image "${IMAGE_URI}" \
  --arg tag "${TAG}" \
  '{env: $env, function: $function, functionUrl: $url, image: $image, tag: $tag, deployedAt: (now | todate)}' \
  > "${OUT}"

log "Function URL: ${FUNCTION_URL}"
log "Outputs: ${OUT}"
