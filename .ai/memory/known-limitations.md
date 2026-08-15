# Known Limitations (v0.2.0-mvp)

## Core

- Regex routing and URL rewrite not implemented (post-MVP M1)
- Gateway 4xx/5xx errors are JSON; upstream/proxy errors may remain stdlib format
- JWT auth applies to all proxied traffic when enabled; no per-path public bypass
- JWKS keys loaded at startup; no background key rotation refresh in MVP
- Configuration is environment-only (no YAML)
- OpenTelemetry enabled but HTTP handlers not instrumented until M4
- In-memory rate limit is per process / per Lambda instance (not global)
- Round robin is per process / per Lambda instance (not global)
- Retry has linear backoff without jitter; non-idempotent methods never retry
- No health-aware upstream skip; failed upstreams still receive traffic
- No circuit breaker, bulkhead, or fallback

## Per-instance scaling (Kubernetes / Lambda)

- Effective global rate limit scales with replica/instance count ([ADR 0004](../docs/adrs/0004-rate-limit-lambda-strategy.md))
- Round robin distribution is per instance, not globally even across replicas

## AWS Lambda

- Rate limit effective RPS scales with concurrent Lambda instances ([ADR 0004](../docs/adrs/0004-rate-limit-lambda-strategy.md))
- Round robin distribution is per instance, not global
- Upstreams must be publicly reachable (no VPC in MVP)
- ~6 MB sync payload limit
- Cold start latency on container Lambda
- `GATEWAY_UPSTREAM_TIMEOUT` plus retries must fit within Lambda timeout minus 2s margin (validated at startup when `GATEWAY_RUNTIME=lambda`)

## Deferred

- OAuth2 / OIDC, mTLS, RBAC
- Distributed rate limiting (Redis / DynamoDB)
- Circuit breaker, caching
- Documentation website (M12)
