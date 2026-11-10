<p align="center">
  <img src="assets/raccoon.svg" width="180" alt="Raccoon LB mascot">
</p>

<h1 align="center">Raccoon LB</h1>

<p align="center">
  A Layer 7 HTTP load balancer and reverse proxy written in Go.
</p>

Raccoon LB distributes incoming HTTP requests across multiple backend servers using configurable load-balancing algorithms. It includes active health checking, weighted backend selection, concurrent request handling, and graceful shutdown.

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [Features](#features)
- [Load-Balancing Algorithms](#load-balancing-algorithms)
- [Installation](#installation)
  - [Download a prebuilt binary](#download-a-prebuilt-binary)
  - [Pull the Docker image](#pull-the-docker-image)
  - [Install with Go](#install-with-go)
  - [Build from Source](#build-from-source)
- [Create your configuration file](#create-your-configuration-file)
- [Run Raccoon LB](#run-raccoon-lb)
  - [Run the binary](#run-the-binary)
  - [Run with Docker](#run-with-docker)
- [Usage](#usage)
- [Configuration](#configuration)
  - [Server](#server)
  - [Balancer](#balancer)
  - [Backends](#backends)
  - [Health Checking](#health-checking)
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


### Download a prebuilt binary

Download the latest Linux AMD64 release from [GitHub Releases](https://github.com/IshaanShivKr/raccoon-lb/releases).

Extract the archive:

```bash
tar -xzf raccoon-lb-linux-amd64.tar.gz
```

Make the binary executable:

```bash
chmod +x raccoon-lb
```

### Pull the Docker image

Pull the latest Raccoon LB image from GitHub Container Registry:

```bash
docker pull ghcr.io/ishaanshivkr/raccoon-lb:latest
```

### Install with Go

Requires Go 1.27.3 or later.

Install Raccoon LB using `go install`:

```bash
go install github.com/IshaanShivKr/raccoon-lb/cmd/raccoon-lb@latest
```

Ensure that your Go binary directory (`$(go env GOPATH)/bin`) is included in your system's `PATH`.

### Build from Source

Requirements:

- Go 1.27.3 or later
- Git
- Make (optional)

Clone the repository:

```bash
git clone https://github.com/IshaanShivKr/raccoon-lb.git
cd raccoon-lb
```

Build the binary:

```bash
make build
```

If you don't have Make, build the binary directly with Go:

```bash
CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/raccoon-lb ./cmd/raccoon-lb
```

This creates the binary at `bin/raccoon-lb`.

## Create your configuration file

Download the example configuration:

```bash
curl -O https://raw.githubusercontent.com/IshaanShivKr/raccoon-lb/main/configs/config.example.yaml
cp config.example.yaml config.yaml
```

Alternatively, you can download config.example.yaml directly from the [GitHub repository](https://github.com/IshaanShivKr/raccoon-lb/blob/main/configs/config.example.yaml).

If you built Raccoon LB from source, the example configuration is already available in the configs/ directory. Copy it to your working directory:

```bash
cp configs/config.example.yaml config.yaml
```

Edit config.yaml to configure your backend servers and load-balancing algorithm. See [Configuration](#configuration) for details about all available fields.

## Run Raccoon LB

### Run the binary

Use the command that matches how you installed it:

- **Prebuilt binary:**

  ```bash
  ./raccoon-lb -config config.yaml
  ```

- **Installed with `go install`:** the binary is on your `PATH`, so no `./` is needed.

  ```bash
  raccoon-lb -config config.yaml
  ```

- **Built from source:** the binary is in `bin/`.

  ```bash
  ./bin/raccoon-lb -config config.yaml
  ```

### Run with Docker

```bash
docker run -d \
  --name raccoon-lb \
  -p 8080:8080 \
  -v "$(pwd)/config.yaml:/etc/raccoon-lb/config.yaml:ro" \
  --restart unless-stopped \
  ghcr.io/ishaanshivkr/raccoon-lb:latest
```

Check the container logs:

```bash
docker logs raccoon-lb
```

Raccoon LB is now accessible on port 8080 of the machine running Docker. Ensure the port is allowed through your firewall or cloud security group.

To stop and remove the container:

```bash
docker stop raccoon-lb && docker rm raccoon-lb
```

## Usage

Once Raccoon LB is running, send HTTP requests to the address set in `server.address` (port 8080 in the example configuration):

```bash
curl http://localhost:8080/
```

Raccoon LB forwards each request to a healthy backend, preserving the request path, so `http://localhost:8080/foo/bar` reaches `/foo/bar` on the selected backend.

To see load balancing in action, send several requests and watch them spread across your backends:

```bash
for i in {1..6}; do curl -s http://localhost:8080/; echo; done
```

How requests are distributed depends on the configured [algorithm](#load-balancing-algorithms) and backend weights. 

If requests fail, check that your backends are running and return `200 OK` on the [health check path](#health-checking). A backend that fails its checks is logged as `backend unhealthy` and removed from rotation. If no backend is healthy, Raccoon LB returns an error response instead of forwarding the request.

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

## Project Structure

```text
raccoon-lb/
├── cmd/
│   └── raccoon-lb/
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
