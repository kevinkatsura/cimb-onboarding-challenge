#!/bin/bash

echo "Initializing database..."

./scripts/init_db.sh

echo "Seeding database..."

./scripts/seed_db.sh

echo "Running Go backend..."

go run cmd/server/main.go