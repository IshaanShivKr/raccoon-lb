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

type ctxKey struct{}

type Proxy struct {
	balancer balancer.Balancer
	logger   *slog.Logger
	rp       *httputil.ReverseProxy
}

func New(b balancer.Balancer, logger *slog.Logger) *Proxy {
	p := &Proxy{
		balancer: b,
		logger:   logger,
	}

	p.rp = &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			be := pr.In.Context().Value(ctxKey{}).(*backend.Backend)
			pr.SetURL(be.URL)
			pr.SetXForwarded()
		},
		Transport:    newTransport(),
		ErrorHandler: p.handleError,
	}

	return p
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	be, err := p.balancer.Next()
	if err != nil {
		if errors.Is(err, balancer.ErrNoHealthyBackends) {
			http.Error(w, "no available backends", http.StatusServiceUnavailable)
			return
		}
		p.logger.Error("balancer failed", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	ctx := context.WithValue(r.Context(), ctxKey{}, be)
	p.rp.ServeHTTP(w, r.WithContext(ctx))
}

func (p *Proxy) handleError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, context.Canceled) {
		return
	}

	backendURL := "unknown"
	if be, ok := r.Context().Value(ctxKey{}).(*backend.Backend); ok {
		backendURL = be.URL.String()
	}

	p.logger.Error("upstream request failed",
		"backend", backendURL,
		"method", r.Method,
		"path", r.URL.Path,
		"error", err,
	)

	var netErr net.Error
	if errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &netErr) && netErr.Timeout()) {
		http.Error(w, "upstream timeout", http.StatusGatewayTimeout)
		return
	}

	http.Error(w, "upstream server error", http.StatusBadGateway)
}

var _ http.Handler = (*Proxy)(nil)
