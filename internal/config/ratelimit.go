package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// RateLimitParams holds token bucket rate and burst for one limit profile.
type RateLimitParams struct {
	RPS   float64
	Burst int
}

// RateLimitRoute maps a path prefix to rate limit parameters.
type RateLimitRoute struct {
	Prefix string
	RPS    float64
	Burst  int
}

// RateLimitConfig holds rate limiting settings from environment variables.
type RateLimitConfig struct {
	Global *RateLimitParams
	Routes []RateLimitRoute
}

// Enabled reports whether rate limiting is configured.
func (c RateLimitConfig) Enabled() bool {
	return c.Global != nil
}

// LimitsForPath returns the effective limit parameters and profile key for path.
// Longest matching route prefix wins; otherwise global limits apply.
func (c RateLimitConfig) LimitsForPath(path string) (RateLimitParams, string) {
	if c.Global == nil {
		return RateLimitParams{}, "global"
	}

	bestPrefix := ""
	var best RateLimitParams

	for _, route := range c.Routes {
		if !strings.HasPrefix(path, route.Prefix) {
			continue
		}
		if len(route.Prefix) <= len(bestPrefix) {
			continue
		}

		bestPrefix = route.Prefix
		best = RateLimitParams{RPS: route.RPS, Burst: route.Burst}
	}

	if bestPrefix != "" {
		return best, bestPrefix
	}

	return *c.Global, "global"
}

func loadRateLimit() (RateLimitConfig, error) {
	rpsRaw := strings.TrimSpace(os.Getenv("GATEWAY_RATE_LIMIT_RPS"))
	if rpsRaw == "" {
		return RateLimitConfig{}, nil
	}

	rps, err := strconv.ParseFloat(rpsRaw, 64)
	if err != nil || rps <= 0 {
		return RateLimitConfig{}, fmt.Errorf("invalid GATEWAY_RATE_LIMIT_RPS: must be a positive number")
	}

	burstRaw := strings.TrimSpace(os.Getenv("GATEWAY_RATE_LIMIT_BURST"))
	if burstRaw == "" {
		return RateLimitConfig{}, fmt.Errorf("GATEWAY_RATE_LIMIT_BURST is required when GATEWAY_RATE_LIMIT_RPS is set")
	}

	burst, err := strconv.Atoi(burstRaw)
	if err != nil || burst <= 0 {
		return RateLimitConfig{}, fmt.Errorf("invalid GATEWAY_RATE_LIMIT_BURST: must be a positive integer")
	}

	routes, err := ParseRateLimitRoutes(os.Getenv("GATEWAY_RATE_LIMIT_ROUTES"))
	if err != nil {
		return RateLimitConfig{}, err
	}

	return RateLimitConfig{
		Global: &RateLimitParams{RPS: rps, Burst: burst},
		Routes: routes,
	}, nil
}

// ParseRateLimitRoutes parses comma-separated prefix=rps:burst entries.
func ParseRateLimitRoutes(raw string) ([]RateLimitRoute, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	entries := strings.Split(raw, ",")
	routes := make([]RateLimitRoute, 0, len(entries))

	for _, entry := range entries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}

		prefix, limits, ok := strings.Cut(entry, "=")
		if !ok {
			return nil, fmt.Errorf("entry %q: expected prefix=rps:burst", entry)
		}

		prefix = strings.TrimSpace(prefix)
		if prefix == "" {
			return nil, fmt.Errorf("entry %q: prefix is required", entry)
		}
		if !strings.HasPrefix(prefix, "/") {
			return nil, fmt.Errorf("entry %q: prefix must start with /", entry)
		}

		rpsRaw, burstRaw, ok := strings.Cut(strings.TrimSpace(limits), ":")
		if !ok {
			return nil, fmt.Errorf("entry %q: expected prefix=rps:burst", entry)
		}

		rps, err := strconv.ParseFloat(strings.TrimSpace(rpsRaw), 64)
		if err != nil || rps <= 0 {
			return nil, fmt.Errorf("entry %q: rps must be a positive number", entry)
		}

		burst, err := strconv.Atoi(strings.TrimSpace(burstRaw))
		if err != nil || burst <= 0 {
			return nil, fmt.Errorf("entry %q: burst must be a positive integer", entry)
		}

		routes = append(routes, RateLimitRoute{
			Prefix: prefix,
			RPS:    rps,
			Burst:  burst,
		})
	}

	return routes, nil
}
