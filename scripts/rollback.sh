#!/bin/bash

# Check if environment is provided
if [ -z "$2" ]; then
  echo "Usage: ./scripts/rollback.sh <version_number> <environment>"
  echo "  version_number: Version to roll back to (e.g. 1 to roll back to baseline)"
  echo "  environment: 'local' or 'prod'"
  exit 1
fi

TARGET_VERSION=$1
ENV=$2

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
else
    CONFIG_FILE="flyway.conf"
fi

# Check if environment variables are set
if [ -z "$DB_HOST" ] || [ -z "$DB_USER" ] || [ -z "$DB_NAME" ] || [ -z "$DB_PASSWORD" ]; then
    echo "Error: Required environment variables are not set."
    echo "Please check your .env file configuration."
    exit 1
fi

# Export environment variables for Flyway
export DB_HOST
export DB_USER
export DB_NAME
export DB_PASSWORD

# Get current version from flyway
echo "Getting current schema version..."
# Get Schema version directly from the info output
FLYWAY_INFO=$(flyway -configFiles=$CONFIG_FILE info)
echo "Flyway info output:"
echo "$FLYWAY_INFO"

# Extract the schema version line
SCHEMA_VERSION_LINE=$(echo "$FLYWAY_INFO" | grep "Schema version:")
CURRENT_VERSION=$(echo "$SCHEMA_VERSION_LINE" | awk '{print $3}')

if [ -z "$CURRENT_VERSION" ]; then
  echo "Error: Could not determine current schema version."
  echo "Please check the Flyway info output above."
  exit 1
fi

echo "Current version: $CURRENT_VERSION"
echo "Target version: $TARGET_VERSION"

# Make sure both are treated as integers
if [ "$TARGET_VERSION" -ge "$CURRENT_VERSION" ] 2>/dev/null; then
  echo "Error: Target version must be less than current version."
  exit 1
fi

# If production environment, ask for confirmation
if [ "$ENV" = "prod" ]; then
  echo "WARNING: You are about to rollback the PRODUCTION database!"
  read -p "Are you sure you want to continue? (y/N) " -n 1 -r
  echo
  if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Rollback cancelled."
    exit 1
  fi
fi

# Execute rollbacks for all versions from current down to target+1
echo "Rolling back from version $CURRENT_VERSION to version $TARGET_VERSION..."

for ((version=CURRENT_VERSION; version>TARGET_VERSION; version--)); do
  ROLLBACK_SCRIPT="pkg/db/rollbacks/R${version}__rollback.sql"
  
  if [ ! -f "$ROLLBACK_SCRIPT" ]; then
    echo "Error: Rollback script $ROLLBACK_SCRIPT not found."
    exit 1
  fi
  
  echo "Executing rollback script for version $version: $ROLLBACK_SCRIPT"
  PGPASSWORD="$DB_PASSWORD" psql \
    --host="$DB_HOST" \
    --username="$DB_USER" \
    --dbname="$DB_NAME" \
    -f "$ROLLBACK_SCRIPT"
  
  if [ $? -ne 0 ]; then
    echo "Error: Rollback for version $version failed."
    exit 1
  fi
done

# Update Flyway schema history table
echo "Updating Flyway schema history table..."
PGPASSWORD="$DB_PASSWORD" psql \
  --host="$DB_HOST" \
  --username="$DB_USER" \
  --dbname="$DB_NAME" \
  -c "DELETE FROM flyway_schema_history WHERE installed_rank > (SELECT installed_rank FROM flyway_schema_history WHERE version = '$TARGET_VERSION')"

if [ $? -ne 0 ]; then
  echo "Error: Failed to update Flyway schema history table."
  exit 1
fi

# Update schema.sql file for sqlc
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
    echo "Rollback to version $TARGET_VERSION completed successfully."
else
    echo "Error: Go code generation failed."
    exit 1
fi