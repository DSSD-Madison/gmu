# better-evidence-project

## Setup
First install golang. The easiest way is with [Brew](brew.sh) using `brew install go`. Next install air and tailwindcss using `make install-all-arm64` or `make install-all-x64` based on your Mac version.

## Development
This project uses a local PostgreSQL database running in Docker for development.

### Local Database Setup
To start the local database:
```bash
docker compose up -d
./scripts/migrate_db.sh # copies prod data into local (optional)
```

### Running the Application
To run the application in development mode:
```bash
air
```

### Database Access
To run psql on your local instance:
```bash
docker exec -it mypostgres psql -U postgres
```

### Frontend Development
To update the CSS:
```bash
./tools/tailwindcss -i ./web/assets/css/input.css -o ./web/assets/css/output.css --minify
```

### Backend Development
To generate Go code from SQL queries:
```bash
sqlc generate
```

## Database Migrations
This project uses Flyway for database migrations. Migrations are stored in `pkg/db/migrations` and rollback scripts in `pkg/db/rollbacks`.

To run migrations on your local database:
```bash
brew install flyway
./scripts/migrate_and_generate.sh local
```

### Creating New Migrations
When you need to make database changes:

1. Create a new migration file in `pkg/db/migrations` with the next version number
2. Create a corresponding rollback file in `pkg/db/rollbacks`
3. Test the migration locally:
   ```bash
   ./scripts/migrate_and_generate.sh local
   ```

### Rolling Back Migrations
If you need to roll back to a previous version:
```bash
./scripts/rollback.sh 2 local # local development
./scripts/rollback.sh 2 prod  # prod database
```
Rollbacks are temporary. If you want a rollback to be permanent, you need to add a new migration.