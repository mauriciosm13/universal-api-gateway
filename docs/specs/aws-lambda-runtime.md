# Specification — AWS Lambda Runtime (MVP Deploy)

## Overview

Run the gateway as an AWS Lambda container function using Lambda Web Adapter (LWA). Deploy and manage resources via AWS CLI scripts in `deploy/aws/`. Complements Docker Compose and Kubernetes targets.

## Responsibilities

| Component | Responsibility |
|---|---|
| `Dockerfile.lambda` | Multi-stage build: gateway binary + LWA |
| `cmd/gateway/main.go` | Branch on `GATEWAY_RUNTIME` |
| `internal/config/runtime.go` | Runtime mode enum + Lambda validation |
| `deploy/aws/scripts/*` | Idempotent AWS resource lifecycle |
| `deploy/aws/config/*.json` | Non-secret env templates per environment |
| `.github/workflows/deploy-aws.yml` | CI deploy on tag / manual |
| `docs/DEPLOYMENT.md` | Operator runbook |

## Runtime modes

| Value | Listen | Shutdown |
|---|---|---|
| `server` | Yes, block on signal | SIGINT/SIGTERM graceful 10s |
| `lambda` | Yes, block forever | Process exit on Lambda freeze (LWA managed) |

Environment variable: `GATEWAY_RUNTIME` (default `server`).

## Container image

Built from `Dockerfile.lambda`:

```dockerfile
# Stage 1: build gateway (same as Dockerfile)
# Stage 2: public.ecr.aws/awsguru/aws-lambda-adapter:0.9.1
#   COPY gateway → /gateway
#   ENV PORT=8080
#   ENV AWS_LWA_READINESS_CHECK_PATH=/health/ready
#   ENTRYPOINT ["/gateway"]
```

Lambda function settings (staging defaults):

| Setting | Value |
|---|---|
| Package type | Image |
| Architecture | x86_64 |
| Memory | 512 MB |
| Timeout | 30 s |
| Handler | N/A (container) |

## AWS resources (naming convention)

Prefix: `uag-{env}-` (universal-api-gateway)

| Resource | Name example |
|---|---|
| ECR repo | `universal-api-gateway` |
| Lambda function | `uag-staging-gateway` |
| IAM role | `uag-staging-lambda-exec` |
| Secret (JWT) | `uag/staging/jwt-hmac-secret` |
| Secret (API keys) | `uag/staging/api-keys` |

Region: `AWS_REGION` env (default `us-east-1` in scripts).

## Function URL

- Auth type: `NONE` (gateway handles auth)
- CORS: disabled at Function URL (gateway handles CORS post-MVP)
- HTTPS URL output stored in `deploy/aws/.outputs/{env}.json` by deploy script

## Secrets

| Secret | Env var injected |
|---|---|
| JWT HMAC | `GATEWAY_JWT_HMAC_SECRET` |
| API keys | `GATEWAY_API_KEYS` |

Deploy script uses `aws lambda update-function-configuration` with secrets from Secrets Manager (not plain env in console).

JWKS mode: no secret; set `GATEWAY_JWT_JWKS_URL` in env JSON only.

## Config templates

`deploy/aws/config/lambda-env.staging.json`:

```json
{
  "GATEWAY_RUNTIME": "lambda",
  "GATEWAY_HOST": "0.0.0.0",
  "GATEWAY_PORT": "8080",
  "GATEWAY_UPSTREAM_TIMEOUT": "25s",
  "GATEWAY_RETRY_MAX": "1",
  "GATEWAY_DEFAULT_UPSTREAM": "https://httpbin.org",
  "OTEL_SDK_DISABLED": "true"
}
```

Operators copy to `lambda-env.staging.local.json` (gitignored) for overrides.

## CI authentication

GitHub Actions OIDC:

- Trust policy on IAM role `uag-github-actions-deploy`
- Permissions: ECR push, Lambda update, pass role
- No `AWS_ACCESS_KEY_ID` in secrets

## Data flow

```text
Client HTTPS
  → Lambda Function URL
  → Lambda runtime
  → LWA → http://127.0.0.1:8080/...
  → gateway middleware + proxy
  → upstream URL
```

## Errors

| Failure | Behavior |
|---|---|
| Config invalid at cold start | Lambda init error → 502 from Function URL |
| Upstream timeout | 504 JSON (Week 6) |
| Auth failure | 401 JSON |

## Security

- Execution role least privilege: logs, ECR pull, secrets read
- Secrets never in git — templates use placeholders
- Function URL public — gateway auth required for protected routes
- Read-only root filesystem not required on Lambda (container writable /tmp)

## Performance

- Cold start target: < 1s p99 (512MB, no VPC)
- No provisioned concurrency in MVP

## Testing

| Test | Type |
|---|---|
| `Dockerfile.lambda` builds locally | Manual / CI |
| Container responds to curl `/health/ready` locally | Manual |
| `smoke-test.sh` against staging Function URL | Post-deploy |
| Config validation lambda mode | Unit |

## Known limitations

- Per-instance rate limit and round robin (ADR 0004)
- No VPC — upstreams must be reachable from public internet
- Max request/response size: Lambda 6MB sync payload (streaming via LWA for larger — document 6MB MVP limit)
- WebSocket not supported

## References

- [RFC MVP-6](../rfcs/mvp-aws-lambda-deployment.md)
- [ADR 0003](../adrs/0003-aws-lambda-runtime-adapter.md)
- [ADR 0005](../adrs/0005-aws-deploy-aws-cli.md)
- [DEPLOYMENT.md](../DEPLOYMENT.md)
