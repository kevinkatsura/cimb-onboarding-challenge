#!/bin/bash

DB_USER=katsuke
DB_NAME=go_db_exercise

echo "Seeding database..."

psql -U $DB_USER -d $DB_NAME -f db/seeds/seed_data.sql

echo "Seed completed."