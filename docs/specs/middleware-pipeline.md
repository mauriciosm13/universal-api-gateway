# Specification — Middleware Pipeline (MVP Week 2)

## Overview

Execute ordered middleware on proxied traffic before routing. Short-circuit with JSON domain responses or continue to reverse proxy.

## Order

1. Error middleware
2. Auth middleware (`Authenticator` port)
3. Rate limit middleware (`Limiter` port)
4. Terminal continue handler

## Rate limit key

After auth, rate limit middleware selects the bucket key:

| Condition | Key |
|---|---|
| Identity subject present and not `anonymous` | `Identity.Subject` |
| Otherwise | `Host + Path` |

Route-specific limits match request path (longest prefix). See [rate-limiting.md](rate-limiting.md).

## Continue semantics

`StatusCode: 0` means continue to routing and proxy.

## Testing

- Pipeline continue path
- Rate limit 429 short-circuit
- Error middleware maps handler errors to JSON 500
