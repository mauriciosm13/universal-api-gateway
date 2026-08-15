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

ROLE_NAME="$(lambda_role_name "${ENV}")"
ACCOUNT="$(account_id)"
TRUST_POLICY="$(cat <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": { "Service": "lambda.amazonaws.com" },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
)"

log "Ensuring IAM role: ${ROLE_NAME}"

if resource_exists aws iam get-role --role-name "${ROLE_NAME}"; then
  log "Role already exists"
else
  aws iam create-role \
    --role-name "${ROLE_NAME}" \
    --assume-role-policy-document "${TRUST_POLICY}" \
    >/dev/null
  log "Created role"
fi

# Managed policies
for policy in arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole; do
  aws iam attach-role-policy --role-name "${ROLE_NAME}" --policy-arn "${policy}" 2>/dev/null || true
done

# Inline policy: ECR pull + Secrets Manager read
POLICY_NAME="uag-${ENV}-lambda-inline"
INLINE_POLICY="$(cat <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ecr:GetDownloadUrlForLayer",
        "ecr:BatchGetImage",
        "ecr:BatchCheckLayerAvailability"
      ],
      "Resource": "arn:aws:ecr:${AWS_REGION}:${ACCOUNT}:repository/${UAG_ECR_REPO}"
    },
    {
      "Effect": "Allow",
      "Action": "ecr:GetAuthorizationToken",
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": "secretsmanager:GetSecretValue",
      "Resource": "arn:aws:secretsmanager:${AWS_REGION}:${ACCOUNT}:secret:uag/${ENV}/*"
    }
  ]
}
EOF
)"

aws iam put-role-policy \
  --role-name "${ROLE_NAME}" \
  --policy-name "${POLICY_NAME}" \
  --policy-document "${INLINE_POLICY}"

log "Role ARN: arn:aws:iam::${ACCOUNT}:role/${ROLE_NAME}"
