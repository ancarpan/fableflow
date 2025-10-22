# FableFlow

A modern ebook management system with web interface and API.

Fableflow is a small experiment. 95+% vibe coded. Including this README. Except this line.

## Features

- 📚 EPUB ebook library management
- 🔍 Search and browse by author, title, ISBN
- 📖 Built-in EPUB reader
- 🔄 Auto-import and scanning
- 📱 Responsive web interface
- 🐳 Docker containerized

## Quick Start

### Using Pre-built Image (Recommended)
```bash
# Pull and run from GitHub Container Registry
docker pull ghcr.io/ancarpan/fableflow:latest
make run

# Access the application
# Frontend: http://localhost:3000
# Backend API: http://localhost:8080
```

### Building Locally
```bash
# Build and run locally
make run
```

## Commands

- `make run` - Start services
- `make stop` - Stop services  
- `make clean` - Clean data and containers
- `make logs` - View logs

## Architecture

- **Backend**: Go API server (port 8080)
- **Frontend**: Static web interface (port 3000)
- **Database**: SQLite
- **Storage**: Local file system

## Docker Images

Pre-built images are available on GitHub Container Registry:
- `ghcr.io/ancarpan/fableflow:latest` - Latest from main branch
- `ghcr.io/ancarpan/fableflow:v1.0.0` - Tagged releases

## Configuration

Copy `config.yaml.template` to `config.yaml` and adjust settings as needed.
