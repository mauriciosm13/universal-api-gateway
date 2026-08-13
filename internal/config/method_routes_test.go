package config

import "testing"

func TestParseMethodRoutesValid(t *testing.T) {
	t.Parallel()

	routes, err := ParseMethodRoutes("GET=http://get:8080, POST=http://post:8080")
	if err != nil {
		t.Fatalf("ParseMethodRoutes() error = %v", err)
	}

	if len(routes) != 2 || routes[0].Method != "GET" {
		t.Fatalf("routes = %+v", routes)
	}
}

func TestParseMethodRoutesInvalid(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		raw  string
	}{
		{name: "invalid method", raw: "INVALID=http://api:8080"},
		{name: "missing method", raw: "=http://api:8080"},
		{name: "missing upstream", raw: "GET="},
		{name: "missing separator", raw: "GET"},
		{name: "invalid upstream", raw: "GET=not-a-url"},
		{name: "empty entries", raw: " , "},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if _, err := ParseMethodRoutes(tc.raw); err == nil {
				t.Fatalf("ParseMethodRoutes(%q) error = nil, want error", tc.raw)
			}
		})
	}
}

func TestParseMethodRoutesEmpty(t *testing.T) {
	t.Parallel()

	routes, err := ParseMethodRoutes("")
	if err != nil {
		t.Fatalf("ParseMethodRoutes() error = %v", err)
	}

	if routes != nil {
		t.Fatalf("routes = %+v, want nil", routes)
	}
}
