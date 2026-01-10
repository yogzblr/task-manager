#!/bin/bash
# Script to run docker-compose in Ubuntu WSL2
# Make sure Docker Desktop WSL integration is enabled for Ubuntu

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "=========================================="
echo "Automation Platform Docker Compose"
echo "=========================================="
echo ""

# Check if docker is available
if ! command -v docker &> /dev/null; then
    echo "Error: Docker is not installed or not in PATH"
    echo ""
    echo "Please enable Docker Desktop WSL integration:"
    echo "1. Open Docker Desktop"
    echo "2. Go to Settings → Resources → WSL Integration"
    echo "3. Enable integration for Ubuntu-22.04"
    echo "4. Click Apply & Restart"
    exit 1
fi

# Check if docker daemon is running
if ! docker info &> /dev/null; then
    echo "Error: Docker daemon is not running"
    echo "Please start Docker Desktop on Windows"
    exit 1
fi

# Determine compose command
if docker compose version &> /dev/null; then
    COMPOSE_CMD="docker compose"
    echo "Using: docker compose (v2)"
elif command -v docker-compose &> /dev/null; then
    COMPOSE_CMD="docker-compose"
    echo "Using: docker-compose (v1)"
else
    echo "Error: docker compose or docker-compose not found"
    exit 1
fi

echo ""
echo "Current directory: $(pwd)"
echo ""

# Use WSL-optimized compose file (excludes Windows agent)
COMPOSE_FILE="docker-compose-wsl.yml"
if [ ! -f "$COMPOSE_FILE" ]; then
    echo "Warning: $COMPOSE_FILE not found, using docker-compose.yml"
    COMPOSE_FILE="docker-compose.yml"
    echo "Note: Windows agent will fail to build in WSL2 (this is expected)"
fi

# Build images
echo "=========================================="
echo "Building Docker images..."
echo "=========================================="
$COMPOSE_CMD -f "$COMPOSE_FILE" build

echo ""
echo "=========================================="
echo "Starting services..."
echo "=========================================="
$COMPOSE_CMD -f "$COMPOSE_FILE" up -d

echo ""
echo "=========================================="
echo "Service Status"
echo "=========================================="
$COMPOSE_CMD -f "$COMPOSE_FILE" ps

echo ""
echo "=========================================="
echo "Useful Commands"
echo "=========================================="
echo "View logs:        $COMPOSE_CMD -f $COMPOSE_FILE logs -f"
echo "View agent logs:  $COMPOSE_CMD -f $COMPOSE_FILE logs -f agent-linux"
echo "Stop services:    $COMPOSE_CMD -f $COMPOSE_FILE down"
echo "Stop + volumes:   $COMPOSE_CMD -f $COMPOSE_FILE down -v"
echo "Restart:          $COMPOSE_CMD -f $COMPOSE_FILE restart"
echo ""
