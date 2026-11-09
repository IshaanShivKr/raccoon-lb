package backend

import (
	"log/slog"
	"net/url"
	"sync"
)

type Backend struct {
	URL       *url.URL
	Logger    *slog.Logger
	mu        sync.Mutex
	unhealthy bool
	draining  bool
	active    int64
}

func (b *Backend) logger() *slog.Logger {
	if b.Logger != nil {
		return b.Logger
	}
	return slog.Default()
}

func (b *Backend) IsHealthy() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	return !b.unhealthy
}

func (b *Backend) SetHealthy(healthy bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.unhealthy = !healthy
}

func (b *Backend) IsDraining() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.draining
}

func (b *Backend) Available() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	return !b.unhealthy && !b.draining
}

func (b *Backend) Drain() bool {
	b.mu.Lock()
	if b.draining {
		b.mu.Unlock()
		return false
	}

	b.draining = true
	active := b.active
	b.mu.Unlock()

	b.logger().Info("backend draining",
		"backend", b.URL.String(),
		"active_requests", active,
	)
	if active == 0 {
		b.logger().Info("backend drained", "backend", b.URL.String())
	}

	return true
}

func (b *Backend) Active() int64 {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.active
}

func (b *Backend) Drained() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.draining && b.active == 0
}

func (b *Backend) TryAcquire() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.unhealthy || b.draining {
		return false
	}

	b.active++
	return true
}

func (b *Backend) Release() {
	b.mu.Lock()
	if b.active <= 0 {
		b.mu.Unlock()
		panic("backend: release called without active request")
	}

	b.active--
	drained := b.draining && b.active == 0
	b.mu.Unlock()

	if drained {
		b.logger().Info("backend drained", "backend", b.URL.String())
	}
}
