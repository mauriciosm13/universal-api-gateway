package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mauriciomendonca/universal-api-gateway/internal/domain"
)

func TestWriteDomainResponse(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	writeDomainResponse(rec, domain.Response{
		StatusCode: http.StatusTooManyRequests,
		Headers:    map[string][]string{"Content-Type": {"application/json"}},
		Body:       []byte(`{"code":429}`),
	})

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want 429", rec.Code)
	}

	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("Content-Type = %q", ct)
	}

	var body map[string]int
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
}

func TestShouldContinue(t *testing.T) {
	t.Parallel()

	if !shouldContinue(domain.Response{StatusCode: 0}) {
		t.Fatal("expected continue for status 0")
	}

	if shouldContinue(domain.Response{StatusCode: http.StatusOK}) {
		t.Fatal("expected no continue for status 200")
	}
}
