package config

import (
	"fmt"
	"strings"
)

// HostRoute maps an HTTP Host to an upstream URL.
type HostRoute struct {
	Host     string
	Upstream string
}

// ParseHostRoutes parses comma-separated host=upstream pairs.
func ParseHostRoutes(raw string) ([]HostRoute, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	entries := strings.Split(raw, ",")
	routes := make([]HostRoute, 0, len(entries))

	for _, entry := range entries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}

		host, upstream, ok := strings.Cut(entry, "=")
		if !ok {
			return nil, fmt.Errorf("entry %q: expected host=upstream", entry)
		}

		host = strings.TrimSpace(host)
		upstream = strings.TrimSpace(upstream)
		if host == "" {
			return nil, fmt.Errorf("entry %q: host is required", entry)
		}
		if upstream == "" {
			return nil, fmt.Errorf("entry %q: upstream is required", entry)
		}
		if err := ValidateUpstreamURL(upstream); err != nil {
			return nil, fmt.Errorf("entry %q: %w", entry, err)
		}

		routes = append(routes, HostRoute{
			Host:     strings.ToLower(host),
			Upstream: upstream,
		})
	}

	if len(routes) == 0 {
		return nil, fmt.Errorf("GATEWAY_HOST_ROUTES contains no valid entries")
	}

	return routes, nil
}
