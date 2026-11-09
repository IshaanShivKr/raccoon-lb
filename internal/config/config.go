package config

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	AlgorithmRoundRobin               = "round_robin"
	AlgorithmLeastConnections         = "least_connections"
	AlgorithmRandom                   = "random"
	AlgorithmP2C                      = "p2c"
	AlgorithmWeightedRoundRobin       = "weighted_round_robin"
	AlgorithmWeightedLeastConnections = "weighted_least_connections"
	AlgorithmWeightedRandom           = "weighted_random"
	AlgorithmWeightedP2C              = "weighted_p2c"
)

type Config struct {
	Server      ServerConfig    `yaml:"server"`
	Backends    []BackendConfig `yaml:"backends"`
	Balancer    BalancerConfig  `yaml:"balancer"`
	Health      HealthConfig    `yaml:"health"`
	Log         LogConfig       `yaml:"log"`
	BackendURLs []*url.URL      `yaml:"-"`
}

type ServerConfig struct {
	Address           string        `yaml:"address"`
	ReadHeaderTimeout time.Duration `yaml:"read_header_timeout"`
	ReadTimeout       time.Duration `yaml:"read_timeout"`
	WriteTimeout      time.Duration `yaml:"write_timeout"`
	IdleTimeout       time.Duration `yaml:"idle_timeout"`
	ShutdownTimeout   time.Duration `yaml:"shutdown_timeout"`
}

type BackendConfig struct {
	URL       string   `yaml:"url"`
	Weight    int      `yaml:"weight"`
	ParsedURL *url.URL `yaml:"-"`
}

type BalancerConfig struct {
	Algorithm string `yaml:"algorithm"`
}

type HealthConfig struct {
	Interval      time.Duration `yaml:"interval"`
	Timeout       time.Duration `yaml:"timeout"`
	Path          string        `yaml:"path"`
	FallThreshold int           `yaml:"fall_threshold"`
	RiseThreshold int           `yaml:"rise_threshold"`
}

type LogConfig struct {
	Level slog.Level `yaml:"level"`
}

func defaults() Config {
	return Config{
		Server: ServerConfig{
			Address:           ":8080",
			ReadHeaderTimeout: 5 * time.Second,
			ReadTimeout:       15 * time.Second,
			WriteTimeout:      60 * time.Second,
			IdleTimeout:       60 * time.Second,
			ShutdownTimeout:   10 * time.Second,
		},
		Balancer: BalancerConfig{
			Algorithm: AlgorithmRoundRobin,
		},
		Health: HealthConfig{
			Interval:      5 * time.Second,
			Timeout:       2 * time.Second,
			Path:          "/",
			FallThreshold: 3,
			RiseThreshold: 2,
		},
	}
}

func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open config: %w", err)
	}
	defer f.Close()

	cfg := defaults()

	dec := yaml.NewDecoder(f)
	dec.KnownFields(true)

	if err := dec.Decode(&cfg); err != nil && !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("parse config %s: %w", path, err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config %s:\n%w", path, err)
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	var errs []error
	addf := func(format string, args ...any) {
		errs = append(errs, fmt.Errorf(format, args...))
	}
	positive := func(name string, d time.Duration) {
		if d <= 0 {
			addf("%s must be positive, got %s", name, d)
		}
	}

	if strings.TrimSpace(c.Server.Address) == "" {
		addf("server.address is required")
	}
	positive("server.read_header_timeout", c.Server.ReadHeaderTimeout)
	positive("server.read_timeout", c.Server.ReadTimeout)
	positive("server.write_timeout", c.Server.WriteTimeout)
	positive("server.idle_timeout", c.Server.IdleTimeout)
	positive("server.shutdown_timeout", c.Server.ShutdownTimeout)

	switch c.Balancer.Algorithm {
	case AlgorithmRoundRobin,
		AlgorithmLeastConnections,
		AlgorithmRandom,
		AlgorithmP2C,
		AlgorithmWeightedRoundRobin,
		AlgorithmWeightedLeastConnections,
		AlgorithmWeightedRandom,
		AlgorithmWeightedP2C:
	default:
		addf("unsupported balancing algorithm: %q", c.Balancer.Algorithm)
	}

	positive("health.interval", c.Health.Interval)
	positive("health.timeout", c.Health.Timeout)
	if !strings.HasPrefix(c.Health.Path, "/") {
		addf("health.path must start with \"/\", got %q", c.Health.Path)
	}
	if c.Health.FallThreshold < 1 {
		addf("health.fall_threshold must be at least 1, got %d", c.Health.FallThreshold)
	}
	if c.Health.RiseThreshold < 1 {
		addf("health.rise_threshold must be at least 1, got %d", c.Health.RiseThreshold)
	}

	if len(c.Backends) == 0 {
		addf("backends: at least one backend URL is required")
	}

	seen := make(map[string]struct{}, len(c.Backends))

	for i := range c.Backends {
		be := &c.Backends[i]

		if be.Weight == 0 {
			be.Weight = 1
		}

		if be.Weight < 1 {
			addf("backends[%d].weight must be positive", i)
			continue
		}

		u, err := parseBackend(be.URL)
		if err != nil {
			addf("backends[%d].url: %v", i, err)
			continue
		}

		key := backendKey(u)

		if _, exists := seen[key]; exists {
			addf("duplicate backend URL %q", be.URL)
			continue
		}

		seen[key] = struct{}{}
		be.ParsedURL = u
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func parseBackend(raw string) (*url.URL, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return nil, err
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, errors.New("scheme must be http or https")
	}
	if u.Hostname() == "" {
		return nil, errors.New("host is required")
	}
	if p := u.Port(); p != "" {
		n, err := strconv.Atoi(p)
		if err != nil || n < 1 || n > 65535 {
			return nil, fmt.Errorf("port %q is out of range (1-65535)", p)
		}
	}
	return u, nil
}

func backendKey(u *url.URL) string {
	return u.Scheme + "://" + strings.ToLower(u.Host) + strings.TrimRight(u.EscapedPath(), "/")
}
