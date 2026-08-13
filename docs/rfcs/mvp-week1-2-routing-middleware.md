# RFC MVP-1 — Host/Method Routing and Middleware Pipeline

## Summary

Complete MVP weeks 1–2: host and method routing adapters with updated chain precedence, plus a request middleware pipeline (error, auth, rate-limit hooks) executed before reverse proxy.

## Goals

- `GATEWAY_HOST_ROUTES` and `GATEWAY_METHOD_ROUTES` env configuration
- Router chain: host → header → method → path → default upstream
- Request pipeline with error, auth, and rate-limit middleware (no-op passthrough until M2/M3)
- JSON error responses from middleware short-circuit
- Tests and spec updates

## Non Goals

- JWT/API key enforcement (week 3+)
- Real rate limiting (week 5)
- YAML config, hot reload

## Router precedence

1. Host
2. Header
3. Method
4. Path
5. Default upstream
6. No route (404)

## Middleware order

1. Error (outermost — catches middleware errors)
2. Auth (uses `Authenticator` port; noop accepts all today)
3. Rate limit (uses `Limiter` port; noop allows all today)
4. Terminal handler (`StatusCode: 0` → continue to proxy)

## Migration

No config changes required. New env vars optional.
