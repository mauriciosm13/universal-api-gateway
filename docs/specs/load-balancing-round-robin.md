# Specification — Load Balancing (Round Robin, MVP Week 7)

## Overview

Distribute requests across multiple upstream URLs configured for a route using in-memory round robin. Per-process selection; not coordinated across Lambda instances or K8s pods.

## Responsibilities

| Package | Responsibility |
|---|---|
| `internal/routing/port/route.go` | Extend `Route` with `Upstreams []string` |
| `internal/routing/adapter/roundrobin/selector.go` | `UpstreamSelector` — pick next URL |
| `internal/config/routing.go` | Parse comma-separated upstream lists in route env vars |
| `internal/server/gateway.go` | Call selector before `proxyFor` |
| Existing routers | Populate `Upstreams` from parsed config |

## Config format

Comma-separated URLs in existing route variables (backward compatible — single URL unchanged):

```bash
GATEWAY_DEFAULT_UPSTREAM="http://upstream-a:8080,http://upstream-b:8080"
GATEWAY_PATH_ROUTES="/api=http://a:8080,http://b:8080,/v2=http://v2:8080"
GATEWAY_HOST_ROUTES="api.example.com=http://a:8080,http://b:8080"
GATEWAY_HEADER_ROUTES="X-Version=v1=http://a:8080,http://b:8080"
GATEWAY_METHOD_ROUTES="GET=http://a:8080,http://b:8080"
```

Parsing rules:

- Split on `=` for route key → upstream list (existing)
- Split upstream list on `,` (trim spaces)
- Each URL validated like today (`http`/`https`, non-empty host)
- Minimum 1 URL; 2+ enables round robin

## Upstream selection

```go
type UpstreamSelector interface {
    Next(routeID string, upstreams []string) (string, error)
}
```

Round robin implementation:

- Key counter by `routeID` (e.g. `path:/api`, `default`)
- Atomic increment modulo `len(upstreams)`
- Thread-safe via `sync.Map` of `*atomic.Uint64`

Gateway flow:

```text
route, err := Router.Resolve(...)
upstream := route.Upstreams[0]  // when len == 1
upstream := selector.Next(route.ID, route.Upstreams)  // when len > 1
proxy.ServeHTTP(w, r) via proxyFor(upstream)
```

## Data flow

Unchanged from reverse proxy except selection step between resolve and proxy.

## Errors

| Condition | Status |
|---|---|
| Empty upstream list after resolve | 500 `invalid upstream` |
| Invalid URL in list | Fail at config load |

## Limitations

- No health checks — unhealthy upstream still receives traffic
- No sticky sessions
- Per-instance round robin — distribution approximate under multiple gateway replicas
- Same limitation on Lambda (see ADR 0004)

## Metrics

Deferred to M4. No new metrics in MVP.

## Testing

| Test | Type |
|---|---|
| Parse single vs multiple upstreams | Unit |
| Round robin alternates A, B, A, B | Unit |
| Integration: 2 mock upstreams, 4 requests → 2 each | Integration |
| Single upstream unchanged behavior | Integration |
| Invalid comma URL fails config load | Unit |

## References

- [RFC MVP-5](../rfcs/mvp-weeks-6-8-reliability-lb-release.md)
- [path-routing.md](path-routing.md)
