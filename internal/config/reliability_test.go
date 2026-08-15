package config

import (
	"strings"
	"testing"
	"time"
)

func TestLoadRuntimeDefaultsToServer(t *testing.T) {
	t.Setenv("GATEWAY_RUNTIME", "")

	cfg, err := loadRuntime()
	if err != nil {
		t.Fatalf("loadRuntime() error = %v", err)
	}

	if cfg.Mode != RuntimeServer || cfg.IsLambda() {
		t.Fatalf("runtime = %+v, want server mode", cfg)
	}
}

func TestLoadRuntimeLambdaUsesDefaultTimeout(t *testing.T) {
	t.Setenv("GATEWAY_RUNTIME", "lambda")
	t.Setenv("AWS_LAMBDA_FUNCTION_TIMEOUT", "")

	cfg, err := loadRuntime()
	if err != nil {
		t.Fatalf("loadRuntime() error = %v", err)
	}

	if !cfg.IsLambda() || cfg.LambdaTimeout != defaultLambdaTimeout {
		t.Fatalf("runtime = %+v, want lambda with %s timeout", cfg, defaultLambdaTimeout)
	}
}

func TestLoadRuntimeLambdaCustomTimeout(t *testing.T) {
	t.Setenv("GATEWAY_RUNTIME", "lambda")
	t.Setenv("AWS_LAMBDA_FUNCTION_TIMEOUT", "45")

	cfg, err := loadRuntime()
	if err != nil {
		t.Fatalf("loadRuntime() error = %v", err)
	}

	if cfg.LambdaTimeout != 45*time.Second {
		t.Fatalf("LambdaTimeout = %s, want 45s", cfg.LambdaTimeout)
	}
}

func TestLoadRuntimeRejectsInvalidMode(t *testing.T) {
	t.Setenv("GATEWAY_RUNTIME", "ecs")

	_, err := loadRuntime()
	if err == nil {
		t.Fatal("expected error for invalid runtime")
	}
}

func TestLoadReliabilityDefaults(t *testing.T) {
	t.Setenv("GATEWAY_UPSTREAM_TIMEOUT", "")
	t.Setenv("GATEWAY_RETRY_MAX", "")
	t.Setenv("GATEWAY_RETRY_BACKOFF", "")

	cfg, err := loadReliability(RuntimeConfig{Mode: RuntimeServer})
	if err != nil {
		t.Fatalf("loadReliability() error = %v", err)
	}

	if cfg.UpstreamTimeout != defaultUpstreamTimeout {
		t.Fatalf("UpstreamTimeout = %s, want %s", cfg.UpstreamTimeout, defaultUpstreamTimeout)
	}
	if cfg.RetryMax != defaultRetryMax || cfg.RetryBackoff != defaultRetryBackoff {
		t.Fatalf("retry = max %d backoff %s, want max %d backoff %s",
			cfg.RetryMax, cfg.RetryBackoff, defaultRetryMax, defaultRetryBackoff)
	}
}

func TestLoadReliabilityCustomValues(t *testing.T) {
	t.Setenv("GATEWAY_UPSTREAM_TIMEOUT", "5s")
	t.Setenv("GATEWAY_RETRY_MAX", "1")
	t.Setenv("GATEWAY_RETRY_BACKOFF", "50ms")

	cfg, err := loadReliability(RuntimeConfig{Mode: RuntimeServer})
	if err != nil {
		t.Fatalf("loadReliability() error = %v", err)
	}

	if cfg.UpstreamTimeout != 5*time.Second || cfg.RetryMax != 1 || cfg.RetryBackoff != 50*time.Millisecond {
		t.Fatalf("cfg = %+v", cfg)
	}
}

func TestLoadReliabilityRejectsInvalidTimeout(t *testing.T) {
	t.Setenv("GATEWAY_UPSTREAM_TIMEOUT", "not-a-duration")

	_, err := loadReliability(RuntimeConfig{Mode: RuntimeServer})
	if err == nil {
		t.Fatal("expected error for invalid timeout")
	}
}

func TestLoadReliabilityRejectsRetryMaxOutOfRange(t *testing.T) {
	t.Setenv("GATEWAY_RETRY_MAX", "9")

	_, err := loadReliability(RuntimeConfig{Mode: RuntimeServer})
	if err == nil {
		t.Fatal("expected error for retry max out of range")
	}
}

func TestLoadReliabilityLambdaValidation(t *testing.T) {
	t.Setenv("GATEWAY_UPSTREAM_TIMEOUT", "20s")
	t.Setenv("GATEWAY_RETRY_MAX", "2")
	t.Setenv("GATEWAY_RETRY_BACKOFF", "100ms")

	runtime := RuntimeConfig{Mode: RuntimeLambda, LambdaTimeout: 30 * time.Second}
	_, err := loadReliability(runtime)
	if err == nil {
		t.Fatal("expected lambda worst-case validation error")
	}
	if !strings.Contains(err.Error(), "lambda timeout") {
		t.Fatalf("error = %v, want lambda timeout mention", err)
	}
}

func TestLoadReliabilityLambdaAcceptsSafeConfig(t *testing.T) {
	t.Setenv("GATEWAY_UPSTREAM_TIMEOUT", "8s")
	t.Setenv("GATEWAY_RETRY_MAX", "1")
	t.Setenv("GATEWAY_RETRY_BACKOFF", "100ms")

	runtime := RuntimeConfig{Mode: RuntimeLambda, LambdaTimeout: 30 * time.Second}
	cfg, err := loadReliability(runtime)
	if err != nil {
		t.Fatalf("loadReliability() error = %v", err)
	}

	if cfg.MaxAttempts() != 2 {
		t.Fatalf("MaxAttempts() = %d, want 2", cfg.MaxAttempts())
	}
}

func TestReliabilityWorstCaseDurationNoRetry(t *testing.T) {
	cfg := ReliabilityConfig{UpstreamTimeout: 10 * time.Second, RetryMax: 0}
	if cfg.WorstCaseDuration() != 10*time.Second {
		t.Fatalf("WorstCaseDuration() = %s, want 10s", cfg.WorstCaseDuration())
	}
}

func TestReliabilityWorstCaseDuration(t *testing.T) {
	cfg := ReliabilityConfig{
		UpstreamTimeout: 10 * time.Second,
		RetryMax:        2,
		RetryBackoff:    100 * time.Millisecond,
	}

	want := 3*10*time.Second + (100*time.Millisecond + 200*time.Millisecond)
	if cfg.WorstCaseDuration() != want {
		t.Fatalf("WorstCaseDuration() = %s, want %s", cfg.WorstCaseDuration(), want)
	}
}

func TestReliabilityWithDefaults(t *testing.T) {
	cfg := ReliabilityConfig{}.WithDefaults()
	if cfg.UpstreamTimeout != defaultUpstreamTimeout || cfg.RetryBackoff != defaultRetryBackoff {
		t.Fatalf("cfg = %+v", cfg)
	}
}

func TestLoadIncludesReliability(t *testing.T) {
	t.Setenv("GATEWAY_UPSTREAM_TIMEOUT", "12s")
	t.Setenv("GATEWAY_RETRY_MAX", "0")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Reliability.UpstreamTimeout != 12*time.Second || cfg.Reliability.RetryMax != 0 {
		t.Fatalf("Reliability = %+v", cfg.Reliability)
	}
}
