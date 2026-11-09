package balancer

import (
	"math/rand"

	"github.com/IshaanShivKr/raccoon-lb/internal/backend"
)

type PowerOfTwoChoices struct {
	pool *backend.Pool
}

var _ Balancer = (*PowerOfTwoChoices)(nil)

func NewPowerOfTwoChoices(pool *backend.Pool) *PowerOfTwoChoices {
	return &PowerOfTwoChoices{
		pool: pool,
	}
}

func (p *PowerOfTwoChoices) Next() (*backend.Backend, error) {
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

	activeA := a.Active()
	activeB := b.Active()

	switch {
	case activeA < activeB:
		return a, nil
	case activeB < activeA:
		return b, nil
	default:
		if rand.Intn(2) == 0 {
			return a, nil
		}
		return b, nil
	}
}
