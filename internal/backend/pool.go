package backend

type Pool struct {
	backends []*Backend
}

func NewPool(backends []*Backend) *Pool {
	b := make([]*Backend, len(backends))
	copy(b, backends)
	return &Pool{backends: b}
}

func (p *Pool) All() []*Backend {
	out := make([]*Backend, len(p.backends))
	copy(out, p.backends)
	return out
}

func (p *Pool) Healthy() []*Backend {
	out := make([]*Backend, 0, len(p.backends))

	for _, b := range p.backends {
		if b.IsHealthy() {
			out = append(out, b)
		}
	}

	return out
}
