package config

import (
	"fmt"
	"strings"
)

// HeaderRoute maps a header name/value pair to an upstream URL.
type HeaderRoute struct {
	Name     string
	Value    string
	Upstream string
}

// ParseHeaderRoutes parses comma-separated HeaderName=HeaderValue=upstream entries.
func ParseHeaderRoutes(raw string) ([]HeaderRoute, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	entries := strings.Split(raw, ",")
	routes := make([]HeaderRoute, 0, len(entries))

	for _, entry := range entries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}

		name, rest, ok := strings.Cut(entry, "=")
		if !ok {
			return nil, fmt.Errorf("entry %q: expected HeaderName=HeaderValue=upstream", entry)
		}

		value, upstream, ok := strings.Cut(rest, "=")
		if !ok {
			return nil, fmt.Errorf("entry %q: expected HeaderName=HeaderValue=upstream", entry)
		}

		name = strings.TrimSpace(name)
		value = strings.TrimSpace(value)
		upstream = strings.TrimSpace(upstream)

		if name == "" {
			return nil, fmt.Errorf("entry %q: header name is required", entry)
		}
		if value == "" {
			return nil, fmt.Errorf("entry %q: header value is required", entry)
		}
		if upstream == "" {
			return nil, fmt.Errorf("entry %q: upstream is required", entry)
		}
		if err := ValidateUpstreamURL(upstream); err != nil {
			return nil, fmt.Errorf("entry %q: %w", entry, err)
		}

		routes = append(routes, HeaderRoute{
			Name:     name,
			Value:    value,
			Upstream: upstream,
		})
	}

	if len(routes) == 0 {
		return nil, fmt.Errorf("GATEWAY_HEADER_ROUTES contains no valid entries")
	}

	return routes, nil
}
