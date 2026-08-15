package config

import (
	"fmt"
	"strings"
)

// HeaderRoute maps a header name/value pair to one or more upstream URLs.
type HeaderRoute struct {
	Name      string
	Value     string
	Upstreams []string
}

// ParseHeaderRoutes parses comma-separated HeaderName=HeaderValue=upstream entries.
// Each upstream side may list multiple comma-separated URLs.
func ParseHeaderRoutes(raw string) ([]HeaderRoute, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	entries := splitRouteEntries(raw, isHeaderRouteEntry)
	routes := make([]HeaderRoute, 0, len(entries))

	for _, entry := range entries {
		name, rest, ok := strings.Cut(entry, "=")
		if !ok {
			return nil, fmt.Errorf("entry %q: expected HeaderName=HeaderValue=upstream", entry)
		}

		value, upstreamRaw, ok := strings.Cut(rest, "=")
		if !ok {
			return nil, fmt.Errorf("entry %q: expected HeaderName=HeaderValue=upstream", entry)
		}

		name = strings.TrimSpace(name)
		value = strings.TrimSpace(value)

		if name == "" {
			return nil, fmt.Errorf("entry %q: header name is required", entry)
		}
		if value == "" {
			return nil, fmt.Errorf("entry %q: header value is required", entry)
		}

		upstreams, err := ParseUpstreamList(upstreamRaw)
		if err != nil {
			return nil, fmt.Errorf("entry %q: %w", entry, err)
		}

		routes = append(routes, HeaderRoute{
			Name:      name,
			Value:     value,
			Upstreams: upstreams,
		})
	}

	if len(routes) == 0 {
		return nil, fmt.Errorf("GATEWAY_HEADER_ROUTES contains no valid entries")
	}

	return routes, nil
}
