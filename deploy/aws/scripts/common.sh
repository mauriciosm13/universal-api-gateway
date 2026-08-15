#!/usr/bin/env bash
# Shared helpers for AWS deploy scripts.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEPLOY_AWS_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
REPO_ROOT="$(cd "${DEPLOY_AWS_DIR}/../.." && pwd)"

: "${AWS_REGION:=us-east-1}"
: "${UAG_PROJECT:=universal-api-gateway}"
: "${UAG_ECR_REPO:=universal-api-gateway}"

log() { printf '[uag-aws] %s\n' "$*"; }
die() { log "ERROR: $*"; exit 1; }

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || die "missing required command: $1"
}

require_aws() {
  require_cmd aws
  aws sts get-caller-identity >/dev/null 2>&1 || die "AWS credentials not configured"
}

account_id() {
  aws sts get-caller-identity --query Account --output text
}

ecr_registry() {
  echo "$(account_id).dkr.ecr.${AWS_REGION}.amazonaws.com"
}

ecr_image_uri() {
  local tag="${1:-latest}"
  echo "$(ecr_registry)/${UAG_ECR_REPO}:${tag}"
}

lambda_name() {
  local env="${1:-staging}"
  echo "uag-${env}-gateway"
}

lambda_role_name() {
  local env="${1:-staging}"
  echo "uag-${env}-lambda-exec"
}

outputs_file() {
  local env="${1:-staging}"
  mkdir -p "${DEPLOY_AWS_DIR}/.outputs"
  echo "${DEPLOY_AWS_DIR}/.outputs/${env}.json"
}

env_config_file() {
  local env="${1:-staging}"
  local local_file="${DEPLOY_AWS_DIR}/config/lambda-env.${env}.local.json"
  local template="${DEPLOY_AWS_DIR}/config/lambda-env.${env}.json"
  if [[ -f "${local_file}" ]]; then
    echo "${local_file}"
  else
    echo "${template}"
  fi
}

resource_exists() {
  # usage: resource_exists aws lambda get-function --function-name NAME
  if "$@" >/dev/null 2>&1; then
    return 0
  fi
  return 1
}
