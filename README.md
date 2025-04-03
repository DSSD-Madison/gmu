# better-evidence-project

## Setup
First install golang. The easiest way is with [Brew](brew.sh) using `brew install go`. Next install air and tailwindcss using `make install-all-arm64` or `make install-all-x64` based on your Mac version.

## Development
To set up your local database, run `docker compose up -d`. Copy the production database into the local database with `./migrate_db.sh`. To run the website, run `air` in your terminal. Changes will hot reload the page. To run your local database instance, run `docker exec -it mypostgres psql -U postgres`. If you need to apply the effects of tailwind changes, run `./tailwindcss -i ./static/css/input.css -o ./static/css/output.css --minify`. If you need to generate go code for your sql queries, run sqlc generate.