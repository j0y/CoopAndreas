#!/bin/bash

# CoopAndreas Dummy Client Launcher
# Launches multiple dummy clients and handles graceful shutdown

set -e

# Configuration
NUM_CLIENTS=127
SERVER_IP="localhost"
SERVER_PORT=6767
CLIENT_BINARY="./bin/dummy-client"
LOG_DIR="./logs/dummy_clients"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Array to store PIDs of client processes
declare -a CLIENT_PIDS=()

# Function to print colored output
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to cleanup all client processes
cleanup() {
    log_warning "Received interrupt signal. Shutting down all clients..."
    
    if [ ${#CLIENT_PIDS[@]} -eq 0 ]; then
        log_info "No clients to shutdown."
        exit 0
    fi
    
    log_info "Terminating ${#CLIENT_PIDS[@]} client processes..."
    
    # Send TERM signal to all clients
    for pid in "${CLIENT_PIDS[@]}"; do
        if kill -0 "$pid" 2>/dev/null; then
            log_info "Terminating client with PID $pid"
            kill -TERM "$pid" 2>/dev/null || true
        fi
    done
    
    # Wait a bit for graceful shutdown
    sleep 2
    
    # Force kill any remaining processes
    for pid in "${CLIENT_PIDS[@]}"; do
        if kill -0 "$pid" 2>/dev/null; then
            log_warning "Force killing client with PID $pid"
            kill -KILL "$pid" 2>/dev/null || true
        fi
    done
    
    log_success "All clients terminated."
    exit 0
}

# Function to check if binary exists and is executable
check_binary() {
    if [ ! -f "$CLIENT_BINARY" ]; then
        log_error "Client binary not found at $CLIENT_BINARY"
        log_info "Please build the dummy client first:"
        log_info "  go build -o bin/dummy-client cmd/dummy-client/main.go"
        exit 1
    fi
    
    if [ ! -x "$CLIENT_BINARY" ]; then
        log_error "Client binary is not executable: $CLIENT_BINARY"
        exit 1
    fi
}

# Function to create log directory
setup_logging() {
    if [ ! -d "$LOG_DIR" ]; then
        log_info "Creating log directory: $LOG_DIR"
        mkdir -p "$LOG_DIR"
    fi
    
    # Clean old logs
    if [ -d "$LOG_DIR" ]; then
        log_info "Cleaning old log files..."
        rm -f "$LOG_DIR"/client_*.log
    fi
}

# Function to launch a single client
launch_client() {
    local client_id=$1
    local client_name="DummyClient_$(printf "%03d" $client_id)"
    local log_file="$LOG_DIR/client_${client_id}.log"
    
    # Launch client in background
    "$CLIENT_BINARY" \
        -ip "$SERVER_IP" \
        -port "$SERVER_PORT" \
        -name "$client_name" \
        > "$log_file" 2>&1 &
    
    local pid=$!
    CLIENT_PIDS+=($pid)
    
    log_info "Started client $client_id ($client_name) with PID $pid"
}

# Function to monitor clients
monitor_clients() {
    local running_count=0
    local failed_count=0
    
    log_info "Monitoring client processes..."
    
    for pid in "${CLIENT_PIDS[@]}"; do
        if kill -0 "$pid" 2>/dev/null; then
            ((running_count++))
        else
            ((failed_count++))
        fi
    done
    
    log_info "Status: $running_count running, $failed_count failed/stopped"
    
    if [ $failed_count -gt 0 ]; then
        log_warning "Some clients have failed. Check log files in $LOG_DIR"
    fi
}

# Main function
main() {
    log_info "CoopAndreas Dummy Client Launcher"
    log_info "Configuration:"
    log_info "  Clients: $NUM_CLIENTS"
    log_info "  Server: $SERVER_IP:$SERVER_PORT"
    log_info "  Binary: $CLIENT_BINARY"
    log_info "  Logs: $LOG_DIR"
    echo
    
    # Setup interrupt handler
    trap cleanup SIGINT SIGTERM
    
    # Pre-flight checks
    check_binary
    setup_logging
    
    log_info "Starting $NUM_CLIENTS dummy clients..."
    echo
    
    # Launch all clients
    for ((i=1; i<=NUM_CLIENTS; i++)); do
        launch_client $i
        
        # Small delay to prevent overwhelming the server
        sleep 0.1
        
        # Progress indicator
        if [ $((i % 10)) -eq 0 ]; then
            log_success "Launched $i/$NUM_CLIENTS clients"
        fi
    done
    
    echo
    log_success "All $NUM_CLIENTS clients started!"
    log_info "Client PIDs: ${CLIENT_PIDS[*]}"
    echo
    log_info "Press Ctrl+C to stop all clients"
    log_info "Logs are available in: $LOG_DIR"
    echo
    
    # Monitor clients periodically
    while true; do
        sleep 10
        monitor_clients
    done
}

# Run main function
main "$@"
