package balancer

import (
	"errors"

	"github.com/IshaanShivKr/raccoon-lb/internal/backend"
)

var ErrNoHealthyBackends = errors.New("no healthy backends available")

type Balancer interface {
	Next() (*backend.Backend, error)
}
