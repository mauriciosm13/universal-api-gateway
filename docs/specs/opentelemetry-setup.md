# Specification — OpenTelemetry Setup (M0)

## Overview

Bootstrap OpenTelemetry tracing with environment-based configuration, a TracerProvider lifecycle in `main`, and an adapter implementing `observability/port.Tracer`. Disabled by default.

## Responsibilities

| Package | Responsibility |
|---|---|
| `internal/config` | Parse `TelemetryConfig` from `OTEL_*` env vars |
| `internal/observability/otelsetup` | Install/shutdown global TracerProvider |
| `internal/observability/adapter` | OTel-backed `port.Tracer` |
| `internal/observability/module` | Wire adapter or noop stub |
| `cmd/gateway/main.go` | Call Install before Build, shutdown on exit |

## API

### config.TelemetryConfig

```go
type TelemetryConfig struct {
    Disabled       bool
    ServiceName    string
    TracesExporter string // stdout, otlp, none
    OTLPEndpoint   string
}
```

### otelsetup

```go
func Install(ctx context.Context, cfg config.TelemetryConfig) (func(context.Context) error, error)
```

Returns shutdown function; no-op shutdown when disabled.

### adapter

```go
func NewTracer() port.Tracer
```

Uses global `otel.Tracer` after Install.

## Data Flow

1. `config.Load` parses gateway and telemetry settings.
2. If telemetry enabled, `otelsetup.Install` configures TracerProvider and exporter.
3. `app.Build` registers `adapter.NewTracer()` or noop stub.
4. Callers use `port.Tracer.StartSpan` (health handlers uninstrumented in M0).
5. On shutdown, TracerProvider flush via returned shutdown function.

## Errors

- Invalid `OTEL_TRACES_EXPORTER` value returns error from Install.
- OTLP exporter without endpoint returns error when exporter is `otlp`.

## Metrics

Deferred to M4.

## Logs

Process logs remain `slog` in `main`. Port `Logger` stays noop.

## Tracing

M0 enables provider only. Example span creation available via injected tracer; full HTTP instrumentation in M4.

## Security

OTLP endpoint configured via env; no secrets in code. TLS for OTLP uses exporter defaults (M4 hardening).

## Performance

When disabled, noop tracer — zero OTel overhead. When enabled, batch span processor defaults from SDK.

## Testing

- Config parsing tests for telemetry env vars.
- Adapter `StartSpan` creates/end span without panic (sdk test exporter or noop provider in test).
- Existing test suite green.

## Environment reference

| Variable | Default | Values |
|---|---|---|
| `OTEL_SDK_DISABLED` | `true` | `true`, `false` |
| `OTEL_SERVICE_NAME` | `universal-api-gateway` | any string |
| `OTEL_TRACES_EXPORTER` | `stdout` | `stdout`, `otlp`, `none` |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | empty | URL e.g. `http://localhost:4318` |

Enable tracing locally:

```bash
OTEL_SDK_DISABLED=false OTEL_TRACES_EXPORTER=stdout make run
```
