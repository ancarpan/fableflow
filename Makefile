# FableFlow Root Makefile

# Variables
BACKEND_DIR=backend
FRONTEND_DIR=frontend

# Default target
.PHONY: all
all: help

# Install dependencies for both services
.PHONY: deps
deps:
	@echo "Installing backend dependencies..."
	cd $(BACKEND_DIR) && go mod tidy
	@echo "Installing frontend dependencies..."
	cd $(FRONTEND_DIR) && go mod tidy

# Build both services
.PHONY: build
build:
	@echo "Building backend..."
	cd $(BACKEND_DIR) && make build
	@echo "Building frontend..."
	cd $(FRONTEND_DIR) && make build

# Run backend only
.PHONY: backend
backend:
	@echo "Starting backend API..."
	cd $(BACKEND_DIR) && make run

# Run frontend only
.PHONY: frontend
frontend:
	@echo "Starting frontend server..."
	cd $(FRONTEND_DIR) && make run

# Run both services (in foreground)
.PHONY: run
run: build
	@./start-services.sh

# Clean all build artifacts
.PHONY: clean
clean:
	@echo "Cleaning backend..."
	cd $(BACKEND_DIR) && make clean
	@echo "Cleaning frontend..."
	cd $(FRONTEND_DIR) && make clean
	@echo "Clean complete"

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting backend code..."
	cd $(BACKEND_DIR) && go fmt ./...
	@echo "Formatting frontend code..."
	cd $(FRONTEND_DIR) && go fmt ./...

# Run tests
.PHONY: test
test:
	@echo "Running backend tests..."
	cd $(BACKEND_DIR) && go test ./...
	@echo "Running frontend tests..."
	cd $(FRONTEND_DIR) && go test ./...

# Development mode (with auto-restart)
.PHONY: dev
dev:
	@./start-services.sh

# Help
.PHONY: help
help:
	@echo "FableFlow - Ebook Manager"
	@echo ""
	@echo "Available targets:"
	@echo "  deps         - Install dependencies for both services"
	@echo "  build        - Build both backend and frontend"
	@echo "  backend      - Run backend API only (port 8080)"
	@echo "  frontend     - Run frontend only (port 3000)"
	@echo "  run          - Run both services"
	@echo "  dev          - Development mode (both services)"
	@echo "  clean        - Clean all build artifacts"
	@echo "  fmt          - Format all code"
	@echo "  test         - Run all tests"
	@echo "  help         - Show this help"
	@echo ""
	@echo "Quick start:"
	@echo "  make deps    # Install dependencies"
	@echo "  make dev     # Start development servers"
	@echo "  make run     # Start production servers"

# Default target
.DEFAULT_GOAL := help