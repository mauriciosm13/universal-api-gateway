package roundrobin

import (
	"errors"
	"sync"
	"sync/atomic"
)

// ErrEmptyUpstreams indicates that no upstream URLs were provided.
var ErrEmptyUpstreams = errors.New("roundrobin: empty upstream list")

// UpstreamSelector picks the next upstream URL for a route.
type UpstreamSelector interface {
	Next(routeID string, upstreams []string) (string, error)
}

// Selector implements round-robin upstream selection per route ID.
type Selector struct {
	counters sync.Map
}

// NewSelector returns a thread-safe round-robin upstream selector.
func NewSelector() *Selector {
	return &Selector{}
}

// Next returns the next upstream URL for the given route.
func (s *Selector) Next(routeID string, upstreams []string) (string, error) {
	if len(upstreams) == 0 {
		return "", ErrEmptyUpstreams
	}
	if len(upstreams) == 1 {
		return upstreams[0], nil
	}

	counterVal, _ := s.counters.LoadOrStore(routeID, &atomic.Uint64{})
	counter := counterVal.(*atomic.Uint64)
	idx := counter.Add(1) - 1
	return upstreams[idx%uint64(len(upstreams))], nil
}
