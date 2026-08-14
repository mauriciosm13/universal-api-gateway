package config

import (
	"strings"
	"testing"
)

func TestJWTConfigDisabledByDefault(t *testing.T) {
	t.Setenv("GATEWAY_JWT_JWKS_URL", "")
	t.Setenv("GATEWAY_JWT_HMAC_SECRET", "")
	t.Setenv("GATEWAY_JWT_ISSUER", "")
	t.Setenv("GATEWAY_JWT_AUDIENCE", "")
	t.Setenv("OTEL_TRACES_EXPORTER", "none")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.JWT.Enabled() {
		t.Fatal("expected JWT disabled when env vars unset")
	}
}

func TestLoadJWTHMACValid(t *testing.T) {
	secret := strings.Repeat("a", minHMACSecretLength)
	t.Setenv("GATEWAY_JWT_HMAC_SECRET", secret)
	t.Setenv("GATEWAY_JWT_ISSUER", "https://issuer.example.com")
	t.Setenv("GATEWAY_JWT_AUDIENCE", "my-api")
	t.Setenv("GATEWAY_JWT_JWKS_URL", "")
	t.Setenv("OTEL_TRACES_EXPORTER", "none")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if !cfg.JWT.Enabled() {
		t.Fatal("expected JWT enabled")
	}

	if cfg.JWT.HMACSecret != secret {
		t.Fatalf("HMAC secret mismatch")
	}

	if cfg.JWT.Issuer != "https://issuer.example.com" || cfg.JWT.Audience != "my-api" {
		t.Fatalf("unexpected issuer/audience: %+v", cfg.JWT)
	}
}

func TestLoadJWKSValid(t *testing.T) {
	t.Setenv("GATEWAY_JWT_JWKS_URL", "https://issuer.example.com/.well-known/jwks.json")
	t.Setenv("GATEWAY_JWT_HMAC_SECRET", "")
	t.Setenv("OTEL_TRACES_EXPORTER", "none")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if !cfg.JWT.Enabled() || cfg.JWT.JWKSURL == "" {
		t.Fatalf("expected JWKS config, got %+v", cfg.JWT)
	}
}

func TestLoadJWTMutuallyExclusive(t *testing.T) {
	t.Setenv("GATEWAY_JWT_JWKS_URL", "https://issuer.example.com/.well-known/jwks.json")
	t.Setenv("GATEWAY_JWT_HMAC_SECRET", strings.Repeat("a", minHMACSecretLength))
	t.Setenv("OTEL_TRACES_EXPORTER", "none")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error when both JWKS and HMAC are set")
	}
}

func TestLoadHMACSecretTooShort(t *testing.T) {
	t.Setenv("GATEWAY_JWT_HMAC_SECRET", "short")
	t.Setenv("GATEWAY_JWT_JWKS_URL", "")
	t.Setenv("OTEL_TRACES_EXPORTER", "none")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for short HMAC secret")
	}
}

func TestLoadJWKSInvalidURL(t *testing.T) {
	t.Setenv("GATEWAY_JWT_JWKS_URL", "not-a-url")
	t.Setenv("GATEWAY_JWT_HMAC_SECRET", "")
	t.Setenv("OTEL_TRACES_EXPORTER", "none")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid JWKS URL")
	}
}

func TestValidateJWKSURL(t *testing.T) {
	t.Parallel()

	if err := validateJWKSURL("https://issuer.example.com/jwks.json"); err != nil {
		t.Fatalf("valid URL: %v", err)
	}

	if err := validateJWKSURL("ftp://issuer.example.com/jwks.json"); err == nil {
		t.Fatal("expected error for ftp scheme")
	}
}
