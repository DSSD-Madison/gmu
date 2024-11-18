install: install-tailwind install-air

# Target to install Tailwind CSS binary
install-tailwind:
	@echo "Installing Tailwind..."
	curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-macos-arm64
	chmod +x tailwindcss-macos-arm64
	mv tailwindcss-macos-arm64 tailwindcss


# Target to install Air binary
install-air:
	@echo "Installing Air..."
	go install github.com/air-verse/air@latest