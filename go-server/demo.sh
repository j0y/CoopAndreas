#!/bin/bash

# Demo script showing zerolog output

echo "=== Starting CoopAndreas Go Server Demo ==="
echo "This demonstrates the new zerolog structured logging"
echo

# Start server in background
echo "Starting server..."
cd /home/yar/Yar/Projects/CoopAndreas/go-server
./bin/coopandreas-server &
SERVER_PID=$!

# Wait for server to start
sleep 2

# Run client test
echo "Running test client..."
./bin/test-client

# Wait a moment
sleep 2

# Kill server
echo "Stopping server..."
kill $SERVER_PID

echo "Demo complete!"
