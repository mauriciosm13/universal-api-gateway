package domain

import "testing"

func TestRequestResponseFields(t *testing.T) {
	t.Parallel()

	req := Request{
		Method:  "GET",
		Path:    "/api/v1/users",
		Host:    "gateway.local",
		Headers: map[string][]string{"Accept": {"application/json"}},
	}

	if req.Method != "GET" {
		t.Fatalf("expected method GET, got %q", req.Method)
	}

	resp := Response{
		StatusCode: 200,
		Headers:    map[string][]string{"Content-Type": {"application/json"}},
		Body:       []byte(`{"ok":true}`),
	}

	if resp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
}
