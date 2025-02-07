install-arm64: install-tailwind-arm64 install-air
install-x64: install-tailwind-x64 install-air

# Target to install Tailwind CSS binary - MacOS arm64 (M1 onwards)
install-tailwind-arm64:
	@echo "Installing Tailwind..."
	curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-macos-arm64
	chmod +x tailwindcss-macos-arm64
	mv tailwindcss-macos-arm64 tailwindcss

# Target to install Tailwind CSS binary - MacOS x64 (Intel)
install-tailwind-x64:
	@echo "Installing Tailwind..."
	curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-macos-x64
	chmod +x tailwindcss-macos-x64
	mv tailwindcss-macos-x64 tailwindcss

# Target to install Tailwind CSS binary for Linux (x86_64 for GitHub Actions)
install-tailwind-linux:
	@echo "Installing Tailwind for Linux x64..."
	curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-x64
	chmod +x tailwindcss-linux-x64
	mv tailwindcss-linux-x64 tailwindcss

# Target to install Air binary
install-air:
	@echo "Installing Air..."
	go install github.com/air-verse/air@latest