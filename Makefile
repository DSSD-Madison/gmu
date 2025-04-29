SQLC_VERSION = v1.24.0
TAILWIND_VERSION = v4.0.6
GOLANGCI_LINT_VERSION = v1.64.8
TAILWIND_BASE_URL = https://github.com/tailwindlabs/tailwindcss/releases/download/$(TAILWIND_VERSION)/

install-all-arm64: install-air install-templ install-sqlc-arm64 install-golangci-lint
	$(MAKE) install-tailwind FILE=tailwindcss-macos-arm64

install-all-x64: install-air install-templ install-sqlc-x64 install-golangci-lint
	$(MAKE) install-tailwind FILE=tailwindcss-macos-x64

install-all-linux: install-air install-templ install-sqlc-linux install-golangci-lint
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

install-golangci-lint:
	@echo "Installing golangci-lint ($(GOLANGCI_LINT_VERSION))..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)


install-sqlc-arm64:
	@echo "Installing sqlc ($(SQLC_VERSION)) for macOS ARM64..."
	curl -sL https://github.com/sqlc-dev/sqlc/releases/download/$(SQLC_VERSION)/sqlc_$(subst v,,$(SQLC_VERSION))_darwin_arm64.tar.gz | tar -xz
	chmod +x sqlc
	sudo mv sqlc /usr/local/bin/sqlc

install-sqlc-x64:
	@echo "Installing sqlc ($(SQLC_VERSION)) for macOS x64..."
	curl -sL https://github.com/sqlc-dev/sqlc/releases/download/$(SQLC_VERSION)/sqlc_$(subst v,,$(SQLC_VERSION))_darwin_amd64.tar.gz | tar -xz
	chmod +x sqlc
	sudo mv sqlc /usr/local/bin/sqlc

install-sqlc-linux:
	@echo "Installing sqlc ($(SQLC_VERSION)) for Linux x64..."
	curl -sL https://github.com/sqlc-dev/sqlc/releases/download/$(SQLC_VERSION)/sqlc_$(subst v,,$(SQLC_VERSION))_linux_amd64.tar.gz | tar -xz
	chmod +x sqlc
	sudo mv sqlc /usr/local/bin/sqlc
