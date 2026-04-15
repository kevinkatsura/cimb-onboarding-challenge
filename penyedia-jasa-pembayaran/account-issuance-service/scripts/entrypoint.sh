#!/bin/sh
set -e
echo "==> Running migrations..."
ais-migrate up
echo "==> Starting Account Issuance Service..."
ais-api &
API_PID=$!
shutdown() {
    echo "==> Shutdown signal received"
    kill -TERM "$API_PID" 2>/dev/null
    wait "$API_PID" 2>/dev/null
    echo "==> Running migration down..."
    ais-migrate down
    echo "==> Shutdown complete"
    exit 0
}
trap shutdown TERM INT
wait "$API_PID"
