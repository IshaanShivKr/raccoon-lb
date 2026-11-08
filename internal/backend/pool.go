package backend

import "sync"

type Pool struct {
	mu       sync.RWMutex
	backends []*Backend
}

func NewPool(backends []*Backend) *Pool {
	return &Pool{
		backends: backends,
	}
}

func (p *Pool) All() []*Backend {
	p.mu.RLock()
	defer p.mu.RUnlock()

	backends := make([]*Backend, len(p.backends))
	copy(backends, p.backends)

	return backends
}
