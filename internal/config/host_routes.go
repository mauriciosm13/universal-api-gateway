package config

import (
	"fmt"
	"strings"
)

// HostRoute maps an HTTP Host to one or more upstream URLs.
type HostRoute struct {
	Host      string
	Upstreams []string
}

// ParseHostRoutes parses comma-separated host=upstream pairs.
// Each upstream side may list multiple comma-separated URLs.
func ParseHostRoutes(raw string) ([]HostRoute, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	entries := splitRouteEntries(raw, isHostRouteEntry)
	routes := make([]HostRoute, 0, len(entries))

	for _, entry := range entries {
		host, upstreamRaw, ok := strings.Cut(entry, "=")
		if !ok {
			return nil, fmt.Errorf("entry %q: expected host=upstream", entry)
		}

		host = strings.TrimSpace(host)
		if host == "" {
			return nil, fmt.Errorf("entry %q: host is required", entry)
		}

		upstreams, err := ParseUpstreamList(upstreamRaw)
		if err != nil {
			return nil, fmt.Errorf("entry %q: %w", entry, err)
		}

		routes = append(routes, HostRoute{
			Host:      strings.ToLower(host),
			Upstreams: upstreams,
		})
	}

	if len(routes) == 0 {
		return nil, fmt.Errorf("GATEWAY_HOST_ROUTES contains no valid entries")
	}

	return routes, nil
}
