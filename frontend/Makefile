# FableFlow Frontend Makefile

# Variables
BINARY_NAME=fableflow-frontend
MAIN_FILE=main.go

# Default target
.PHONY: all
all: build

# Build the application
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) $(MAIN_FILE)
	@echo "Build complete: $(BINARY_NAME)"

# Run the application
.PHONY: run
run: build
	@echo "Starting $(BINARY_NAME)..."
	./$(BINARY_NAME)

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -f $(BINARY_NAME)
	@echo "Clean complete"

# Help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build        - Build the application"
	@echo "  run          - Build and run the application"
	@echo "  clean        - Clean build artifacts"
	@echo "  help         - Show this help"

# Default target
.DEFAULT_GOAL := help
