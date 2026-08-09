# Specification — Dependency Injection (M0)

## Overview

Introduce a manual dependency-injection container (`internal/di`) with per-module providers over the hexagonal scaffold. Replaces centralized stub wiring with an explicit, validated graph that supports test overrides without changing `cmd/gateway/main.go`.

## Responsibilities

| Package | Responsibility |
|---|---|
| `internal/di` | `Module` interface, `Builder`, `Provide*` setters, `Build()` validation |
| `internal/auth/module.go` | Register `Authenticator` provider (M0 stub) |
| `internal/routing/module.go` | Register `Router` provider |
| `internal/ratelimit/module.go` | Register `Limiter` provider |
| `internal/middleware/module.go` | Register `Pipeline` provider |
| `internal/observability/module.go` | Register `Logger` and `Tracer` providers |
| `internal/server/module.go` | M0 placeholder for HTTP-edge providers (future adapters) |
| `internal/deps` | Shared `Gateway` struct consumed by DI and HTTP adapter |
| `internal/app` | `DefaultModules()`, `Build()` — application composition entry |
| `cmd/gateway` | Bootstrap: config load, DI build, server lifecycle |

## API

### di

```go
type Module interface {
    Register(b *Builder)
}

type Builder struct { /* unexported fields */ }

func NewBuilder(cfg config.Config, modules ...Module) *Builder

func (b *Builder) ProvideAuthenticator(authport.Authenticator)
func (b *Builder) ProvideRouter(routingport.Router)
func (b *Builder) ProvideLimiter(ratelimitport.Limiter)
func (b *Builder) ProvidePipeline(middlewareport.Pipeline)
func (b *Builder) ProvideLogger(observabilityport.Logger)
func (b *Builder) ProvideTracer(observabilityport.Tracer)

func (b *Builder) Build() (deps.Gateway, error)

// FuncModule adapts a function for tests and one-off overrides.
type FuncModule struct {
    Name string
    Fn   func(b *Builder)
}
```

### Per-module provider

```go
// Example: internal/auth/module.go
type Module struct{}

func (Module) Register(b *di.Builder) {
    b.ProvideAuthenticator(authstub.NewNoOpAuthenticator())
}
```

### app

```go
func DefaultModules() []di.Module

func Build(cfg config.Config, modules ...di.Module) (deps.Gateway, error)
```

When `modules` is empty, `DefaultModules()` is used.

### Override semantics

Modules register in order. Later `Provide*` calls overwrite earlier values for the same port. Tests append a `FuncModule` after defaults to swap implementations.

## Data Flow

1. `main` loads config via `config.Load()`.
2. `app.Build(cfg)` creates `di.NewBuilder(cfg, DefaultModules()...)`.
3. Each module's `Register` calls `Provide*` on the builder.
4. `Builder.Build()` validates all ports are non-nil and returns `deps.Gateway`.
5. `server.New(deps)` constructs the HTTP adapter (`server.Dependencies` is a type alias).
6. Health requests continue to use direct handlers; gateway pipeline wiring remains M1.

## Errors

- `Build()` returns an error if any required port is nil after module registration.
- `main` logs the error and exits with code 1 on build failure.

## Metrics

Not applicable for M0 DI layer.

## Logs

Process lifecycle logging remains in `main` via `slog`. Port `Logger` is wired but not used by health handlers in M0.

## Tracing

Port `Tracer` is wired via observability module; OpenTelemetry integration deferred to M4.

## Security

DI container holds no secrets. Config continues to load from environment variables.

## Performance

Module registration runs once at startup. No reflection; negligible overhead.

## Testing

- `di` builder: full graph builds with all ports non-nil.
- `di` builder: missing provider returns error from `Build()`.
- `di` builder: later module overrides earlier provider (stub swap).
- `app.Build`: default modules produce valid `server.Dependencies`.
- Existing stub, server health, and domain tests remain green.
- `go test ./...` and `go build ./...` pass.
