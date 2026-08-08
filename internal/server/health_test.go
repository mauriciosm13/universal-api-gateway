package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthEndpoints(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		path           string
		expectedStatus string
	}{
		{name: "health", path: "/health", expectedStatus: "ok"},
		{name: "liveness", path: "/health/live", expectedStatus: "alive"},
		{name: "readiness", path: "/health/ready", expectedStatus: "ready"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			switch tt.path {
			case "/health":
				handleHealth(rec, req)
			case "/health/live":
				handleLive(rec, req)
			case "/health/ready":
				handleReady(rec, req)
			}

			if rec.Code != http.StatusOK {
				t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
			}

			var body healthResponse
			if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
				t.Fatalf("decode response: %v", err)
			}

			if body.Status != tt.expectedStatus {
				t.Fatalf("expected status %q, got %q", tt.expectedStatus, body.Status)
			}

			if body.Service != "universal-api-gateway" {
				t.Fatalf("expected service name universal-api-gateway, got %q", body.Service)
			}
		})
	}
}
