package composite

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

func TestCompositeJWTValid(t *testing.T) {
	t.Parallel()

	auth := newComposite(t)

	resp, err := auth.AuthenticateRequest(context.Background(), domain.Request{
		Headers: map[string][]string{"Authorization": {"Bearer " + signHS256Token(t, "user-jwt")}},
	})
	if err != nil {
		t.Fatalf("AuthenticateRequest() error = %v", err)
	}

	if resp.Subject != "user-jwt" {
		t.Fatalf("subject = %q, want user-jwt", resp.Subject)
	}
}

func TestCompositeAPIKeyValid(t *testing.T) {
	t.Parallel()

	auth := newComposite(t)

	resp, err := auth.AuthenticateRequest(context.Background(), domain.Request{
		Headers: map[string][]string{"X-API-Key": {"dev-key-abc"}},
	})
	if err != nil {
		t.Fatalf("AuthenticateRequest() error = %v", err)
	}

	if resp.Subject != "local-dev" {
		t.Fatalf("subject = %q, want local-dev", resp.Subject)
	}
}

func TestCompositeBothFail(t *testing.T) {
	t.Parallel()

	auth := newComposite(t)

	if _, err := auth.AuthenticateRequest(context.Background(), domain.Request{}); err == nil {
		t.Fatal("AuthenticateRequest() error = nil, want error")
	}
}

func TestCompositeMalformedBearerRejected(t *testing.T) {
	t.Parallel()

	auth := newComposite(t)

	_, err := auth.AuthenticateRequest(context.Background(), domain.Request{
		Headers: map[string][]string{"Authorization": {"Basic dGVzdA=="}},
	})
	if err == nil {
		t.Fatal("AuthenticateRequest() error = nil, want error")
	}
}

func TestCompositeInvalidJWTDoesNotFallThroughToAPIKey(t *testing.T) {
	t.Parallel()

	auth := newComposite(t)

	_, err := auth.AuthenticateRequest(context.Background(), domain.Request{
		Headers: map[string][]string{
			"Authorization": {"Bearer bad-token"},
			"X-API-Key":     {"dev-key-abc"},
		},
	})
	if err == nil {
		t.Fatal("AuthenticateRequest() error = nil, want error")
	}
}

func newComposite(t *testing.T) *Authenticator {
	t.Helper()

	jwt, err := jwtadapter.NewValidator(config.JWTConfig{HMACSecret: testHMACSecret})
	if err != nil {
		t.Fatalf("NewValidator() error = %v", err)
	}

	apiKey, err := apikeyadapter.NewValidator(config.APIKeyConfig{
		Keys:       map[string]string{"dev-key-abc": "local-dev"},
		HeaderName: "X-API-Key",
		QueryParam: "api_key",
	})
	if err != nil {
		t.Fatalf("NewValidator() error = %v", err)
	}

	return New(jwt, apiKey, config.APIKeyConfig{
		HeaderName: "X-API-Key",
		QueryParam: "api_key",
	})
}

func signHS256Token(t *testing.T, subject string) string {
	t.Helper()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": subject,
		"exp": time.Now().Add(time.Hour).Unix(),
	})

	signed, err := token.SignedString([]byte(testHMACSecret))
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}

	return signed
}
