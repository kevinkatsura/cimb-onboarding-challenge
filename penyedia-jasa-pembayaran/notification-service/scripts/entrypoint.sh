#!/bin/sh
set -e
echo "==> Running migrations..."
ns-migrate up
echo "==> Starting Notification Service..."
ns-consumer &
PID=$!
shutdown() { echo "==> Shutdown"; kill -TERM "$PID" 2>/dev/null; wait "$PID" 2>/dev/null; ns-migrate down; exit 0; }
trap shutdown TERM INT
wait "$PID"
