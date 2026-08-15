package config

import (
	"fmt"
	"strings"
)

// PathRoute maps a path prefix to one or more upstream URLs.
type PathRoute struct {
	Prefix    string
	Upstreams []string
}

// ParsePathRoutes parses comma-separated prefix=upstream pairs.
// Each upstream side may list multiple comma-separated URLs.
func ParsePathRoutes(raw string) ([]PathRoute, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	entries := splitRouteEntries(raw, isPathRouteEntry)
	routes := make([]PathRoute, 0, len(entries))

	for _, entry := range entries {
		prefix, upstreamRaw, ok := strings.Cut(entry, "=")
		if !ok {
			return nil, fmt.Errorf("entry %q: expected prefix=upstream", entry)
		}

		prefix = strings.TrimSpace(prefix)
		if prefix == "" {
			return nil, fmt.Errorf("entry %q: prefix is required", entry)
		}
		if !strings.HasPrefix(prefix, "/") {
			return nil, fmt.Errorf("entry %q: prefix must start with /", entry)
		}

		upstreams, err := ParseUpstreamList(upstreamRaw)
		if err != nil {
			return nil, fmt.Errorf("entry %q: %w", entry, err)
		}

		routes = append(routes, PathRoute{
			Prefix:    prefix,
			Upstreams: upstreams,
		})
	}

	if len(routes) == 0 {
		return nil, fmt.Errorf("GATEWAY_PATH_ROUTES contains no valid entries")
	}

	return routes, nil
}
