package balancer

import (
	"net/http"
	"sync"

	"github.com/IshaanShivKr/raccoon-lb/internal/backend"
)

type WeightedRoundRobin struct {
	pool *backend.Pool

	mu             sync.Mutex
	currentWeights map[*backend.Backend]int64
}

var _ Balancer = (*WeightedRoundRobin)(nil)

func NewWeightedRoundRobin(pool *backend.Pool) *WeightedRoundRobin {
	return &WeightedRoundRobin{
		pool:           pool,
		currentWeights: make(map[*backend.Backend]int64),
	}
}

func (wrr *WeightedRoundRobin) Next(_ *http.Request) (*backend.Backend, error) {
	backends := wrr.pool.Available()
	if len(backends) == 0 {
		return nil, ErrNoAvailableBackends
	}

	wrr.mu.Lock()
	defer wrr.mu.Unlock()

	available := make(map[*backend.Backend]struct{}, len(backends))
	for _, be := range backends {
		available[be] = struct{}{}
	}

	for be := range wrr.currentWeights {
		if _, ok := available[be]; !ok {
			delete(wrr.currentWeights, be)
		}
	}

	var totalWeight int64
	var selected *backend.Backend
	var maxWeight int64

	for _, be := range backends {
		weight := int64(be.Weight)

		wrr.currentWeights[be] += weight
		totalWeight += weight

		if selected == nil || wrr.currentWeights[be] > maxWeight {
			selected = be
			maxWeight = wrr.currentWeights[be]
		}
	}

	wrr.currentWeights[selected] -= totalWeight

	return selected, nil
}
