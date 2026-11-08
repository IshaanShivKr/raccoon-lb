package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IshaanShivKr/raccoon-lb/internal/backend"
	"github.com/IshaanShivKr/raccoon-lb/internal/balancer"
	"github.com/IshaanShivKr/raccoon-lb/internal/health"
	"github.com/IshaanShivKr/raccoon-lb/internal/proxy"
)

const (
	listenAddr      = ":8080"
	shutdownTimeout = 10 * time.Second

	readHeaderTimeout = 5 * time.Second
	readTimeout       = 15 * time.Second
	writeTimeout      = 60 * time.Second
	idleTimeout       = 60 * time.Second

	healthInterval = 5 * time.Second
	healthTimeout  = 2 * time.Second
	healthPath     = "/"
	healthFall     = 3
	healthRise     = 2
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	if err := run(logger); err != nil {
		logger.Error("application failed", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	targets := []string{
		"http://localhost:9001",
		"http://localhost:9002",
		"http://localhost:9003",
	}

	backends := make([]*backend.Backend, 0, len(targets))
	for _, t := range targets {
		u, err := url.Parse(t)
		if err != nil {
			return fmt.Errorf("parse backend %q: %w", t, err)
		}
		backends = append(backends, &backend.Backend{URL: u})
	}

	pool := backend.NewPool(backends)
	lb := balancer.NewRoundRobin(pool)
	handler := proxy.New(lb, logger)

	checker, err := health.New(
		pool,
		logger,
		healthInterval,
		healthTimeout,
		healthPath,
		healthFall,
		healthRise,
	)
	if err != nil {
		return fmt.Errorf("create health checker: %w", err)
	}

	server := &http.Server{
		Addr:              listenAddr,
		Handler:           handler,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)

	healthDone := make(chan struct{})
	go func() {
		defer close(healthDone)
		checker.Run(ctx)
	}()

	defer func() {
		stop()
		<-healthDone
	}()

	serverErr := make(chan error, 1)

	go func() {
		logger.Info("starting HTTP server", "address", server.Addr, "backends", len(backends))

		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
			return
		}

		serverErr <- nil
	}()

	select {
	case err := <-serverErr:
		if err != nil {
			return fmt.Errorf("http server: %w", err)
		}
		return nil

	case <-ctx.Done():
		stop()
		logger.Info("shutdown signal received, shutting down server", "timeout", shutdownTimeout)
	}

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		shutdownTimeout,
	)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)

		if closeErr := server.Close(); closeErr != nil {
			return errors.Join(err, closeErr)
		}

		return err
	}

	logger.Info("server stopped gracefully")
	return nil
}
