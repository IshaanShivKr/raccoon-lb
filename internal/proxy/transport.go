package proxy

import (
	"net"
	"net/http"
	"time"
)

const dialTimeout = 5 * time.Second

func newTransport() *http.Transport {
	t := http.DefaultTransport.(*http.Transport).Clone()

	t.DialContext = (&net.Dialer{
		Timeout:   dialTimeout,
		KeepAlive: 30 * time.Second,
	}).DialContext
	t.MaxIdleConns = 100
	t.MaxIdleConnsPerHost = 100
	t.IdleConnTimeout = 90 * time.Second
	t.ResponseHeaderTimeout = 30 * time.Second

	return t
}
