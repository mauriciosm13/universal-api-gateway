# Observability

- Structured JSON logs (process lifecycle via `slog` in `main`)
- OpenTelemetry tracing bootstrap (M0) — disabled by default
- Prometheus metrics (planned — M4)
- Health checks: liveness and readiness
- Correlation / request IDs (planned — M4)

## OpenTelemetry (M0)

Tracing is configured via standard `OTEL_*` environment variables. Disabled by default (`OTEL_SDK_DISABLED=true`).

| Variable | Default | Description |
|---|---|---|
| `OTEL_SDK_DISABLED` | `true` | Set to `false` to enable tracing |
| `OTEL_SERVICE_NAME` | `universal-api-gateway` | Service name resource attribute |
| `OTEL_TRACES_EXPORTER` | `stdout` | `stdout`, `otlp`, or `none` |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | — | Required when exporter is `otlp` (e.g. `http://localhost:4318`) |

### Local example

```bash
OTEL_SDK_DISABLED=false OTEL_TRACES_EXPORTER=stdout make run
```

See [RFC 0005](rfcs/0005-opentelemetry-setup.md), [ADR 0001](adrs/0001-opentelemetry-go-sdk.md), and [Spec](specs/opentelemetry-setup.md).
