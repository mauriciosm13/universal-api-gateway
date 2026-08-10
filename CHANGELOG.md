# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

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
