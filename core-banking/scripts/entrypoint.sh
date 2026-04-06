#!/bin/sh
set -e

echo "==> Running migrations..."
core-banking-migrate up

echo "==> Running seeder..."
core-banking-seed

echo "==> Starting API server..."
core-banking-api &
API_PID=$!

# Trap SIGTERM/SIGINT for graceful shutdown
shutdown() {
    echo "==> Shutdown signal received"

    echo "==> Stopping API server (PID $API_PID)..."
    kill -TERM "$API_PID" 2>/dev/null
    wait "$API_PID" 2>/dev/null
    echo "==> API server stopped"

    echo "==> Running migration down..."
    core-banking-migrate down
    echo "==> Migration down completed"

    echo "==> Shutdown complete"
    exit 0
}

trap shutdown TERM INT

# Wait for API process
wait "$API_PID"
