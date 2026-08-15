# AWS Lambda Deploy — universal-api-gateway

Deploy the gateway to AWS Lambda using **AWS CLI scripts** (see [ADR 0005](../docs/adrs/0005-aws-deploy-aws-cli.md)).

## Prerequisites

- AWS CLI v2 (`aws --version`)
- Docker
- `curl`, `jq`
- AWS credentials configured (`aws sts get-caller-identity`)

## Quick start (staging)

```bash
export AWS_REGION=us-east-1
export AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)

# One-time setup
./deploy/aws/scripts/setup-ecr.sh
./deploy/aws/scripts/setup-iam.sh --env staging
./deploy/aws/scripts/setup-secrets.sh --env staging

# Configure non-secret env (copy template, edit upstream)
cp deploy/aws/config/lambda-env.staging.json deploy/aws/config/lambda-env.staging.local.json
# edit lambda-env.staging.local.json

# Build, push, deploy
./deploy/aws/scripts/build-and-push.sh --tag latest
./deploy/aws/scripts/deploy-lambda.sh --env staging --tag latest

# Verify
./deploy/aws/scripts/smoke-test.sh --env staging
```

Function URL is printed by deploy script and saved to `deploy/aws/.outputs/staging.json`.

## Layout

```text
deploy/aws/
  README.md                 This file
  scripts/
    common.sh                 Shared helpers
    setup-ecr.sh              ECR repository
    setup-iam.sh              Lambda execution role
    setup-secrets.sh          Secrets Manager
    build-and-push.sh         Docker build + ECR push
    deploy-lambda.sh          Lambda function + Function URL
    smoke-test.sh             Post-deploy checks
  config/
    lambda-env.staging.json   Non-secret env template
    lambda-env.prod.json      Production template
  .outputs/                   Generated URLs (gitignored)
```

## Environments

| Env | Lambda name | Config file |
|---|---|---|
| `staging` | `uag-staging-gateway` | `lambda-env.staging.local.json` or `.staging.json` |
| `prod` | `uag-prod-gateway` | `lambda-env.prod.local.json` or `.prod.json` |

Local override files (`*.local.json`) are gitignored.

## Secrets

Created by `setup-secrets.sh`:

| Secret ID | Maps to |
|---|---|
| `uag/{env}/jwt-hmac-secret` | `GATEWAY_JWT_HMAC_SECRET` |
| `uag/{env}/api-keys` | `GATEWAY_API_KEYS` |

Set values interactively or via env `UAG_JWT_SECRET` / `UAG_API_KEYS` when running setup.

## CI/CD

GitHub Actions workflow `.github/workflows/deploy-aws.yml`:

- Manual dispatch or push tag `v0.*`
- Requires OIDC IAM role (see workflow comments)

## Limitations (MVP)

- Rate limit and round robin are **per Lambda instance** — see [ADR 0004](../docs/adrs/0004-rate-limit-lambda-strategy.md)
- Upstreams must be **publicly reachable** (no VPC)
- Do not edit resources in AWS Console without updating scripts

## Troubleshooting

| Symptom | Check |
|---|---|
| Init error on cold start | CloudWatch Logs; config validation |
| 502 from Function URL | LWA readiness `/health/ready`; container port 8080 |
| 504 on proxy | Reduce `GATEWAY_UPSTREAM_TIMEOUT`; check Lambda timeout |
| Auth 401 | Secrets Manager values; redeploy after secret update |

## References

- [RFC MVP-6](../docs/rfcs/mvp-aws-lambda-deployment.md)
- [Spec — AWS Lambda runtime](../docs/specs/aws-lambda-runtime.md)
- [DEPLOYMENT.md](../docs/DEPLOYMENT.md)
