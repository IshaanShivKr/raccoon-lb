package balancer

import (
	"errors"

	"github.com/IshaanShivKr/raccoon-lb/internal/backend"
)

var ErrNoAvailableBackends = errors.New("no available backends")

type Balancer interface {
	Next() (*backend.Backend, error)
}
