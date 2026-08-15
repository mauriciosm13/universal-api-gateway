#!/usr/bin/env bash
set -euo pipefail
source "$(dirname "$0")/common.sh"

TAG="latest"
while [[ $# -gt 0 ]]; do
  case "$1" in
    --tag) TAG="$2"; shift 2 ;;
    *) die "unknown arg: $1" ;;
  esac
done

require_aws
require_cmd docker

IMAGE_URI="$(ecr_image_uri "${TAG}")"
REGISTRY="$(ecr_registry)"

log "Logging in to ECR: ${REGISTRY}"
aws ecr get-login-password --region "${AWS_REGION}" | \
  docker login --username AWS --password-stdin "${REGISTRY}"

log "Building Lambda image from Dockerfile.lambda (linux/amd64)"
docker build --platform linux/amd64 \
  --provenance=false \
  --sbom=false \
  -f "${REPO_ROOT}/Dockerfile.lambda" \
  -t "${IMAGE_URI}" \
  "${REPO_ROOT}"

log "Pushing ${IMAGE_URI}"
docker push "${IMAGE_URI}"

log "Done: ${IMAGE_URI}"
