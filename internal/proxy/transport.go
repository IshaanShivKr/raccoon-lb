package proxy

import (
	"net/http"
	"time"
)

func newTransport() *http.Transport {
	t := http.DefaultTransport.(*http.Transport).Clone()

	t.MaxIdleConns = 100
	t.MaxIdleConnsPerHost = 100
	t.IdleConnTimeout = 90 * time.Second
	t.ResponseHeaderTimeout = 30 * time.Second

	return t
}
