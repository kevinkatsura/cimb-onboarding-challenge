#!/bin/sh
set -e
echo "==> Running migrations..."
ns-migrate up
echo "==> Starting Notification Service..."
exec ns-consumer
