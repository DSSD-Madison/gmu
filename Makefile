TAILWIND_VERSION = v4.0.6
TAILWIND_BASE_URL = https://github.com/tailwindlabs/tailwindcss/releases/download/$(TAILWIND_VERSION)/

# call install-arm64, -x64, or -linux, based on your OS
install-all-arm64: install-air install-templ install-sqlc
	$(MAKE) install-tailwind FILE=tailwindcss-macos-arm64

install-all-x64: install-air install-templ install-sqlc
	$(MAKE) install-tailwind FILE=tailwindcss-macos-x64

install-all-linux: install-air install-templ install-sqlc
	$(MAKE) install-tailwind FILE=tailwindcss-linux-x64

install-tailwind:
	@echo "Downloading Tailwind..."
	curl -sLO $(TAILWIND_BASE_URL)$(FILE)
	chmod +x $(FILE)
	mkdir -p tools
	mv $(FILE) tools/tailwindcss

install-air:
	@echo "Installing Air..."
	go install github.com/air-verse/air@latest

install-templ:
	@echo "Installing Templ..."
	go install github.com/a-h/templ/cmd/templ@v0.3.833

install-sqlc:
	@echo "Installing sqlc..."
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest