#!/bin/bash

# Check if environment argument is provided
if [ "$1" != "local" ] && [ "$1" != "prod" ]; then
    echo "Usage: $0 [local|prod]"
    exit 1
fi

ENV=$1

# Load environment variables
if [ -f .env ]; then
    source .env
    echo "Using $ENV environment"
else
    echo "Error: .env file not found"
    exit 1
fi

# Set the appropriate Flyway config file
if [ "$ENV" = "prod" ]; then
    CONFIG_FILE="flyway.prod.conf"
    
    # Confirm before proceeding with production
    read -p "WARNING: You are about to run migrations on the PRODUCTION database. Are you sure? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Operation cancelled."
        exit 1
    fi
else
    CONFIG_FILE="flyway.conf"
fi

# Check if required environment variables are set
if [ -z "$DB_HOST" ] || [ -z "$DB_USER" ] || [ -z "$DB_NAME" ] || [ -z "$DB_PASSWORD" ]; then
    echo "Error: Required environment variables are not set."
    echo "Please ensure DB_HOST, DB_USER, DB_NAME, and DB_PASSWORD are set in your .env file."
    exit 1
fi

echo "Running Flyway migrations for $ENV environment..."

# First, try to baseline the database if needed
echo "Attempting to baseline the database..."
flyway -configFiles=$CONFIG_FILE baseline

# Then run migrations
flyway -configFiles=$CONFIG_FILE migrate

if [ $? -eq 0 ]; then
    echo "Migrations completed successfully."
    
    # Create a schema.sql file for sqlc
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