#!/bin/bash

# VeChain Mesh API Docker Startup Script

set -e

echo "🚀 Starting VeChain Mesh API with Docker Compose..."

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker first."
    exit 1
fi

# Check if docker-compose is available
if ! command -v docker-compose &> /dev/null; then
    echo "❌ docker-compose is not installed. Please install docker-compose first."
    exit 1
fi

# Parse command line arguments
NETWORK="test"
MODE="online"

while [[ $# -gt 0 ]]; do
    case $1 in
        --network)
            NETWORK="$2"
            shift 2
            ;;
        --mode)
            MODE="$2"
            shift 2
            ;;
        --help)
            echo "Usage: $0 [--network main|test] [--mode online|offline]"
            echo ""
            echo "Options:"
            echo "  --network    Set the VeChain network (main or test, default: test)"
            echo "  --mode       Set the server mode (online or offline, default: online)"
            echo "  --help       Show this help message"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

echo "📋 Configuration:"
echo "   Network: $NETWORK"
echo "   Mode: $MODE"
echo "   Thor Image: vechain/thor (official)"
echo ""

# Set environment variables
export NETWORK=$NETWORK
export MODE=$MODE

# Start services
echo "🔨 Building and starting services..."
docker-compose up --build -d

echo ""
echo "✅ Services started successfully!"
echo ""
echo "📊 Service Status:"
docker-compose ps

echo ""
echo "🌐 Access Points:"
echo "   VeChain Thor API:    http://localhost:8669"
echo "   VeChain Mesh API:    http://localhost:8000"
echo "   Health Check:        http://localhost:8000/health"
echo ""

echo "📝 Useful Commands:"
echo "   View logs:           docker-compose logs -f"
echo "   Stop services:       docker-compose down"
echo "   Restart services:    docker-compose restart"
echo "   View service status: docker-compose ps"
echo ""

# Wait a moment for services to be ready
echo "⏳ Waiting for services to be ready..."
sleep 10

# Check if services are healthy
echo "🔍 Checking service health..."
if curl -s http://localhost:8000/health > /dev/null; then
    echo "✅ VeChain Mesh API is healthy"
else
    echo "⚠️  VeChain Mesh API is not responding yet"
fi

if curl -s http://localhost:8669/blocks/best > /dev/null; then
    echo "✅ VeChain Thor node is healthy"
else
    echo "⚠️  VeChain Thor node is not responding yet"
fi

echo ""
echo "🎉 Setup complete! You can now use the VeChain Mesh API."
