.DEFAULT_GOAL := help

.PHONY: help build run clean fmt vet check tidy

help:
	@echo "Raccoon LB"
	@echo "  make build  - Build binary"
	@echo "  make run    - Run application"
	@echo "  make clean  - Remove build artifacts"
	@echo "  make fmt    - Format Go files"
	@echo "  make vet    - Run go vet"
	@echo "  make check  - Run formatting check, vet and build"
	@echo "  make tidy   - Tidy dependencies"

build:
	mkdir -p bin
	go build -trimpath -o bin/raccoon-lb ./cmd/raccoon-lb

run:
	go run ./cmd/raccoon-lb

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
