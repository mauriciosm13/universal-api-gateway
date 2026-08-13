package config

import "testing"

func TestParseHostRoutesValid(t *testing.T) {
	t.Parallel()

	routes, err := ParseHostRoutes("api.example.com=http://api:8080, admin.example.com=https://admin.example.com")
	if err != nil {
		t.Fatalf("ParseHostRoutes() error = %v", err)
	}

	if len(routes) != 2 || routes[0].Host != "api.example.com" {
		t.Fatalf("routes = %+v", routes)
	}
}

func TestParseHostRoutesInvalid(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		raw  string
	}{
		{name: "missing host", raw: "=http://api:8080"},
		{name: "missing upstream", raw: "api.example.com="},
		{name: "missing separator", raw: "api.example.com"},
		{name: "invalid upstream", raw: "api.example.com=not-a-url"},
		{name: "empty entries", raw: " , "},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if _, err := ParseHostRoutes(tc.raw); err == nil {
				t.Fatalf("ParseHostRoutes(%q) error = nil, want error", tc.raw)
			}
		})
	}
}

func TestParseHostRoutesEmpty(t *testing.T) {
	t.Parallel()

	routes, err := ParseHostRoutes("")
	if err != nil {
		t.Fatalf("ParseHostRoutes() error = %v", err)
	}

	if routes != nil {
		t.Fatalf("routes = %+v, want nil", routes)
	}
}
