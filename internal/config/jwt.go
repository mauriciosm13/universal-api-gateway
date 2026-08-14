package config

import (
	"fmt"
	"net/url"
	"os"
)

const minHMACSecretLength = 32

// JWTConfig holds JWT authentication settings from environment variables.
type JWTConfig struct {
	JWKSURL    string
	HMACSecret string
	Issuer     string
	Audience   string
}

// Enabled reports whether JWT authentication is configured.
func (c JWTConfig) Enabled() bool {
	return c.JWKSURL != "" || c.HMACSecret != ""
}

func loadJWT() (JWTConfig, error) {
	jwksURL := os.Getenv("GATEWAY_JWT_JWKS_URL")
	hmacSecret := os.Getenv("GATEWAY_JWT_HMAC_SECRET")

	if jwksURL == "" && hmacSecret == "" {
		return JWTConfig{}, nil
	}

	if jwksURL != "" && hmacSecret != "" {
		return JWTConfig{}, fmt.Errorf("GATEWAY_JWT_JWKS_URL and GATEWAY_JWT_HMAC_SECRET are mutually exclusive")
	}

	if jwksURL != "" {
		if err := validateJWKSURL(jwksURL); err != nil {
			return JWTConfig{}, fmt.Errorf("invalid GATEWAY_JWT_JWKS_URL: %w", err)
		}
	}

	if hmacSecret != "" {
		if len(hmacSecret) < minHMACSecretLength {
			return JWTConfig{}, fmt.Errorf("GATEWAY_JWT_HMAC_SECRET must be at least %d characters", minHMACSecretLength)
		}
	}

	return JWTConfig{
		JWKSURL:    jwksURL,
		HMACSecret: hmacSecret,
		Issuer:     os.Getenv("GATEWAY_JWT_ISSUER"),
		Audience:   os.Getenv("GATEWAY_JWT_AUDIENCE"),
	}, nil
}

func validateJWKSURL(raw string) error {
	parsed, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("parse URL: %w", err)
	}

	switch parsed.Scheme {
	case "http", "https":
	default:
		return fmt.Errorf("scheme must be http or https, got %q", parsed.Scheme)
	}

	if parsed.Host == "" {
		return fmt.Errorf("host is required")
	}

	return nil
}
