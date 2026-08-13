# Specification — Header Routing (M1)

## Overview

Route requests by HTTP header name and exact value. Header routes compose with path and default upstream routers in fixed precedence.

## Responsibilities

| Package | Responsibility |
|---|---|
| `internal/config` | Parse and validate `GATEWAY_HEADER_ROUTES` |
| `internal/routing/adapter/header` | `HeaderRouter` — header name/value match |
| `internal/routing/adapter/chain` | Ordered router composition |
| `internal/routing/module` | Build router chain from config |

## API

### Config

```go
type HeaderRoute struct {
    Name     string
    Value    string
    Upstream string
}
```

Format: `HeaderName=HeaderValue=upstream` per entry, comma-separated.

### HeaderRouter

```go
func NewRouter(routes []config.HeaderRoute) (*Router, error)
func (r *Router) Resolve(ctx context.Context, req domain.Request) (port.Route, error)
```

Returns `port.Route{ID: "header:Name=Value", Upstream: upstream}` on match, else `port.ErrNoRoute`.

### Chain router

```go
func NewRouter(routers ...port.Router) *Router
```

Tries each router; first success wins. Propagates non-`ErrNoRoute` errors immediately.

## Precedence

1. Header routes
2. Path routes
3. Default upstream
4. No route (404)

## Testing

- Config parse valid/invalid entries
- Header router: match, no match, canonical header name
- Chain: header beats path; path beats default
- Integration test with httptest upstreams
