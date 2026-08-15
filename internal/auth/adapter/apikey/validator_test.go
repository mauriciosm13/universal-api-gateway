package apikey

import (
	"context"
	"testing"

	"github.com/mauriciomendonca/universal-api-gateway/internal/config"
)

func TestValidatorValidKey(t *testing.T) {
	t.Parallel()

	validator, err := NewValidator(config.APIKeyConfig{
		Keys: map[string]string{"dev-key-abc": "local-dev"},
	})
	if err != nil {
		t.Fatalf("NewValidator() error = %v", err)
	}

	id, err := validator.Authenticate(context.Background(), "dev-key-abc")
	if err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}

	if id.Subject != "local-dev" || id.Claims["source"] != "apikey" {
		t.Fatalf("identity = %+v", id)
	}
}

func TestValidatorInvalidKey(t *testing.T) {
	t.Parallel()

	validator, err := NewValidator(config.APIKeyConfig{
		Keys: map[string]string{"dev-key-abc": "local-dev"},
	})
	if err != nil {
		t.Fatalf("NewValidator() error = %v", err)
	}

	if _, err := validator.Authenticate(context.Background(), "wrong"); err == nil {
		t.Fatal("Authenticate() error = nil, want error")
	}
}

func TestValidatorEmptyKey(t *testing.T) {
	t.Parallel()

	validator, err := NewValidator(config.APIKeyConfig{
		Keys: map[string]string{"dev-key-abc": "local-dev"},
	})
	if err != nil {
		t.Fatalf("NewValidator() error = %v", err)
	}

	if _, err := validator.Authenticate(context.Background(), ""); err == nil {
		t.Fatal("Authenticate() error = nil, want error")
	}
}

func TestNewValidatorRequiresConfig(t *testing.T) {
	t.Parallel()

	if _, err := NewValidator(config.APIKeyConfig{}); err == nil {
		t.Fatal("NewValidator() error = nil, want error")
	}
}
