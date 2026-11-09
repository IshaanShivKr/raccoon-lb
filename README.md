# Raccoon LB

A Layer 7 HTTP load balancer and reverse proxy written in Go.

Raccoon LB distributes incoming HTTP requests across multiple backend servers using configurable load-balancing algorithms. It includes active health checking, weighted backend selection, concurrent request handling, and graceful shutdown.

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [Features](#features)
- [Load-Balancing Algorithms](#load-balancing-algorithms)
- [Installation](#installation)
  - [Option 1: Download a prebuilt binary](#option-1-download-a-prebuilt-binary)
- [Configuration](#configuration)
  - [Server](#server)
  - [Balancer](#balancer)
  - [Backends](#backends)
  - [Health Checking](#health-checking)
- [Usage](#usage)
- [Project Structure](#project-structure)
- [License](#license)

<!-- END doctoc -->

## Features

- Eight load-balancing algorithms (four standard and four weighted variants).
- Active backend health checking with automatic failure detection and recovery.
- Configurable backend weights.
- Concurrent request handling and active request tracking.
- Graceful shutdown.
- Structured logging.
- YAML-based configuration.

## Load-Balancing Algorithms

Raccoon LB supports eight algorithms:

| Algorithm | Description |
|---|---|
| Round Robin | Distributes requests sequentially across available backends. |
| Least Connections | Selects the backend with the fewest active connections. |
| Random | Randomly selects an available backend. |
| Power of Two Choices (P2C) | Randomly samples two backends and selects the one with fewer active connections. |
| Weighted Round Robin | Distributes requests proportionally to backend weights using smooth weighted round robin. |
| Weighted Least Connections | Selects the backend with the lowest active-connection-to-weight ratio. |
| Weighted Random | Randomly selects a backend proportionally to its weight. |
| Weighted P2C | Samples two backends and selects based on their active-connection-to-weight ratios. |

## Installation


### Option 1: Download a prebuilt binary

Download the latest Linux AMD64 release from [GitHub Releases](https://github.com/IshaanShivKr/raccoon-lb/releases).

Extract the archive:

```bash
tar -xzf raccoon-lb-linux-amd64.tar.gz
```

Make the binary executable:

```bash
chmod +x raccoon-lb
```

Create your configuration file using the provided example:

```bash
cp config.example.yaml config.yaml
```

Edit `config.yaml` to configure your backend servers and load-balancing algorithm.

Run Raccoon LB:

```bash
./raccoon-lb -config config.yaml
```

## Configuration

Raccoon LB uses YAML configuration.

Example:

```yaml
server:
  address: ":8080"
  read_header_timeout: 5s
  read_timeout: 15s
  write_timeout: 60s
  idle_timeout: 60s
  shutdown_timeout: 10s

balancer:
  algorithm: round_robin

backends:
  - url: http://localhost:9001
    weight: 3
  - url: http://localhost:9002
    weight: 2
  - url: http://localhost:9003
    weight: 1

health:
  interval: 5s
  timeout: 2s
  path: /health
  fall_threshold: 3
  rise_threshold: 2
```

### Server

| Field | Description |
|---|---|
| `address` | Address and port on which the load balancer listens. |
| `read_header_timeout` | Maximum time to read request headers. |
| `read_timeout` | Maximum time to read the entire request. |
| `write_timeout` | Maximum time before timing out response writes. |
| `idle_timeout` | Maximum time to wait for the next request on an idle connection. |
| `shutdown_timeout` | Maximum time allowed for graceful shutdown. |

### Balancer

The `algorithm` field selects the load-balancing strategy.

Supported values:

- `round_robin`
- `least_connections`
- `random`
- `p2c`
- `weighted_round_robin`
- `weighted_least_connections`
- `weighted_random`
- `weighted_p2c`

### Backends

Each backend has:

| Field | Description |
|---|---|
| `url` | Backend server URL. |
| `weight` | Backend selection weight, used by weighted algorithms. |

### Health Checking

| Field | Description |
|---|---|
| `interval` | Time between health checks. |
| `timeout` | Maximum duration of an individual health check. |
| `path` | HTTP endpoint used for health checks. |
| `fall_threshold` | Consecutive failures required to mark a backend unhealthy. |
| `rise_threshold` | Consecutive successes required to mark an unhealthy backend healthy again. |

Raccoon LB periodically sends HTTP requests to each backend's configured health-check endpoint. Unhealthy backends are excluded from request distribution and can rejoin once they recover.

## Usage

Once Raccoon LB is running, send HTTP requests to its listening address.

For example:

```bash
curl http://localhost:8080/foo/bar
```

Raccoon LB forwards the request to a selected backend, preserving the request path.

## Project Structure

```text
raccoon-lb/
├── cmd/
│   └── lb/
├── configs/
│   └── config.example.yaml
├── internal/
│   ├── backend/
│   ├── balancer/
│   ├── config/
│   ├── health/
│   ├── middleware/
│   └── proxy/
├── Dockerfile
├── Makefile
├── go.mod
└── README.md
```

## License

Raccoon LB is licensed under the [MIT License](LICENSE).
