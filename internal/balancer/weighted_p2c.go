package balancer

import (
	"math/rand"
	"net/http"

	"github.com/IshaanShivKr/raccoon-lb/internal/backend"
)

type WeightedP2C struct {
	pool *backend.Pool
}

var _ Balancer = (*WeightedP2C)(nil)

func NewWeightedP2C(pool *backend.Pool) *WeightedP2C {
	return &WeightedP2C{pool: pool}
}

func (p *WeightedP2C) Next(_ *http.Request) (*backend.Backend, error) {
	backends := p.pool.Available()
	if len(backends) == 0 {
		return nil, ErrNoAvailableBackends
	}

	if len(backends) == 1 {
		return backends[0], nil
	}

	first := rand.Intn(len(backends))
	second := rand.Intn(len(backends) - 1)

	if second >= first {
		second++
	}

	a := backends[first]
	b := backends[second]

	loadA := float64(a.Active()) / float64(a.Weight)
	loadB := float64(b.Active()) / float64(b.Weight)

	switch {
	case loadA < loadB:
		return a, nil
	case loadB < loadA:
		return b, nil
	default:
		if rand.Intn(2) == 0 {
			return a, nil
		}
		return b, nil
	}
}
