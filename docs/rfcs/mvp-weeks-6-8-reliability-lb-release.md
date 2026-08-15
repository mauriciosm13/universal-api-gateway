# RFC MVP-5 — Reliability, Load Balancing, and Release (Weeks 6–8)

## Summary

Complete the MVP feature track (M3 subset) with upstream timeout and idempotent retry (Week 6), round-robin load balancing across multiple upstreams (Week 7), and release polish including E2E demo, documentation, and tag `v0.2.0-mvp` (Week 8). Runs in parallel with [mvp-aws-lambda-deployment.md](mvp-aws-lambda-deployment.md).

## Motivation

Weeks 1–5 delivered routing, auth, and rate limiting. MVP success criteria require traffic control beyond rate limits: bounded upstream latency, safe retries, and basic load distribution. Week 8 closes the loop with a reproducible demo and release artifacts.

## Goals

### Week 6 — Timeout and retry

- Configurable upstream request timeout via env
- Retry for idempotent methods (GET, HEAD, OPTIONS) with max attempts and backoff
- JSON 502/504 for upstream failures where appropriate
- `internal/reliability/` module (port + adapters)
- Spec: [upstream-timeout-retry.md](../specs/upstream-timeout-retry.md)
- Tests + `make quality` passes

### Week 7 — Load balancing

- Multiple upstream URLs per route (comma-separated)
- Round-robin selection per route (in-memory, per process)
- Spec: [load-balancing-round-robin.md](../specs/load-balancing-round-robin.md)
- Tests + `make quality` passes

### Week 8 — Release polish

- `docker-compose.demo.yml` — 2 mock upstreams + gateway with auth + rate limit
- `scripts/demo.sh` — documented curl flow (route → auth → rate limit → upstream)
- Update README, DEPLOYMENT, CHANGELOG, known-limitations
- Mark MVP items done in ROADMAP / MVP_SCOPE
- Tag `v0.2.0-mvp`
- Bootstrap integration test in `quality/` (Phase 2 placeholder or first real E2E)

## Non Goals

- Circuit breaker, bulkhead, fallback (post-MVP M3)
- Health-aware upstream skip
- Least connections / weighted / sticky load balancing
- Distributed rate limiting (see ADR 0004)
- YAML config, hot reload
- Full observability stack (M4)

## Detailed Design

### Week 6 — Reliability module

```text
internal/reliability/
  port/transport.go       Configurable HTTP transport for ReverseProxy
  adapter/
    timeout/              http.Transport with ResponseHeaderTimeout + overall timeout
    retry/                RoundTripper wrapper: retry idempotent methods only
  module.go               DI wiring (optional; may wire in server/gateway first)
```

Integration point: `gatewayHandler.proxyFor()` sets `ReverseProxy.Transport` to the reliability transport built from config.

| Variable | Default | Description |
|---|---|---|
| `GATEWAY_UPSTREAM_TIMEOUT` | `30s` | Max time for a single upstream attempt (connect + response headers) |
| `GATEWAY_RETRY_MAX` | `2` | Extra attempts after first failure. `0` = no retry |
| `GATEWAY_RETRY_BACKOFF` | `100ms` | Base backoff between retries (linear, no jitter for MVP) |

Retry triggers: network error, `502`, `503`, `504` from upstream. Non-idempotent methods (POST, PUT, PATCH, DELETE) never retry.

504 when gateway exhausts timeout; 502 when upstream connection fails after retries.

### Week 7 — Round robin

Extend route upstream parsing to accept comma-separated URLs:

```bash
GATEWAY_PATH_ROUTES="/api=http://upstream-a:8080,http://upstream-b:8080"
GATEWAY_DEFAULT_UPSTREAM="http://a:8080,http://b:8080"
```

New adapter `internal/routing/adapter/roundrobin/`:

- Wraps a list of upstream URLs
- `Resolve` returns next upstream via atomic counter per route ID
- Integrates in chain: when route config has multiple URLs, use RoundRobinRouter for that segment

Alternative (simpler MVP): resolve to single upstream in a new `UpstreamSelector` port called from `gatewayHandler` before `proxyFor`. Prefer selector in gateway to avoid changing every router adapter.

**Decision:** `internal/routing/port/UpstreamSelector` + `roundrobin` adapter. Routers return `Route{ID, Upstreams []string}`; gateway selects one URL before proxy.

### Week 8 — Demo stack

```text
docker-compose.demo.yml
  gateway       — full env: routes, JWT or API key, rate limit
  upstream-a    — nginx or simple Go echo on :8081
  upstream-b    — nginx or simple Go echo on :8082
scripts/demo.sh — health, auth fail, auth pass, rate limit 429, LB distribution
```

## Parallel Work Streams

| Stream | Weeks | Can start after |
|---|---|---|
| A — Timeout/retry | 6 | This RFC + spec approved |
| B — Round robin | 7 | Spec approved; transport interface stable (Week 6 spec, not necessarily merged) |
| C — AWS Lambda deploy | 6–8 | [mvp-aws-lambda-deployment.md](mvp-aws-lambda-deployment.md) |
| D — E2E demo + docs | 8 | Streams A + B merged |

Streams A and C start in parallel. Stream B starts when upstream selector interface is defined (Day 1 of Week 6 design).

## Alternatives

1. **Retry in middleware pipeline** — rejected; retry belongs at transport/proxy layer after route resolution.
2. **External load balancer only** — rejected for MVP; gateway must support multiple upstreams in config.
3. **Single RFC per week** — rejected; one combined RFC reduces doc overhead (RFC-lite rule).

## Risks

| Risk | Mitigation |
|---|---|
| Retry amplifies load on failing upstream | Cap `GATEWAY_RETRY_MAX` at 3; document ops guidance |
| Timeout vs Lambda timeout (AWS deploy) | Validate `GATEWAY_UPSTREAM_TIMEOUT < Lambda timeout - 2s` at startup when `GATEWAY_RUNTIME=lambda` |
| Round robin uneven across Lambda instances | Document per-instance selection (same as rate limit) |

## Migration

No breaking changes. Existing single-URL routes continue to work. New env vars are optional with safe defaults.

## Open Questions

- Add jitter to retry backoff post-MVP?
- `Retry-After` header on 429 — deferred from Week 5, still post-MVP.

## References

- [MVP_SCOPE.md](../MVP_SCOPE.md)
- [MVP_EXECUTION_PLAN.md](../MVP_EXECUTION_PLAN.md)
- [upstream-timeout-retry.md](../specs/upstream-timeout-retry.md)
- [load-balancing-round-robin.md](../specs/load-balancing-round-robin.md)
