# FableFlow Backend Makefile

# Variables
BINARY_NAME=fableflow-api
MAIN_FILE=main.go
CONFIG_FILE=config.yaml

# Default target
.PHONY: all
all: build

# Build the application
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	CGO_ENABLED=1 go build -o $(BINARY_NAME) $(MAIN_FILE)
	@echo "Build complete: $(BINARY_NAME)"

# Run the application
.PHONY: run
run: build
	@echo "Starting $(BINARY_NAME)..."
	./$(BINARY_NAME)

# Install dependencies
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	go mod tidy
	go mod download

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -f $(BINARY_NAME)
	@echo "Clean complete"

# Run tests (if any)
.PHONY: test
test:
	@echo "Running tests..."
	go test ./...

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Lint code
.PHONY: lint
lint:
	@echo "Linting code..."
	golangci-lint run

# Cross-platform builds
.PHONY: build-linux
build-linux:
	@echo "Building for Linux..."
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME)-linux $(MAIN_FILE)

.PHONY: build-darwin
build-darwin:
	@echo "Building for macOS..."
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -o $(BINARY_NAME)-darwin $(MAIN_FILE)

# Build all platforms
.PHONY: build-all
build-all: build-linux build-darwin
	@echo "Cross-platform builds complete"

# Development mode (build and run with auto-restart)
.PHONY: dev
dev:
	@echo "Starting development mode..."
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "Air not found. Install with: go install github.com/cosmtrek/air@latest"; \
		echo "Falling back to regular run..."; \
		$(MAKE) run; \
	fi

# Help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build        - Build the application"
	@echo "  run          - Build and run the application"
	@echo "  deps         - Install dependencies"
	@echo "  clean        - Clean build artifacts"
	@echo "  test         - Run tests"
	@echo "  fmt          - Format code"
	@echo "  lint         - Lint code"
	@echo "  build-linux  - Build for Linux"
	@echo "  build-darwin - Build for macOS"
	@echo "  build-all    - Build for all platforms"
	@echo "  dev          - Development mode with auto-restart"
	@echo "  help         - Show this help"

# Default target
.DEFAULT_GOAL := help
