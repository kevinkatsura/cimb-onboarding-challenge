#!/bin/sh
set -e
echo "==> Running migrations..."
pias-migrate up
echo "==> Starting Payment Initiation Service..."
pias-api &
PID=$!
shutdown() { echo "==> Shutdown"; kill -TERM "$PID" 2>/dev/null; wait "$PID" 2>/dev/null; pias-migrate down; exit 0; }
trap shutdown TERM INT
wait "$PID"
