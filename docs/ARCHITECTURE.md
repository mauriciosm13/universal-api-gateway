# Architecture

## Style

Hexagonal Architecture (Ports & Adapters).

## Goals

- Cloud agnostic deployment
- Dependency inversion between core logic and infrastructure
- Testable domain without external dependencies
- Clear module boundaries for auth, routing, and rate limiting

## Current Structure (v0.1)

```text
cmd/gateway/                    Entry point (bootstrap only)
internal/
  di/                           DI container — Builder, Module, Provide*, Build()
  deps/                         Shared Gateway dependency struct (breaks di/server cycle)
  app/                          Application composition — DefaultModules(), Build()
  domain/                       Pure request/response types
  auth/
    port/                       Authenticator interface
    stub/                       NoOpAuthenticator
    module.go                   DI provider for auth
  routing/
    port/                       Router interface
    stub/                       NoOpRouter
    adapter/static/             StaticRouter (default upstream)
    module.go                   DI provider for routing
  ratelimit/
    port/                       Limiter interface
    stub/                       NoOpLimiter
    module.go                   DI provider for ratelimit
  middleware/
    port/                       Pipeline and Middleware interfaces
    stub/                       PassthroughPipeline
    module.go                   DI provider for middleware
  observability/
    port/                       Logger and Tracer interfaces
    stub/                       NoOp telemetry
    adapter/                    OTel port.Tracer adapter
    otelsetup/                  TracerProvider bootstrap
    module.go                   DI provider for observability
  config/                       Configuration from environment variables
  server/
    module.go                   DI placeholder for HTTP-edge providers
    server.go                   HTTP adapter (health endpoints)
docs/
  rfcs/                         Design proposals
  specs/                        Technical specifications
deploy/
  kubernetes/
    base/               Deployment, Service, ConfigMap, Namespace
    overlays/           dev, staging, prod (Kustomize)
```

## Layer Boundaries

| Layer | Packages | May import |
|---|---|---|
| Domain | `domain` | stdlib only (no `net/http`, `os`) |
| Ports | `*/port` | `domain`, `context`, stdlib |
| Stubs | `*/stub` | matching `port`, `domain` |
| DI | `di` | ports, `config`, `deps` |
| Shared deps | `deps` | ports, `config` |
| Module providers | `*/module.go` | `di`, matching `stub`/`adapter`, `port` |
| Composition | `app` | `di`, module packages, `config`, `server` |
| HTTP edge | `server` | `net/http`, ports, `config`, `di` |
| Entry | `cmd/gateway` | `app`, `server`, `config` |

Port packages never import stub or adapter packages. Infrastructure depends inward on ports.

## Dependency Flow

```text
cmd/gateway/main.go
  └── config.Load()
  └── app.Build(cfg)                    application composition
        └── di.NewBuilder(cfg, modules...)
              ├── auth.Module           → Authenticator
              ├── routing.Module        → Router
              ├── ratelimit.Module      → Limiter
              ├── middleware.Module     → Pipeline
              ├── observability.Module  → Logger, Tracer
              └── server.Module         → (M0: HTTP edge uses Dependencies post-Build)
  └── server.New(deps)                  HTTP adapter + lifecycle
```

Config enters the graph via the `Builder` constructor (environment bootstrap), not via a DI module.

Module registration order matters for overrides: later `Provide*` calls replace earlier values. Tests append a `di.FuncModule` to swap stubs without changing `main.go`.

Health routes (`/health`, `/health/live`, `/health/ready`) register directly in `internal/server`. Non-health traffic is proxied via `gatewayHandler` when `GATEWAY_DEFAULT_UPSTREAM` is set; otherwise the catch-all returns 404.

## Target Modules

| Module | Responsibility | M0 status |
|---|---|---|
| `auth` | JWT, OAuth2, API keys, mTLS | Port + NoOp stub + DI module |
| `routing` | Reverse proxy, path/host routing | Port + StaticRouter adapter + DI module |
| `ratelimit` | Token bucket, sliding window | Port + NoOp stub + DI module |
| `middleware` | Request/response pipeline | Port + passthrough stub + DI module |
| `observability` | Logs, metrics, traces | Port + NoOp stub + DI module |

## Design Rules

- One responsibility per package
- Small interfaces (ports)
- Constructor injection, no global state
- Adapters live at the edges; core stays pure
- DI graph explicit and validated at build time

## References

- [RFC 0001 — Hexagonal Architecture Scaffold](rfcs/0001-hexagonal-architecture-scaffold.md)
- [RFC 0002 — Dependency Injection](rfcs/0002-dependency-injection.md)
- [Spec — Hexagonal Architecture Scaffold](specs/hexagonal-architecture-scaffold.md)
- [RFC 0003 — Kubernetes Base Manifests](rfcs/0003-kubernetes-base-manifests.md)
- [RFC 0005 — OpenTelemetry Setup](rfcs/0005-opentelemetry-setup.md)
- [RFC 0007 — Reverse Proxy](rfcs/0007-reverse-proxy.md)
- [Spec — Reverse Proxy](specs/reverse-proxy.md)
- [ADR 0001 — OpenTelemetry Go SDK](adrs/0001-opentelemetry-go-sdk.md)
- [Spec — OpenTelemetry Setup](specs/opentelemetry-setup.md)
