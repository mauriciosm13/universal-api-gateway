# RFC MVP-4 — Token Bucket Rate Limiting (Week 5)

## Summary

Add in-memory token bucket rate limiting with env-based global and per-route limits. Excess traffic receives JSON 429 before upstream is called.

## Goals

- Token bucket limiter adapter implementing `ratelimitport.Limiter`
- Config via `GATEWAY_RATE_LIMIT_*` env vars
- Rate limit key from authenticated identity subject, fallback host+path
- JSON 429 short-circuit in existing middleware pipeline
- No-op limiter when rate limit env is unset
- Unit and integration tests; `make quality` passes

## Non Goals

- Distributed rate limiting (Redis)
- Sliding window / leaky bucket
- Rate limit by client IP
- `Retry-After` / `X-RateLimit-*` headers
- Rate limit before auth
- Public route bypass

## Config

| Variable | Default | Description |
|---|---|---|
| `GATEWAY_RATE_LIMIT_RPS` | — | Global rate (tokens/s). Empty = rate limit off |
| `GATEWAY_RATE_LIMIT_BURST` | — | Global burst. Required when RPS set |
| `GATEWAY_RATE_LIMIT_ROUTES` | — | Per path prefix: `/api=10:20,/admin=2:4` (rps:burst) |

## Middleware behavior

1. Auth runs first (unchanged)
2. Build rate limit key: identity subject when present and not `anonymous`, else `host+path`
3. Match request path to route override (longest prefix), else global limits
4. `Allow` consumes one token; `(false, nil)` when exhausted
5. Blocked → JSON `{"code":429,"message":"rate limit exceeded"}`; upstream not called

## Limitations

- In-memory buckets are per process; not shared across replicas
- Stateless between instances by design for MVP

## Migration

No breaking change. Deployments without rate limit env vars keep allow-all behavior.
