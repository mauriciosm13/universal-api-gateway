# Specification — API Keys and Identity (MVP Week 4)

## Overview

Validate API keys alongside JWT authentication, attach caller identity to request context, and propagate identity to upstream via `X-User-Id`.

## Responsibilities

- Parse API keys from configurable header or query parameter
- Validate keys against env-configured allowlist
- Compose JWT and API key authenticators when both enabled
- Store `authport.Identity` in context after successful auth
- Forward `X-User-Id` to upstream without overwriting client header
- Preserve no-op auth when no JWT or API key configuration

## API

### Environment variables

| Variable | Default | Description |
|---|---|---|
| `GATEWAY_API_KEYS` | — | `key1:name1,key2:name2` or JSON object/map |
| `GATEWAY_API_KEY_HEADER` | `X-API-Key` | Header for API key |
| `GATEWAY_API_KEY_QUERY` | `api_key` | Query param for API key |

JWT variables unchanged (see [jwt-authentication.md](jwt-authentication.md)).

### RequestAuthenticator port

```go
AuthenticateRequest(ctx context.Context, req domain.Request) (Identity, error)
```

### Identity context

Package `internal/auth/context`:

```go
WithIdentity(ctx context.Context, id Identity) context.Context
IdentityFrom(ctx context.Context) (Identity, bool)
```

### Upstream headers

| Header | Source | Precedence |
|---|---|---|
| `X-User-Id` | `Identity.Subject` | Gateway sets only when client did not send header |

## Data Flow

1. HTTP request → domain request (headers + query)
2. Pipeline: error → auth → rate limit → continue
3. Auth middleware calls `AuthenticateRequest`
4. Success → identity in context → routing → proxy with `X-User-Id`
5. Failure → JSON 401

## Composite credential order

When JWT and API keys both enabled:

1. If `Authorization` header present → require valid Bearer JWT (malformed → 401)
2. Else extract API key from header or query
3. No credentials → 401

## Errors

| Condition | HTTP | Body |
|---|---|---|
| Missing credentials (auth enabled) | 401 | `{"code":401,"message":"unauthorized"}` |
| Malformed Bearer | 401 | same |
| Invalid JWT or API key | 401 | same |
| Invalid startup API key config | panic at module register | process exit |

## Security

- Do not leak validation failure details in responses
- API keys env-only in MVP (no hashing at rest in gateway)
- Never overwrite inbound `X-User-Id`

## Testing

- Config: key parsing (comma and JSON), header/query defaults
- Unit: API key validator (valid, invalid, empty)
- Unit: composite (JWT ok, key ok, both fail, single-mode wiring)
- Unit: identity context helpers
- Integration: valid API key proxies; upstream receives `X-User-Id`
- Integration: JWT or API key accepted when both enabled
- Integration: `/auth/validate` returns identity for both credential types

## Known limitations

- No per-route public bypass
- API keys plain text in env
- No upstream claims headers beyond `X-User-Id` in MVP
