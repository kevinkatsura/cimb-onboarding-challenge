#!/bin/sh
set -e
echo "==> Running migrations..."
ains-migrate up
echo "==> Starting Account Issuance Service..."
ains-api &
API_PID=$!
shutdown() {
    echo "==> Shutdown signal received"
    kill -TERM "$API_PID" 2>/dev/null
    wait "$API_PID" 2>/dev/null
    echo "==> Running migration down..."
    ains-migrate down
    echo "==> Shutdown complete"
    exit 0
}
trap shutdown TERM INT
wait "$API_PID"
