package config

import "testing"

func TestLoadTelemetryDefaults(t *testing.T) {
	t.Setenv("OTEL_SDK_DISABLED", "")
	t.Setenv("OTEL_TRACES_EXPORTER", "")
	t.Setenv("OTEL_SERVICE_NAME", "")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if !cfg.Telemetry.Disabled {
		t.Fatal("expected telemetry disabled by default")
	}

	if cfg.Telemetry.ServiceName != "universal-api-gateway" {
		t.Fatalf("expected default service name, got %q", cfg.Telemetry.ServiceName)
	}

	if cfg.Telemetry.TracesExporter != "stdout" {
		t.Fatalf("expected stdout exporter default, got %q", cfg.Telemetry.TracesExporter)
	}
}

func TestLoadTelemetryEnabled(t *testing.T) {
	t.Setenv("OTEL_SDK_DISABLED", "false")
	t.Setenv("OTEL_TRACES_EXPORTER", "otlp")
	t.Setenv("OTEL_SERVICE_NAME", "gateway-test")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4318")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Telemetry.Disabled {
		t.Fatal("expected telemetry enabled")
	}

	if cfg.Telemetry.ServiceName != "gateway-test" {
		t.Fatalf("expected gateway-test, got %q", cfg.Telemetry.ServiceName)
	}

	if cfg.Telemetry.TracesExporter != "otlp" {
		t.Fatalf("expected otlp exporter, got %q", cfg.Telemetry.TracesExporter)
	}

	if cfg.Telemetry.OTLPEndpoint != "http://localhost:4318" {
		t.Fatalf("unexpected OTLP endpoint %q", cfg.Telemetry.OTLPEndpoint)
	}
}

func TestLoadTelemetryInvalidExporter(t *testing.T) {
	t.Setenv("OTEL_TRACES_EXPORTER", "invalid")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid OTEL_TRACES_EXPORTER")
	}
}
