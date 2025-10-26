# FableFlow Development Setup

This document explains how to set up and run FableFlow in development mode for faster iteration.

## Quick Start

```bash
# Start development mode
make dev
```

## Development Mode

### Local Development Mode (`make dev`)
- Runs Go backend locally on port 8080
- Serves frontend with Python proxy server on port 3000
- Requires manual restart for backend changes

## Setup Instructions

### Prerequisites
- Go 1.21+ installed locally
- Python 3.x installed locally
- Make installed

### First Time Setup
```bash
# Create necessary directories
mkdir -p data/ebooks data/import data/quarantine data/logs

# Start development mode
make dev
```

## Development Workflow

1. **Start Development Server**
   ```bash
   make dev
   ```

2. **Access the Application**
   - Frontend: http://localhost:3000
   - Backend API: http://localhost:8080
   - Backend Health: http://localhost:8080/api/health

3. **Make Changes**
   - Frontend changes: Edit files in `frontend/` - changes are served immediately
   - Backend changes: Press Ctrl+C to stop, then `make dev` to restart

4. **Stop Development Server**
   - Press **Ctrl+C** to stop all services gracefully
   - Or use `make dev-stop` if needed

## File Structure

```
fableflow/
├── backend/
│   ├── config.dev.yaml    # Development configuration
│   └── ...                # Go backend code
├── frontend/
│   ├── static/            # Static assets (CSS, JS, images)
│   └── templates/         # HTML templates
├── dev/
│   ├── dev-server.py      # Development proxy server
│   ├── start-dev.sh       # Development startup script
│   └── DEVELOPMENT.md      # This documentation
└── Makefile              # Development commands
```

## Configuration

### Backend Configuration (`backend/config.dev.yaml`)
- Uses local paths for development
- Disables static file serving (handled by proxy)
- Points to local data directories


## Troubleshooting

### Port Already in Use
```bash
# Kill processes on ports 3000 and 8080
lsof -ti:3000 | xargs kill -9
lsof -ti:8080 | xargs kill -9
```

### Backend Not Starting
```bash
# Check if Go is installed
go version

# Check backend configuration
cd backend && go run . -c config.dev.yaml
```

### Frontend Not Loading
```bash
# Check if Python is installed
python3 --version

# Test the development server directly
cd dev && python3 dev-server.py
```

## Production vs Development

| Feature | Development | Production |
|---------|-------------|------------|
| Backend | Local Go binary | Docker container |
| Frontend | Python proxy server | Caddy reverse proxy |
| Build Time | ~2-3 seconds | ~30-60 seconds |
| Static Files | Served by Python | Served by Caddy |

## Advanced Options

### Custom Backend Port
```bash
# Edit backend/config.dev.yaml
server:
  port: "8081"  # Change port
```

### Custom Frontend Port
```bash
# Edit dev/dev-server.py
PORT = 3001  # Change port
```

### Database Location
The development setup uses `../data/ebooks.db` by default. You can change this in `backend/config.dev.yaml`.

## Performance Tips

1. **Use Local Database**: Keep database file in `data/` directory for persistence
2. **Monitor Logs**: Check terminal output for errors and debugging info
3. **Quick Restart**: Press Ctrl+C then `make dev` for backend changes
