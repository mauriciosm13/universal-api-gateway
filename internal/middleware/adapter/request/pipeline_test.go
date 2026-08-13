package request

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/mauriciomendonca/universal-api-gateway/internal/domain"
	"github.com/mauriciomendonca/universal-api-gateway/internal/middleware/port"
)

func TestPipelineContinues(t *testing.T) {
	t.Parallel()

	pipeline := NewPipeline(ContinueHandler, NewErrorMiddleware())

	resp, err := pipeline.Execute(context.Background(), domain.Request{Path: "/api"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if resp.StatusCode != ContinueStatus {
		t.Fatalf("status = %d, want continue", resp.StatusCode)
	}
}

func TestRateLimitMiddlewareBlocks(t *testing.T) {
	t.Parallel()

	pipeline := NewPipeline(
		ContinueHandler,
		NewRateLimitMiddleware(blockLimiter{}),
	)

	resp, err := pipeline.Execute(context.Background(), domain.Request{Host: "localhost", Path: "/api"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if resp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want 429", resp.StatusCode)
	}
}

type blockLimiter struct{}

func (blockLimiter) Allow(context.Context, string) (bool, error) {
	return false, nil
}

func TestRateLimitMiddlewareError(t *testing.T) {
	t.Parallel()

	pipeline := NewPipeline(
		ContinueHandler,
		NewRateLimitMiddleware(errorLimiter{}),
	)

	resp, err := pipeline.Execute(context.Background(), domain.Request{Host: "localhost", Path: "/api"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

type errorLimiter struct{}

func (errorLimiter) Allow(context.Context, string) (bool, error) {
	return false, errors.New("limiter unavailable")
}

func TestErrorMiddlewareMapsError(t *testing.T) {
	t.Parallel()

	pipeline := NewPipeline(
		ContinueHandler,
		NewErrorMiddleware(),
		failingMiddleware{},
	)

	resp, err := pipeline.Execute(context.Background(), domain.Request{})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

type failingMiddleware struct{}

func (failingMiddleware) Wrap(next port.Handler) port.Handler {
	return func(context.Context, domain.Request) (domain.Response, error) {
		return domain.Response{}, errors.New("boom")
	}
}
