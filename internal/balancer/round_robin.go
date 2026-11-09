package balancer

import (
	"net/http"
	"sync/atomic"

	"github.com/IshaanShivKr/raccoon-lb/internal/backend"
)

type RoundRobin struct {
	pool    *backend.Pool
	counter atomic.Uint64
}

var _ Balancer = (*RoundRobin)(nil)

func NewRoundRobin(pool *backend.Pool) *RoundRobin {
	return &RoundRobin{pool: pool}
}

func (rr *RoundRobin) Next(_ *http.Request) (*backend.Backend, error) {
	backends := rr.pool.Available()

	if len(backends) == 0 {
		return nil, ErrNoAvailableBackends
	}

	index := rr.counter.Add(1) - 1

	return backends[index%uint64(len(backends))], nil
}
