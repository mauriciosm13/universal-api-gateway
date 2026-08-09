# RFC 0002 — Dependency Injection

## Summary

Formalize constructor-based dependency injection over the M0 hexagonal scaffold using a manual container (`internal/di`) with per-module providers. Replaces the hardcoded composition in `internal/app/app.go` with an explicit, testable, and extensible dependency graph — no external DI framework.

## Motivation

RFC 0001 established ports, stubs, and a manual composition root that wires every stub in one function. That works for five modules but does not scale: adding a module requires editing the central wiring file, swapping implementations for tests requires copy-pasting structs, and the dependency graph is implicit.

Milestone 0 requires a DI layer that keeps constructor injection, avoids global state, and lets M1+ swap stubs for real adapters without changing `cmd/gateway/main.go`.

## Goals

- Explicit dependency graph assembled from registered modules
- Per-module `module.go` providers (auth, routing, ratelimit, middleware, observability, server)
- `cmd/gateway/main.go` limited to bootstrap: config, DI build, server lifecycle
- Testable wiring: non-nil deps, correct interfaces, stub replaceable via module override
- Preserve health endpoints and hexagonal boundaries
- No new external dependencies

## Non Goals

- Real auth, routing, or rate-limit implementations (M1+)
- OpenTelemetry setup (M4)
- Runtime service locator or global mutable registry
- Kubernetes manifests or CI changes

## Detailed Design

### Approach: manual container (chosen)

Introduce `internal/di` with:

- **`Module` interface** — `Register(b *Builder)` registers providers for one bounded context
- **`Builder`** — holds config and port instances; `Build()` validates and returns `server.Dependencies`
- **`Provide*`** methods — modules set port implementations on the builder (last registration wins for overrides)

`internal/app` becomes the application composition entry: `DefaultModules()` lists M0 modules; `Build(cfg, modules...)` delegates to `di.NewBuilder`.

### Dependency graph

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
              └── server.Module         → (M0: no extra providers; HTTP edge uses Dependencies)
  └── server.New(deps)                HTTP adapter + lifecycle
```

Config enters the graph via `Builder` constructor (not a module) because it originates from the environment at bootstrap.

### Module registration

Each feature package exports a zero-value `Module` struct implementing `di.Module`:

```go
// internal/auth/module.go
func (Module) Register(b *di.Builder) {
    b.ProvideAuthenticator(authstub.NewNoOpAuthenticator())
}
```

Adding a new module in M1+: create `module.go`, append to `app.DefaultModules()`. No edits to `main.go` or other modules' providers.

### Test overrides

Tests register an additional module after defaults (or pass a custom module list) to replace a port:

```go
di.FuncModule("test-auth", func(b *di.Builder) {
    b.ProvideAuthenticator(fakeAuth)
})
```

Later modules overwrite earlier provider values — explicit, no reflection.

### Package boundaries

| Package | Role | May import |
|---|---|---|
| `di` | Builder, Module interface | ports, `config`, `deps` |
| `deps` | Shared `Gateway` dependency struct | ports, `config` |
| `*/module.go` | Module provider | `di`, matching `stub`/`adapter`, `port` |
| `app` | Default module set, `Build()` | `di`, all `*/module.go` packages, `config`, `server` |
| `cmd/gateway` | Bootstrap only | `app`, `config`, `server` |

Ports and domain remain free of `di` imports.

`internal/deps` holds the shared `Gateway` struct so `di` and `server` do not import each other (avoids an import cycle while `server/module.go` registers on the builder).

## Alternatives

### 1. Google Wire (compile-time codegen)

**Pros:** Compile-time graph validation; no runtime overhead; popular in Go ecosystem.

**Cons:** Adds `wire` as a build dependency and `//go:generate` step; codegen indirection conflicts with "explicit code over magic"; graph is small (~7 deps) and stable for M0.

**Decision:** Rejected for M0. Revisit if the graph exceeds ~20 providers or cyclic dependencies become painful.

### 2. Uber Fx (runtime DI framework)

**Pros:** Lifecycle hooks, module grouping, rich ecosystem.

**Cons:** External dependency (requires ADR); reflection-based; encourages framework-specific patterns; overkill for a stateless gateway with a flat port graph.

**Decision:** Rejected. Violates "no dependency without need" and adds runtime magic.

### 3. Manual container (chosen)

**Pros:** Stdlib only; fully explicit graph; easy stub swap in tests; each module owns its provider; no codegen or reflection.

**Cons:** Manual `Provide*` methods and validation must stay in sync as the graph grows.

**Decision:** Accepted. Best fit for current graph size, project principles, and M0 scope.

## Risks

- **Provider drift** — new port added without `Provide*` and validation. Mitigated by `Builder.validate()` and wiring tests.
- **Module order confusion** — last-wins override semantics must be documented. Mitigated by spec and test demonstrating override.
- **app package as god-wiring** — `DefaultModules()` could grow large. Mitigated by one line per module; alternative Fx-style grouping deferred.

## Migration

1. Add `internal/di` and per-module `module.go` files.
2. Change `app.New(cfg)` → `app.Build(cfg)` returning `(server.Dependencies, error)`.
3. Update `main.go` to handle build error.
4. Remove hardcoded stub imports from `app.go`.
5. Update ARCHITECTURE.md dependency flow.

No breaking changes to HTTP API or config env vars.

## Open Questions

- None for M0. Wire reconsideration tracked as future ADR if graph complexity warrants it.
