package config

import (
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
