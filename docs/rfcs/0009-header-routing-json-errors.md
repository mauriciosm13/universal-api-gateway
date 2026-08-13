# RFC 0009 — Header Routing and JSON Error Responses

## Summary

Add HTTP header-based routing via `GATEWAY_HEADER_ROUTES`, compose routing adapters in precedence order (header → path → default upstream), and return structured JSON bodies for gateway-generated 4xx/5xx errors.

## Motivation

M1 path routing (RFC 0008) covers prefix matching only. Operators also route by headers (API version, tenant, environment). Plain-text error bodies (`no route matched`) are hard to consume from clients and API tools like Postman.

## Goals

- Match requests by header name and exact value
- Configure routes through `GATEWAY_HEADER_ROUTES`
- Chain routers: header → path → static default → NoOp
- Return JSON error responses with `Content-Type: application/json` for gateway errors
- Preserve backward compatibility when header routes are unset
- Unit and integration tests

## Non Goals

- Host, method, or regex routing
- Header regex or prefix matching
- JSON errors for upstream/proxy failures (stdlib ReverseProxy responses unchanged)
- YAML configuration or hot reload

## Detailed Design

### Header routing configuration

- `GATEWAY_HEADER_ROUTES` — comma-separated `HeaderName=HeaderValue=upstream` entries
- Example: `X-Version=v1=http://v1:8080,X-Version=v2=http://v2:8080`
- Header name lookup uses canonical HTTP header keys; value match is case-sensitive

### Router chain

```text
internal/routing/adapter/chain/
  router.go    tries routers in order until one resolves
```

`routing.Module` builds:

1. Header router (if `GATEWAY_HEADER_ROUTES` set)
2. Path router without embedded default (if `GATEWAY_PATH_ROUTES` set)
3. Static router (if `GATEWAY_DEFAULT_UPSTREAM` set)
4. NoOp router (if none configured)

### JSON errors

```text
internal/server/json_error.go
```

Response shape:

```json
{
  "error": "Not Found",
  "code": 404,
  "message": "no route matched"
}
```

Used by `gatewayHandler` for routing and upstream configuration errors only.

## Alternatives

| Alternative | Rejected because |
|---|---|
| Single monolithic router | Violates hexagonal adapter pattern; harder to test |
| RFC 7807 Problem Details | Heavier than needed for M1; can evolve later |
| Header routing in server layer | Belongs in routing port adapters |

## Risks

- **Ambiguous `=` in config** — parse with `SplitN(entry, "=", 3)` so upstream URLs stay intact
- **Multiple matching headers** — first configured route wins (documented)

## Migration

No breaking changes for existing env-only deployments. Error response format changes from plain text to JSON for gateway errors.

## Open Questions

- RFC 7807 adoption deferred to API management milestone
