
.DEFAULT_GOAL := help

.PHONY: help build build-linux-amd64 run clean fmt vet check tidy deps all

help:
	@echo "Raccoon LB"
	@echo "  make build              - Build binary for current platform"
	@echo "  make build-linux-amd64  - Build Linux AMD64 binary"
	@echo "  make run                - Run application"
	@echo "  make clean              - Remove build artifacts"
	@echo "  make fmt                - Format Go files"
	@echo "  make vet                - Run go vet"
	@echo "  make check              - Run formatting check, vet and build"
	@echo "  make tidy               - Tidy dependencies"
	@echo "  make deps               - Download dependencies"

build:
	mkdir -p bin
	go build -trimpath -o bin/lb ./cmd/lb

build-linux-amd64:
	mkdir -p dist
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o dist/raccoon-lb-linux-amd64 ./cmd/lb

run:
	go run ./cmd/lb

clean:
	rm -rf bin dist

fmt:
	gofmt -w .

vet:
	go vet ./...

check:
	test -z "$$(gofmt -l .)"
	go vet ./...
	go build ./...

tidy:
	go mod tidy

deps:
	go mod download

all: fmt tidy vet build
