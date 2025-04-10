#!/bin/bash

# Load environment variables
if [ -f .env ]; then
    source .env
    echo "Using .env configuration"
else
    echo "Error: .env file not found"
    exit 1
fi

# Check if required environment variables are set
if [ -z "$DB_HOST" ] || [ -z "$DB_USER" ] || [ -z "$DB_NAME" ] || [ -z "$DB_PASSWORD" ]; then
    echo "Error: Required environment variables are not set."
    echo "Please ensure DB_HOST, DB_USER, DB_NAME, and DB_PASSWORD are set in your .env file."
    exit 1
fi

echo "Running Flyway migrations..."

export DB_HOST
export DB_USER
export DB_NAME
export DB_PASSWORD

echo "Attempting to baseline the database..."
flyway -configFiles=flyway.conf baseline

flyway -configFiles=flyway.conf migrate

if [ $? -eq 0 ]; then
    echo "Migrations completed successfully."

    echo "Creating schema.sql file for sqlc..."
    PGPASSWORD="$DB_PASSWORD" pg_dump \
        --host="$DB_HOST" \
        --username="$DB_USER" \
        --dbname="$DB_NAME" \
        --schema-only \
        --no-owner \
        --no-privileges \
        > schema.sql

    if [ $? -ne 0 ]; then
        echo "Error: Failed to create schema.sql file."
        exit 1
    fi

    echo "Generating Go code from database schema..."
    sqlc generate

    if [ $? -eq 0 ]; then
        echo "Go code generation completed successfully."
    else
        echo "Error: Go code generation failed."
        exit 1
    fi
else
    echo "Error: Migrations failed."
    exit 1
fi
