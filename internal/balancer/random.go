package balancer

import (
	"math/rand"
	"net/http"

	"github.com/IshaanShivKr/raccoon-lb/internal/backend"
)

type Random struct {
	pool *backend.Pool
}

var _ Balancer = (*Random)(nil)

func NewRandom(pool *backend.Pool) *Random {
	return &Random{pool: pool}
}

func (r *Random) Next(_ *http.Request) (*backend.Backend, error) {
	backends := r.pool.Available()
	if len(backends) == 0 {
		return nil, ErrNoAvailableBackends
	}

	return backends[rand.Intn(len(backends))], nil
}
