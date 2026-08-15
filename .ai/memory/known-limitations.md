# Known Limitations (v0.1 → v0.2.0-mvp target)

## Core (current)

- Regex routing and URL rewrite not implemented (post-MVP M1)
- Gateway 4xx/5xx errors are JSON; upstream/proxy errors may remain stdlib format
- JWT auth applies to all proxied traffic when enabled; no per-path public bypass
- JWKS keys loaded at startup; no background key rotation refresh in MVP
- Configuration is environment-only (no YAML)
- OpenTelemetry enabled but HTTP handlers not instrumented until M4
- In-memory rate limit is per process / per Lambda instance (not global)

## Planned for v0.2.0-mvp (design complete, implementation pending)

- Upstream timeout and retry — [spec](../docs/specs/upstream-timeout-retry.md)
- Round robin load balancing — [spec](../docs/specs/load-balancing-round-robin.md)
- AWS Lambda deploy — [spec](../docs/specs/aws-lambda-runtime.md)

## AWS Lambda (when deployed)

- Rate limit effective RPS scales with concurrent Lambda instances ([ADR 0004](../docs/adrs/0004-rate-limit-lambda-strategy.md))
- Round robin distribution is per instance, not global
- Upstreams must be publicly reachable (no VPC in MVP)
- ~6 MB sync payload limit
- Cold start latency on container Lambda

## Deferred

- OAuth2 / OIDC, mTLS, RBAC
- Distributed rate limiting (Redis / DynamoDB)
- Circuit breaker, caching
- Documentation website (M12)
