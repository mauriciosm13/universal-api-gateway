# Specification — Kubernetes Base Manifests (M0)

## Overview

Cloud-agnostic Kubernetes deployment for universal-api-gateway using Kustomize. Base manifests define Deployment, Service, and ConfigMap; overlays adjust replicas, resources, and image tags per environment. Namespace is fixed to `gateway`.

## Responsibilities

| Path | Responsibility |
|---|---|
| `deploy/kubernetes/base/` | Shared Deployment, Service, ConfigMap, namespace |
| `deploy/kubernetes/overlays/dev/` | 1 replica, minimum resources, `:latest` |
| `deploy/kubernetes/overlays/staging/` | 2 replicas, base resources, `:staging` |
| `deploy/kubernetes/overlays/prod/` | 3 replicas, increased resources, versioned tag |

## API

Not applicable — Kubernetes YAML resources, no Go API.

### Resource summary

| Resource | Name | Key settings |
|---|---|---|
| Namespace | `gateway` | Fixed for all overlays |
| ConfigMap | `gateway-config` | Gateway env vars |
| Deployment | `universal-api-gateway` | Probes, non-root, port 8080 |
| Service | `universal-api-gateway` | ClusterIP :8080 |

### ConfigMap keys

| Key | Value | Maps to |
|---|---|---|
| `GATEWAY_HOST` | `0.0.0.0` | `config.defaultHost` |
| `GATEWAY_PORT` | `8080` | `config.defaultPort` |
| `GATEWAY_READ_TIMEOUT` | `15s` | `config.defaultReadTimeout` |
| `GATEWAY_WRITE_TIMEOUT` | `15s` | `config.defaultWriteTimeout` |
| `GATEWAY_IDLE_TIMEOUT` | `60s` | `config.defaultIdleTimeout` |

### Probes

| Probe | Path | Port | initialDelay | period |
|---|---|---|---|---|
| Liveness | `/health/live` | 8080 | 5s | 10s |
| Readiness | `/health/ready` | 8080 | 3s | 5s |

## Data Flow

1. Operator builds container image (`docker build` or CI).
2. `kubectl kustomize deploy/kubernetes/overlays/<env>` renders manifests.
3. `kubectl apply -k deploy/kubernetes/overlays/<env>` creates namespace, ConfigMap, Deployment, Service.
4. Pod loads env from ConfigMap; gateway listens on 8080.
5. kubelet probes `/health/live` and `/health/ready`.
6. In-cluster clients reach gateway via Service `universal-api-gateway.gateway.svc:8080`.

## Errors

- Missing ConfigMap prevents pod env injection — Deployment references existing ConfigMap name.
- Failed readiness probe removes pod from Service endpoints until `/health/ready` returns 200.

## Metrics

Not applicable for M0 manifests.

## Logs

Container stdout (JSON slog from gateway process); cluster log collection deferred to M4/M8.

## Tracing

Not applicable for M0 manifests.

## Security

- Pod runs as non-root (matches distroless Dockerfile).
- `readOnlyRootFilesystem: true`, `allowPrivilegeEscalation: false`.
- No Secrets in M0; no sensitive values in ConfigMap.

## Performance

Overlay resource patches:

| Overlay | CPU request | Memory request | CPU limit | Memory limit |
|---|---|---|---|---|
| dev | 50m | 64Mi | 200m | 128Mi |
| staging | 50m | 64Mi | 200m | 128Mi |
| prod | 100m | 128Mi | 500m | 256Mi |

## Testing

- `kubectl kustomize deploy/kubernetes/base` succeeds.
- `kubectl kustomize` succeeds for dev, staging, prod overlays.
- CI runs the same commands on every push/PR.
- Manual smoke (optional): `kubectl apply -k overlays/dev`, port-forward, `curl /health/ready`.

## Commands

```bash
# Render manifests
kubectl kustomize deploy/kubernetes/overlays/dev

# Apply to cluster
kubectl apply -k deploy/kubernetes/overlays/dev

# Port-forward and test
kubectl port-forward -n gateway svc/universal-api-gateway 8080:8080
curl http://localhost:8080/health/ready
```
