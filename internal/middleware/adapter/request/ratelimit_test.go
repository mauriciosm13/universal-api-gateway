package request

import (
	"context"
	"testing"

	authctx "github.com/mauriciomendonca/universal-api-gateway/internal/auth/context"
	authport "github.com/mauriciomendonca/universal-api-gateway/internal/auth/port"
	"github.com/mauriciomendonca/universal-api-gateway/internal/domain"
)

type captureKeyLimiter struct {
	lastKey string
}

func (l *captureKeyLimiter) Allow(_ context.Context, key string) (bool, error) {
	l.lastKey = key
	return true, nil
}

func TestRateLimitKeyUsesIdentitySubject(t *testing.T) {
	t.Parallel()

	limiter := &captureKeyLimiter{}
	pipeline := NewPipeline(
		ContinueHandler,
		NewRateLimitMiddleware(limiter),
	)

	ctx := authctx.WithIdentity(context.Background(), authport.Identity{Subject: "user-123"})
	_, _, err := pipeline.Execute(ctx, domain.Request{Host: "localhost", Path: "/api"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if limiter.lastKey != "user-123" {
		t.Fatalf("key = %q, want user-123", limiter.lastKey)
	}
}

func TestRateLimitKeyIgnoresAnonymousIdentity(t *testing.T) {
	t.Parallel()

	limiter := &captureKeyLimiter{}
	pipeline := NewPipeline(
		ContinueHandler,
		NewRateLimitMiddleware(limiter),
	)

	ctx := authctx.WithIdentity(context.Background(), authport.Identity{Subject: "anonymous"})
	_, _, err := pipeline.Execute(ctx, domain.Request{Host: "localhost", Path: "/api"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if limiter.lastKey != "localhost/api" {
		t.Fatalf("key = %q, want localhost/api", limiter.lastKey)
	}
}

func TestRateLimitKeyFallbackHostPath(t *testing.T) {
	t.Parallel()

	limiter := &captureKeyLimiter{}
	pipeline := NewPipeline(
		ContinueHandler,
		NewRateLimitMiddleware(limiter),
	)

	_, _, err := pipeline.Execute(context.Background(), domain.Request{Host: "api.example.com", Path: "/v1/items"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if limiter.lastKey != "api.example.com/v1/items" {
		t.Fatalf("key = %q, want api.example.com/v1/items", limiter.lastKey)
	}
}
