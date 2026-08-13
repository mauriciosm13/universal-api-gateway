package request

import (
	"context"
	"errors"
	"net/http"
	"testing"

	authport "github.com/mauriciomendonca/universal-api-gateway/internal/auth/port"
	"github.com/mauriciomendonca/universal-api-gateway/internal/domain"
)

func TestAuthMiddlewareAllowsNoOpAuthenticator(t *testing.T) {
	t.Parallel()

	pipeline := NewPipeline(
		ContinueHandler,
		NewAuthMiddleware(stubAuthenticator{}),
	)

	resp, err := pipeline.Execute(context.Background(), domain.Request{})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if resp.StatusCode != ContinueStatus {
		t.Fatalf("status = %d, want continue", resp.StatusCode)
	}
}

func TestAuthMiddlewareRejectsInvalidToken(t *testing.T) {
	t.Parallel()

	pipeline := NewPipeline(
		ContinueHandler,
		NewAuthMiddleware(rejectAuthenticator{}),
	)

	resp, err := pipeline.Execute(context.Background(), domain.Request{
		Headers: map[string][]string{"Authorization": {"Bearer bad"}},
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
}

type stubAuthenticator struct{}

func (stubAuthenticator) Authenticate(context.Context, string) (authport.Identity, error) {
	return authport.Identity{Subject: "anonymous"}, nil
}

type rejectAuthenticator struct{}

func (rejectAuthenticator) Authenticate(context.Context, string) (authport.Identity, error) {
	return authport.Identity{}, errors.New("invalid token")
}
