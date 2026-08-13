package config

import "testing"

func TestParseHeaderRoutesValid(t *testing.T) {
	t.Parallel()

	routes, err := ParseHeaderRoutes("X-Version=v1=http://v1:8080, X-Tenant=acme=https://acme.example.com")
	if err != nil {
		t.Fatalf("ParseHeaderRoutes() error = %v", err)
	}

	if len(routes) != 2 {
		t.Fatalf("len(routes) = %d, want 2", len(routes))
	}

	if routes[0].Name != "X-Version" || routes[0].Value != "v1" || routes[0].Upstream != "http://v1:8080" {
		t.Fatalf("routes[0] = %+v", routes[0])
	}

	if routes[1].Name != "X-Tenant" || routes[1].Value != "acme" {
		t.Fatalf("routes[1] = %+v", routes[1])
	}
}

func TestParseHeaderRoutesEmpty(t *testing.T) {
	t.Parallel()

	routes, err := ParseHeaderRoutes("")
	if err != nil {
		t.Fatalf("ParseHeaderRoutes() error = %v", err)
	}

	if routes != nil {
		t.Fatalf("routes = %v, want nil", routes)
	}
}

func TestParseHeaderRoutesInvalid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
	}{
		{name: "missing upstream", raw: "X-Version=v1="},
		{name: "missing value", raw: "X-Version="},
		{name: "missing separator", raw: "X-Version=v1"},
		{name: "invalid upstream", raw: "X-Version=v1=not-a-url"},
		{name: "only commas", raw: ",,,"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if _, err := ParseHeaderRoutes(tt.raw); err == nil {
				t.Fatalf("ParseHeaderRoutes(%q) error = nil, want error", tt.raw)
			}
		})
	}
}
