package request

import (
	"context"
	"errors"
	"net/http"
	"testing"

	authctx "github.com/mauriciomendonca/universal-api-gateway/internal/auth/context"
	authport "github.com/mauriciomendonca/universal-api-gateway/internal/auth/port"
	"github.com/mauriciomendonca/universal-api-gateway/internal/domain"
)

func TestAuthMiddlewareAllowsNoOpAuthenticator(t *testing.T) {
	t.Parallel()

	pipeline := NewPipeline(
		ContinueHandler,
		NewAuthMiddleware(stubRequestAuthenticator{}),
	)

	_, resp, err := pipeline.Execute(context.Background(), domain.Request{})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if resp.StatusCode != ContinueStatus {
		t.Fatalf("status = %d, want continue", resp.StatusCode)
	}
}

func TestAuthMiddlewareStoresIdentityInContext(t *testing.T) {
	t.Parallel()

	pipeline := NewPipeline(
		ContinueHandler,
		NewAuthMiddleware(stubRequestAuthenticator{}),
	)

	ctx, _, err := pipeline.Execute(context.Background(), domain.Request{})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	id, ok := authctx.IdentityFrom(ctx)
	if !ok || id.Subject != "anonymous" {
		t.Fatalf("identity = %+v, ok = %v", id, ok)
	}
}

func TestAuthMiddlewareRejectsInvalidToken(t *testing.T) {
	t.Parallel()

	pipeline := NewPipeline(
		ContinueHandler,
		NewAuthMiddleware(rejectRequestAuthenticator{}),
	)

	_, resp, err := pipeline.Execute(context.Background(), domain.Request{
		Headers: map[string][]string{"Authorization": {"Bearer bad"}},
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
}

func TestAuthMiddlewareRejectsMissingToken(t *testing.T) {
	t.Parallel()

	pipeline := NewPipeline(
		ContinueHandler,
		NewAuthMiddleware(rejectRequestAuthenticator{}),
	)

	_, resp, err := pipeline.Execute(context.Background(), domain.Request{})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
}

func TestAuthMiddlewareRejectsMalformedAuthorizationHeader(t *testing.T) {
	t.Parallel()

	pipeline := NewPipeline(
		ContinueHandler,
		NewAuthMiddleware(rejectRequestAuthenticator{}),
	)

	_, resp, err := pipeline.Execute(context.Background(), domain.Request{
		Headers: map[string][]string{"Authorization": {"Basic dGVzdA=="}},
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
}

type stubRequestAuthenticator struct{}

func (stubRequestAuthenticator) AuthenticateRequest(context.Context, domain.Request) (authport.Identity, error) {
	return authport.Identity{Subject: "anonymous"}, nil
}

type rejectRequestAuthenticator struct{}

func (rejectRequestAuthenticator) AuthenticateRequest(_ context.Context, req domain.Request) (authport.Identity, error) {
	if len(req.Headers["Authorization"]) == 0 {
		return authport.Identity{}, errors.New("missing token")
	}

	return authport.Identity{}, errors.New("invalid token")
}
