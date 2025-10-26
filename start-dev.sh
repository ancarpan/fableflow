#!/bin/bash

# FableFlow Development Startup Script
# Handles Ctrl+C properly to stop all processes

echo "🚀 Starting FableFlow in development mode..."
echo "Backend: http://localhost:8080"
echo "Frontend: http://localhost:3000"
echo ""
echo "Starting backend and frontend servers..."
echo "Press Ctrl+C to stop all services"
echo ""

# Function to cleanup processes
cleanup() {
    echo ""
    echo "🛑 Stopping development servers..."
    
    # Try graceful shutdown first
    kill $BACKEND_PID 2>/dev/null || true
    kill $FRONTEND_PID 2>/dev/null || true
    
    # Wait a moment for graceful shutdown
    sleep 2
    
    # Force kill if still running
    pkill -f "./fableflow-api" 2>/dev/null || true
    pkill -f "python3 dev-server.py" 2>/dev/null || true
    pkill -f "go run" 2>/dev/null || true
    
    # Kill any processes on the ports
    lsof -ti:8080 | xargs kill -9 2>/dev/null || true
    lsof -ti:3000 | xargs kill -9 2>/dev/null || true
    
    echo "✅ Development servers stopped"
    exit 0
}

# Set up signal handlers
trap cleanup INT TERM

# Start backend
cd backend && ./fableflow-api -c config.dev.yaml &
BACKEND_PID=$!

# Start frontend
python3 dev-server.py &
FRONTEND_PID=$!

# Wait for both processes
wait $BACKEND_PID $FRONTEND_PID
