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
cmd/gateway/          Entry point
internal/config/      Configuration port
internal/server/      HTTP adapter
docs/                 Product and engineering docs
deploy/               Deployment artifacts (future)
```

## Target Modules

| Module | Responsibility |
|---|---|
| `auth` | JWT, OAuth2, API keys, mTLS |
| `routing` | Reverse proxy, path/host routing |
| `ratelimit` | Token bucket, sliding window |
| `middleware` | Request/response pipeline |
| `observability` | Logs, metrics, traces |

## Design Rules

- One responsibility per package
- Small interfaces (ports)
- Constructor injection, no global state
- Adapters live at the edges; core stays pure
