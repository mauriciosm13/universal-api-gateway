package config

import (
	"fmt"
	"strings"
)

// ParseUpstreamList splits comma-separated upstream URLs and validates each.
func ParseUpstreamList(raw string) ([]string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("upstream list is required")
	}

	parts := strings.Split(raw, ",")
	upstreams := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if err := ValidateUpstreamURL(part); err != nil {
			return nil, err
		}
		upstreams = append(upstreams, part)
	}

	if len(upstreams) == 0 {
		return nil, fmt.Errorf("upstream list is required")
	}

	return upstreams, nil
}

func splitRouteEntries(raw string, isNewEntry func(segment string) bool) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	parts := strings.Split(raw, ",")
	entries := make([]string, 0)
	var current string

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if current == "" {
			current = part
			continue
		}
		if isNewEntry(part) {
			entries = append(entries, current)
			current = part
		} else {
			current = current + "," + part
		}
	}

	if current != "" {
		entries = append(entries, current)
	}

	return entries
}

func isPathRouteEntry(segment string) bool {
	return strings.HasPrefix(segment, "/") && strings.Contains(segment, "=")
}

func isHostRouteEntry(segment string) bool {
	host, _, ok := strings.Cut(segment, "=")
	if !ok {
		return false
	}

	host = strings.TrimSpace(host)
	if host == "" {
		return false
	}

	return !strings.Contains(host, "://")
}

func isMethodRouteEntry(segment string) bool {
	method, _, ok := strings.Cut(segment, "=")
	if !ok {
		return false
	}

	return isValidHTTPMethod(strings.TrimSpace(strings.ToUpper(method)))
}

func isHeaderRouteEntry(segment string) bool {
	return strings.Count(segment, "=") >= 2
}
