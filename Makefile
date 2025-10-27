.PHONY: test test-integration test-all build install lint clean help

# Default target
.DEFAULT_GOAL := help

## test: Run unit tests
test:
	go test -v -short ./...

## test-integration: Run integration tests (requires terraform binary)
test-integration:
	go test -v -tags=integration ./...

## test-e2e: Run end-to-end tests (requires AWS credentials)
test-e2e:
	go test -v -tags=e2e -timeout=30m ./...

## test-all: Run all tests except e2e
test-all: test test-integration

## build: Build the forge binary
build:
	go build -o bin/forge ./cmd/forge

## install: Install forge to GOPATH/bin
install:
	go install ./cmd/forge

## lint: Run linter
lint:
	@which golangci-lint > /dev/null || (echo "golangci-lint not found, install from https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run ./...

## fmt: Format code
fmt:
	go fmt ./...
	gofmt -s -w .

## clean: Remove build artifacts
clean:
	rm -rf bin/
	rm -f forge
	find . -name "*.test" -delete
	find . -name "*.out" -delete

## tidy: Tidy go modules
tidy:
	go mod tidy

## generate: Run go generate
generate:
	go generate ./...

## deps: Download dependencies
deps:
	go mod download

## verify: Run tests and linting
verify: test lint

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'
