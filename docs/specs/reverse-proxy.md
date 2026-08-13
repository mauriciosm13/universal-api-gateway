# Specification — Reverse Proxy (M1)

## Overview

Enable HTTP reverse proxy for non-health traffic. The routing port resolves an upstream target; the server HTTP adapter forwards the request with `httputil.ReverseProxy`. A single default upstream is configured via `GATEWAY_DEFAULT_UPSTREAM`.

## Responsibilities

| Package | Responsibility |
|---|---|
| `internal/config` | Load and validate `GATEWAY_DEFAULT_UPSTREAM` |
| `internal/routing/adapter/static` | `StaticRouter` — resolve all requests to default upstream |
| `internal/routing/module` | Wire `StaticRouter` or `NoOpRouter` based on config |
| `internal/server/handler` | Root handler: dispatch health vs gateway traffic |
| `internal/server/gateway` | Proxy handler: resolve route, forward via ReverseProxy |
| `internal/server/request` | Convert `*http.Request` to `domain.Request` |

## API

### Config

```go
type Config struct {
    // ...existing fields...
    DefaultUpstream string // GATEWAY_DEFAULT_UPSTREAM
}
```

Validation: when non-empty, must parse as URL with `http` or `https` scheme and non-empty host.

### StaticRouter

```go
func NewRouter(upstream string) (*Router, error)

func (r *Router) Resolve(ctx context.Context, req domain.Request) (port.Route, error)
```

Returns `port.Route{ID: "default", Upstream: upstream}` for every request.

### Gateway handler

Internal `gatewayHandler` on `server.Server`:

```go
func (h *gatewayHandler) ServeHTTP(w http.ResponseWriter, r *http.Request)
```

## Data Flow

```text
Client → rootHandler
  → GET /health* → local health handlers
  → otherwise → gatewayHandler
      → domain.Request from http.Request
      → Router.Resolve()
      → ReverseProxy.ServeHTTP → upstream
      → response streamed to client
```

## Errors

| Condition | HTTP status | Body |
|---|---|---|
| No route (`ErrNoRoute`) | 404 | JSON — see [JSON Error Responses](json-error-responses.md) |
| Resolve error (other) | 500 | JSON — see [JSON Error Responses](json-error-responses.md) |
| Upstream unreachable | 502 | (stdlib ReverseProxy) |
| Invalid upstream in config | — | process exit at startup |

## Metrics

Not applicable until M4 Prometheus integration.

## Logs

No new structured logs in this deliverable. Process lifecycle logging remains in `main`.

## Tracing

HTTP handlers not instrumented until M4. Port `Tracer` remains wired but unused by proxy handler.

## Security

- Upstream URL validated at startup (scheme + host required)
- No SSRF beyond operator-configured upstream
- Health endpoints remain unproxied

## Performance

- ReverseProxy instances cached per upstream URL (`sync.Map`)
- Streaming passthrough; no body buffering in domain layer

## Testing

- `config`: valid/invalid `GATEWAY_DEFAULT_UPSTREAM`
- `static.Router`: resolves default route; rejects invalid upstream in constructor
- `server`: integration test — httptest upstream, proxy forwards request and returns upstream response
- `server`: no upstream configured → 404 on non-health path
- `server`: health endpoints still 200 when upstream configured
- `go test ./...` and `make quality` pass
