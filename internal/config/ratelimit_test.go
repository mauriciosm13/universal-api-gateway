package config

import (
	"testing"
)

func TestLoadRateLimitDisabledByDefault(t *testing.T) {
	t.Setenv("GATEWAY_RATE_LIMIT_RPS", "")
	t.Setenv("GATEWAY_RATE_LIMIT_BURST", "")
	t.Setenv("GATEWAY_RATE_LIMIT_ROUTES", "")

	cfg, err := loadRateLimit()
	if err != nil {
		t.Fatalf("loadRateLimit() error = %v", err)
	}

	if cfg.Enabled() {
		t.Fatal("expected rate limit disabled")
	}
}

func TestLoadRateLimitGlobal(t *testing.T) {
	t.Setenv("GATEWAY_RATE_LIMIT_RPS", "5")
	t.Setenv("GATEWAY_RATE_LIMIT_BURST", "10")
	t.Setenv("GATEWAY_RATE_LIMIT_ROUTES", "")

	cfg, err := loadRateLimit()
	if err != nil {
		t.Fatalf("loadRateLimit() error = %v", err)
	}

	if !cfg.Enabled() {
		t.Fatal("expected rate limit enabled")
	}

	if cfg.Global.RPS != 5 || cfg.Global.Burst != 10 {
		t.Fatalf("global = %+v, want rps=5 burst=10", cfg.Global)
	}
}

func TestLoadRateLimitRequiresBurst(t *testing.T) {
	t.Setenv("GATEWAY_RATE_LIMIT_RPS", "5")
	t.Setenv("GATEWAY_RATE_LIMIT_BURST", "")

	_, err := loadRateLimit()
	if err == nil {
		t.Fatal("expected error when burst missing")
	}
}

func TestLoadRateLimitInvalidRPS(t *testing.T) {
	t.Setenv("GATEWAY_RATE_LIMIT_RPS", "0")
	t.Setenv("GATEWAY_RATE_LIMIT_BURST", "10")

	_, err := loadRateLimit()
	if err == nil {
		t.Fatal("expected error for invalid RPS")
	}
}

func TestParseRateLimitRoutes(t *testing.T) {
	t.Parallel()

	routes, err := ParseRateLimitRoutes("/api=10:20,/admin=2:4")
	if err != nil {
		t.Fatalf("ParseRateLimitRoutes() error = %v", err)
	}

	if len(routes) != 2 {
		t.Fatalf("routes len = %d, want 2", len(routes))
	}

	if routes[0].Prefix != "/api" || routes[0].RPS != 10 || routes[0].Burst != 20 {
		t.Fatalf("first route = %+v", routes[0])
	}
}

func TestParseRateLimitRoutesInvalidPrefix(t *testing.T) {
	t.Parallel()

	_, err := ParseRateLimitRoutes("api=10:20")
	if err == nil {
		t.Fatal("expected error for prefix without leading slash")
	}
}

func TestParseRateLimitRoutesInvalidLimits(t *testing.T) {
	t.Parallel()

	_, err := ParseRateLimitRoutes("/api=10")
	if err == nil {
		t.Fatal("expected error for missing burst")
	}
}

func TestRateLimitConfigLimitsForPath(t *testing.T) {
	t.Parallel()

	cfg := RateLimitConfig{
		Global: &RateLimitParams{RPS: 100, Burst: 100},
		Routes: []RateLimitRoute{
			{Prefix: "/api", RPS: 2, Burst: 4},
			{Prefix: "/api/admin", RPS: 1, Burst: 1},
		},
	}

	params, profile := cfg.LimitsForPath("/other")
	if profile != "global" || params.RPS != 100 {
		t.Fatalf("other path = (%+v, %q)", params, profile)
	}

	params, profile = cfg.LimitsForPath("/api/users")
	if profile != "/api" || params.RPS != 2 || params.Burst != 4 {
		t.Fatalf("/api/users = (%+v, %q)", params, profile)
	}

	params, profile = cfg.LimitsForPath("/api/admin/settings")
	if profile != "/api/admin" || params.RPS != 1 || params.Burst != 1 {
		t.Fatalf("/api/admin/settings = (%+v, %q)", params, profile)
	}
}

func TestConfigLoadIncludesRateLimit(t *testing.T) {
	t.Setenv("GATEWAY_RATE_LIMIT_RPS", "3")
	t.Setenv("GATEWAY_RATE_LIMIT_BURST", "6")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if !cfg.RateLimit.Enabled() {
		t.Fatal("expected rate limit enabled in Load()")
	}
}
