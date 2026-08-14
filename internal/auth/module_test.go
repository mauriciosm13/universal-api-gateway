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

func TestBuildAuthenticatorUsesNoOpWhenJWTDisabled(t *testing.T) {
	auth, err := buildAuthenticator(config.Config{})
	if err != nil {
		t.Fatalf("buildAuthenticator() error = %v", err)
	}

	noop := authstub.NewNoOpAuthenticator()
	want, err := noop.Authenticate(context.Background(), "")
	if err != nil {
		t.Fatalf("noop Authenticate() error = %v", err)
	}

	got, err := auth.Authenticate(context.Background(), "")
	if err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}

	if got.Subject != want.Subject {
		t.Fatalf("subject = %q, want %q", got.Subject, want.Subject)
	}
}

func TestBuildAuthenticatorUsesJWTWhenHMACConfigured(t *testing.T) {
	auth, err := buildAuthenticator(config.Config{
		JWT: config.JWTConfig{
			HMACSecret: strings.Repeat("a", 32),
		},
	})
	if err != nil {
		t.Fatalf("buildAuthenticator() error = %v", err)
	}

	if _, err := auth.Authenticate(context.Background(), ""); err == nil {
		t.Fatal("expected JWT authenticator to reject empty token")
	}
}
