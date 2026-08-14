# RFC MVP-2 — JWT Authentication (Week 3)

## Summary

Add real JWT validation to the auth middleware via a new `Authenticator` adapter. Support asymmetric keys (RS256/ES256 via JWKS) and symmetric keys (HS256 via HMAC secret). Preserve no-op auth when no JWT env vars are set.

## Goals

- `Authorization: Bearer <token>` validation before routing/proxy
- Config: `GATEWAY_JWT_JWKS_URL` or `GATEWAY_JWT_HMAC_SECRET` (mutually exclusive)
- Optional `GATEWAY_JWT_ISSUER` and `GATEWAY_JWT_AUDIENCE`
- Invalid, expired, or missing tokens return JSON 401 when JWT is enabled
- Unit and gateway integration tests; `make quality` passes

## Non Goals

- API keys (Week 4)
- Identity propagation to upstream headers (Week 4)
- OAuth2/OIDC login flows
- Public route bypass / path-based skip auth
- RBAC, mTLS

## Config

| Variable | Required | Description |
|---|---|---|
| `GATEWAY_JWT_JWKS_URL` | One of JWKS or HMAC | JWKS document URL for RS256/ES256 |
| `GATEWAY_JWT_HMAC_SECRET` | One of JWKS or HMAC | Shared secret for HS256 (min 32 chars) |
| `GATEWAY_JWT_ISSUER` | No | Expected `iss` claim |
| `GATEWAY_JWT_AUDIENCE` | No | Expected `aud` claim |

When neither JWKS nor HMAC is set, `NoOpAuthenticator` remains (backward compatible).

## Middleware behavior

1. Extract Bearer token from `Authorization` header
2. Malformed header (no `Bearer ` prefix) treated as missing token
3. Call `Authenticator.Authenticate`
4. On error → JSON `{"code":401,"message":"unauthorized"}`
5. On success → continue pipeline (`StatusCode: 0`)

## Migration

No breaking change for deployments without JWT env vars. Enabling JWT requires setting exactly one of JWKS URL or HMAC secret.
