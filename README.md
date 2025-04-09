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

### Frontend Development
To update the CSS:
```bash
./tools/tailwindcss -i ./web/assets/css/input.css -o ./web/assets/css/output.css --minify
```
