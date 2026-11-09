.DEFAULT_GOAL := help

.PHONY: help build build-linux-amd64 build-linux-arm64 run clean fmt vet check tidy deps all

help:
	@echo "Raccoon LB"
	@echo "  make build  - Build binary"
	@echo "  make run    - Run application"
	@echo "  make clean  - Remove build artifacts"
	@echo "  make fmt    - Format Go files"
	@echo "  make vet    - Run go vet"
	@echo "  make check  - Run formatting check and vet"
	@echo "  make tidy   - Tidy dependencies"
	@echo "  make deps   - Download dependencies"

build:
	mkdir -p bin
	go build -trimpath -o bin/lb ./cmd/lb

build-linux-amd64:
	mkdir -p bin
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -o bin/raccoon-lb-linux-amd64 ./cmd/lb

build-linux-arm64:
	mkdir -p bin
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -o bin/raccoon-lb-linux-arm64 ./cmd/lb

run:
	go run ./cmd/lb

clean:
	rm -rf bin

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
