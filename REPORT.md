# Conventions Report — universal-api-gateway

Companion to `CONVENTIONS.md`. Written from the repository at release `v0.2.0-mvp` (`main`, clean tree).

## 1. What this repository is (and is not)

A cloud-agnostic HTTP API gateway in Go 1.25, built with hexagonal architecture, that does path/header/host/method routing, reverse proxying with round-robin load balancing, JWT and API-key authentication, in-memory token-bucket rate limiting, and upstream timeout/retry. It runs as a long-lived server or on AWS Lambda through the Lambda Web Adapter.

It is not a full API platform yet. There is no distributed rate limiting, no OAuth2/OIDC/mTLS/RBAC, no circuit breaker, no YAML configuration (environment variables only), and HTTP handlers are not instrumented with tracing yet (`.ai/memory/known-limitations.md`). It is also not a library: everything lives under `internal/`.

## 2. How code is organized and where a new feature enters

- `cmd/gateway/main.go` only bootstraps: `config.Load()`, telemetry, `app.Build(cfg)`, then `server.New(deps)`.
- `internal/app/app.go` lists the DI modules (`DefaultModules()`); `internal/di/builder.go` holds `Provide*` methods and validates the graph; `internal/deps/deps.go` is the struct handed to the HTTP edge.
- Each bounded context (`auth`, `routing`, `ratelimit`, `middleware`, `observability`, plus `reliability`) has `port/` (small interfaces), `stub/` (no-op defaults), `adapter/<name>/` (real implementations) and `module.go` (picks stub or adapter from config).
- `internal/config/` turns `GATEWAY_*` environment variables into typed structs, one file per concern.
- `internal/server/` is the `net/http` edge; routes are registered in `internal/server/routes.go`.

A new feature starts with docs (RFC in `docs/rfcs/`, an ADR in `docs/adrs/` if needed, spec in `docs/specs/`). Then it gets a config file in `internal/config/`, an adapter under the owning module (or a new port + stub if it is a new capability), wiring in that module's `module.go`, and tests. It ends with a `CHANGELOG.md` entry. Request-path behavior goes into the middleware pipeline (`internal/middleware/adapter/request/`).

## 3. How a change is proven

The single merge gate is `make quality`, which runs `quality/scripts/quality-gate.sh`; CI (`.github/workflows/ci.yml`) runs the same script. Hard gates:

- formatting: `gofmt -l .` must be empty (`make fmt` fixes it)
- static analysis: `go vet ./...` (no golangci-lint is configured, so vet plus compilation is the lint and type check)
- unit tests: `go test -race -count=1 -covermode=atomic` on every package that has tests
- coverage, via `quality/covercheck`: at least 85% of `internal/` statements, no regression against `quality/baselines/quality-baseline.json` (currently 90.0%), at least 90% of changed lines on PRs, and every executable `internal/` package must have tests

I ran the gate on `main` at `v0.2.0-mvp` using a local Go 1.25.0 toolchain (put on `PATH`, `GOTOOLCHAIN=local`). It finished with `QUALITY GATE: PASSED (Phase 1)` in about 7 seconds:

- Real, enforcing steps: [1] formatting PASS, [2] `go vet` PASS, [3] tests and coverage PASS. Global coverage was 90.0% against a minimum of 85%, and it matched the 90.0% baseline. Changed-code coverage reported SKIP because no `internal/` file had changed.
- Placeholders that only print a message and never fail: [4] mutation, [5] integration, [6] regression, [7] E2E, and [9] security are `TODO`. [8] dependencies only writes a bootstrap inventory. [10] complexity/module size and [11] API compatibility/performance are "Bootstrap placeholder" messages.

Faster local loop: `make test`, `make vet`, `make fmt-check`, `make build`. Opt-in integration and E2E: `bash quality/scripts/e2e-bootstrap.sh` (needs Docker), with tests behind `//go:build integration`.

