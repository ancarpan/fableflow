# FableFlow - Makefile

# Default target
.PHONY: all
all: help

# Build and run with Docker
.PHONY: run
run:
	@echo "🚀 Starting FableFlow with Docker..."
	docker-compose up --build

# Stop services
.PHONY: stop
stop:
	@echo "🛑 Stopping FableFlow..."
	docker-compose down

# Clean everything
.PHONY: clean
clean:
	rm -rf data/database/* data/ebooks/* data/logs/* data/quarantine/* data/ebooks.db

# Show logs
.PHONY: logs
logs:
	docker-compose logs -f

# Development mode - local Go backend
.PHONY: dev
dev: build-backend
	@./start-dev.sh

# Build backend locally
.PHONY: build-backend
build-backend:
	@echo "🔨 Building backend..."
	@cd backend && make build
	@echo "✅ Backend built successfully"


# Help
.PHONY: help
help:
	@echo "FableFlow - Ebook Manager"
	@echo ""
	@echo "Production commands:"
	@echo "  make run     - Start production mode (Docker)"
	@echo "  make stop    - Stop services"
	@echo "  make clean   - Clean everything"
	@echo "  make logs    - Show logs"
	@echo ""
	@echo "Development commands:"
	@echo "  make dev         - Start development mode (local Go + Python server)"
	@echo "  make build-backend - Build backend locally"
	@echo ""
	@echo "Quick start: make dev"

# Default target
.DEFAULT_GOAL := help