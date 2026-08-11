# RFC 0007 — Reverse Proxy

## Summary

Add HTTP reverse proxy to the gateway: resolve upstream via the routing port, forward non-health traffic with `net/http/httputil.ReverseProxy`, and configure a default upstream through environment variables.

## Motivation

Milestone 1 starts with a production-ready reverse proxy. M0 scaffolded routing ports and health endpoints but did not forward traffic. Operators need a working proxy before path, host, and header routing land as separate deliverables.

## Goals

- Forward non-health HTTP requests to a configured upstream
- Use hexagonal boundaries: `Router` resolves routes; `server` adapts HTTP
- Validate upstream URL at config load time
- Preserve streaming (no full-body buffering in domain types)
- Keep health endpoints (`/health`, `/health/live`, `/health/ready`) local
- Unit and integration tests with `httptest` upstream

## Non Goals

- Path, host, header, method, or regex routing (separate M1 items)
- YAML configuration or hot reload
- Middleware pipeline integration for proxied traffic (deferred to middleware pipeline item)
- Authentication, rate limiting, or request transformation
- OpenTelemetry HTTP instrumentation (M4)
- TLS termination or mutual TLS

## Detailed Design

### Configuration

- `GATEWAY_DEFAULT_UPSTREAM` — optional absolute URL (e.g. `http://backend:8080`)
- When unset, gateway keeps M0 behavior: non-health traffic returns 404 via `NoOpRouter`
- When set, `routing.Module` registers `adapter/static.Router` that resolves every request to the default upstream

### Routing adapter

```text
internal/routing/adapter/static/
  router.go    StaticRouter — always resolves to configured upstream
```

`StaticRouter` implements `routing/port.Router`. Upstream URL is validated in `config.Load()` and again in constructor.

### HTTP edge

```text
internal/server/
  gateway.go   catch-all handler: Resolve → ReverseProxy
  request.go   http.Request → domain.Request conversion
```

Registration in `server.New`:

- Existing health routes (method-specific, higher precedence)
- `/{path...}` catch-all for all methods → gateway handler

Flow:

1. Convert `*http.Request` to `domain.Request`
2. `deps.Router.Resolve(ctx, req)`
3. `ErrNoRoute` → 404
4. Success → cached `httputil.ReverseProxy` for `route.Upstream` → `ServeHTTP`

Reverse proxy instances are cached per upstream URL string (`sync.Map`).

### Module wiring

`routing.Module.Register` reads `b.Config().DefaultUpstream`:

- empty → `stub.NoOpRouter`
- set → `static.NewRouter(upstream)`

## Alternatives

| Alternative | Rejected because |
|---|---|
| Domain-level proxy port returning `domain.Response` | Breaks streaming; buffers entire body |
| Third-party proxy library (Caddy, etc.) | New dependency; stdlib sufficient for M1 |
| Single upstream hardcoded in server | Bypasses routing port; blocks future multi-route |

## Risks

| Risk | Mitigation |
|---|---|
| Catch-all intercepts health routes | Go 1.22+ ServeMux specificity; health patterns registered first |
| Invalid upstream at runtime | Validated at config load; StaticRouter constructor double-checks |
| Upstream unreachable | ReverseProxy returns 502; covered in integration test |

## Migration

No breaking changes. Existing deployments without `GATEWAY_DEFAULT_UPSTREAM` behave as M0 (404 on non-health paths). Set the env var to enable proxying.

## Open Questions

- None for this deliverable. Path-based routing will replace `StaticRouter` in the next M1 item.
