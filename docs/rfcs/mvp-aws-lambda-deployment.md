# RFC MVP-6 — AWS Lambda Deployment (AWS CLI)

## Summary

Deploy the gateway to AWS Lambda as a container image using the Lambda Web Adapter (LWA), managed entirely via **AWS CLI scripts** (no SAM/Terraform for MVP). Includes ECR, Function URL, Secrets Manager, GitHub Actions CI, and staging smoke tests. Runs in parallel with Weeks 6–8 feature work.

## Motivation

MVP requires a production-like deploy target. AWS Lambda minimizes ops overhead for initial traffic while staying compatible with the existing HTTP server model. AWS CLI keeps tooling simple and avoids new IaC frameworks in the repo.

## Goals

- `Dockerfile.lambda` — container image with LWA + gateway binary
- `GATEWAY_RUNTIME=lambda|server` — skip signal/graceful shutdown on Lambda
- `deploy/aws/scripts/` — idempotent setup, build/push, deploy scripts
- Secrets Manager for JWT HMAC / API keys; env for non-secrets
- GitHub Actions: test → build image → push ECR → update Lambda (manual or on tag)
- Function URL for HTTPS endpoint
- Spec: [aws-lambda-runtime.md](../specs/aws-lambda-runtime.md)
- ADRs: [0003](../adrs/0003-aws-lambda-runtime-adapter.md), [0004](../adrs/0004-rate-limit-lambda-strategy.md), [0005](../adrs/0005-aws-deploy-aws-cli.md)
- Document limitations (per-instance rate limit, per-instance round robin)

## Non Goals

- AWS API Gateway as the product (may add in front of Function URL post-MVP for custom domain / WAF)
- DynamoDB distributed rate limiting (post-MVP; see ADR 0004)
- VPC-attached Lambda (post-MVP unless upstreams are private)
- Multi-region, blue-green Lambda aliases
- Terraform/SAM/CDK in MVP
- EKS/ECS deploy (K8s manifests remain secondary target)

## Architecture

```text
Internet
  └── Lambda Function URL (HTTPS)
        └── Lambda (container)
              ├── Lambda Web Adapter (invocation → HTTP)
              └── universal-api-gateway :8080
                    ├── middleware (auth, rate limit)
                    ├── routing + round robin
                    └── upstream HTTP(S) targets
```

### AWS resources

| Resource | Purpose |
|---|---|
| ECR repository | `universal-api-gateway` container images |
| Lambda function | Container image, 512–1024 MB, timeout 30s (configurable) |
| Function URL | Public HTTPS endpoint |
| IAM role | Lambda execution + ECR pull + Secrets Manager read |
| Secrets Manager | `gateway/jwt-hmac-secret`, `gateway/api-keys` |
| CloudWatch Logs | Lambda stdout (JSON logs) |
| (Optional) SSM Parameter Store | Non-secret config overrides |

### Runtime mode

| `GATEWAY_RUNTIME` | Behavior |
|---|---|
| `server` (default) | Current behavior: listen, SIGTERM graceful shutdown |
| `lambda` | Listen only; LWA handles lifecycle; validate Lambda-specific config |

Lambda-specific validation at startup:

- `GATEWAY_UPSTREAM_TIMEOUT` must be ≤ Lambda timeout − 2s (read from `AWS_LAMBDA_FUNCTION_TIMEOUT` or env override)

## Deploy Flow

```bash
# One-time (per account/region)
./deploy/aws/scripts/setup-ecr.sh
./deploy/aws/scripts/setup-iam.sh
./deploy/aws/scripts/setup-secrets.sh --env staging

# Every release
./deploy/aws/scripts/build-and-push.sh --tag v0.2.0-mvp
./deploy/aws/scripts/deploy-lambda.sh --env staging --tag v0.2.0-mvp

# Verify
./deploy/aws/scripts/smoke-test.sh --env staging
```

## Configuration on Lambda

Non-secrets via Lambda environment variables (from `deploy/aws/config/lambda-env.staging.json` template):

- `GATEWAY_RUNTIME=lambda`
- `GATEWAY_DEFAULT_UPSTREAM`, route env vars
- `GATEWAY_RATE_LIMIT_*`
- `GATEWAY_JWT_ISSUER`, `GATEWAY_JWT_AUDIENCE`, `GATEWAY_JWT_JWKS_URL`
- `OTEL_SDK_DISABLED=true` (MVP)

Secrets referenced at deploy time:

- `GATEWAY_JWT_HMAC_SECRET` from Secrets Manager
- `GATEWAY_API_KEYS` from Secrets Manager

Deploy script merges env JSON + secret ARNs into `aws lambda update-function-configuration`.

## CI/CD (GitHub Actions)

Workflow `.github/workflows/deploy-aws.yml`:

- Trigger: `workflow_dispatch` + push tag `v0.*`
- Auth: OIDC → IAM role (no long-lived access keys)
- Jobs: `test` → `docker build Dockerfile.lambda` → `ecr push` → `deploy-lambda.sh`

Staging deploy on tag; production requires manual workflow dispatch with `--env prod`.

## Parallel Timeline

| Week | Lambda stream |
|---|---|
| 1 | ADRs, spec, `Dockerfile.lambda`, `setup-*.sh`, local `sam` not used |
| 1 | Deploy to staging Lambda (minimal config, health only) |
| 2 | Secrets, full env, CI workflow |
| 2 | Smoke test with auth + upstream |
| 3 | Production deploy runbook, tag `v0.2.0-mvp` triggers staging + optional prod |

Feature streams (timeout, retry, LB) merge independently; redeploy Lambda after each merge.

## Alternatives

| Option | Verdict |
|---|---|
| Lambda Web Adapter | **Selected** — minimal code change (ADR 0003) |
| Native `lambda.Start` handler | Rejected for MVP — rewrites HTTP edge |
| SAM | Rejected — user chose AWS CLI (ADR 0005) |
| Terraform | Rejected for MVP — add post-MVP if drift becomes painful |

## Risks

| Risk | Mitigation |
|---|---|
| Per-instance rate limit | Document; ADR 0004 accepts for MVP |
| Cold start latency | 512MB+ memory; monitor p99 |
| Upstream in private VPC | Document public upstream requirement for MVP |
| Script drift vs console changes | Scripts idempotent; README warns against manual edits |
| Lambda 30s default timeout | Set 30s; align `GATEWAY_UPSTREAM_TIMEOUT=25s` |

## Migration

K8s and Docker Compose remain supported. Lambda is an additional deploy target, not a replacement.

## Open Questions

- Custom domain via CloudFront + Function URL — post-MVP?
- Provisioned concurrency for cold start — post-MVP?

## References

- [MVP_EXECUTION_PLAN.md](../MVP_EXECUTION_PLAN.md)
- [DEPLOYMENT.md](../DEPLOYMENT.md)
- [aws-lambda-runtime.md](../specs/aws-lambda-runtime.md)
- [Lambda Web Adapter](https://github.com/awslabs/aws-lambda-web-adapter)
