# ADR 0004 — Rate Limiting Strategy on AWS Lambda

## Context

MVP rate limiting uses in-memory token buckets (`internal/ratelimit/adapter/tokenbucket`). On Kubernetes or a single process, operators accept per-replica limits. Lambda scales to N concurrent execution environments, each with isolated memory — effective global limit becomes `configured_RPS × concurrent_instances`.

Options for Lambda MVP:

| Strategy | Complexity | Accuracy |
|---|---|---|
| A — Per-instance (current) | None | Approximate |
| B — DynamoDB conditional writes | High | Global |
| C — API Gateway throttling in front | Medium | Edge-only |

## Decision

Adopt **Strategy A — per-instance in-memory token bucket** for Lambda MVP.

- No new dependencies
- Same code path as Docker/K8s
- Document effective limit formula in DEPLOYMENT.md and known-limitations
- Revisit DynamoDB adapter post-MVP if global limits are required

Optional operator mitigation (documented, not automated in MVP):

- Set `GATEWAY_RATE_LIMIT_RPS` lower than desired global limit divided by expected max concurrency
- Add API Gateway HTTP API stage throttling post-MVP as a second line of defense

## Consequences

**Positive**

- Ship Lambda deploy without new ports/adapters
- Consistent behavior across all deploy targets
- ADR creates explicit follow-up path

**Negative**

- Rate limit is not globally exact under Lambda autoscaling
- Operators must understand per-instance semantics

## Alternatives

1. **DynamoDB token bucket** — rejected for MVP; new dependency, ADR, and ops cost.
2. **ElastiCache Redis** — rejected; VPC + cost + dependency.
3. **Disable rate limit on Lambda** — rejected; feature is MVP-in-scope.

## Tradeoffs

Accept approximate limits until traffic patterns justify distributed state. Round-robin load balancing has the same per-instance caveat (documented together).
