package credential

import (
	"testing"

	"github.com/mauriciomendonca/universal-api-gateway/internal/domain"
)

func TestBearerTokenValid(t *testing.T) {
	t.Parallel()

	token := BearerToken(domain.Request{
		Headers: map[string][]string{"Authorization": {"Bearer abc.def"}},
	})
	if token != "abc.def" {
		t.Fatalf("token = %q, want abc.def", token)
	}
}

func TestBearerTokenMalformed(t *testing.T) {
	t.Parallel()

	if BearerToken(domain.Request{
		Headers: map[string][]string{"Authorization": {"Basic x"}},
	}) != "" {
		t.Fatal("expected empty token for malformed header")
	}
}

func TestAPIKeyFromHeader(t *testing.T) {
	t.Parallel()

	key := APIKey(domain.Request{
		Headers: map[string][]string{"X-API-Key": {"secret"}},
	}, "X-API-Key", "api_key")
	if key != "secret" {
		t.Fatalf("key = %q, want secret", key)
	}
}

func TestAPIKeyFromQuery(t *testing.T) {
	t.Parallel()

	key := APIKey(domain.Request{
		Query: map[string][]string{"api_key": {"secret"}},
	}, "X-API-Key", "api_key")
	if key != "secret" {
		t.Fatalf("key = %q, want secret", key)
	}
}

func TestHasAuthorizationHeader(t *testing.T) {
	t.Parallel()

	if !HasAuthorizationHeader(domain.Request{
		Headers: map[string][]string{"Authorization": {"Bearer x"}},
	}) {
		t.Fatal("HasAuthorizationHeader() = false, want true")
	}
}
