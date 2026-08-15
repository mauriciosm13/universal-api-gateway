package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultUpstreamTimeout = 30 * time.Second
	defaultRetryMax        = 2
	defaultRetryBackoff    = 100 * time.Millisecond
	maxUpstreamTimeout     = 5 * time.Minute
	maxRetryMax            = 3
)

// ReliabilityConfig holds upstream timeout and retry settings.
type ReliabilityConfig struct {
	UpstreamTimeout time.Duration
	RetryMax        int
	RetryBackoff    time.Duration
}

// MaxAttempts returns total upstream attempts including the first try.
func (c ReliabilityConfig) MaxAttempts() int {
	return c.RetryMax + 1
}

// WorstCaseDuration estimates upper bound latency for all retry attempts.
func (c ReliabilityConfig) WorstCaseDuration() time.Duration {
	if c.RetryMax <= 0 {
		return c.UpstreamTimeout
	}

	backoffTotal := c.RetryBackoff * time.Duration(c.RetryMax*(c.RetryMax+1)/2)
	return time.Duration(c.MaxAttempts())*c.UpstreamTimeout + backoffTotal
}

// WithDefaults returns cfg with zero values replaced by MVP defaults.
func (c ReliabilityConfig) WithDefaults() ReliabilityConfig {
	if c.UpstreamTimeout <= 0 {
		c.UpstreamTimeout = defaultUpstreamTimeout
	}
	if c.RetryBackoff <= 0 {
		c.RetryBackoff = defaultRetryBackoff
	}
	if c.RetryMax < 0 {
		c.RetryMax = defaultRetryMax
	}
	return c
}

func loadReliability(runtime RuntimeConfig) (ReliabilityConfig, error) {
	timeout, err := envDurationStrict("GATEWAY_UPSTREAM_TIMEOUT", defaultUpstreamTimeout)
	if err != nil {
		return ReliabilityConfig{}, fmt.Errorf("invalid GATEWAY_UPSTREAM_TIMEOUT: %w", err)
	}
	if timeout <= 0 || timeout > maxUpstreamTimeout {
		return ReliabilityConfig{}, fmt.Errorf("invalid GATEWAY_UPSTREAM_TIMEOUT: must be > 0 and <= %s", maxUpstreamTimeout)
	}

	retryMax, err := envIntInRange("GATEWAY_RETRY_MAX", defaultRetryMax, 0, maxRetryMax)
	if err != nil {
		return ReliabilityConfig{}, fmt.Errorf("invalid GATEWAY_RETRY_MAX: %w", err)
	}

	backoff, err := envDurationStrict("GATEWAY_RETRY_BACKOFF", defaultRetryBackoff)
	if err != nil {
		return ReliabilityConfig{}, fmt.Errorf("invalid GATEWAY_RETRY_BACKOFF: %w", err)
	}
	if backoff <= 0 {
		return ReliabilityConfig{}, fmt.Errorf("invalid GATEWAY_RETRY_BACKOFF: must be > 0")
	}

	cfg := ReliabilityConfig{
		UpstreamTimeout: timeout,
		RetryMax:        retryMax,
		RetryBackoff:    backoff,
	}

	if runtime.IsLambda() {
		margin := 2 * time.Second
		if cfg.WorstCaseDuration() > runtime.LambdaTimeout-margin {
			return ReliabilityConfig{}, fmt.Errorf(
				"reliability worst-case duration %s exceeds lambda timeout %s (need %s margin)",
				cfg.WorstCaseDuration(),
				runtime.LambdaTimeout,
				margin,
			)
		}
	}

	return cfg, nil
}

func envDurationStrict(key string, fallback time.Duration) (time.Duration, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, err
	}
	return parsed, nil
}

func envIntInRange(key string, fallback, min, max int) (int, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}
	if parsed < min || parsed > max {
		return 0, fmt.Errorf("must be between %d and %d", min, max)
	}
	return parsed, nil
}
