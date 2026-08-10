package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultHost              = "0.0.0.0"
	defaultPort              = 8080
	defaultReadTimeout       = 15 * time.Second
	defaultWriteTimeout      = 15 * time.Second
	defaultIdleTimeout       = 60 * time.Second
	defaultServiceName       = "universal-api-gateway"
	defaultTracesExporter    = "stdout"
	defaultOTELSDKDisabled   = true
)

// Config holds runtime configuration loaded from environment variables.
type Config struct {
	Host         string
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	Telemetry    TelemetryConfig
}

// TelemetryConfig holds OpenTelemetry settings from OTEL_* environment variables.
type TelemetryConfig struct {
	Disabled       bool
	ServiceName    string
	TracesExporter string
	OTLPEndpoint   string
}

// Load reads configuration from the environment with sensible defaults.
func Load() (Config, error) {
	port, err := envInt("GATEWAY_PORT", defaultPort)
	if err != nil {
		return Config{}, fmt.Errorf("invalid GATEWAY_PORT: %w", err)
	}

	telemetry, err := loadTelemetry()
	if err != nil {
		return Config{}, err
	}

	return Config{
		Host:         envString("GATEWAY_HOST", defaultHost),
		Port:         port,
		ReadTimeout:  envDuration("GATEWAY_READ_TIMEOUT", defaultReadTimeout),
		WriteTimeout: envDuration("GATEWAY_WRITE_TIMEOUT", defaultWriteTimeout),
		IdleTimeout:  envDuration("GATEWAY_IDLE_TIMEOUT", defaultIdleTimeout),
		Telemetry:    telemetry,
	}, nil
}

func loadTelemetry() (TelemetryConfig, error) {
	exporter := strings.ToLower(envString("OTEL_TRACES_EXPORTER", defaultTracesExporter))
	if exporter != "stdout" && exporter != "otlp" && exporter != "none" {
		return TelemetryConfig{}, fmt.Errorf("invalid OTEL_TRACES_EXPORTER %q: want stdout, otlp, or none", exporter)
	}

	return TelemetryConfig{
		Disabled:       envBool("OTEL_SDK_DISABLED", defaultOTELSDKDisabled),
		ServiceName:    envString("OTEL_SERVICE_NAME", defaultServiceName),
		TracesExporter: exporter,
		OTLPEndpoint:   os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
	}, nil
}

// Addr returns the listen address in host:port form.
func (c Config) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func envString(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envInt(key string, fallback int) (int, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}
	return parsed, nil
}

func envBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func envDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
}
