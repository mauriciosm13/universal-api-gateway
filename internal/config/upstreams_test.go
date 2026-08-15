package config

import "testing"

func TestParseUpstreamListSingle(t *testing.T) {
	t.Parallel()

	upstreams, err := ParseUpstreamList("http://a:8080")
	if err != nil {
		t.Fatalf("ParseUpstreamList() error = %v", err)
	}

	if len(upstreams) != 1 || upstreams[0] != "http://a:8080" {
		t.Fatalf("upstreams = %v, want [http://a:8080]", upstreams)
	}
}

func TestParseUpstreamListMultiple(t *testing.T) {
	t.Parallel()

	upstreams, err := ParseUpstreamList("http://a:8080, http://b:8080")
	if err != nil {
		t.Fatalf("ParseUpstreamList() error = %v", err)
	}

	if len(upstreams) != 2 {
		t.Fatalf("len(upstreams) = %d, want 2", len(upstreams))
	}

	if upstreams[0] != "http://a:8080" || upstreams[1] != "http://b:8080" {
		t.Fatalf("upstreams = %v", upstreams)
	}
}

func TestParseUpstreamListInvalid(t *testing.T) {
	t.Parallel()

	if _, err := ParseUpstreamList("http://a:8080,not-a-url"); err == nil {
		t.Fatal("ParseUpstreamList() error = nil, want invalid URL error")
	}
}

func TestParsePathRoutesMultipleUpstreams(t *testing.T) {
	t.Parallel()

	routes, err := ParsePathRoutes("/api=http://a:8080,http://b:8080,/v2=http://v2:8080")
	if err != nil {
		t.Fatalf("ParsePathRoutes() error = %v", err)
	}

	if len(routes) != 2 {
		t.Fatalf("len(routes) = %d, want 2", len(routes))
	}

	if len(routes[0].Upstreams) != 2 {
		t.Fatalf("routes[0].Upstreams = %v, want 2 URLs", routes[0].Upstreams)
	}

	if routes[0].Prefix != "/api" || routes[1].Prefix != "/v2" {
		t.Fatalf("routes = %+v", routes)
	}
}

func TestParseHostRoutesMultipleUpstreams(t *testing.T) {
	t.Parallel()

	routes, err := ParseHostRoutes("api.example.com=http://a:8080,http://b:8080")
	if err != nil {
		t.Fatalf("ParseHostRoutes() error = %v", err)
	}

	if len(routes) != 1 || len(routes[0].Upstreams) != 2 {
		t.Fatalf("routes = %+v", routes)
	}
}

func TestParseMethodRoutesMultipleUpstreams(t *testing.T) {
	t.Parallel()

	routes, err := ParseMethodRoutes("GET=http://a:8080,http://b:8080")
	if err != nil {
		t.Fatalf("ParseMethodRoutes() error = %v", err)
	}

	if len(routes) != 1 || len(routes[0].Upstreams) != 2 {
		t.Fatalf("routes = %+v", routes)
	}
}

func TestParseHeaderRoutesMultipleUpstreams(t *testing.T) {
	t.Parallel()

	routes, err := ParseHeaderRoutes("X-Version=v1=http://a:8080,http://b:8080")
	if err != nil {
		t.Fatalf("ParseHeaderRoutes() error = %v", err)
	}

	if len(routes) != 1 || len(routes[0].Upstreams) != 2 {
		t.Fatalf("routes = %+v", routes)
	}
}

func TestLoadDefaultUpstreamsMultiple(t *testing.T) {
	t.Setenv("GATEWAY_DEFAULT_UPSTREAM", "http://a:8080,http://b:8080")
	t.Setenv("OTEL_TRACES_EXPORTER", "none")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(cfg.DefaultUpstreams) != 2 {
		t.Fatalf("DefaultUpstreams = %v, want 2 URLs", cfg.DefaultUpstreams)
	}
}
