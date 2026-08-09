# RFC 0001 — Hexagonal Architecture Scaffold

## Summary

Introduce Ports & Adapters package structure for auth, routing, ratelimit, middleware, and observability modules. Scaffold only — no business logic for M1+ features.

## Motivation

Milestone 0 requires a testable, cloud-agnostic foundation before reverse proxy, authentication, and rate limiting land in later milestones. Hexagonal architecture enforces dependency inversion: core contracts stay free of HTTP and infrastructure imports.

## Goals

- Scaffold `internal/` modules with small port interfaces and no-op stubs
- Keep `internal/server` as the HTTP adapter at the system edge
- Wire dependencies via constructor injection in `cmd/gateway/main.go`
- Preserve existing health endpoints and configuration behavior
- Document real package boundaries in `docs/ARCHITECTURE.md`

## Non Goals

- Reverse proxy or real routing (M1)
- DI frameworks (wire, fx) — next roadmap item
- OpenTelemetry instrumentation
- New external dependencies
- Kubernetes manifests

## Detailed Design

### Package layout

```text
internal/
  domain/                 Shared pure types (Request, Response, Identity)
  app/                    Composition root — builds dependency graph
  auth/
    port/                 Authenticator interface
    stub/                 NoOpAuthenticator
  routing/
    port/                 Router interface
    stub/                 NoOpRouter
  ratelimit/
    port/                 Limiter interface
    stub/                 NoOpLimiter (always allow)
  middleware/
    port/                 Pipeline, Middleware interfaces
    stub/                 Passthrough pipeline
  observability/
    port/                 Logger, Tracer interfaces
    stub/                 NoOp logger and tracer
  config/                 Configuration (existing)
  server/                 HTTP adapter (existing, extended with injected deps)
```

### Boundaries

| Layer | Packages | May import |
|---|---|---|
| Domain | `domain` | stdlib only (no `net/http`, `os`) |
| Ports | `*/port` | `domain`, `context`, stdlib |
| Stubs | `*/stub` | matching `port`, `domain` |
| Composition | `app` | all ports and stubs, `config` |
| HTTP edge | `server` | `net/http`, ports, `config` |
| Entry | `cmd/gateway` | `app`, `server`, `config` |

Port packages never import adapter or stub packages. Adapters depend inward on ports, not the reverse.

### Dependency graph

```text
cmd/gateway/main.go
  └── app.New(cfg)           composition root
        ├── auth/stub
        ├── routing/stub
        ├── ratelimit/stub
        ├── middleware/stub
        └── observability/stub
  └── server.New(deps)       HTTP adapter
        └── health handlers (unchanged)
```

`server.New` receives a `server.Dependencies` struct holding config and port interfaces. Health routes register immediately; gateway pipeline wiring happens in M1.

### Port contracts (M0)

- **auth/port.Authenticator** — `Authenticate(ctx, token) (Identity, error)`
- **routing/port.Router** — `Resolve(ctx, Request) (Route, error)`
- **ratelimit/port.Limiter** — `Allow(ctx, key) (bool, error)`
- **middleware/port.Pipeline** — `Execute(ctx, Request) (Response, error)`
- **observability/port.Logger** — structured log methods
- **observability/port.Tracer** — span lifecycle hooks

Stubs satisfy interfaces with safe defaults: anonymous identity, route-not-found, allow-all, not-implemented response, no-op telemetry.

## Alternatives

1. **Single `internal/domain` with all interfaces** — rejected; couples unrelated modules and grows into a god package.
2. **Per-module `adapter/http` now** — premature; only health HTTP exists today. HTTP stays in `internal/server`.
3. **DI framework (wire/fx)** — deferred to next M0 task per roadmap.

## Risks

- Package proliferation before features land — mitigated by clear RFC/spec and documented tree in ARCHITECTURE.md.
- Stubs diverging from future real adapters — mitigated by spec-defined contracts and smoke tests.

## Migration

No breaking API changes. Health endpoints and env-based config unchanged. `server.New(cfg)` becomes `server.New(deps)`; callers update through `app` composition root.

## Open Questions

- None for M0 scaffold. Module-specific adapter layout (e.g. `auth/adapter/jwt`) decided in M2 RFCs.
