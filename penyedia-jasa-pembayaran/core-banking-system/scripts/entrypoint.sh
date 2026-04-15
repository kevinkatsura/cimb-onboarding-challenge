#!/bin/sh
set -e

echo "==> Running migrations..."
cbs-migrate up

echo "==> Starting Core Banking System gRPC server..."
cbs-server &
SERVER_PID=$!

shutdown() {
    echo "==> Shutdown signal received"
    kill -TERM "$SERVER_PID" 2>/dev/null
    wait "$SERVER_PID" 2>/dev/null
    echo "==> Running migration down..."
    cbs-migrate down
    echo "==> Shutdown complete"
    exit 0
}

trap shutdown TERM INT
wait "$SERVER_PID"
