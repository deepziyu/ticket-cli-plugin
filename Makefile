# Makefile for ticket-cli-plugin
.PHONY: build build-linux build-windows build-darwin build-linux-amd64 build-linux-arm64 build-windows-amd64 build-darwin-amd64 build-darwin-arm64 build-all clean help

BINARY_DIR=plugin/commands
BINARY_NAME=ticket

# Check OS to set exe suffix
ifeq ($(OS),Windows_NT)
    EXE_SUFFIX=.exe
else
    EXE_SUFFIX=
endif

build:
	@echo "🚀 Building ticket-cli-plugin for host OS..."
	@mkdir -p $(BINARY_DIR)
	go build -ldflags "-s -w" -o $(BINARY_DIR)/$(BINARY_NAME)$(EXE_SUFFIX) main.go
	@echo "✨ Build complete! Binary written to $(BINARY_DIR)/$(BINARY_NAME)$(EXE_SUFFIX)"

build-linux:
	@echo "🚀 Building ticket-cli-plugin for Linux amd64..."
	@mkdir -p $(BINARY_DIR)
	GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o $(BINARY_DIR)/$(BINARY_NAME) main.go
	@echo "✨ Build complete! Binary written to $(BINARY_DIR)/$(BINARY_NAME)"

build-windows:
	@echo "🚀 Building ticket-cli-plugin for Windows amd64..."
	@mkdir -p $(BINARY_DIR)
	GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o $(BINARY_DIR)/$(BINARY_NAME).exe main.go
	@echo "✨ Build complete! Binary written to $(BINARY_DIR)/$(BINARY_NAME).exe"

build-darwin:
	@echo "🚀 Building ticket-cli-plugin for macOS arm64 (M1/M2/M3)..."
	@mkdir -p $(BINARY_DIR)
	GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w" -o $(BINARY_DIR)/$(BINARY_NAME) main.go
	@echo "✨ Build complete! Binary written to $(BINARY_DIR)/$(BINARY_NAME)"

build-linux-amd64:
	@echo "🚀 Building ticket-cli-plugin for Linux amd64 (with suffix)..."
	@mkdir -p $(BINARY_DIR)
	GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o $(BINARY_DIR)/$(BINARY_NAME)-linux-amd64 main.go
	@echo "✨ Build complete! Binary written to $(BINARY_DIR)/$(BINARY_NAME)-linux-amd64"

build-linux-arm64:
	@echo "🚀 Building ticket-cli-plugin for Linux arm64 (with suffix)..."
	@mkdir -p $(BINARY_DIR)
	GOOS=linux GOARCH=arm64 go build -ldflags "-s -w" -o $(BINARY_DIR)/$(BINARY_NAME)-linux-arm64 main.go
	@echo "✨ Build complete! Binary written to $(BINARY_DIR)/$(BINARY_NAME)-linux-arm64"

build-windows-amd64:
	@echo "🚀 Building ticket-cli-plugin for Windows amd64 (with suffix)..."
	@mkdir -p $(BINARY_DIR)
	GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o $(BINARY_DIR)/$(BINARY_NAME)-windows-amd64.exe main.go
	@echo "✨ Build complete! Binary written to $(BINARY_DIR)/$(BINARY_NAME)-windows-amd64.exe"

build-darwin-amd64:
	@echo "🚀 Building ticket-cli-plugin for macOS amd64 (with suffix)..."
	@mkdir -p $(BINARY_DIR)
	GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -o $(BINARY_DIR)/$(BINARY_NAME)-darwin-amd64 main.go
	@echo "✨ Build complete! Binary written to $(BINARY_DIR)/$(BINARY_NAME)-darwin-amd64"

build-darwin-arm64:
	@echo "🚀 Building ticket-cli-plugin for macOS arm64 (with suffix)..."
	@mkdir -p $(BINARY_DIR)
	GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w" -o $(BINARY_DIR)/$(BINARY_NAME)-darwin-arm64 main.go
	@echo "✨ Build complete! Binary written to $(BINARY_DIR)/$(BINARY_NAME)-darwin-arm64"

build-all: build-linux-amd64 build-linux-arm64 build-windows-amd64 build-darwin-amd64 build-darwin-arm64

