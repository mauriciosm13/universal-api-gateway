# Deployment

## Supported Targets

- Docker
- Docker Compose
- Kubernetes (Kustomize base + overlays)
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

## Kubernetes

Manifests live under `deploy/kubernetes/`. Namespace is fixed to `gateway`.

### Layout

```text
deploy/kubernetes/
  base/                 Deployment, Service, ConfigMap, Namespace
  overlays/
    dev/                1 replica, :latest
    staging/            2 replicas, :staging
    prod/               3 replicas, :0.1.0
```

### Prerequisites

- `kubectl` configured for your cluster
- Container image available to the cluster (build locally or push to a registry)

```bash
docker build -t universal-api-gateway:latest .
# For minikube/kind: load the image into the cluster
```

### Render manifests

```bash
kubectl kustomize deploy/kubernetes/overlays/dev
```

### Apply

```bash
kubectl apply -k deploy/kubernetes/overlays/dev
```

### Verify

```bash
kubectl get pods -n gateway
kubectl port-forward -n gateway svc/universal-api-gateway 8080:8080
curl http://localhost:8080/health/ready
```

### Overlays

| Overlay | Replicas | Image tag | Use case |
|---|---|---|---|
| `dev` | 1 | `:latest` | Local / development cluster |
| `staging` | 2 | `:staging` | Pre-production |
| `prod` | 3 | `:0.1.0` | Production (pin semver in overlay) |

Only one overlay should run per cluster — all use namespace `gateway`.

See [RFC 0003](rfcs/0003-kubernetes-base-manifests.md) and [Spec](specs/kubernetes-base-manifests.md) for details.

## Environment

All configuration is injected via environment variables. In Kubernetes, non-secret settings are provided by ConfigMap `gateway-config`. See [README](../README.md#configuration).

## Cloud Agnostic Rule

No cloud-specific SDK in core modules. Cloud integrations are adapters behind ports. Kubernetes manifests use standard resources without cloud-provider annotations.
