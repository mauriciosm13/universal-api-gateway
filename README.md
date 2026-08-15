# universal-api-gateway

Cloud-agnostic API Gateway for authentication, routing, and rate limiting.

**Status: v0.1 — foundation in progress**

This project is under active development. The current release focuses on project foundation: health endpoints, configuration, Docker, and engineering documentation.

## What works today

- HTTP server with graceful shutdown
- Health endpoints: `/health`, `/health/live`, `/health/ready`
- JWT validation endpoint: `GET /auth/validate` (runs auth pipeline; returns identity JSON)
- HTTP reverse proxy with default, path, and header routing
- JSON error responses for gateway-generated 4xx/5xx errors
- JWT authentication (HS256 HMAC or RS256/ES256 JWKS) when configured via env
- API key authentication (header or query) when configured via env
- Composite auth when JWT and API keys both configured
- Upstream identity header `X-User-Id` from authenticated subject
- Environment-based configuration
- Docker and Docker Compose
- Kubernetes manifests (Kustomize base + dev/staging/prod overlays)
- Makefile and Dev Container for local development
- OpenTelemetry tracing bootstrap (disabled by default)
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

### Kubernetes

```bash
docker build -t universal-api-gateway:latest .
kubectl apply -k deploy/kubernetes/overlays/dev
kubectl port-forward -n gateway svc/universal-api-gateway 8080:8080
curl http://localhost:8080/health/ready
```

See [Deployment](docs/DEPLOYMENT.md) for overlay details.

### Configuration

| Variable | Default | Description |
|---|---|---|
| `GATEWAY_HOST` | `0.0.0.0` | Listen host |
| `GATEWAY_PORT` | `8080` | Listen port |
| `GATEWAY_READ_TIMEOUT` | `15s` | Read timeout |
| `GATEWAY_WRITE_TIMEOUT` | `15s` | Write timeout |
| `GATEWAY_IDLE_TIMEOUT` | `60s` | Idle timeout |
| `GATEWAY_DEFAULT_UPSTREAM` | — | Default upstream URL for reverse proxy |
| `GATEWAY_PATH_ROUTES` | — | Comma-separated path routes: `/api=http://backend:8080` |
| `GATEWAY_HEADER_ROUTES` | — | Comma-separated header routes: `X-Version=v1=http://v1:8080` |
| `GATEWAY_HOST_ROUTES` | — | Comma-separated host routes: `api.example.com=http://api:8080` |
| `GATEWAY_METHOD_ROUTES` | — | Comma-separated method routes: `GET=http://get:8080` |
| `GATEWAY_JWT_JWKS_URL` | — | JWKS URL for RS256/ES256 JWT validation (mutually exclusive with HMAC) |
| `GATEWAY_JWT_HMAC_SECRET` | — | HMAC secret for HS256 JWT validation (min 32 chars; mutually exclusive with JWKS) |
| `GATEWAY_JWT_ISSUER` | — | Optional expected JWT `iss` claim |
| `GATEWAY_JWT_AUDIENCE` | — | Optional expected JWT `aud` claim |
| `GATEWAY_API_KEYS` | — | Allowed API keys: `key1:name1,key2:name2` or JSON |
| `GATEWAY_API_KEY_HEADER` | `X-API-Key` | Header name for API key |
| `GATEWAY_API_KEY_QUERY` | `api_key` | Query param name for API key |
| `GATEWAY_RATE_LIMIT_RPS` | — | Global rate limit (tokens/s). Empty = off |
| `GATEWAY_RATE_LIMIT_BURST` | — | Global burst. Required when RPS set |
| `GATEWAY_RATE_LIMIT_ROUTES` | — | Per path prefix limits: `/api=10:20` |
| `GATEWAY_UPSTREAM_TIMEOUT` | `30s` | Upstream request timeout per attempt |
| `GATEWAY_RETRY_MAX` | `2` | Extra retry attempts for GET/HEAD/OPTIONS (`0` = off) |
| `GATEWAY_RETRY_BACKOFF` | `100ms` | Linear backoff between retries |
| `GATEWAY_RUNTIME` | `server` | `server` or `lambda` (AWS Lambda Web Adapter) |
| `OTEL_SDK_DISABLED` | `true` | Disable OpenTelemetry tracing |
| `OTEL_TRACES_EXPORTER` | `stdout` | Trace exporter when OTel enabled |

See [Observability](docs/OBSERVABILITY.md) for all `OTEL_*` variables.

## Roadmap

See [docs/ROADMAP.md](docs/ROADMAP.md) for the full product vision and milestone progress.

**MVP track (M0–M3):** see [docs/MVP_SCOPE.md](docs/MVP_SCOPE.md) for accelerated scope and 8-week backlog.

## Documentation

| Document | Description |
|---|---|
| [Project Bible](docs/PROJECT_BIBLE.md) | Vision, mission, principles |
| [Architecture](docs/ARCHITECTURE.md) | Hexagonal architecture overview |
| [Engineering Manual](docs/ENGINEERING_MANUAL.md) | Development process index |
| [Deployment](docs/DEPLOYMENT.md) | Docker, Kubernetes, cloud targets |
| [MVP Scope](docs/MVP_SCOPE.md) | Accelerated M0–M3 scope and 8-week backlog |
| [MVP Execution Plan](docs/MVP_EXECUTION_PLAN.md) | Weeks 6–8 parallel schedule + AWS Lambda |

## Development

```bash
make help     # list targets
make test     # run tests
make run      # start gateway
make build    # build bin/gateway
```

Or without Make:

```bash
go test ./...
go vet ./...
gofmt -l .
go build -o bin/gateway ./cmd/gateway
```

### Dev Container

Open the repository in VS Code / Cursor and select **Reopen in Container** (`.devcontainer/devcontainer.json` — Go 1.25, Docker-in-Docker, port 8080 forwarded).

CI runs the same checks on every push and pull request to `main`.

## License

MIT — see [LICENSE](LICENSE).
