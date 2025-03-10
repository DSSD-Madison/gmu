#!/bin/bash

# Move to the script's parent directory
cd "$(dirname "$0")/.."

# Load environment variables from .env file
set -a
source .env
set +a

DUMP_FILE="prod_dump.sqlc"

echo "Dumping production database..."
PGPASSWORD=$PROD_PASSWORD pg_dump -h $PROD_HOST -U $PROD_USER -d $PROD_DB -F c -f $DUMP_FILE

if [ $? -ne 0 ]; then
  echo "Error: Failed to dump the database."
  exit 1
fi

echo "Copying dump to Docker container..."
docker cp $DUMP_FILE $DOCKER_CONTAINER:/tmp/$DUMP_FILE

if [ $? -ne 0 ]; then
  echo "Error: Failed to copy dump to Docker container."
  exit 1
fi

echo "Dropping existing local database..."
# Terminate active connections and drop the target database
docker exec -it $DOCKER_CONTAINER psql -U $DOCKER_USER -d postgres -c "SELECT pg_terminate_backend(pg_stat_activity.pid) FROM pg_stat_activity WHERE datname = '$DOCKER_DB' AND pid <> pg_backend_pid();"
docker exec -it $DOCKER_CONTAINER psql -U $DOCKER_USER -d postgres -c "DROP DATABASE IF EXISTS $DOCKER_DB;"

if [ $? -ne 0 ]; then
  echo "Error: Failed to drop the existing database."
  exit 1
fi

echo "Creating new local database..."
docker exec -it $DOCKER_CONTAINER psql -U $DOCKER_USER -d postgres -c "CREATE DATABASE $DOCKER_DB OWNER $DOCKER_USER;"

if [ $? -ne 0 ]; then
  echo "Error: Failed to create the new database."
  exit 1
fi

echo "Restoring dump into local PostgreSQL database..."
docker exec -it $DOCKER_CONTAINER pg_restore -U $DOCKER_USER -d $DOCKER_DB /tmp/$DUMP_FILE

if [ $? -ne 0 ]; then
  echo "Error: Failed to restore the database."
  exit 1
fi

echo "Cleaning up dump file in Docker container..."
docker exec -it $DOCKER_CONTAINER rm /tmp/$DUMP_FILE

echo "Database migration complete!"

