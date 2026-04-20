#!/bin/sh
set -e

echo "=== Fraud Detection Service ==="
echo "Running Alembic migrations..."
export PYTHONPATH=.
cd /app
alembic upgrade head
echo "Migrations complete."

echo "Starting FastAPI server on port ${HTTP_PORT:-8085}..."
exec uvicorn app.main:app \
    --host 0.0.0.0 \
    --port "${HTTP_PORT:-8085}" \
    --log-level info
