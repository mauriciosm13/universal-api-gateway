# Conventions

Short house rules for agents working in this repo. Sources of truth: `AGENTS.md`, `docs/ARCHITECTURE.md`, `docs/TESTING.md`, `docs/GIT_WORKFLOW.md`, `quality/AI_REMEDIATION.md`. If this file and those disagree, those win.

## Layers (hexagonal, Go 1.25, module `github.com/mauriciomendonca/universal-api-gateway`)

Internal imports below are what the code does at v0.2.0-mvp, verified with `go list -f '{{.ImportPath}}: {{join .Imports " "}}' ./internal/... ./cmd/...`.

| Layer | Where | Imports (observed) |
|---|---|---|
| Domain | `internal/domain` | nothing (`internal/domain/request.go`) |
| Ports | `internal/<module>/port` | `context`, `errors`, `domain` only (`internal/routing/port/router.go`) |
| Stubs | `internal/<module>/stub` | stdlib, `domain`, usually the matching `port`; `internal/ratelimit/stub/noop.go` and `internal/middleware/stub/passthrough.go` satisfy their port structurally without importing it |
| Adapters | `internal/<module>/adapter/<name>` (`internal/observability/adapter/tracer.go` has no `<name>`) | matching `port`, `domain`, `config`, stdlib incl. `net/http`; same-module helpers (`internal/auth/credential` from `internal/auth/adapter/request/authenticator.go`); other modules' ports and context packages (`internal/middleware/adapter/request/auth.go`, `ratelimit.go`); third-party SDKs only with an ADR (`internal/auth/adapter/jwt`, `internal/observability/adapter`) |
| Module provider | `internal/<module>/module.go` | `di` and the module's own adapters/stub; usually its `port` and `config` (`internal/ratelimit/module.go`); `internal/middleware/module.go` imports only `di` + its adapter |
| Context / helpers | `internal/auth/context`, `internal/ratelimit/context`, `internal/auth/credential` | own `port` or `domain` and stdlib only (`internal/auth/context/identity.go`) |
| DI / shared deps | `internal/di`, `internal/deps` | all module ports, `config`; `di` also imports `deps` (`internal/di/builder.go`) |
| Composition | `internal/app` | `di`, module packages, `config`, `server` (`internal/app/app.go`) |
| HTTP edge | `internal/server` | `net/http` family, `deps`, `di`, `domain`, `routing/port`, `auth/context`, `internal/reliability`; config only via `deps.Gateway.Config` (`internal/server/server.go`). Also two adapters directly, see below |
| Reliability | `internal/reliability` | `config` + its own `adapter/retry`, `adapter/timeout`; no `module.go`, used only by `internal/server/gateway.go` |
| Config | `internal/config` | stdlib only; reads `GATEWAY_*` plus `OTEL_*` (`internal/config/config.go`) and `AWS_LAMBDA_FUNCTION_TIMEOUT` (`internal/config/runtime.go`) |
| Entry | `cmd/gateway` | `app`, `config`, `server`, `observability/otelsetup` (`cmd/gateway/main.go`) |

Ports never import stubs or adapters, and no production adapter imports another adapter (both verified). Infrastructure depends inward on ports.

Open exception, not a precedent: `internal/server/gateway.go` imports `internal/routing/adapter/roundrobin` and `internal/server/response.go` imports `internal/middleware/adapter/request`. `docs/ANTI_PATTERNS.md` forbids handlers calling infrastructure directly; whether these two imports count is unresolved (`REPORT.md` section 7).

## Where a new feature goes

1. Docs first: RFC in `docs/rfcs/`, ADR in `docs/adrs/` if it adds a dependency or a hard decision, spec in `docs/specs/` using `docs/templates/SPEC.md`.
2. Config: new `internal/config/<feature>.go` with a typed struct and a `load<Feature>()` reading `GATEWAY_*` env vars, plus `<feature>_test.go`.
3. Behavior: a new adapter under the owning module, e.g. `internal/ratelimit/adapter/<name>/`, implementing the existing port. New capability = new small port in `internal/<module>/port/` + no-op stub in `internal/<module>/stub/`.
4. Wiring: select stub vs adapter in `internal/<module>/module.go`. A new port also needs a `Provide*` method in `internal/di/builder.go` and a field in `internal/deps/deps.go`. New modules are listed in `DefaultModules()` in `internal/app/app.go`.
5. HTTP: request-path behavior goes in the middleware pipeline (`internal/middleware/adapter/request/`); new edge endpoints register in `internal/server/routes.go`.
6. Finish with tests, docs (`docs/ARCHITECTURE.md` if structure changes), and an `[Unreleased]` entry in `CHANGELOG.md`.

