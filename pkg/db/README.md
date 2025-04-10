# Database README.md

### Database Access
To run psql on your local instance:
```bash
docker exec -it mypostgres psql -U postgres
```

### Backend Development
To generate Go code from SQL queries:
```bash
sqlc generate
```

## Database Migrations
This project uses Flyway for database migrations. Migrations are stored in `pkg/db/migrations` and rollback scripts in `pkg/db/rollbacks`.

To run migrations on your database:
```bash
brew install flyway
./scripts/migrate_and_generate.sh
```

### Creating New Migrations
When you need to make database changes:

1. Create a new migration file in `pkg/db/migrations` with the next version number
2. Create a corresponding rollback file in `pkg/db/rollbacks`
3. Test the migration locally:
   ```bash
   ./scripts/migrate_and_generate.sh
   ```

### Rolling Back Migrations
If you need to roll back to a previous version:
```bash
./scripts/rollback.sh 2 local # local development
./scripts/rollback.sh 2 prod  # prod database
```
Rollbacks are temporary. If you want a rollback to be permanent, you need to add a new migration.