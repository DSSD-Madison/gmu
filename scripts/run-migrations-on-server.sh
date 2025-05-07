#!/bin/bash
set -e # Exit on error

echo "Running Flyway migrations on server..."
echo "Current Directory: $(pwd)"

if [ -z "$DB_HOST" ] || [ -z "$DB_USER" ] || [ -z "$DB_NAME" ] | [ -z "$DB_PASSWORD" ]; then
    echo "Error: DB_HOST, DB_USER, DB_NAME, or DB_PASSWORD not set in environment"
    exit 1
fi

echo "Attempting to baseline the database..."
flyway baseline -baselineOnMigrate=true

echo "Applying migrations..."
flyway migrate

if [ $? -eq 0 ]; then
    echo "Migrations completed successfully on server."
else
    echo "Error: Migrations failed on server."
    exit 1
fi
