# Specification — JWT Authentication (MVP Week 3)

## Overview

Validate JWT bearer tokens in the auth middleware before routing and reverse proxy. When JWT configuration is absent, preserve no-op authentication.

## Responsibilities

- Parse `Authorization: Bearer <token>` (middleware)
- Validate signature and expiration (JWT adapter)
- Optionally validate `iss` and `aud` claims
- Return `authport.Identity` with `Subject` and stringified claims
- Startup failure on invalid or conflicting JWT configuration

## API

### Environment variables

| Variable | Description |
|---|---|
| `GATEWAY_JWT_JWKS_URL` | JWKS URL for RS256/ES256 (mutually exclusive with HMAC) |
| `GATEWAY_JWT_HMAC_SECRET` | HMAC secret for HS256, minimum 32 characters |
| `GATEWAY_JWT_ISSUER` | Optional expected issuer |
| `GATEWAY_JWT_AUDIENCE` | Optional expected audience |

### Authenticator port

```go
Authenticate(ctx context.Context, token string) (Identity, error)
```

Empty or invalid token returns error. Success returns `Identity{Subject, Claims}`.

## Data Flow

1. HTTP request enters gateway handler
2. Middleware pipeline: error → auth → rate limit → continue
3. Auth middleware extracts Bearer token
4. JWT adapter validates token
5. Success → routing → proxy; failure → JSON 401

## Errors

| Condition | HTTP | Body |
|---|---|---|
| Missing token (JWT enabled) | 401 | `{"code":401,"message":"unauthorized"}` |
| Malformed Authorization (no Bearer) | 401 | same |
| Invalid/expired/wrong iss/aud | 401 | same |
| Invalid startup JWT config | panic at module register | process exit |

## Security

- Do not leak validation failure details in responses
- HMAC secret minimum length enforced at config load
- JWKS URL must be absolute http/https

## Testing

- Config: mutual exclusion, URL validation, HMAC length
- Unit: valid HS256, expired, bad signature, wrong iss/aud
- Unit: JWKS RS256 with mock HTTP JWKS server
- Integration: valid JWT proxies; invalid/missing JWT returns 401 JSON
- No JWT config: no-op behavior unchanged

## Known limitations

- No per-route public bypass (all proxied traffic requires JWT when enabled)
- JWKS keys loaded at startup; no background refresh in MVP
- Identity forwarded to upstream via `X-User-Id` (see [api-keys-and-identity.md](docs/specs/api-keys-and-identity.md))
