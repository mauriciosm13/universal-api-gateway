package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

const (
	defaultAPIKeyHeader = "X-API-Key"
	defaultAPIKeyQuery  = "api_key"
)

// APIKeyConfig holds API key authentication settings from environment variables.
type APIKeyConfig struct {
	Keys       map[string]string
	HeaderName string
	QueryParam string
}

// Enabled reports whether API key authentication is configured.
func (c APIKeyConfig) Enabled() bool {
	return len(c.Keys) > 0
}

func loadAPIKeys() (APIKeyConfig, error) {
	raw := os.Getenv("GATEWAY_API_KEYS")
	if raw == "" {
		return APIKeyConfig{
			HeaderName: defaultAPIKeyHeader,
			QueryParam: defaultAPIKeyQuery,
		}, nil
	}

	keys, err := ParseAPIKeys(raw)
	if err != nil {
		return APIKeyConfig{}, fmt.Errorf("invalid GATEWAY_API_KEYS: %w", err)
	}

	if len(keys) == 0 {
		return APIKeyConfig{}, fmt.Errorf("GATEWAY_API_KEYS must contain at least one key")
	}

	headerName := envString("GATEWAY_API_KEY_HEADER", defaultAPIKeyHeader)
	if strings.TrimSpace(headerName) == "" {
		return APIKeyConfig{}, fmt.Errorf("GATEWAY_API_KEY_HEADER must not be empty")
	}

	queryParam := envString("GATEWAY_API_KEY_QUERY", defaultAPIKeyQuery)
	if strings.TrimSpace(queryParam) == "" {
		return APIKeyConfig{}, fmt.Errorf("GATEWAY_API_KEY_QUERY must not be empty")
	}

	return APIKeyConfig{
		Keys:       keys,
		HeaderName: headerName,
		QueryParam: queryParam,
	}, nil
}

// ParseAPIKeys parses comma-separated key:name pairs or JSON object/array.
func ParseAPIKeys(raw string) (map[string]string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	if strings.HasPrefix(raw, "{") || strings.HasPrefix(raw, "[") {
		return parseAPIKeysJSON(raw)
	}

	return parseAPIKeysCSV(raw)
}

func parseAPIKeysCSV(raw string) (map[string]string, error) {
	keys := make(map[string]string)
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		key, name, ok := strings.Cut(part, ":")
		key = strings.TrimSpace(key)
		if key == "" {
			return nil, fmt.Errorf("empty API key in %q", part)
		}

		if !ok || strings.TrimSpace(name) == "" {
			name = key
		} else {
			name = strings.TrimSpace(name)
		}

		if _, exists := keys[key]; exists {
			return nil, fmt.Errorf("duplicate API key %q", key)
		}

		keys[key] = name
	}

	if len(keys) == 0 {
		return nil, fmt.Errorf("no API keys parsed from %q", raw)
	}

	return keys, nil
}

func parseAPIKeysJSON(raw string) (map[string]string, error) {
	var object map[string]string
	if err := json.Unmarshal([]byte(raw), &object); err == nil {
		return normalizeAPIKeyMap(object)
	}

	var entries []struct {
		Key  string `json:"key"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal([]byte(raw), &entries); err != nil {
		return nil, fmt.Errorf("parse JSON: invalid object or array")
	}

	keys := make(map[string]string, len(entries))
	for _, entry := range entries {
		key := strings.TrimSpace(entry.Key)
		if key == "" {
			return nil, fmt.Errorf("empty API key in JSON array entry")
		}

		name := strings.TrimSpace(entry.Name)
		if name == "" {
			name = key
		}

		if _, exists := keys[key]; exists {
			return nil, fmt.Errorf("duplicate API key %q", key)
		}

		keys[key] = name
	}

	if len(keys) == 0 {
		return nil, fmt.Errorf("no API keys in JSON array")
	}

	return keys, nil
}

func normalizeAPIKeyMap(object map[string]string) (map[string]string, error) {
	if len(object) == 0 {
		return nil, fmt.Errorf("no API keys in JSON object")
	}

	keys := make(map[string]string, len(object))
	for key, name := range object {
		key = strings.TrimSpace(key)
		if key == "" {
			return nil, fmt.Errorf("empty API key in JSON object")
		}

		name = strings.TrimSpace(name)
		if name == "" {
			name = key
		}

		if _, exists := keys[key]; exists {
			return nil, fmt.Errorf("duplicate API key %q", key)
		}

		keys[key] = name
	}

	return keys, nil
}

// AuthEnabled reports whether any authentication mechanism is configured.
func (c Config) AuthEnabled() bool {
	return c.JWT.Enabled() || c.APIKeys.Enabled()
}
