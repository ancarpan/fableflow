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
	@echo "🧹 Cleaning Docker containers and volumes..."
	docker-compose down -v
	rm -rf data/database/* data/ebooks/* data/logs/* data/quarantine/*
	@echo "✅ Clean complete"

# Show logs
.PHONY: logs
logs:
	docker-compose logs -f

# Help
.PHONY: help
help:
	@echo "FableFlow - Ebook Manager (Docker-only)"
	@echo ""
	@echo "Available commands:"
	@echo "  make run     - Start production mode"
	@echo "  make stop    - Stop services"
	@echo "  make clean   - Clean everything"
	@echo "  make logs    - Show logs"
	@echo "  make help    - Show this help"
	@echo ""
	@echo "Quick start: make run"

# Default target
.DEFAULT_GOAL := help