# RFC MVP-3 — API Keys and Identity Propagation (Week 4)

## Summary

Add API key authentication alongside existing JWT validation, compose credentials when both are configured, attach authenticated identity to request context, and forward `X-User-Id` to upstream services.

## Goals

- API key validation from configurable header (default `X-API-Key`) or query param (default `api_key`)
- Keys loaded from `GATEWAY_API_KEYS` env (comma `key:name` pairs or JSON)
- Composite auth: JWT Bearer first, then API key when both enabled
- Preserve no-op auth when neither JWT nor API keys are configured
- Identity in `context.Context` after successful auth
- Upstream header `X-User-Id` set from identity subject (never overwrite client value)
- `GET /auth/validate` returns identity for JWT or API key
- Unit and integration tests; `make quality` passes

## Non Goals

- Rate limiting (Week 5)
- Per-path public bypass
- OAuth2/OIDC, RBAC, mTLS
- API key hashing, rotation, or DB storage
- Granular key scopes

## Config

| Variable | Default | Description |
|---|---|---|
| `GATEWAY_API_KEYS` | — | Allowed keys: `key1:name1,key2:name2` or JSON map/array |
| `GATEWAY_API_KEY_HEADER` | `X-API-Key` | Header name for API key |
| `GATEWAY_API_KEY_QUERY` | `api_key` | Query param name for API key |

Existing JWT variables unchanged. Authenticator selection:

| JWT | API keys | Result |
|---|---|---|
| off | off | `NoOpAuthenticator` |
| on | off | JWT only |
| off | on | API key only |
| on | on | Composite (JWT then API key) |

## Middleware behavior

1. Call `RequestAuthenticator.AuthenticateRequest`
2. JWT path when `Authorization: Bearer` present (malformed Bearer → 401)
3. API key path when no Bearer header (header or query key)
4. On success → store identity in context, continue pipeline
5. On failure → JSON `{"code":401,"message":"unauthorized"}`

## Upstream propagation

After auth, before reverse proxy:

- Set `X-User-Id: <subject>` when absent on inbound request
- Do not overwrite existing client `X-User-Id`

## Migration

No breaking change for deployments without API key env vars. JWT-only behavior unchanged.
