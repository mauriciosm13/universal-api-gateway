# ADR 0003 — AWS Lambda Runtime Adapter

## Context

MVP deploy target is AWS Lambda. The gateway is implemented as a long-running `net/http` server (`cmd/gateway`). Lambda invokes functions per request (or via response streaming in newer modes). We need a runtime strategy that reuses the existing HTTP stack without rewriting the hexagonal core.

## Decision

Use **AWS Lambda Web Adapter (LWA)** with a **container image** deployment:

- Base image includes LWA binary + gateway binary
- Gateway listens on `0.0.0.0:8080` as today
- LWA translates Lambda invocation events to HTTP requests to localhost
- Expose via **Lambda Function URL** for MVP (HTTPS out of the box)
- Add `GATEWAY_RUNTIME=lambda|server` to skip graceful shutdown / signal handling in Lambda mode

LWA image reference: `public.ecr.aws/awsguru/aws-lambda-adapter:0.9.1` (pin in Dockerfile.lambda; bump via PR).

Readiness: `AWS_LWA_READINESS_CHECK_PATH=/health/ready`

## Consequences

**Positive**

- Minimal changes to `internal/server` and handlers
- Same integration tests and demo compose apply to container locally
- Container image also runnable on ECS/K8s if needed later
- No new Go dependencies

**Negative**

- Container images have larger cold start than zip-based Go Lambda
- LWA adds one process in the container
- Lambda-specific config validation needed (timeout alignment)

## Alternatives

1. **Native `aws-lambda-go-api-proxy` or custom handler** — rejected for MVP; requires refactoring HTTP edge and middleware wiring.
2. **API Gateway HTTP API → Lambda (non-container zip)** — rejected; still requires handler refactor; container + LWA is closer to current server model.
3. **ECS Fargate instead of Lambda** — rejected as initial target; higher ops cost for MVP; Lambda chosen by product owner.

## Tradeoffs

Cold start (~200–500ms) acceptable for MVP. Provisioned concurrency deferred. VPC attachment deferred (public upstreams for MVP).
