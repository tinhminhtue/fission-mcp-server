.PHONY: build build-linux build-darwin build-windows test test-smoke clean run help

# Binary name
BINARY_NAME=fission-mcp-server

# Build flags
LDFLAGS=-ldflags "-s -w"

# Default target
.DEFAULT_GOAL := help

## build: Build binary for current platform
build:
	@echo "Building $(BINARY_NAME)..."
	@go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/server
	@echo "Build complete: $(BINARY_NAME)"

## build-linux: Build binary for Linux
build-linux:
	@echo "Building $(BINARY_NAME) for Linux..."
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME)-linux ./cmd/server
	@echo "Build complete: $(BINARY_NAME)-linux"

## build-darwin: Build binary for macOS
build-darwin:
	@echo "Building $(BINARY_NAME) for macOS..."
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME)-darwin ./cmd/server
	@echo "Build complete: $(BINARY_NAME)-darwin"

## build-windows: Build binary for Windows
build-windows:
	@echo "Building $(BINARY_NAME) for Windows..."
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME)-windows.exe ./cmd/server
	@echo "Build complete: $(BINARY_NAME)-windows.exe"

## build-all: Build binaries for all platforms
build-all: build-linux build-darwin build-windows
	@echo "Build complete for all platforms"

## test: Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

## test-smoke: Run Function API smoke tests
test-smoke:
	@echo "Running Function API smoke tests..."
	@echo "Make sure fission-mcp-server is running before executing smoke tests"
	@go test -v ./tests/

## test-coverage: Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## run: Run application
run:
	@echo "Running server with go run..."
	@go run ./cmd/server

## clean: Remove build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f $(BINARY_NAME)
	@rm -f $(BINARY_NAME)-linux
	@rm -f $(BINARY_NAME)-darwin
	@rm -f $(BINARY_NAME)-windows.exe
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

## fmt: Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "Format complete"

## lint: Run linter
lint:
	@echo "Running linter..."
	@go vet ./...
	@echo "Lint complete"

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies updated"

## help: Show this help message
help:
	@echo "Available targets:"
	@grep -E '^##' Makefile | sed 's/## //'
