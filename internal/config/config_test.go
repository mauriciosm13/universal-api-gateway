package config

import (
	"testing"
	"time"
)

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

func TestLoadDefaults(t *testing.T) {
	t.Setenv("GATEWAY_HOST", "")
	t.Setenv("GATEWAY_PORT", "")
	t.Setenv("GATEWAY_READ_TIMEOUT", "")
	t.Setenv("GATEWAY_WRITE_TIMEOUT", "")
	t.Setenv("GATEWAY_IDLE_TIMEOUT", "")
	t.Setenv("OTEL_TRACES_EXPORTER", "none")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Host != "0.0.0.0" {
		t.Fatalf("expected default host, got %q", cfg.Host)
	}

	if cfg.Port != 8080 {
		t.Fatalf("expected default port 8080, got %d", cfg.Port)
	}

	if cfg.ReadTimeout != 15*time.Second {
		t.Fatalf("expected default read timeout, got %v", cfg.ReadTimeout)
	}

	if cfg.Telemetry.TracesExporter != "none" {
		t.Fatalf("expected none exporter, got %q", cfg.Telemetry.TracesExporter)
	}
}

func TestLoadInvalidPort(t *testing.T) {
	t.Setenv("GATEWAY_PORT", "not-a-port")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid GATEWAY_PORT")
	}
}

func TestLoadEnvOverrides(t *testing.T) {
	t.Setenv("GATEWAY_HOST", "127.0.0.1")
	t.Setenv("GATEWAY_PORT", "9090")
	t.Setenv("GATEWAY_READ_TIMEOUT", "5s")
	t.Setenv("GATEWAY_WRITE_TIMEOUT", "6s")
	t.Setenv("GATEWAY_IDLE_TIMEOUT", "7s")
	t.Setenv("OTEL_TRACES_EXPORTER", "stdout")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Host != "127.0.0.1" || cfg.Port != 9090 {
		t.Fatalf("unexpected host/port: %s:%d", cfg.Host, cfg.Port)
	}

	if cfg.ReadTimeout != 5*time.Second || cfg.WriteTimeout != 6*time.Second || cfg.IdleTimeout != 7*time.Second {
		t.Fatal("unexpected timeout overrides")
	}
}

func TestConfigAddr(t *testing.T) {
	cfg := Config{Host: "127.0.0.1", Port: 8080}
	if cfg.Addr() != "127.0.0.1:8080" {
		t.Fatalf("unexpected addr %q", cfg.Addr())
	}
}

func TestEnvBoolInvalidFallback(t *testing.T) {
	t.Setenv("OTEL_SDK_DISABLED", "not-a-bool")
	t.Setenv("OTEL_TRACES_EXPORTER", "stdout")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if !cfg.Telemetry.Disabled {
		t.Fatal("expected invalid bool to fall back to default true")
	}
}

func TestLoadDefaultUpstreamValid(t *testing.T) {
	t.Setenv("GATEWAY_DEFAULT_UPSTREAM", "http://backend:8080")
	t.Setenv("OTEL_TRACES_EXPORTER", "none")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.DefaultUpstreams[0] != "http://backend:8080" {
		t.Fatalf("expected default upstream, got %q", cfg.DefaultUpstreams[0])
	}
}

func TestLoadDefaultUpstreamInvalid(t *testing.T) {
	t.Setenv("GATEWAY_DEFAULT_UPSTREAM", "not-a-url")
	t.Setenv("OTEL_TRACES_EXPORTER", "none")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid GATEWAY_DEFAULT_UPSTREAM")
	}
}

func TestLoadHostRoutesValid(t *testing.T) {
	t.Setenv("GATEWAY_HOST_ROUTES", "api.example.com=http://api:8080")
	t.Setenv("OTEL_TRACES_EXPORTER", "none")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(cfg.HostRoutes) != 1 {
		t.Fatalf("unexpected host routes: %+v", cfg.HostRoutes)
	}
}

func TestLoadMethodRoutesValid(t *testing.T) {
	t.Setenv("GATEWAY_METHOD_ROUTES", "GET=http://get:8080")
	t.Setenv("OTEL_TRACES_EXPORTER", "none")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(cfg.MethodRoutes) != 1 || cfg.MethodRoutes[0].Method != "GET" {
		t.Fatalf("unexpected method routes: %+v", cfg.MethodRoutes)
	}
}

func TestLoadHeaderRoutesValid(t *testing.T) {
	t.Setenv("GATEWAY_HEADER_ROUTES", "X-Version=v1=http://v1:8080")
	t.Setenv("OTEL_TRACES_EXPORTER", "none")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(cfg.HeaderRoutes) != 1 || cfg.HeaderRoutes[0].Name != "X-Version" {
		t.Fatalf("unexpected header routes: %+v", cfg.HeaderRoutes)
	}
}

func TestLoadHeaderRoutesInvalid(t *testing.T) {
	t.Setenv("GATEWAY_HEADER_ROUTES", "X-Version=v1")
	t.Setenv("OTEL_TRACES_EXPORTER", "none")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid GATEWAY_HEADER_ROUTES")
	}
}

func TestLoadPathRoutesValid(t *testing.T) {
	t.Setenv("GATEWAY_PATH_ROUTES", "/api=http://api:8080")
	t.Setenv("OTEL_TRACES_EXPORTER", "none")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(cfg.PathRoutes) != 1 || cfg.PathRoutes[0].Prefix != "/api" {
		t.Fatalf("unexpected path routes: %+v", cfg.PathRoutes)
	}
}

func TestLoadPathRoutesInvalid(t *testing.T) {
	t.Setenv("GATEWAY_PATH_ROUTES", "api=http://api:8080")
	t.Setenv("OTEL_TRACES_EXPORTER", "none")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid GATEWAY_PATH_ROUTES")
	}
}

func TestValidateUpstreamURL(t *testing.T) {
	t.Parallel()

	if err := ValidateUpstreamURL("http://localhost:8080"); err != nil {
		t.Fatalf("valid http URL: %v", err)
	}

	if err := ValidateUpstreamURL("https://api.example.com"); err != nil {
		t.Fatalf("valid https URL: %v", err)
	}

	if err := ValidateUpstreamURL("ftp://files.example.com"); err == nil {
		t.Fatal("expected error for ftp scheme")
	}

	if err := ValidateUpstreamURL("http:///path"); err == nil {
		t.Fatal("expected error for missing host")
	}
}

func TestLoadInvalidDurationFallback(t *testing.T) {
	t.Setenv("GATEWAY_READ_TIMEOUT", "invalid")
	t.Setenv("OTEL_TRACES_EXPORTER", "none")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.ReadTimeout != defaultReadTimeout {
		t.Fatalf("ReadTimeout = %v, want default %v", cfg.ReadTimeout, defaultReadTimeout)
	}
}