package-windows-amd64:
	@echo "📦 Packaging plugin for Windows amd64..."
	@rm -rf dist/temp
	@mkdir -p dist/temp/plugin/commands
	@cp -r plugin/skills dist/temp/plugin/
	@cp plugin/plugin.json dist/temp/plugin/
	@cp -r plugin/.claude-plugin dist/temp/plugin/ 2>/dev/null || true
	GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o dist/temp/plugin/commands/ticket.exe main.go
	@cd dist/temp && zip -r ../ticket-management-plugin-windows-amd64.zip plugin
	@rm -rf dist/temp
	@echo "✨ Package created at dist/ticket-management-plugin-windows-amd64.zip"

package-darwin-arm64:
	@echo "📦 Packaging plugin for macOS arm64..."
	@rm -rf dist/temp
	@mkdir -p dist/temp/plugin/commands
	@cp -r plugin/skills dist/temp/plugin/
	@cp plugin/plugin.json dist/temp/plugin/
	@cp -r plugin/.claude-plugin dist/temp/plugin/ 2>/dev/null || true
	GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w" -o dist/temp/plugin/commands/ticket main.go
	@cd dist/temp && zip -r ../ticket-management-plugin-darwin-arm64.zip plugin
	@rm -rf dist/temp
	@echo "✨ Package created at dist/ticket-management-plugin-darwin-arm64.zip"

package-darwin-amd64:
	@echo "📦 Packaging plugin for macOS amd64..."
	@rm -rf dist/temp
	@mkdir -p dist/temp/plugin/commands
	@cp -r plugin/skills dist/temp/plugin/
	@cp plugin/plugin.json dist/temp/plugin/
	@cp -r plugin/.claude-plugin dist/temp/plugin/ 2>/dev/null || true
	GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -o dist/temp/plugin/commands/ticket main.go
	@cd dist/temp && zip -r ../ticket-management-plugin-darwin-amd64.zip plugin
	@rm -rf dist/temp
	@echo "✨ Package created at dist/ticket-management-plugin-darwin-amd64.zip"

package-linux-amd64:
	@echo "📦 Packaging plugin for Linux amd64..."
	@rm -rf dist/temp
	@mkdir -p dist/temp/plugin/commands
	@cp -r plugin/skills dist/temp/plugin/
	@cp plugin/plugin.json dist/temp/plugin/
	@cp -r plugin/.claude-plugin dist/temp/plugin/ 2>/dev/null || true
	GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o dist/temp/plugin/commands/ticket main.go
	@cd dist/temp && tar -czf ../ticket-management-plugin-linux-amd64.tar.gz plugin
	@rm -rf dist/temp
	@echo "✨ Package created at dist/ticket-management-plugin-linux-amd64.tar.gz"

package-all: package-windows-amd64 package-darwin-arm64 package-darwin-amd64 package-linux-amd64

clean:
	@echo "🧹 Cleaning built binaries and packages..."
	rm -rf $(BINARY_DIR) dist
	@echo "✨ Clean complete!"

help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build                 Build the binary for the host OS"
	@echo "  build-linux           Build the Linux amd64 binary as 'ticket'"
	@echo "  build-windows         Build the Windows amd64 binary as 'ticket.exe'"
	@echo "  build-darwin          Build the macOS arm64 (M1/M2/M3) binary as 'ticket'"
	@echo "  build-linux-amd64     Build Linux amd64 binary with arch suffix"
	@echo "  build-linux-arm64     Build Linux arm64 binary with arch suffix"
	@echo "  build-windows-amd64   Build Windows amd64 binary with arch suffix"
	@echo "  build-darwin-amd64    Build macOS amd64 binary with arch suffix"
	@echo "  build-darwin-arm64    Build macOS arm64 binary with arch suffix"
	@echo "  build-all             Build all platforms with arch suffixes"
	@echo "  package-windows-amd64 Build and package plugin as ZIP for Windows amd64"
	@echo "  package-darwin-arm64  Build and package plugin as ZIP for macOS arm64 (M1/M2/M3)"
	@echo "  package-darwin-amd64  Build and package plugin as ZIP for macOS amd64 (Intel)"
	@echo "  package-linux-amd64   Build and package plugin as tar.gz for Linux amd64"
	@echo "  package-all           Build and package for all platforms"
	@echo "  clean                 Remove built binaries and packages"
	@echo "  help                  Show this help message"
