# Specification — Hexagonal Architecture Scaffold (M0)

## Overview

Scaffold Ports & Adapters structure for five gateway modules without implementing proxy, auth, or rate-limit business logic. Establishes contracts, no-op stubs, and constructor-injected composition for Milestone 1 work.

## Responsibilities

| Package | Responsibility |
|---|---|
| `internal/domain` | Pure request/response/identity types |
| `internal/app` | Build dependency graph from config and stubs |
| `internal/auth/port` | Authentication contract |
| `internal/auth/stub` | No-op authenticator |
| `internal/routing/port` | Route resolution contract |
| `internal/routing/stub` | No-op router |
| `internal/ratelimit/port` | Rate limit contract |
| `internal/ratelimit/stub` | Allow-all limiter |
| `internal/middleware/port` | Pipeline contract |
| `internal/middleware/stub` | Passthrough pipeline |
| `internal/observability/port` | Logging and tracing contracts |
| `internal/observability/stub` | No-op telemetry |
| `internal/server` | HTTP adapter; health routes unchanged |

## API

### domain

```go
type Request struct {
    Method  string
    Path    string
    Host    string
    Headers map[string][]string
}

type Response struct {
    StatusCode int
    Headers    map[string][]string
    Body       []byte
}
```

### auth/port

```go
type Identity struct {
    Subject string
    Claims  map[string]string
}

type Authenticator interface {
    Authenticate(ctx context.Context, token string) (Identity, error)
}
```

### routing/port

```go
type Route struct {
    ID       string
    Upstream string
}

var ErrNoRoute = errors.New("routing: no route matched")

type Router interface {
    Resolve(ctx context.Context, req domain.Request) (Route, error)
}
```

### ratelimit/port

```go
type Limiter interface {
    Allow(ctx context.Context, key string) (bool, error)
}
```

### middleware/port

```go
type Handler func(ctx context.Context, req domain.Request) (domain.Response, error)

type Middleware interface {
    Wrap(next Handler) Handler
}

type Pipeline interface {
    Execute(ctx context.Context, req domain.Request) (domain.Response, error)
}
```

### observability/port

```go
type Logger interface {
    Info(msg string, attrs ...any)
    Error(msg string, attrs ...any)
}

type Tracer interface {
    StartSpan(ctx context.Context, name string) (context.Context, func())
}
```

### app

```go
func New(cfg config.Config) server.Dependencies
```

### server

```go
type Dependencies struct {
    Config        config.Config
    Authenticator authport.Authenticator
    Router        routingport.Router
    Limiter       ratelimitport.Limiter
    Pipeline      middlewareport.Pipeline
    Logger        observabilityport.Logger
    Tracer        observabilityport.Tracer
}

func New(deps Dependencies) *Server
```

## Data Flow

1. `main` loads config via `config.Load`.
2. `app.New(cfg)` constructs stubs and returns `server.Dependencies`.
3. `server.New(deps)` registers health handlers and holds port references for M1 pipeline wiring.
4. Incoming health requests bypass gateway pipeline (direct handlers).

## Errors

- Stubs return documented sentinel errors where applicable (`routing/port.ErrNoRoute`).
- No-op auth returns anonymous identity with nil error.
- Passthrough pipeline returns `501 Not Implemented` response body for non-health traffic (not wired to HTTP yet).

## Metrics

Not applicable for M0 scaffold.

## Logs

No-op logger satisfies `observability/port.Logger`; `main` continues using `slog` for process lifecycle logs.

## Tracing

No-op tracer satisfies `observability/port.Tracer`; OpenTelemetry integration deferred to M4.

## Security

Port packages contain no secrets handling. Auth stub does not validate tokens.

## Performance

No runtime impact; stubs are in-memory no-ops.

## Testing

- Smoke test per stub: implements port interface, predictable behavior.
- `domain` request/response construction test.
- `app.New` returns non-nil dependencies.
- Existing `internal/server` health tests remain green.
- `go test ./...` and `go build ./...` pass.
