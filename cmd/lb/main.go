package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/IshaanShivKr/raccoon-lb/internal/backend"
	"github.com/IshaanShivKr/raccoon-lb/internal/balancer"
	"github.com/IshaanShivKr/raccoon-lb/internal/config"
	"github.com/IshaanShivKr/raccoon-lb/internal/health"
	"github.com/IshaanShivKr/raccoon-lb/internal/middleware"
	"github.com/IshaanShivKr/raccoon-lb/internal/proxy"
)

const configPath = "configs/config.yaml"

func main() {
	level := new(slog.LevelVar)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))

	cfg, err := config.Load(configPath)
	if err != nil {
		logger.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	level.Set(cfg.Log.Level)

	if err := run(logger, cfg); err != nil {
		logger.Error("application failed", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger, cfg *config.Config) error {
	backends := make([]*backend.Backend, 0, len(cfg.BackendURLs))
	for _, u := range cfg.BackendURLs {
		backends = append(backends, &backend.Backend{URL: u, Logger: logger})
	}

	pool := backend.NewPool(backends)

	var lb balancer.Balancer

	switch cfg.Balancer.Algorithm {
	case config.AlgorithmRoundRobin:
		lb = balancer.NewRoundRobin(pool)

	case config.AlgorithmLeastConnections:
		lb = balancer.NewLeastConnections(pool)

	case config.AlgorithmRandom:
		lb = balancer.NewRandom(pool)

	case config.AlgorithmP2C:
		lb = balancer.NewPowerOfTwoChoices(pool)

	case config.AlgorithmWeightedRoundRobin:
		lb = balancer.NewWeightedRoundRobin(pool)

	case config.AlgorithmWeightedLeastConnections:
		lb = balancer.NewWeightedLeastConnections(pool)

	case config.AlgorithmWeightedRandom:
		lb = balancer.NewWeightedRandom(pool)

	case config.AlgorithmWeightedP2C:
		lb = balancer.NewWeightedP2C(pool)
	}

	handler := middleware.Logging(logger, proxy.New(lb, logger))

	checker, err := health.New(
		pool,
		logger,
		cfg.Health.Interval,
		cfg.Health.Timeout,
		cfg.Health.Path,
		cfg.Health.FallThreshold,
		cfg.Health.RiseThreshold,
	)
	if err != nil {
		return fmt.Errorf("create health checker: %w", err)
	}

	server := &http.Server{
		Addr:              cfg.Server.Address,
		Handler:           handler,
		ReadHeaderTimeout: cfg.Server.ReadHeaderTimeout,
		ReadTimeout:       cfg.Server.ReadTimeout,
		WriteTimeout:      cfg.Server.WriteTimeout,
		IdleTimeout:       cfg.Server.IdleTimeout,
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
		logger.Info("shutdown signal received, shutting down server", "timeout", cfg.Server.ShutdownTimeout)
	}

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		cfg.Server.ShutdownTimeout,
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
