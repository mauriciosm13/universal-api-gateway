package roundrobin

import "testing"

func TestSelectorAlternatesUpstreams(t *testing.T) {
	t.Parallel()

	selector := NewSelector()
	upstreams := []string{"http://a:8080", "http://b:8080"}

	got := make([]string, 4)
	for i := range got {
		var err error
		got[i], err = selector.Next("route-a", upstreams)
		if err != nil {
			t.Fatalf("Next() error = %v", err)
		}
	}

	want := []string{
		"http://a:8080",
		"http://b:8080",
		"http://a:8080",
		"http://b:8080",
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got[%d] = %q, want %q (full=%v)", i, got[i], want[i], got)
		}
	}
}

func TestSelectorSingleUpstream(t *testing.T) {
	t.Parallel()

	selector := NewSelector()
	upstream, err := selector.Next("route-a", []string{"http://only:8080"})
	if err != nil {
		t.Fatalf("Next() error = %v", err)
	}

	if upstream != "http://only:8080" {
		t.Fatalf("upstream = %q, want http://only:8080", upstream)
	}
}

func TestSelectorEmptyUpstreams(t *testing.T) {
	t.Parallel()

	selector := NewSelector()
	_, err := selector.Next("route-a", nil)
	if err == nil {
		t.Fatal("Next() error = nil, want ErrEmptyUpstreams")
	}
}

func TestSelectorIndependentRouteCounters(t *testing.T) {
	t.Parallel()

	selector := NewSelector()
	upstreams := []string{"http://a:8080", "http://b:8080"}

	first, err := selector.Next("route-a", upstreams)
	if err != nil {
		t.Fatalf("Next(route-a) error = %v", err)
	}

	second, err := selector.Next("route-b", upstreams)
	if err != nil {
		t.Fatalf("Next(route-b) error = %v", err)
	}

	if first != "http://a:8080" || second != "http://a:8080" {
		t.Fatalf("first=%q second=%q, want both http://a:8080", first, second)
	}
}
