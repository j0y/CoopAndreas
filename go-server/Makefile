# CoopAndreas Go Server Makefile

.PHONY: build run test clean server client dummy-client

# Default target
all: build

# Build server, test client, and dummy client
build: server client dummy-client

# Build server
server:
	@echo "Building server..."
	@go build -o bin/coopandreas-server cmd/server/main.go
	@echo "Server built: bin/coopandreas-server"

# Build test client
client:
	@echo "Building test client..."
	@go build -o bin/test-client cmd/test-client/main.go
	@echo "Test client built: bin/test-client"

# Build dummy client
dummy-client:
	@echo "Building dummy client..."
	@go build -o bin/dummy-client cmd/dummy-client/main.go
	@echo "Dummy client built: bin/dummy-client"

# Build with optimizations for production
release:
	@echo "Building optimized server..."
	@go build -ldflags="-s -w" -o bin/coopandreas-server cmd/server/main.go
	@echo "Building optimized test client..."
	@go build -ldflags="-s -w" -o bin/test-client cmd/test-client/main.go
	@echo "Building optimized dummy client..."
	@go build -ldflags="-s -w" -o bin/dummy-client cmd/dummy-client/main.go
	@echo "Release builds complete"

# Run server
run-server:
	@echo "Starting server..."
	@go run cmd/server/main.go

# Run test client
run-client:
	@echo "Starting test client..."
	@go run cmd/test-client/main.go

# Run dummy client with default settings
run-dummy:
	@echo "Starting dummy client..."
	@go run cmd/dummy-client/main.go

# Run dummy client with custom settings (usage: make run-dummy-custom NAME=YourName IP=192.168.1.1 PORT=7777)
run-dummy-custom:
	@echo "Starting dummy client with custom settings..."
	@go run cmd/dummy-client/main.go -name="$(NAME)" -ip="$(IP)" -port="$(PORT)"

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
server client dummy-client release: | bin

# Development workflow - format, test, build
dev: fmt test build

# Show help
help:
	@echo "Available targets:"
	@echo "  build        - Build server, test client, and dummy client"
	@echo "  server       - Build server only"
	@echo "  client       - Build test client only"
	@echo "  dummy-client - Build dummy client only"
	@echo "  release      - Build optimized binaries"
	@echo "  run-server   - Run server in development mode"
	@echo "  run-client   - Run test client"
	@echo "  run-dummy    - Run dummy client with default settings (localhost:6767, name: DummyPlayer)"
	@echo "  run-dummy-custom NAME=<name> IP=<ip> PORT=<port> - Run dummy client with custom settings"
	@echo "  test         - Run tests"
	@echo "  fmt          - Format code"
	@echo "  clean        - Remove built binaries"
	@echo "  dev          - Format, test, and build"
	@echo "  help         - Show this help"
