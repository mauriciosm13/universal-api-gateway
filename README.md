# universal-api-gateway

[![CI](https://github.com/mauriciomendonca/universal-api-gateway/actions/workflows/ci.yml/badge.svg)](https://github.com/mauriciomendonca/universal-api-gateway/actions/workflows/ci.yml)
[![Docker](https://github.com/mauriciomendonca/universal-api-gateway/actions/workflows/docker.yml/badge.svg)](https://github.com/mauriciomendonca/universal-api-gateway/actions/workflows/docker.yml)

Cloud-agnostic API Gateway for authentication, routing, and rate limiting.

**Status: v0.1 — foundation in progress**

This project is under active development. The current release focuses on project foundation: health endpoints, configuration, Docker, and engineering documentation.

## What works today

- HTTP server with graceful shutdown
- Health endpoints: `/health`, `/health/live`, `/health/ready`
- Environment-based configuration
- Docker and Docker Compose
- Structured JSON logging
- GitHub Actions CI (test, vet, build + Docker build)

## Quick start

### Local

```bash
go run ./cmd/gateway
curl http://localhost:8080/health
```

### Docker

```bash
docker compose up --build
curl http://localhost:8080/health/ready
```

### Configuration

| Variable | Default | Description |
|---|---|---|
| `GATEWAY_HOST` | `0.0.0.0` | Listen host |
| `GATEWAY_PORT` | `8080` | Listen port |
| `GATEWAY_READ_TIMEOUT` | `15s` | Read timeout |
| `GATEWAY_WRITE_TIMEOUT` | `15s` | Write timeout |
| `GATEWAY_IDLE_TIMEOUT` | `60s` | Idle timeout |

## Roadmap

See [docs/ROADMAP.md](docs/ROADMAP.md) for the full product vision and milestone progress.

## Documentation

| Document | Description |
|---|---|
| [Project Bible](docs/PROJECT_BIBLE.md) | Vision, mission, principles |
| [Architecture](docs/ARCHITECTURE.md) | Hexagonal architecture overview |
| [Engineering Manual](docs/ENGINEERING_MANUAL.md) | Development process index |
| [Deployment](docs/DEPLOYMENT.md) | Docker, Kubernetes, cloud targets |

## Development

```bash
go test ./...
go vet ./...
gofmt -l .
go build -o bin/gateway ./cmd/gateway
```

CI runs the same checks on every push and pull request to `main`.

## License

MIT — see [LICENSE](LICENSE).
