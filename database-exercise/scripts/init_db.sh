#!/bin/bash

DB_USER=katsuke
DB_NAME=go_db_exercise

echo "Creating database if not exists..."

psql -U $DB_USER -f db/init/001_create_database.sql

echo "Running migrations..."

for file in db/migrations/*.sql
do
  echo "Running $file"
  psql -U $DB_USER -d $DB_NAME -f $file
done

echo "Database initialized."