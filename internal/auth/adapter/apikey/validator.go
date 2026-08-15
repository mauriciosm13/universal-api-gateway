package apikey

import (
	"context"
	"errors"

	authport "github.com/mauriciomendonca/universal-api-gateway/internal/auth/port"
	"github.com/mauriciomendonca/universal-api-gateway/internal/config"
)

var errInvalidKey = errors.New("invalid api key")

// Validator validates API keys against configured allowlist.
type Validator struct {
	keys map[string]string
}

// NewValidator builds an API key authenticator from configuration.
func NewValidator(cfg config.APIKeyConfig) (*Validator, error) {
	if !cfg.Enabled() {
		return nil, errors.New("apikey: configuration is disabled")
	}

	keys := make(map[string]string, len(cfg.Keys))
	for key, name := range cfg.Keys {
		keys[key] = name
	}

	return &Validator{keys: keys}, nil
}

// Authenticate validates key and returns caller identity.
func (v *Validator) Authenticate(_ context.Context, key string) (authport.Identity, error) {
	if key == "" {
		return authport.Identity{}, errInvalidKey
	}

	name, ok := v.keys[key]
	if !ok {
		return authport.Identity{}, errInvalidKey
	}

	return authport.Identity{
		Subject: name,
		Claims: map[string]string{
			"source": "apikey",
		},
	}, nil
}
