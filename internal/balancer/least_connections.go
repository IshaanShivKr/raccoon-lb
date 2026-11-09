package balancer

import (
	"sync/atomic"

	"github.com/IshaanShivKr/raccoon-lb/internal/backend"
)

type LeastConnections struct {
	pool    *backend.Pool
	counter atomic.Uint64
}

var _ Balancer = (*LeastConnections)(nil)

func NewLeastConnections(pool *backend.Pool) *LeastConnections {
	return &LeastConnections{pool: pool}
}

func (lc *LeastConnections) Next() (*backend.Backend, error) {
	backends := lc.pool.Available()

	if len(backends) == 0 {
		return nil, ErrNoAvailableBackends
	}

	minActive := backends[0].Active()
	candidates := []*backend.Backend{backends[0]}

	for _, be := range backends[1:] {
		active := be.Active()

		switch {
		case active < minActive:
			minActive = active
			candidates = []*backend.Backend{be}
		case active == minActive:
			candidates = append(candidates, be)
		}
	}

	index := lc.counter.Add(1) - 1

	return candidates[index%uint64(len(candidates))], nil
}
