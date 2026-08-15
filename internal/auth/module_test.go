package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	authstub "github.com/mauriciomendonca/universal-api-gateway/internal/auth/stub"
	"github.com/mauriciomendonca/universal-api-gateway/internal/config"
	"github.com/mauriciomendonca/universal-api-gateway/internal/di"
	"github.com/mauriciomendonca/universal-api-gateway/internal/domain"
)

func TestModuleRegisterPanicsOnInvalidJWKS(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"keys":[]}`))
	}))
	t.Cleanup(server.Close)

	defer func() {
		if recover() == nil {
			t.Fatal("Register() did not panic on invalid JWKS")
		}
	}()

	di.NewBuilder(config.Config{
		JWT: config.JWTConfig{JWKSURL: server.URL},
	}, Module{})
}

func TestBuildRequestAuthenticatorUsesNoOpWhenAuthDisabled(t *testing.T) {
	auth, err := buildRequestAuthenticator(config.Config{})
	if err != nil {
		t.Fatalf("buildRequestAuthenticator() error = %v", err)
	}

	noop := authstub.NewNoOpAuthenticator()
	want, err := noop.AuthenticateRequest(context.Background(), domain.Request{})
	if err != nil {
		t.Fatalf("noop AuthenticateRequest() error = %v", err)
	}

	got, err := auth.AuthenticateRequest(context.Background(), domain.Request{})
	if err != nil {
		t.Fatalf("AuthenticateRequest() error = %v", err)
	}

	if got.Subject != want.Subject {
		t.Fatalf("subject = %q, want %q", got.Subject, want.Subject)
	}
}

func TestBuildRequestAuthenticatorUsesJWTWhenHMACConfigured(t *testing.T) {
	auth, err := buildRequestAuthenticator(config.Config{
		JWT: config.JWTConfig{
			HMACSecret: strings.Repeat("a", 32),
		},
	})
	if err != nil {
		t.Fatalf("buildRequestAuthenticator() error = %v", err)
	}

	if _, err := auth.AuthenticateRequest(context.Background(), domain.Request{}); err == nil {
		t.Fatal("expected JWT authenticator to reject missing token")
	}
}

func TestBuildRequestAuthenticatorUsesAPIKeyWhenConfigured(t *testing.T) {
	auth, err := buildRequestAuthenticator(config.Config{
		APIKeys: config.APIKeyConfig{
			Keys:       map[string]string{"k1": "n1"},
			HeaderName: "X-API-Key",
			QueryParam: "api_key",
		},
	})
	if err != nil {
		t.Fatalf("buildRequestAuthenticator() error = %v", err)
	}

	if _, err := auth.AuthenticateRequest(context.Background(), domain.Request{}); err == nil {
		t.Fatal("expected API key authenticator to reject missing key")
	}
}

func TestBuildRequestAuthenticatorUsesCompositeWhenBothConfigured(t *testing.T) {
	auth, err := buildRequestAuthenticator(config.Config{
		JWT: config.JWTConfig{
			HMACSecret: strings.Repeat("a", 32),
		},
		APIKeys: config.APIKeyConfig{
			Keys:       map[string]string{"k1": "n1"},
			HeaderName: "X-API-Key",
			QueryParam: "api_key",
		},
	})
	if err != nil {
		t.Fatalf("buildRequestAuthenticator() error = %v", err)
	}

	if _, err := auth.AuthenticateRequest(context.Background(), domain.Request{
		Headers: map[string][]string{"X-API-Key": {"k1"}},
	}); err != nil {
		t.Fatalf("AuthenticateRequest() error = %v", err)
	}
}
