package backend

import (
	"net/url"
	"sync/atomic"
)

type Backend struct {
	URL       *url.URL
	unhealthy atomic.Bool
	draining  atomic.Bool
	active    atomic.Int64
}

func (b *Backend) IsHealthy() bool {
	return !b.unhealthy.Load()
}

func (b *Backend) SetHealthy(healthy bool) {
	b.unhealthy.Store(!healthy)
}

func (b *Backend) IsDraining() bool {
	return b.draining.Load()
}

func (b *Backend) Available() bool {
	return b.IsHealthy() && !b.IsDraining()
}

func (b *Backend) Drain() bool {
	return b.draining.CompareAndSwap(false, true)
}

func (b *Backend) Drained() bool {
	return b.IsDraining() && b.active.Load() == 0
}

func (b *Backend) TryAcquire() bool {
	b.active.Add(1)
	if !b.Available() {
		b.active.Add(-1)
		return false
	}
	return true
}

func (b *Backend) Release() {
	b.active.Add(-1)
}
