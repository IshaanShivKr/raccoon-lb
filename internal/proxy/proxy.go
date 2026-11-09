package proxy

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"

	"github.com/IshaanShivKr/raccoon-lb/internal/backend"
	"github.com/IshaanShivKr/raccoon-lb/internal/balancer"
)

const (
	maxAcquireAttempts        = 3
	statusClientClosedRequest = 499
)

type ctxKey struct{}

func withBackend(ctx context.Context, be *backend.Backend) context.Context {
	return context.WithValue(ctx, ctxKey{}, be)
}

func backendFrom(ctx context.Context) (*backend.Backend, bool) {
	be, ok := ctx.Value(ctxKey{}).(*backend.Backend)
	return be, ok && be != nil
}

type Proxy struct {
	balancer balancer.Balancer
	logger   *slog.Logger
	rp       *httputil.ReverseProxy
}

var _ http.Handler = (*Proxy)(nil)

func New(b balancer.Balancer, logger *slog.Logger) *Proxy {
	p := &Proxy{
		balancer: b,
		logger:   logger,
	}

	p.rp = &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			be, ok := backendFrom(pr.In.Context())
			if !ok {
				p.logger.Error("backend missing from request context")
				return
			}

			pr.SetURL(be.URL)
			pr.SetXForwarded()
		},
		Transport:    newTransport(),
		ErrorHandler: p.handleError,
	}

	return p
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	be, err := p.acquire(r)
	if err != nil {
		if errors.Is(err, balancer.ErrNoAvailableBackends) {
			http.Error(w, "no available backends", http.StatusServiceUnavailable)
			return
		}
		p.logger.Error("balancer failed", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer be.Release()

	p.rp.ServeHTTP(w, r.WithContext(withBackend(r.Context(), be)))
}

func (p *Proxy) acquire(r *http.Request) (*backend.Backend, error) {
	for range maxAcquireAttempts {
		be, err := p.balancer.Next(r)
		if err != nil {
			return nil, err
		}
		if be.TryAcquire() {
			return be, nil
		}
	}
	return nil, balancer.ErrNoAvailableBackends
}

func (p *Proxy) handleError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, context.Canceled) {
		w.WriteHeader(statusClientClosedRequest)
		return
	}

	backendURL := "unknown"
	if be, ok := backendFrom(r.Context()); ok {
		backendURL = be.URL.String()
	}

	var netErr net.Error
	if errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &netErr) && netErr.Timeout()) {
		p.logger.Warn("upstream request timed out",
			"backend", backendURL,
			"method", r.Method,
			"path", r.URL.Path,
			"error", err,
		)
		http.Error(w, "upstream timeout", http.StatusGatewayTimeout)
		return
	}

	p.logger.Error("upstream request failed",
		"backend", backendURL,
		"method", r.Method,
		"path", r.URL.Path,
		"error", err,
	)
	http.Error(w, "upstream server error", http.StatusBadGateway)
}
