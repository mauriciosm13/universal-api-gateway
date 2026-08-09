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
cmd/gateway/                    Entry point (constructor injection via app)
internal/
  app/                          Composition root — builds dependency graph
  domain/                       Pure request/response types
  auth/
    port/                       Authenticator interface
    stub/                       NoOpAuthenticator
  routing/
    port/                       Router interface
    stub/                       NoOpRouter
  ratelimit/
    port/                       Limiter interface
    stub/                       NoOpLimiter
  middleware/
    port/                       Pipeline and Middleware interfaces
    stub/                       PassthroughPipeline
  observability/
    port/                       Logger and Tracer interfaces
    stub/                       NoOp telemetry
  config/                       Configuration from environment variables
  server/                       HTTP adapter (health endpoints)
docs/
  rfcs/                         Design proposals
  specs/                        Technical specifications
deploy/                         Deployment artifacts (future)
```

## Layer Boundaries

| Layer | Packages | May import |
|---|---|---|
| Domain | `domain` | stdlib only (no `net/http`, `os`) |
| Ports | `*/port` | `domain`, `context`, stdlib |
| Stubs | `*/stub` | matching `port`, `domain` |
| Composition | `app` | ports, stubs, `config`, `server` |
| HTTP edge | `server` | `net/http`, ports, `config` |
| Entry | `cmd/gateway` | `app`, `server`, `config` |

Port packages never import stub or adapter packages. Infrastructure depends inward on ports.

## Dependency Flow

```text
cmd/gateway/main.go
  └── config.Load()
  └── app.New(cfg)              stubs wired to port interfaces
  └── server.New(deps)          HTTP adapter with injected ports
```

Health routes (`/health`, `/health/live`, `/health/ready`) register directly in `internal/server`. Gateway pipeline wiring for proxied traffic is deferred to M1.

## Target Modules

| Module | Responsibility | M0 status |
|---|---|---|
| `auth` | JWT, OAuth2, API keys, mTLS | Port + NoOp stub |
| `routing` | Reverse proxy, path/host routing | Port + NoOp stub |
| `ratelimit` | Token bucket, sliding window | Port + NoOp stub |
| `middleware` | Request/response pipeline | Port + passthrough stub |
| `observability` | Logs, metrics, traces | Port + NoOp stub |

## Design Rules

- One responsibility per package
- Small interfaces (ports)
- Constructor injection, no global state
- Adapters live at the edges; core stays pure

## References

- [RFC 0001 — Hexagonal Architecture Scaffold](rfcs/0001-hexagonal-architecture-scaffold.md)
- [Spec — Hexagonal Architecture Scaffold](specs/hexagonal-architecture-scaffold.md)
