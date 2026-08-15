# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Round-robin load balancing across comma-separated upstream URLs per route (`GATEWAY_DEFAULT_UPSTREAM`, path/host/header/method routes)
- In-memory per-process upstream selection via `internal/routing/adapter/roundrobin`

### Added

- Upstream timeout and idempotent retry via `GATEWAY_UPSTREAM_TIMEOUT`, `GATEWAY_RETRY_MAX`, and `GATEWAY_RETRY_BACKOFF`
- JSON 502/504 for upstream failures; `internal/reliability/` transport chain
- `GATEWAY_RUNTIME=lambda|server` for AWS Lambda Web Adapter hosting
- MVP execution plan, AWS Lambda deploy design (RFC MVP-5/6, ADRs 0003–0005, specs)
- `Dockerfile.lambda`, `deploy/aws/scripts/`, `docker-compose.demo.yml`, `scripts/demo.sh`

### Added

- Token bucket rate limiting via `GATEWAY_RATE_LIMIT_RPS`, `GATEWAY_RATE_LIMIT_BURST`, and optional `GATEWAY_RATE_LIMIT_ROUTES`
- Rate limit key from authenticated identity subject with host+path fallback
- RFC MVP-4 and rate limiting specification

### Added

- API key authentication via `GATEWAY_API_KEYS` (header or query)
- Composite JWT + API key auth when both configured
- Identity in request context and upstream `X-User-Id` propagation
- RFC MVP-3 and API keys / identity specification

### Changed

- Gateway-generated 4xx/5xx error bodies are JSON (`Content-Type: application/json`) instead of plain text

### Added

- JWT authentication via `GATEWAY_JWT_JWKS_URL` (RS256/ES256) or `GATEWAY_JWT_HMAC_SECRET` (HS256)
- Optional JWT issuer and audience validation (`GATEWAY_JWT_ISSUER`, `GATEWAY_JWT_AUDIENCE`)
- RFC MVP-2, ADR 0002, and JWT authentication specification
- Host routing via `GATEWAY_HOST_ROUTES` and method routing via `GATEWAY_METHOD_ROUTES`
- Request middleware pipeline with auth and rate-limit hooks (no-op passthrough until M2/M3)
- RFC MVP-1 and middleware pipeline specification
- Header-based routing via `GATEWAY_HEADER_ROUTES` with router chain precedence (RFC 0009)
- JSON error responses for gateway-generated 4xx/5xx errors
- Chain routing adapter (`internal/routing/adapter/chain`) and header adapter
- RFC 0009 and specifications for header routing and JSON errors
- Prefix-based path routing via `GATEWAY_PATH_ROUTES` with optional default upstream fallback (RFC 0008)
- Centralized HTTP route registration in `internal/server/routes.go`
- Path routing adapter (`internal/routing/adapter/path`)
- RFC 0008 and specification for M1 path routing
- HTTP reverse proxy with default upstream via `GATEWAY_DEFAULT_UPSTREAM` (RFC 0007)
- Static routing adapter (`internal/routing/adapter/static`) for single-upstream proxying
- RFC 0007 and specification for M1 reverse proxy
- Engineering Intelligence quality layer (Phase 1): `quality/` scripts, coverage gate, CI integration, RFC 0006
- Makefile and Dev Container (`.devcontainer/`) for local development workflow
- OpenTelemetry tracing bootstrap with stdout/OTLP exporters (disabled by default)
- ADR 0001, RFC 0005, and specification for M0 OpenTelemetry setup
- Go toolchain bumped to 1.25 (required by OTel SDK dependencies)
- Conventional Commits enforcement in CI (commit messages and PR titles)
- RFC 0004 for commit message validation workflow
- Kubernetes base manifests with Kustomize overlays (dev, staging, prod) under `deploy/kubernetes/`
- RFC 0003 and specification for M0 Kubernetes deployment
- Manual dependency injection container (`internal/di`) with per-module providers and validated `Build()` graph
- RFC 0002 and specification for M0 dependency injection
- Hexagonal architecture scaffold (ports, stubs, composition root) for auth, routing, ratelimit, middleware, and observability modules
- RFC 0001 and specification for M0 package boundaries
- Project foundation (v0.1)
- Health endpoints: `/health`, `/health/live`, `/health/ready`
- Environment-based configuration
- Docker and Docker Compose support
- Unified engineering documentation
- Product roadmap with milestone tracking
- GitHub Actions CI (test, vet, build) and Docker build workflow
