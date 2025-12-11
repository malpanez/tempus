# Tempus - Multilingual ICS Calendar Generator
# Build configuration

APP_NAME := tempus
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Go build settings
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
GO_LDFLAGS := -ldflags="-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(BUILD_DATE)"

# Directories
BUILD_DIR := build
DIST_DIR := dist
LOCALES_DIR := locales

.PHONY: all build clean test lint install uninstall help deps release

# Default target
all: build

# Build the application
build:
	@echo "Building $(APP_NAME) $(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(GO_LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME) .

# Build for multiple platforms
build-all: clean
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	
	# Linux
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(GO_LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 .
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build $(GO_LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-linux-arm64 .
	
	# macOS
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build $(GO_LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-darwin-amd64 .
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build $(GO_LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-darwin-arm64 .
	
	# Windows
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build $(GO_LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe .
	CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build $(GO_LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-windows-arm64.exe .

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Lint the code
lint:
	@echo "Running linter..."
	golangci-lint run

# Format the code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Install the application to GOPATH/bin
install: build
	@echo "Installing $(APP_NAME)..."
	go install $(GO_LDFLAGS) .

# Uninstall the application
uninstall:
	@echo "Uninstalling $(APP_NAME)..."
	rm -f $(shell go env GOPATH)/bin/$(APP_NAME)

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -rf $(DIST_DIR)
	rm -f coverage.out coverage.html

# Create release packages
release: build-all
	@echo "Creating release packages..."
	@mkdir -p $(DIST_DIR)
	
	# Create archives for each platform
	cd $(BUILD_DIR) && \
	for binary in $(APP_NAME)-*; do \
		if [[ $binary == *".exe" ]]; then \
			platform=${binary%.exe}; \
			zip ../$(DIST_DIR)/$platform.zip $binary ../README.md ../LICENSE; \
		else \
			platform=$binary; \
			tar -czf ../$(DIST_DIR)/$platform.tar.gz $binary ../README.md ../LICENSE; \
		fi \
	done

# Development: run with live reload
dev:
	@echo "Starting development mode..."
	air

# Show version
version:
	@echo "$(APP_NAME) version $(VERSION) ($(COMMIT)) built on $(BUILD_DATE)"

# Initialize translation files
init-translations:
	@echo "Initializing translation files..."
	@mkdir -p $(LOCALES_DIR)
	@if [ ! -f $(LOCALES_DIR)/en.json ]; then \
		echo '{}' > $(LOCALES_DIR)/en.json; \
	fi
	@if [ ! -f $(LOCALES_DIR)/es.json ]; then \
		echo '{}' > $(LOCALES_DIR)/es.json; \
	fi
	@if [ ! -f $(LOCALES_DIR)/ga.json ]; then \
		echo '{}' > $(LOCALES_DIR)/ga.json; \
	fi

# Run example commands
examples:
	@echo "Running examples..."
	@echo "1. Create a simple event:"
	@echo "   ./$(BUILD_DIR)/$(APP_NAME) create 'Meeting with team' -s '2024-03-15 10:00' -e '2024-03-15 11:00' -t 'Europe/Madrid'"
	@echo ""
	@echo "2. Create a flight:"
	@echo "   ./$(BUILD_DIR)/$(APP_NAME) template create flight"
	@echo ""
	@echo "3. Configure timezone:"
	@echo "   ./$(BUILD_DIR)/$(APP_NAME) config set timezone 'Europe/Dublin'"
	@echo ""
	@echo "4. Set language to Spanish:"
	@echo "   ./$(BUILD_DIR)/$(APP_NAME) config set language 'es'"

# Generate documentation
docs:
	@echo "Generating documentation..."
	@mkdir -p docs
	@echo "# Tempus Documentation" > docs/README.md
	@echo "" >> docs/README.md
	@echo "## Commands" >> docs/README.md
	@echo "" >> docs/README.md
	@./$(BUILD_DIR)/$(APP_NAME) --help >> docs/README.md 2>&1 || true

# Benchmark tests
bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

# Security scan
security:
	@echo "Running security scan..."
	gosec ./...

# Check for outdated dependencies
deps-check:
	@echo "Checking for outdated dependencies..."
	go list -u -m all

# Help
help:
	@echo "Available targets:"
	@echo "  build          - Build the application"
	@echo "  build-all      - Build for all platforms"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage"
	@echo "  lint           - Run linter"
	@echo "  fmt            - Format code"
	@echo "  deps           - Install dependencies"
	@echo "  install        - Install application"
	@echo "  uninstall      - Uninstall application"
	@echo "  clean          - Clean build artifacts"
	@echo "  release        - Create release packages"
	@echo "  dev            - Start development mode"
	@echo "  version        - Show version"
	@echo "  examples       - Show usage examples"
	@echo "  docs           - Generate documentation"
	@echo "  bench          - Run benchmarks"
	@echo "  security       - Run security scan"
	@echo "  deps-check     - Check for outdated dependencies"
	@echo "  init-translations - Initialize translation files"
	@echo "  help           - Show this help"