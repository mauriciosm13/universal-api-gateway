package context

import (
	"context"
	"testing"

	authport "github.com/mauriciomendonca/universal-api-gateway/internal/auth/port"
)

func TestWithIdentityRoundTrip(t *testing.T) {
	t.Parallel()

	want := authport.Identity{
		Subject: "user-1",
		Claims:  map[string]string{"source": "apikey"},
	}

	ctx := WithIdentity(context.Background(), want)
	got, ok := IdentityFrom(ctx)
	if !ok {
		t.Fatal("IdentityFrom() = false, want true")
	}

	if got.Subject != want.Subject || got.Claims["source"] != "apikey" {
		t.Fatalf("identity = %+v, want %+v", got, want)
	}
}

func TestIdentityFromMissing(t *testing.T) {
	t.Parallel()

	if _, ok := IdentityFrom(context.Background()); ok {
		t.Fatal("IdentityFrom() = true, want false")
	}
}
