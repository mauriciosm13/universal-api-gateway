package stub

import (
	"context"
	"testing"

	"github.com/mauriciomendonca/universal-api-gateway/internal/domain"
)

func TestNoOpAuthenticatorReturnsAnonymousIdentity(t *testing.T) {
	t.Parallel()

	auth := NewNoOpAuthenticator()
	identity, err := auth.Authenticate(context.Background(), "any-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if identity.Subject != "anonymous" {
		t.Fatalf("expected anonymous subject, got %q", identity.Subject)
	}
}

func TestNoOpAuthenticatorRequestReturnsAnonymousIdentity(t *testing.T) {
	t.Parallel()

	auth := NewNoOpAuthenticator()
	identity, err := auth.AuthenticateRequest(context.Background(), domain.Request{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if identity.Subject != "anonymous" {
		t.Fatalf("expected anonymous subject, got %q", identity.Subject)
	}
}
