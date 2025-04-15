#!/bin/bash

# Load environment variables from .env file
set -a
source ../.env
set +a

# File containing UUIDs (one per line)
UUID_FILE="duplicate_uuids.txt"

if [ ! -f "$UUID_FILE" ]; then
  echo "Error: UUID file '$UUID_FILE' not found."
  exit 1
fi

# Construct SQL query to batch update all UUIDs
SQL="UPDATE documents SET has_duplicate = true WHERE uuid IN ("

while IFS= read -r uuid; do
  SQL+="'$uuid',"
done < "$UUID_FILE"

# Remove trailing comma and close parentheses
SQL=${SQL%,})
SQL+=");"

# Execute the update query against production database
PGPASSWORD=$PROD_PASSWORD psql -h $PROD_HOST -U $PROD_USER -d $PROD_NAME -c "$SQL"

if [ $? -ne 0 ]; then
  echo "Error: Failed to update production database."
  exit 1
fi

echo "Successfully marked duplicates in production database."
