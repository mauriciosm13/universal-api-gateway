# ADR 0002 — JWT Validation Library

## Context

MVP Week 3 requires JWT signature and expiration validation (HS256, RS256, ES256) with optional issuer and audience checks. The gateway already defines `auth/port.Authenticator`; we need a production-grade JWT parser without reinventing JWS/JWT semantics.

## Decision

Adopt **github.com/golang-jwt/jwt/v5** as the JWT validation library.

- Mature, widely used in the Go ecosystem
- Supports HS256, RS256, ES256 and standard claim validation (`exp`, `iss`, `aud`)
- Integrates with custom key resolution for JWKS (fetch and cache at startup in our adapter)
- No vendor lock-in; pure Go

JWKS fetching and key parsing remain in `internal/auth/adapter/jwt/` to keep the dependency surface minimal (no separate JWKS library for MVP).

## Consequences

**Positive**

- Standard JWT parsing and validation
- Community-maintained security fixes
- Clear upgrade path for key rotation (refresh JWKS document)

**Negative**

- New external dependency (module update, CI cache)
- Operators must configure JWT env vars correctly; misconfiguration fails at startup

## Alternatives

1. **Custom JWT parser** — rejected; error-prone for crypto and claim edge cases.
2. **github.com/lestrrat-go/jwx** — rejected for MVP; larger API surface than needed for validate-only use case.
3. **github.com/auth0/go-jwt-middleware** — rejected; middleware-oriented; we already have hexagonal auth port and middleware.

## Tradeoffs

Startup-time JWKS fetch is acceptable for MVP. Background JWKS refresh can be added post-MVP if issuers rotate keys frequently.
