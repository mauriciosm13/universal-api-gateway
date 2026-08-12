package config

import "testing"

func TestParsePathRoutesValid(t *testing.T) {
	t.Parallel()

	routes, err := ParsePathRoutes("/api=http://api:8080, /v2=https://v2.example.com")
	if err != nil {
		t.Fatalf("ParsePathRoutes() error = %v", err)
	}

	if len(routes) != 2 {
		t.Fatalf("len(routes) = %d, want 2", len(routes))
	}

	if routes[0].Prefix != "/api" || routes[0].Upstream != "http://api:8080" {
		t.Fatalf("routes[0] = %+v, want /api -> http://api:8080", routes[0])
	}

	if routes[1].Prefix != "/v2" || routes[1].Upstream != "https://v2.example.com" {
		t.Fatalf("routes[1] = %+v, want /v2 -> https://v2.example.com", routes[1])
	}
}

func TestParsePathRoutesEmpty(t *testing.T) {
	t.Parallel()

	routes, err := ParsePathRoutes("")
	if err != nil {
		t.Fatalf("ParsePathRoutes() error = %v", err)
	}

	if routes != nil {
		t.Fatalf("routes = %v, want nil", routes)
	}
}

func TestParsePathRoutesInvalid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
	}{
		{name: "missing upstream", raw: "/api="},
		{name: "missing prefix", raw: "=http://api:8080"},
		{name: "missing separator", raw: "/api"},
		{name: "prefix without slash", raw: "api=http://api:8080"},
		{name: "invalid upstream", raw: "/api=not-a-url"},
		{name: "only commas", raw: ",,,"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if _, err := ParsePathRoutes(tt.raw); err == nil {
				t.Fatalf("ParsePathRoutes(%q) error = nil, want error", tt.raw)
			}
		})
	}
}