Representative existing tests: `internal/routing/adapter/path/router_test.go` (adapter unit test), `internal/ratelimit/module_register_test.go` (DI wiring), `internal/server/gateway_test.go` (HTTP edge), `quality/e2e/bootstrap_integration_test.go` (integration).

## 4. Commit and PR practice

The practice is written down in `docs/GIT_WORKFLOW.md`. Commits follow Conventional Commits, enforced by `.github/workflows/conventional-commits.yml` using `.commitlintrc.yaml` (header at most 100 characters). The allowed types are `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `build`, `ci`, `chore`, and `revert`, and a scope is recommended. Work happens on feature branches (the history shows `feat/mvp-week5-rate-limiting`, `chore/cursor-skills`, and similar). PR titles use the same format because squash merges use the PR title as the commit message. Every commit in a PR is also linted. The review checklists are `docs/REVIEW_CHECKLIST.md` and `.ai/checklists/pull-request.md`.

## 5. Three patterns copied from existing files

1. A module provider picks the stub when the feature is disabled and the adapter when it is enabled, and panics on config errors during wiring. See `internal/ratelimit/module.go` (`Register` → `buildLimiter(b.Config())` → `b.ProvideLimiter`).
2. Each port is a small interface with one method, paired with a no-op stub that has a `New*` constructor. See `internal/ratelimit/port/limiter.go` (`Allow(ctx, key) (bool, error)`) and `internal/ratelimit/stub/noop.go` (`NewNoOpLimiter`).
3. Tests use only the stdlib `testing` package, live in the same package, call `t.Parallel()`, are table-driven, and use `X = %v, want %v` failure messages. See `internal/routing/adapter/path/router_test.go` and `internal/ratelimit/module_test.go`.

## 6. Three forbidden things

1. Bypassing ports or reversing layer direction is forbidden. That covers handlers calling infrastructure directly, ports importing stubs or adapters, and `domain` importing `net/http` or `os`. The rules are in `docs/ANTI_PATTERNS.md` and the layer table in `docs/ARCHITECTURE.md`. `internal/ratelimit/port/limiter.go`, which imports only `context`, shows the expected shape.
2. Adding a dependency without an ADR is forbidden (`AGENTS.md`). As a precedent, the `github.com/golang-jwt/jwt/v5` entry in `go.mod` is backed by `docs/adrs/0002-jwt-validation-library.md`.
3. Weakening the quality gate to make it pass is forbidden. That includes lowering `quality/config/coverage.yaml`, editing `quality/baselines/quality-baseline.json`, deleting tests, weakening assertions, or disabling checks (`quality/AI_REMEDIATION.md`). After 3 failed repair attempts, stop and ask for human review.

## 7. What was ambiguous (the task proceeded anyway)

- **There is no Go on the default `PATH`.** On this machine `make quality` fails at step 1 with `gofmt: command not found` unless a Go 1.25 toolchain is put on `PATH` first. Nothing in the repo pins or installs the toolchain locally. Only `.devcontainer/` and CI's `go-version-file: go.mod` do that.
- **Running the gate dirties the tree.** `quality/reports/latest.md` and `latest.json` are tracked in git even though `.gitignore` lists `quality/reports/`. In the observed run, the gate rewrote both files, but only `latest.md` changed: one table row went from `Changed-code coverage | PASS` to `SKIP` because no `internal/` file was changed. `latest.json` came out byte-identical. The committed report therefore records an older run, from a branch with changed files. No doc says whether report updates should be committed. For this task the file was restored with `git checkout -- quality/reports/`, and `CONVENTIONS.md` says not to commit that churn.
- **The architecture doc is stale.** `docs/ARCHITECTURE.md` still says "Current Structure (v0.1)", omits `internal/reliability/`, `internal/auth/adapter/`, `internal/routing/adapter/roundrobin/`, and `internal/middleware/adapter/request/`, and has no "Adapters" row in the import table. The Layers table in `CONVENTIONS.md` is no longer inferred: it now records the real import graph from `go list` (follow-up recon). `.ai/context/architecture.md` is even older (describes only health endpoints). `docs/ARCHITECTURE.md` itself was not edited.
- **The HTTP edge imports two adapters directly.** `internal/server/gateway.go` imports `internal/routing/adapter/roundrobin` and `internal/server/response.go` imports `internal/middleware/adapter/request` (for `ContinueStatus`). `docs/ANTI_PATTERNS.md` forbids "bypassing ports to call infrastructure directly from handlers", but no doc says whether a load-balancing selector or a pipeline status constant counts as infrastructure. It is recorded as an open exception, not as a rule or a precedent.
- **Layer boundaries the code leaves open** (from the same `go list` graph):
  - Cross-module imports in adapters: `internal/middleware/adapter/request/auth.go` and `ratelimit.go` import `auth/port`, `auth/context`, `ratelimit/port` and `ratelimit/context`. No doc says whether an adapter may depend on another module's port.
  - Context packages (`internal/auth/context`, `internal/ratelimit/context`) are imported by adapters, by `internal/server` and by `internal/ratelimit/adapter/tokenbucket`. They could belong to the port layer or be a layer of their own.
  - `internal/reliability/transport.go` wires its own adapters from `config` without `di` or a `module.go`, and only `internal/server/gateway.go` uses it. It could be a module provider or a server-only helper.
  - `internal/observability/otelsetup` is imported by `cmd/gateway/main.go`, outside the DI graph.
  - Importing the matching port is not uniform: `internal/ratelimit/stub/noop.go`, `internal/middleware/stub/passthrough.go` and `internal/ratelimit/adapter/tokenbucket/limiter.go` satisfy their ports structurally, without importing them.
  - `internal/config` reads `OTEL_*` (`internal/config/config.go`) and `AWS_LAMBDA_FUNCTION_TIMEOUT` (`internal/config/runtime.go`) alongside `GATEWAY_*`. Whether these are deliberate exceptions to "`GATEWAY_*` only" is not written anywhere.
- **The spec path is unclear.** `AGENTS.md` says the spec goes "in `docs/templates/SPEC.md` format", but the real specs live in `docs/specs/`. The conventions file treats `docs/specs/` as the location and the template as the format.
- **There are two places for anti-patterns.** `docs/ANTI_PATTERNS.md` and `.ai/memory/anti-patterns.md` overlap but differ: one forbids bypassing ports, the other forbids cloud SDKs in core. Both lists were merged.
- **`helpers` naming is inconsistent.** The `helpers.go` catch-all is forbidden, yet `internal/server/helpers_test.go` exists. It reads as an accepted test-fixture exception, not a precedent for production code.
- **Package-level error sentinels are allowed.** "No global state" was read as no *mutable* globals, since sentinels like `ErrNoRoute` in `internal/routing/port/router.go` are an established pattern.
- **Go version decision.** `.ai/memory/project-decisions.md` lists "Go as implementation language" with "Pending ADR". It is treated as settled.

## 8. What CONVENTIONS.md still does not cover

- Observability conventions (metric names, span names, log fields). `docs/OBSERVABILITY.md` was not distilled, and handlers are not yet instrumented.
- Benchmarks. `docs/TESTING.md` asks for them on hot paths, but there is no `make bench` target and no benchmark gate yet (`quality/config/performance.yaml` is a placeholder).
- Security review rules (`docs/SECURITY.md`) and the planned Phase 2–6 gates (mutation, SAST, complexity, API compatibility). In the observed run these only printed TODO or bootstrap placeholder messages, so no convention can rest on them yet.
- Error-response conventions for new endpoints (JSON errors in `internal/server/json_error.go`, `docs/specs/json-error-responses.md`) beyond "use the existing helper".
- Deployment changes: Kubernetes overlays (`make k8s-validate`), `Dockerfile.lambda`, and `deploy/aws/`.
- Release and versioning steps beyond "SemVer + Keep a Changelog" (no release workflow file was found).
- Agent workflow artifacts: how `CONTEXT.md` and spec-driven tasks interact with the RFC/ADR/Spec flow.
