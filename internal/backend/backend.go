package backend

import (
	"net/url"
	"sync/atomic"
)

type Backend struct {
	URL       *url.URL
	unhealthy atomic.Bool
}

func (b *Backend) IsHealthy() bool {
	return !b.unhealthy.Load()
}

func (b *Backend) SetHealthy(healthy bool) {
	b.unhealthy.Store(!healthy)
}
