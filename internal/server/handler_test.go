package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mauriciomendonca/universal-api-gateway/internal/config"
	routingstub "github.com/mauriciomendonca/universal-api-gateway/internal/routing/stub"
)

func TestRootHandlerDispatchesHealth(t *testing.T) {
	t.Parallel()

	deps := Dependencies{
		Config: config.Config{
			ReadTimeout:  time.Second,
			WriteTimeout: time.Second,
			IdleTimeout:  time.Second,
		},
		Router: routingstub.NewNoOpRouter(),
	}

	handler := newRootHandler(deps)

	for _, path := range []string{"/health", "/health/live", "/health/ready"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("GET %q status = %d, want 200", path, rec.Code)
		}
	}
}

func TestRootHandlerNonHealthUsesGateway(t *testing.T) {
	t.Parallel()

	deps := Dependencies{
		Config: config.Config{
			ReadTimeout:  time.Second,
			WriteTimeout: time.Second,
			IdleTimeout:  time.Second,
		},
		Router: routingstub.NewNoOpRouter(),
	}

	handler := newRootHandler(deps)

	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestRootHandlerNonGetHealthUsesGateway(t *testing.T) {
	t.Parallel()

	deps := Dependencies{
		Config: config.Config{
			ReadTimeout:  time.Second,
			WriteTimeout: time.Second,
			IdleTimeout:  time.Second,
		},
		Router: routingstub.NewNoOpRouter(),
	}

	handler := newRootHandler(deps)

	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("POST /health status = %d, want 404 via gateway", rec.Code)
	}
}
