# RFC 0008 — Path Routing and Centralized HTTP Route Registration

## Summary

Centralize HTTP route registration in `internal/server/routes.go` and add prefix-based path routing via `GATEWAY_PATH_ROUTES`, with optional fallback to `GATEWAY_DEFAULT_UPSTREAM`.

## Motivation

M1 reverse proxy (RFC 0007) forwards all traffic to a single upstream. Production gateways route by path prefix. Health dispatch currently lives inline in `handler.go`; the roadmap calls for a dedicated registration file before richer routing rules.

## Goals

- Register health and gateway handlers in `internal/server/routes.go` using `http.ServeMux`
- Match requests by longest path prefix via a new routing adapter
- Configure path routes through `GATEWAY_PATH_ROUTES` (comma-separated `prefix=upstream` pairs)
- Fall back to `GATEWAY_DEFAULT_UPSTREAM` when no prefix matches
- Preserve M0/M1 behavior when path routes are unset
- Unit and integration tests

## Non Goals

- Host, header, method, or regex routing (separate M1 items)
- YAML configuration or hot reload
- URL rewrite or strip-prefix (separate roadmap item)
- Middleware pipeline changes

## Detailed Design

### Configuration

- `GATEWAY_PATH_ROUTES` — optional comma-separated list: `/api=http://api:8080,/v2=http://v2:8080`
- Each entry: path prefix (must start with `/`) and validated upstream URL, separated by the first `=`
- `GATEWAY_DEFAULT_UPSTREAM` — unchanged; used as fallback when path routes are configured

Router selection in `routing.Module` (see RFC 0009 for chain precedence):

| Header routes | Path routes | Default upstream | Router |
|---|---|---|---|
| empty | empty | empty | `NoOpRouter` |
| empty | empty | set | `StaticRouter` |
| any | any | any | `ChainRouter` (header → path → static) |

### Path router adapter

```text
internal/routing/adapter/path/
  router.go    PathRouter — longest prefix match with boundary check
```

Prefix `/api` matches `/api` and `/api/users` but not `/apiv2`. Routes sorted by prefix length descending before match.

### HTTP route registration

```text
internal/server/
  routes.go    ServeMux: GET /health*, catch-all gateway handler
```

`server.New` uses `newRoutes(deps)` instead of inline health switch in `handler.go`.

## Alternatives

| Alternative | Rejected because |
|---|---|
| Regex-only routing | Higher complexity; prefix routing covers common cases first |
| JSON env var | Comma-separated pairs are simpler for M1 env-only config |
| Path rules in server package | Violates hexagonal boundaries; belongs in routing adapter |

## Risks

- **Overlapping prefixes** — longest match resolves ambiguity; document in spec
- **Misconfigured prefix without leading `/`** — rejected at config load

## Migration

No breaking changes. Existing `GATEWAY_DEFAULT_UPSTREAM`-only deployments behave identically.

## Open Questions

- Strip-prefix forwarding deferred until URL rewrite item
