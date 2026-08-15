package config

import (
	"strings"
	"testing"
)

func TestParseAPIKeysCSV(t *testing.T) {
	t.Parallel()

	keys, err := ParseAPIKeys("dev-key:local-dev,prod-key:partner-a")
	if err != nil {
		t.Fatalf("ParseAPIKeys() error = %v", err)
	}

	if keys["dev-key"] != "local-dev" || keys["prod-key"] != "partner-a" {
		t.Fatalf("keys = %v", keys)
	}
}

func TestParseAPIKeysCSVWithoutNames(t *testing.T) {
	t.Parallel()

	keys, err := ParseAPIKeys("solo-key")
	if err != nil {
		t.Fatalf("ParseAPIKeys() error = %v", err)
	}

	if keys["solo-key"] != "solo-key" {
		t.Fatalf("keys = %v", keys)
	}
}

func TestParseAPIKeysJSONObject(t *testing.T) {
	t.Parallel()

	keys, err := ParseAPIKeys(`{"dev-key":"local-dev","prod-key":"partner-a"}`)
	if err != nil {
		t.Fatalf("ParseAPIKeys() error = %v", err)
	}

	if len(keys) != 2 {
		t.Fatalf("len(keys) = %d, want 2", len(keys))
	}
}

func TestParseAPIKeysJSONArray(t *testing.T) {
	t.Parallel()

	keys, err := ParseAPIKeys(`[{"key":"dev-key","name":"local-dev"}]`)
	if err != nil {
		t.Fatalf("ParseAPIKeys() error = %v", err)
	}

	if keys["dev-key"] != "local-dev" {
		t.Fatalf("keys = %v", keys)
	}
}

func TestParseAPIKeysDuplicateRejected(t *testing.T) {
	t.Parallel()

	if _, err := ParseAPIKeys("a:one,a:two"); err == nil {
		t.Fatal("ParseAPIKeys() error = nil, want duplicate error")
	}
}

func TestParseAPIKeysEmptyKeyRejected(t *testing.T) {
	t.Parallel()

	if _, err := ParseAPIKeys(":name"); err == nil {
		t.Fatal("ParseAPIKeys() error = nil, want empty key error")
	}
}

func TestAPIKeyConfigEnabled(t *testing.T) {
	t.Parallel()

	cfg := APIKeyConfig{Keys: map[string]string{"k": "n"}}
	if !cfg.Enabled() {
		t.Fatal("Enabled() = false, want true")
	}
}

func TestLoadAPIKeysDisabled(t *testing.T) {
	t.Setenv("GATEWAY_API_KEYS", "")

	cfg, err := loadAPIKeys()
	if err != nil {
		t.Fatalf("loadAPIKeys() error = %v", err)
	}

	if cfg.Enabled() || cfg.HeaderName != defaultAPIKeyHeader || cfg.QueryParam != defaultAPIKeyQuery {
		t.Fatalf("cfg = %+v", cfg)
	}
}

func TestLoadAPIKeysFromEnv(t *testing.T) {
	t.Setenv("GATEWAY_API_KEYS", "dev-key:local-dev")
	t.Setenv("GATEWAY_API_KEY_HEADER", "X-Custom-Key")
	t.Setenv("GATEWAY_API_KEY_QUERY", "key")

	cfg, err := loadAPIKeys()
	if err != nil {
		t.Fatalf("loadAPIKeys() error = %v", err)
	}

	if !cfg.Enabled() || cfg.Keys["dev-key"] != "local-dev" {
		t.Fatalf("cfg = %+v", cfg)
	}
	if cfg.HeaderName != "X-Custom-Key" || cfg.QueryParam != "key" {
		t.Fatalf("header/query = %q / %q", cfg.HeaderName, cfg.QueryParam)
	}
}

func TestLoadAPIKeysRejectsEmptyHeader(t *testing.T) {
	t.Setenv("GATEWAY_API_KEYS", "k:n")
	t.Setenv("GATEWAY_API_KEY_HEADER", "   ")

	_, err := loadAPIKeys()
	if err == nil {
		t.Fatal("expected error for empty header name")
	}
}

func TestLoadAPIKeysRejectsInvalidKeys(t *testing.T) {
	t.Setenv("GATEWAY_API_KEYS", `{invalid`)

	_, err := loadAPIKeys()
	if err == nil {
		t.Fatal("expected error for invalid GATEWAY_API_KEYS")
	}
}

func TestConfigAuthEnabled(t *testing.T) {
	t.Parallel()

	if (Config{}).AuthEnabled() {
		t.Fatal("expected auth disabled on empty config")
	}

	jwtCfg := Config{JWT: JWTConfig{HMACSecret: strings.Repeat("a", 32)}}
	if !jwtCfg.AuthEnabled() {
		t.Fatal("expected JWT to enable auth")
	}

	apiCfg := Config{APIKeys: APIKeyConfig{Keys: map[string]string{"k": "n"}}}
	if !apiCfg.AuthEnabled() {
		t.Fatal("expected API keys to enable auth")
	}
}

func TestParseAPIKeysEmptyJSONArrayRejected(t *testing.T) {
	t.Parallel()

	if _, err := ParseAPIKeys(`[]`); err == nil {
		t.Fatal("expected error for empty JSON array")
	}
}

func TestParseAPIKeysEmptyJSONObjectRejected(t *testing.T) {
	t.Parallel()

	if _, err := ParseAPIKeys(`{}`); err == nil {
		t.Fatal("expected error for empty JSON object")
	}
}
