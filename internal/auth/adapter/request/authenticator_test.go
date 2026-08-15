package request

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	apikeyadapter "github.com/mauriciomendonca/universal-api-gateway/internal/auth/adapter/apikey"
	jwtadapter "github.com/mauriciomendonca/universal-api-gateway/internal/auth/adapter/jwt"
	"github.com/mauriciomendonca/universal-api-gateway/internal/config"
	"github.com/mauriciomendonca/universal-api-gateway/internal/domain"
)

const testHMACSecret = "dev-secret-at-least-32-chars-long"

func TestJWTAuthenticatorValidBearer(t *testing.T) {
	t.Parallel()

	validator, err := jwtadapter.NewValidator(config.JWTConfig{HMACSecret: testHMACSecret})
	if err != nil {
		t.Fatalf("NewValidator() error = %v", err)
	}

	auth := NewJWTAuthenticator(validator)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "jwt-user",
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	signed, err := token.SignedString([]byte(testHMACSecret))
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}

	id, err := auth.AuthenticateRequest(context.Background(), domain.Request{
		Headers: map[string][]string{"Authorization": {"Bearer " + signed}},
	})
	if err != nil {
		t.Fatalf("AuthenticateRequest() error = %v", err)
	}

	if id.Subject != "jwt-user" {
		t.Fatalf("subject = %q, want jwt-user", id.Subject)
	}
}

func TestJWTAuthenticatorMissingBearer(t *testing.T) {
	t.Parallel()

	validator, err := jwtadapter.NewValidator(config.JWTConfig{HMACSecret: testHMACSecret})
	if err != nil {
		t.Fatalf("NewValidator() error = %v", err)
	}

	auth := NewJWTAuthenticator(validator)
	if _, err := auth.AuthenticateRequest(context.Background(), domain.Request{}); err == nil {
		t.Fatal("AuthenticateRequest() error = nil, want error")
	}
}

func TestAPIKeyAuthenticatorHeader(t *testing.T) {
	t.Parallel()

	cfg := config.APIKeyConfig{
		Keys:       map[string]string{"k1": "name1"},
		HeaderName: "X-API-Key",
		QueryParam: "api_key",
	}

	validator, err := apikeyadapter.NewValidator(cfg)
	if err != nil {
		t.Fatalf("NewValidator() error = %v", err)
	}

	auth := NewAPIKeyAuthenticator(validator, cfg)
	id, err := auth.AuthenticateRequest(context.Background(), domain.Request{
		Headers: map[string][]string{"X-API-Key": {"k1"}},
	})
	if err != nil {
		t.Fatalf("AuthenticateRequest() error = %v", err)
	}

	if id.Subject != "name1" {
		t.Fatalf("subject = %q, want name1", id.Subject)
	}
}

func TestAPIKeyAuthenticatorQuery(t *testing.T) {
	t.Parallel()

	cfg := config.APIKeyConfig{
		Keys:       map[string]string{"k1": "name1"},
		HeaderName: "X-API-Key",
		QueryParam: "api_key",
	}

	validator, err := apikeyadapter.NewValidator(cfg)
	if err != nil {
		t.Fatalf("NewValidator() error = %v", err)
	}

	auth := NewAPIKeyAuthenticator(validator, cfg)
	id, err := auth.AuthenticateRequest(context.Background(), domain.Request{
		Query: map[string][]string{"api_key": {"k1"}},
	})
	if err != nil {
		t.Fatalf("AuthenticateRequest() error = %v", err)
	}

	if id.Subject != "name1" {
		t.Fatalf("subject = %q, want name1", id.Subject)
	}
}
