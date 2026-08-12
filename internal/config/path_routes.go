package config

import (
	"fmt"
	"strings"
)

// PathRoute maps a path prefix to an upstream URL.
type PathRoute struct {
	Prefix   string
	Upstream string
}

// ParsePathRoutes parses comma-separated prefix=upstream pairs.
func ParsePathRoutes(raw string) ([]PathRoute, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	entries := strings.Split(raw, ",")
	routes := make([]PathRoute, 0, len(entries))

	for _, entry := range entries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}

		prefix, upstream, ok := strings.Cut(entry, "=")
		if !ok {
			return nil, fmt.Errorf("entry %q: expected prefix=upstream", entry)
		}

		prefix = strings.TrimSpace(prefix)
		upstream = strings.TrimSpace(upstream)
		if prefix == "" {
			return nil, fmt.Errorf("entry %q: prefix is required", entry)
		}
		if !strings.HasPrefix(prefix, "/") {
			return nil, fmt.Errorf("entry %q: prefix must start with /", entry)
		}
		if upstream == "" {
			return nil, fmt.Errorf("entry %q: upstream is required", entry)
		}
		if err := ValidateUpstreamURL(upstream); err != nil {
			return nil, fmt.Errorf("entry %q: %w", entry, err)
		}

		routes = append(routes, PathRoute{
			Prefix:   prefix,
			Upstream: upstream,
		})
	}

	if len(routes) == 0 {
		return nil, fmt.Errorf("GATEWAY_PATH_ROUTES contains no valid entries")
	}

	return routes, nil
}
