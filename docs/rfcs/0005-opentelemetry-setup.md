# RFC 0005 — OpenTelemetry Setup (M0)

## Summary

Bootstrap OpenTelemetry tracing for the gateway: SDK initialization in `main`, `port.Tracer` adapter, and environment-based configuration. Tracing is **disabled by default**; enabling requires explicit env vars. Logger metrics and full request instrumentation remain M4.

## Motivation

M0 lists OpenTelemetry setup as a foundation deliverable. The Project Bible mandates OpenTelemetry First. Stubs proved the hexagonal boundary; M0 wires a real TracerProvider without instrumenting the full proxy pipeline (M1/M4).

## Goals

- Install and shut down OTel TracerProvider from `cmd/gateway/main.go`
- `observability/adapter` implements `port.Tracer` via OTel
- Configuration via standard `OTEL_*` environment variables
- Default off (`OTEL_SDK_DISABLED=true`) — zero surprise for Docker/K8s deploys
- ADR 0001 for Go SDK dependencies
- Tests for adapter and config parsing

## Non Goals

- Prometheus metrics (M4)
- Structured logs bridged to OTel (M4)
- Span per proxied request (M1 pipeline + M4)
- Auto-instrumentation of `net/http`

## Detailed Design

### Configuration

| Variable | Default | Purpose |
|---|---|---|
| `OTEL_SDK_DISABLED` | `true` | When true, use no-op tracer |
| `OTEL_SERVICE_NAME` | `universal-api-gateway` | Service resource attribute |
| `OTEL_TRACES_EXPORTER` | `stdout` | `stdout`, `otlp`, or `none` |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | — | OTLP/HTTP endpoint when exporter is `otlp` |

### Bootstrap flow

```text
main
  └── config.Load()               includes TelemetryConfig
  └── otelsetup.Install(ctx, cfg)  sets global TracerProvider, returns shutdown
  └── app.Build(cfg)               observability module picks OTel or noop tracer
  └── server lifecycle
  └── shutdown(ctx)                flush TracerProvider
```

### Packages

| Package | Role |
|---|---|
| `internal/config` | `TelemetryConfig` loaded from env |
| `internal/observability/otelsetup` | TracerProvider install/shutdown |
| `internal/observability/adapter` | `port.Tracer` OTel implementation |
| `internal/observability/module` | Registers adapter or stub based on config |

### Exporters

- `stdout` — development, JSON/text spans to stdout
- `otlp` — OTLP/HTTP to collector (Jaeger, Tempo, etc.)
- `none` — provider without exporter (sampling tests)

When `OTEL_SDK_DISABLED=true`, skip Install; module registers noop stub.

## Alternatives

See ADR 0001.

## Risks

- Forgetting shutdown loses spans — mitigated by deferred shutdown in `main`.
- OTLP misconfiguration at startup — fail fast with clear error log.

## Migration

No breaking HTTP or config changes. Existing deployments without OTEL vars keep noop tracing.

## Open Questions

- None for M0.
