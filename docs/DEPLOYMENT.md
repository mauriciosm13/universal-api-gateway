# Deployment

## Supported Targets

- Docker
- Docker Compose
- Kubernetes / Helm (planned)
- AWS, GCP, Azure (planned)

## Docker

```bash
docker build -t universal-api-gateway .
docker run --rm -p 8080:8080 universal-api-gateway
```

## Docker Compose

```bash
docker compose up --build
```

## Environment

All configuration is injected via environment variables. See [README](../README.md#configuration).

## Cloud Agnostic Rule

No cloud-specific SDK in core modules. Cloud integrations are adapters behind ports.
