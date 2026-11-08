package health

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/IshaanShivKr/raccoon-lb/internal/backend"
)

const (
	maxDrainBytes = 4 << 10
	userAgent     = "raccoon-lb-healthcheck/1.0"
)

type streak struct {
	oks   int
	fails int
}

type Checker struct {
	pool          *backend.Pool
	client        *http.Client
	logger        *slog.Logger
	interval      time.Duration
	timeout       time.Duration
	path          string
	fallThreshold int
	riseThreshold int
	streaks       map[*backend.Backend]*streak
}

func New(
	pool *backend.Pool,
	logger *slog.Logger,
	interval, timeout time.Duration,
	path string,
	fallThreshold, riseThreshold int,
) (*Checker, error) {
	if interval <= 0 {
		return nil, errors.New("health: interval must be positive")
	}
	if timeout <= 0 {
		return nil, errors.New("health: timeout must be positive")
	}
	if fallThreshold < 1 {
		return nil, errors.New("health: fall threshold must be at least 1")
	}
	if riseThreshold < 1 {
		return nil, errors.New("health: rise threshold must be at least 1")
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	client := &http.Client{
		Transport: transport,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	streaks := make(map[*backend.Backend]*streak)
	for _, b := range pool.All() {
		streaks[b] = &streak{}
	}

	return &Checker{
		pool:          pool,
		client:        client,
		logger:        logger,
		interval:      interval,
		timeout:       timeout,
		path:          path,
		fallThreshold: fallThreshold,
		riseThreshold: riseThreshold,
		streaks:       streaks,
	}, nil
}

func (c *Checker) Run(ctx context.Context) {
	defer c.client.CloseIdleConnections()

	c.checkAll(ctx)

	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.checkAll(ctx)
		}
	}
}

func (c *Checker) checkAll(ctx context.Context) {
	var wg sync.WaitGroup

	for _, b := range c.pool.All() {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.checkOne(ctx, b)
		}()
	}

	wg.Wait()
}

func (c *Checker) checkOne(ctx context.Context, b *backend.Backend) {
	err := c.probe(ctx, b)

	if ctx.Err() != nil {
		return
	}

	s := c.streaks[b]

	if err == nil {
		s.fails = 0
		s.oks++
		if !b.IsHealthy() && s.oks >= c.riseThreshold {
			b.SetHealthy(true)
			c.logger.Info("backend recovered", "backend", b.URL.String())
		}
		return
	}

	s.oks = 0
	s.fails++
	if b.IsHealthy() && s.fails >= c.fallThreshold {
		b.SetHealthy(false)
		c.logger.Warn("backend unhealthy",
			"backend", b.URL.String(),
			"consecutive_failures", s.fails,
			"reason", err,
		)
	}
}

func (c *Checker) probe(ctx context.Context, b *backend.Backend) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	target := b.URL.JoinPath(c.path).String()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, maxDrainBytes))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	return nil
}
