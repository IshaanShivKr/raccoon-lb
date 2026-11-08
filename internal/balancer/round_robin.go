package balancer

import (
	"sync/atomic"

	"github.com/IshaanShivKr/raccoon-lb/internal/backend"
)

type RoundRobin struct {
	pool    *backend.Pool
	counter atomic.Uint64
}

var _ Balancer = (*RoundRobin)(nil)

func NewRoundRobin(pool *backend.Pool) *RoundRobin {
	return &RoundRobin{
		pool: pool,
	}
}

func (rr *RoundRobin) Next() (*backend.Backend, error) {
	backends := rr.pool.All()

	if len(backends) == 0 {
		return nil, ErrNoHealthyBackends
	}

	index := rr.counter.Add(1) - 1

	return backends[index%uint64(len(backends))], nil
}
