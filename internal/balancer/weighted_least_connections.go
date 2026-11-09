package balancer

import (
	"net/http"
	"sync/atomic"

	"github.com/IshaanShivKr/raccoon-lb/internal/backend"
)

type WeightedLeastConnections struct {
	pool    *backend.Pool
	counter atomic.Uint64
}

var _ Balancer = (*WeightedLeastConnections)(nil)

func NewWeightedLeastConnections(pool *backend.Pool) *WeightedLeastConnections {
	return &WeightedLeastConnections{pool: pool}
}

func (wlc *WeightedLeastConnections) Next(_ *http.Request) (*backend.Backend, error) {
	backends := wlc.pool.Available()
	if len(backends) == 0 {
		return nil, ErrNoAvailableBackends
	}

	var candidates []*backend.Backend
	var minLoad float64

	for i, be := range backends {
		load := float64(be.Active()) / float64(be.Weight)

		switch {
		case i == 0 || load < minLoad:
			minLoad = load
			candidates = []*backend.Backend{be}

		case load == minLoad:
			candidates = append(candidates, be)
		}
	}

	index := wlc.counter.Add(1) - 1

	return candidates[index%uint64(len(candidates))], nil
}
