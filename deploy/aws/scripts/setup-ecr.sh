#!/usr/bin/env bash
set -euo pipefail
source "$(dirname "$0")/common.sh"

require_aws

log "Ensuring ECR repository: ${UAG_ECR_REPO} (${AWS_REGION})"

if resource_exists aws ecr describe-repositories --repository-names "${UAG_ECR_REPO}" --region "${AWS_REGION}"; then
  log "ECR repository already exists"
else
  aws ecr create-repository \
    --repository-name "${UAG_ECR_REPO}" \
    --image-scanning-configuration scanOnPush=true \
    --region "${AWS_REGION}" \
    >/dev/null
  log "Created ECR repository"
fi

log "Registry: $(ecr_registry)"
