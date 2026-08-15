# Specification — Rate Limiting (MVP Week 5)

## Overview

Token bucket rate limiting in the request middleware pipeline. Blocks excess traffic with JSON 429 before routing and upstream proxy.

## Order

Unchanged from [middleware-pipeline.md](middleware-pipeline.md): error → auth → rate limit → continue.

## Responsibilities

| Package | Responsibility |
|---|---|
| `internal/config/ratelimit.go` | Parse `GATEWAY_RATE_LIMIT_*` env vars |
| `internal/ratelimit/adapter/tokenbucket` | Token bucket `Limiter` port adapter |
| `internal/ratelimit/context` | Request path on context for route-aware limits |
| `internal/middleware/adapter/request/ratelimit.go` | Key selection and 429 short-circuit |
| `internal/ratelimit/module.go` | No-op vs token bucket wiring |

## Config

| Variable | Required | Format |
|---|---|---|
| `GATEWAY_RATE_LIMIT_RPS` | No | Positive float. Empty disables rate limiting |
| `GATEWAY_RATE_LIMIT_BURST` | When RPS set | Positive integer |
| `GATEWAY_RATE_LIMIT_ROUTES` | No | Comma-separated `prefix=rps:burst` |

Route prefix must start with `/`. Longest matching prefix wins.

Example:

```bash
export GATEWAY_RATE_LIMIT_RPS="5"
export GATEWAY_RATE_LIMIT_BURST="10"
export GATEWAY_RATE_LIMIT_ROUTES="/api=2:4"
```

## Rate limit key

| Condition | Key |
|---|---|
| Identity subject present and not `anonymous` | `Identity.Subject` |
| Otherwise | `Host + Path` |

Per-client limiting when authenticated; per host+path fallback for anonymous traffic.

Route overrides apply by request path, independent of key source.

## 429 response

```json
{
  "error": "Too Many Requests",
  "code": 429,
  "message": "rate limit exceeded"
}
```

- `Content-Type: application/json`
- Upstream not invoked when blocked

## Token bucket

- Algorithm: in-memory token bucket (stdlib only)
- Parameters: rate (tokens/s), burst (max capacity)
- Thread-safe via `sync.Map` of buckets keyed by client key + limit profile
- `Allow(ctx, key)`: consumes one token; `(false, nil)` when limit exceeded; error only on internal failure
- Not shared across gateway instances (documented limitation)

## Testing

- Config parsing (valid, missing burst, invalid routes)
- Token bucket: allow within burst, block when empty, refill over time
- Middleware key selection (identity vs host+path)
- Integration: burst then 429; refill then 200
- Integration: route limit stricter than global
- Integration: no config → no-op passthrough
