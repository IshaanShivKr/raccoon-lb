package balancer

import (
	"math/rand"
	"net/http"

	"github.com/IshaanShivKr/raccoon-lb/internal/backend"
)

type WeightedRandom struct {
	pool *backend.Pool
}

var _ Balancer = (*WeightedRandom)(nil)

func NewWeightedRandom(pool *backend.Pool) *WeightedRandom {
	return &WeightedRandom{
		pool: pool,
	}
}

func (wr *WeightedRandom) Next(_ *http.Request) (*backend.Backend, error) {
	backends := wr.pool.Available()
	if len(backends) == 0 {
		return nil, ErrNoAvailableBackends
	}

	var totalWeight int
	for _, be := range backends {
		totalWeight += be.Weight
	}

	choice := rand.Intn(totalWeight)

	for _, be := range backends {
		if choice < be.Weight {
			return be, nil
		}
		choice -= be.Weight
	}

	return nil, ErrNoAvailableBackends
}
