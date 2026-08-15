# Specification — Upstream Timeout and Retry (MVP Week 6)

## Overview

Bound upstream latency and retry transient failures for idempotent HTTP methods. Implemented as a custom `http.RoundTripper` used by `httputil.ReverseProxy` in the gateway handler.

## Responsibilities

| Package | Responsibility |
|---|---|
| `internal/config/reliability.go` | Parse `GATEWAY_UPSTREAM_*` and `GATEWAY_RETRY_*` env vars |
| `internal/reliability/port/transport.go` | `TransportFactory` or `RoundTripper` port |
| `internal/reliability/adapter/timeout` | `http.Transport` with timeouts |
| `internal/reliability/adapter/retry` | Retry wrapper RoundTripper |
| `internal/server/gateway.go` | Attach transport to `ReverseProxy` |
| `internal/config/config.go` | Embed reliability config; Lambda timeout validation |

## Config

| Variable | Default | Required | Description |
|---|---|---|---|
| `GATEWAY_UPSTREAM_TIMEOUT` | `30s` | No | Per-attempt timeout (connect + response headers) |
| `GATEWAY_RETRY_MAX` | `2` | No | Additional attempts after first failure. `0` disables retry |
| `GATEWAY_RETRY_BACKOFF` | `100ms` | No | Wait between attempts (linear) |

Validation:

- Timeout must be > 0 and ≤ `5m`
- `GATEWAY_RETRY_MAX` range `0..3`
- When `GATEWAY_RUNTIME=lambda`: upstream timeout ≤ Lambda function timeout − 2s

## Retry policy

| Method | Retry |
|---|---|
| GET, HEAD, OPTIONS | Yes (when `GATEWAY_RETRY_MAX` > 0) |
| POST, PUT, PATCH, DELETE | Never |

Retry when:

- `RoundTrip` returns network error (connection refused, timeout, reset)
- Upstream response status is `502`, `503`, or `504`

Do not retry:

- `4xx` (except optionally `408` — not in MVP)
- Successful `2xx`/`3xx`
- Body read errors after headers received (stream already started)

Backoff: attempt `n` waits `GATEWAY_RETRY_BACKOFF × n` before next try.

## Error responses (gateway-generated)

| Condition | Status | Message |
|---|---|---|
| All attempts timeout | 504 | `upstream timeout` |
| Connection failed after retries | 502 | `upstream unavailable` |
| Internal transport error | 500 | `upstream error` |

JSON body matches existing gateway error format (`writeJSONError`).

Upstream response bodies for `502/503/504` from origin are not retried after full body receipt — only when status received within timeout.

## Data flow

```text
gatewayHandler.ServeHTTP
  → pipeline + router (unchanged)
  → proxyFor(upstream) with shared Transport
       → retryRoundTripper
            → timeoutTransport → upstream
```

Transport instances cached per `gatewayHandler` (same pattern as `sync.Map` for proxies).

## Lambda interaction

Lambda default timeout 30s. Recommended staging config:

```bash
GATEWAY_UPSTREAM_TIMEOUT=25s
GATEWAY_RETRY_MAX=1
GATEWAY_RETRY_BACKOFF=200ms
```

Worst case: 25s + 200ms + 25s ≈ 50s — exceeds 30s Lambda timeout. With `RETRY_MAX=1`, document max single-attempt timeout as `Lambda_timeout - backoff - 1s`.

Startup validation enforces: `(GATEWAY_UPSTREAM_TIMEOUT × (1 + GATEWAY_RETRY_MAX)) + backoff_total ≤ lambda_timeout - 2s` or fail fast.

## Security

- Retries only on idempotent methods — no duplicate POST side effects
- No retry on auth failures from upstream

## Performance

- Reuse `http.Transport` with `MaxIdleConnsPerHost` ≥ number of upstreams
- No retry body buffering for GET (no body)

## Testing

| Test | Type |
|---|---|
| Config parse defaults and bounds | Unit |
| Timeout returns 504 | Integration |
| Retry succeeds on second attempt (mock upstream) | Integration |
| POST does not retry | Integration |
| Lambda timeout validation fails startup | Unit |
| `make quality` passes | Gate |

## References

- [RFC MVP-5](../rfcs/mvp-weeks-6-8-reliability-lb-release.md)
- [reverse-proxy.md](reverse-proxy.md)
