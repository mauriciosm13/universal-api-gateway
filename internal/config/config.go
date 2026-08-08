package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

const (
	defaultHost         = "0.0.0.0"
	defaultPort         = 8080
	defaultReadTimeout  = 15 * time.Second
	defaultWriteTimeout = 15 * time.Second
	defaultIdleTimeout  = 60 * time.Second
)

// Config holds runtime configuration loaded from environment variables.
type Config struct {
	Host         string
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// Load reads configuration from the environment with sensible defaults.
func Load() (Config, error) {
	port, err := envInt("GATEWAY_PORT", defaultPort)
	if err != nil {
		return Config{}, fmt.Errorf("invalid GATEWAY_PORT: %w", err)
	}

	return Config{
		Host:         envString("GATEWAY_HOST", defaultHost),
		Port:         port,
		ReadTimeout:  envDuration("GATEWAY_READ_TIMEOUT", defaultReadTimeout),
		WriteTimeout: envDuration("GATEWAY_WRITE_TIMEOUT", defaultWriteTimeout),
		IdleTimeout:  envDuration("GATEWAY_IDLE_TIMEOUT", defaultIdleTimeout),
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
