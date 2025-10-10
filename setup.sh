#!/bin/bash

# FableFlow Setup Script

echo "🚀 Setting up FableFlow..."

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go first."
    exit 1
fi

echo "✅ Go is installed"

# Install dependencies
echo "📦 Installing dependencies..."
make deps

# Build both services
echo "🔨 Building services..."
make build

echo ""
echo "✅ Setup complete!"
echo ""
echo "To start the application:"
echo "  make dev     # Development mode (both services)"
echo "  make run     # Production mode (both services)"
echo ""
echo "Individual services:"
echo "  make backend  # Backend API only (port 8080)"
echo "  make frontend # Frontend only (port 3000)"
echo ""
echo "Access the application:"
echo "  Frontend: http://localhost:3000"
echo "  Backend API: http://localhost:8080"
