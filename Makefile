SQLC_VERSION = v1.24.0
TAILWIND_VERSION = v4.0.6
TAILWIND_BASE_URL = https://github.com/tailwindlabs/tailwindcss/releases/download/$(TAILWIND_VERSION)/

install-all-arm64: install-air install-templ install-sqlc-arm64
	$(MAKE) install-tailwind FILE=tailwindcss-macos-arm64

install-all-x64: install-air install-templ install-sqlc-x64
	$(MAKE) install-tailwind FILE=tailwindcss-macos-x64

install-all-linux: install-air install-templ install-sqlc-linux
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

install-sqlc-arm64:
	@echo "Installing sqlc ($(SQLC_VERSION)) for macOS ARM64..."
	curl -sL https://github.com/sqlc-dev/sqlc/releases/download/$(SQLC_VERSION)/sqlc_$(subst v,,$(SQLC_VERSION))_darwin_arm64.tar.gz | tar -xz
	chmod +x sqlc
	mkdir -p tools
	mv sqlc tools/sqlc
	$(MAKE) add-tools-to-path

install-sqlc-x64:
	@echo "Installing sqlc ($(SQLC_VERSION)) for macOS x64..."
	curl -sL https://github.com/sqlc-dev/sqlc/releases/download/$(SQLC_VERSION)/sqlc_$(subst v,,$(SQLC_VERSION))_darwin_amd64.tar.gz | tar -xz
	chmod +x sqlc
	mkdir -p tools
	mv sqlc tools/sqlc
	$(MAKE) add-tools-to-path

install-sqlc-linux:
	@echo "Installing sqlc ($(SQLC_VERSION)) for Linux x64..."
	curl -sL https://github.com/sqlc-dev/sqlc/releases/download/$(SQLC_VERSION)/sqlc_$(subst v,,$(SQLC_VERSION))_linux_amd64.tar.gz | tar -xz
	chmod +x sqlc
	mkdir -p tools
	mv sqlc tools/sqlc
	$(MAKE) add-tools-to-path

add-tools-to-path:
	@echo "Ensuring tools/ directory is in PATH..."
	@if ! echo $$PATH | grep -q "$(pwd)/tools"; then \
		echo "export PATH=\$$PATH:$(pwd)/tools" >> ~/.bashrc 2>/dev/null || true; \
		echo "export PATH=\$$PATH:$(pwd)/tools" >> ~/.zshrc 2>/dev/null || true; \
	fi
	@export PATH=$$PATH:$(pwd)/tools