## How a change is proven

- Merge gate (same as CI `.github/workflows/ci.yml`): `make quality` (runs `quality/scripts/quality-gate.sh`).
  - `gofmt -l .` must be empty (fix with `make fmt`).
  - `go vet ./...` must be clean. There is no golangci-lint; vet + compile is the static/type check.
  - `go test -race -count=1 -covermode=atomic` over packages with tests.
  - Coverage of `internal/` ≥ 85% global, no drop vs `quality/baselines/quality-baseline.json`, ≥ 90% on changed lines in PRs, and every executable `internal/` package must have tests.
- Fast loop: `make test` (`go test -race -count=1 ./...`), `make vet`, `make fmt-check`, `make build`.
- Integration (opt-in, needs Docker): `bash quality/scripts/e2e-bootstrap.sh`; tests use `//go:build integration` (see `quality/e2e/bootstrap_integration_test.go`).
- Every bug fix ships a permanent regression test.
- Only steps 1–3 of the gate (format, vet, tests + coverage) enforce anything today; steps 4–11 print TODO/bootstrap placeholders. Verified on `main` at v0.2.0-mvp: `QUALITY GATE: PASSED (Phase 1)`, global coverage 90.0%. Changed-code coverage reports SKIP when no `internal/` file changed.
- If `go`/`gofmt` are not on `PATH`, put a Go 1.25 toolchain first on `PATH` and set `GOTOOLCHAIN=local`. Without Go, the gate fails at step 1 with `gofmt: command not found`.
- The gate rewrites the tracked `quality/reports/latest.md` (and `latest.json`). Do not commit that churn unless the task is about reports; restore with `git checkout -- quality/reports/`.
- On gate failure: read `quality/reports/latest.md`, fix root cause with the smallest change, rerun. Max 3 attempts, then stop for human review.

## Commits and PRs

- Conventional Commits, enforced by `.github/workflows/conventional-commits.yml` and `.commitlintrc.yaml` (header ≤ 100 chars). Types: `feat fix docs style refactor perf test build ci chore revert`. Scope recommended, e.g. `feat(ratelimit): ...`.
- Feature branches (`feat/...`, `chore/...`), PR title uses the same format (squash merge uses it), every commit in the PR is also linted.
- PR checklist: `docs/REVIEW_CHECKLIST.md`, `.ai/checklists/pull-request.md`.

## Patterns to copy

1. **Module provider chooses stub vs adapter from config; config errors panic at wiring.** `internal/ratelimit/module.go`:
   ```go
   func (Module) Register(b *di.Builder) {
   	limiter, err := buildLimiter(b.Config())
   	if err != nil {
   		panic(fmt.Sprintf("ratelimit: %v", err))
   	}
   	b.ProvideLimiter(limiter)
   }
   ```
2. **One-method port + no-op stub with a `New*` constructor.** `internal/ratelimit/port/limiter.go` and `internal/ratelimit/stub/noop.go`:
   ```go
   type Limiter interface {
   	Allow(ctx context.Context, key string) (bool, error)
   }
   func NewNoOpLimiter() *NoOpLimiter { return &NoOpLimiter{} }
   ```
3. **Tests: same package, stdlib `testing` only, `t.Parallel()`, table-driven, `got = %v, want %v` messages.** `internal/routing/adapter/path/router_test.go`, `internal/ratelimit/module_test.go`:
   ```go
   t.Parallel()
   ...
   if route.ID != tt.wantID {
   	t.Fatalf("Resolve(%q) ID = %q, want %q", tt.path, route.ID, tt.wantID)
   }
   ```

## Forbidden

1. **Bypassing ports or breaking layer direction** (handlers calling infrastructure directly, ports importing stubs/adapters, I/O in `domain`). Rule: `docs/ANTI_PATTERNS.md`, `docs/ARCHITECTURE.md`. Reference shape: `internal/ratelimit/port/limiter.go` imports only `context`.
2. **Adding a Go dependency without an ADR.** Rule: `AGENTS.md`. Precedent: `github.com/golang-jwt/jwt/v5` in `go.mod` is justified by `docs/adrs/0002-jwt-validation-library.md`.
3. **Weakening the gate to pass**: lowering `quality/config/coverage.yaml`, editing `quality/baselines/quality-baseline.json`, deleting tests, weakening assertions, disabling checks. Rule: `quality/AI_REMEDIATION.md`.

Also avoid: `utils.go`/`helpers.go` catch-alls, global mutable state, giant interfaces, cloud SDKs in core modules (`docs/ANTI_PATTERNS.md`, `.ai/memory/anti-patterns.md`).
