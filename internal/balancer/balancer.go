package balancer

import (
	"errors"
	"net/http"

	"github.com/IshaanShivKr/raccoon-lb/internal/backend"
)

var ErrNoAvailableBackends = errors.New("no available backends")

type Balancer interface {
	Next(r *http.Request) (*backend.Backend, error)
}
