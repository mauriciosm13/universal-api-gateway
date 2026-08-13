package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mauriciomendonca/universal-api-gateway/internal/config"
	routingstub "github.com/mauriciomendonca/universal-api-gateway/internal/routing/stub"
)

func TestRoutesDispatchesHealth(t *testing.T) {
	t.Parallel()

	deps := newTestDependencies(config.Config{
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second,
		IdleTimeout:  time.Second,
	}, routingstub.NewNoOpRouter())

	handler := newRoutes(deps)

	for _, path := range []string{"/health", "/health/live", "/health/ready"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("GET %q status = %d, want 200", path, rec.Code)
		}
	}
}

func TestRoutesNonHealthUsesGateway(t *testing.T) {
	t.Parallel()

	deps := newTestDependencies(config.Config{
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second,
		IdleTimeout:  time.Second,
	}, routingstub.NewNoOpRouter())

	handler := newRoutes(deps)

	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}

	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", ct)
	}

	var body errorBody
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	if body.Code != 404 || body.Message != "no route matched" {
		t.Fatalf("body = %+v", body)
	}
}

func TestRoutesNonGetHealthUsesGateway(t *testing.T) {
	t.Parallel()

	deps := newTestDependencies(config.Config{
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second,
		IdleTimeout:  time.Second,
	}, routingstub.NewNoOpRouter())

	handler := newRoutes(deps)

	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("POST /health status = %d, want 404 via gateway", rec.Code)
	}
}
