#!/bin/bash
set -e

# Script to build and run docker-compose in WSL2
# Usage: ./run-docker-compose.sh

echo "Building and starting Docker Compose services..."

# Navigate to the deploy directory
cd "$(dirname "$0")"

# Check if docker-compose is available
if ! command -v docker-compose &> /dev/null; then
    echo "docker-compose not found. Trying 'docker compose'..."
    if ! command -v docker &> /dev/null; then
        echo "Error: Docker is not installed or not in PATH"
        exit 1
    fi
    DOCKER_COMPOSE_CMD="docker compose"
else
    DOCKER_COMPOSE_CMD="docker-compose"
fi

# Build images
echo "Building Docker images..."
$DOCKER_COMPOSE_CMD build

# Start services
echo "Starting services..."
$DOCKER_COMPOSE_CMD up -d

# Show status
echo ""
echo "Service status:"
$DOCKER_COMPOSE_CMD ps

echo ""
echo "To view logs, run: $DOCKER_COMPOSE_CMD logs -f"
echo "To stop services, run: $DOCKER_COMPOSE_CMD down"
echo "To stop and remove volumes, run: $DOCKER_COMPOSE_CMD down -v"
