# CoopAndreas Go Server Makefile

.PHONY: build run test clean server client

# Default target
all: build

# Build both server and client
build: server client

# Build server
server:
	@echo "Building server..."
	@go build -o bin/coopandreas-server cmd/server/main.go
	@echo "Server built: bin/coopandreas-server"

# Build client
client:
	@echo "Building test client..."
	@go build -o bin/test-client cmd/test-client/main.go
	@echo "Test client built: bin/test-client"

# Build with optimizations for production
release:
	@echo "Building optimized server..."
	@go build -ldflags="-s -w" -o bin/coopandreas-server cmd/server/main.go
	@echo "Building optimized test client..."
	@go build -ldflags="-s -w" -o bin/test-client cmd/test-client/main.go
	@echo "Release builds complete"

# Run server
run-server:
	@echo "Starting server..."
	@go run cmd/server/main.go

# Run test client
run-client:
	@echo "Starting test client..."
	@go run cmd/test-client/main.go

# Run tests
test:
	@echo "Running tests..."
	@go test ./...

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Clean built binaries
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@echo "Clean complete"

# Create bin directory
bin:
	@mkdir -p bin

# Ensure bin directory exists for builds
server client release: | bin

# Development workflow - format, test, build
dev: fmt test build

# Show help
help:
	@echo "Available targets:"
	@echo "  build        - Build server and client"
	@echo "  server       - Build server only"
	@echo "  client       - Build client only"
	@echo "  release      - Build optimized binaries"
	@echo "  run-server   - Run server in development mode"
	@echo "  run-client   - Run test client"
	@echo "  test         - Run tests"
	@echo "  fmt          - Format code"
	@echo "  clean        - Remove built binaries"
	@echo "  dev          - Format, test, and build"
	@echo "  help         - Show this help"
