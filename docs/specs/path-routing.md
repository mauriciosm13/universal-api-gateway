# Specification — Path Routing (M1)

## Overview

Add prefix-based path routing and centralize HTTP route registration. Path routes are configured via `GATEWAY_PATH_ROUTES`; unmatched requests optionally fall back to `GATEWAY_DEFAULT_UPSTREAM`.

## Responsibilities

| Package | Responsibility |
|---|---|
| `internal/config` | Parse and validate `GATEWAY_PATH_ROUTES` |
| `internal/routing/adapter/path` | `PathRouter` — longest prefix match |
| `internal/routing/module` | Wire `PathRouter`, `StaticRouter`, or `NoOpRouter` |
| `internal/server/routes` | Register health and gateway handlers on `http.ServeMux` |

## API

### Config

```go
type PathRoute struct {
    Prefix   string
    Upstream string
}

type Config struct {
    // ...existing fields...
    PathRoutes []PathRoute // from GATEWAY_PATH_ROUTES
}
```

`GATEWAY_PATH_ROUTES` format: comma-separated `prefix=upstream` entries.

Validation per entry:

- Prefix non-empty, starts with `/`
- Upstream passes `ValidateUpstreamURL`

### PathRouter

```go
func NewRouter(routes []config.PathRoute, defaultUpstream string) (*Router, error)

func (r *Router) Resolve(ctx context.Context, req domain.Request) (port.Route, error)
```

Returns `port.Route{ID: prefix, Upstream: upstream}` on match. When no prefix matches and `defaultUpstream` is set, returns route ID `default`. Otherwise `port.ErrNoRoute`.

### HTTP registration

```go
func newRoutes(deps Dependencies) http.Handler
```

Patterns (Go 1.22+):

- `GET /health`, `GET /health/live`, `GET /health/ready` — local handlers
- `/` — gateway handler (all non-health traffic)

## Data Flow

```text
Client → ServeMux (routes.go)
  → GET /health* → health handlers
  → otherwise → gatewayHandler
      → domain.Request
      → PathRouter.Resolve() (or Static/NoOp)
      → ReverseProxy → upstream
```

## Errors

| Condition | HTTP status | Body |
|---|---|---|
| No prefix match, no default upstream | 404 | `no route matched` |
| Invalid `GATEWAY_PATH_ROUTES` at startup | — | process exit |
| Resolve error (other) | 500 | `routing error` |

## Metrics

Not applicable until M4.

## Logs

No new structured logs in this deliverable.

## Tracing

HTTP handlers not instrumented until M4.

## Security

- Upstream URLs validated at startup
- Prefix boundary check prevents `/api` matching `/apiv2`

## Performance

- Path routes sorted once at router construction
- Linear scan over typically small route tables; acceptable for M1 env config

## Testing

- `config`: parse valid/invalid `GATEWAY_PATH_ROUTES`
- `path.Router`: longest prefix, boundary, default fallback, no match
- `server`: integration — path route proxies to correct upstream; default fallback; health unchanged
- `go test ./...` and `make quality` pass
