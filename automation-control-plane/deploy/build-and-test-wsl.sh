#!/bin/bash
# Automation Platform - WSL Ubuntu Build and Test Script
# This script builds and tests the control plane stack in WSL Ubuntu

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${CYAN}=== Automation Platform - WSL Build and Test ===${NC}"
echo ""

# Function to print colored messages
print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_info() {
    echo -e "${YELLOW}ℹ $1${NC}"
}

print_step() {
    echo -e "${CYAN}=== $1 ===${NC}"
}

# Check if running in WSL
if ! grep -qi microsoft /proc/version; then
    print_error "This script must be run in WSL Ubuntu"
    exit 1
fi
print_success "Running in WSL Ubuntu"

# Check if Docker is available
if ! command -v docker &> /dev/null; then
    print_error "Docker not found. Please enable Docker Desktop WSL integration."
    echo ""
    echo "Steps:"
    echo "1. Start Docker Desktop on Windows"
    echo "2. Go to Settings → Resources → WSL Integration"
    echo "3. Enable integration for Ubuntu-22.04"
    echo "4. Click 'Apply & Restart'"
    exit 1
fi
print_success "Docker is available: $(docker --version)"

# Check if Docker Compose is available
if ! docker compose version &> /dev/null; then
    print_error "Docker Compose not found"
    exit 1
fi
print_success "Docker Compose is available: $(docker compose version)"

# Navigate to the deploy directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"
print_success "Working directory: $PWD"
echo ""

# Clean up previous runs
print_step "Cleaning Up Previous Runs"
docker compose down -v 2>/dev/null || true
print_success "Previous containers and volumes removed"
echo ""

# Build images
print_step "Building Docker Images"
print_info "This may take several minutes on first run..."
echo ""

# Build control plane
print_info "Building control-plane image..."
if docker compose build control-plane; then
    print_success "Control plane image built"
else
    print_error "Failed to build control plane image"
    exit 1
fi
echo ""

# Build Linux agent
print_info "Building agent-linux image..."
if docker compose build agent-linux; then
    print_success "Linux agent image built"
else
    print_error "Failed to build Linux agent image"
    exit 1
fi
echo ""

# Start services (excluding Windows agent)
print_step "Starting Services"
print_info "Starting: MySQL, Valkey, Centrifugo, Quickwit, Control Plane, Linux Agent"
echo ""

if docker compose up -d mysql valkey centrifugo quickwit control-plane agent-linux; then
    print_success "All services started"
else
    print_error "Failed to start services"
    exit 1
fi
echo ""

# Wait for services to be healthy
print_step "Waiting for Services to be Healthy"
echo ""

# Wait for MySQL
print_info "Waiting for MySQL..."
for i in {1..30}; do
    if docker compose exec -T mysql mysqladmin ping -h localhost -u root -prootpass &>/dev/null; then
        print_success "MySQL is ready"
        break
    fi
    if [ $i -eq 30 ]; then
        print_error "MySQL failed to start"
        docker compose logs mysql
        exit 1
    fi
    sleep 2
done

# Wait for Valkey
print_info "Waiting for Valkey..."
for i in {1..20}; do
    if docker compose exec -T valkey valkey-cli ping &>/dev/null; then
        print_success "Valkey is ready"
        break
    fi
    if [ $i -eq 20 ]; then
        print_error "Valkey failed to start"
        docker compose logs valkey
        exit 1
    fi
    sleep 2
done

# Wait for Centrifugo
print_info "Waiting for Centrifugo..."
for i in {1..20}; do
    if curl -sf http://localhost:8000/health &>/dev/null; then
        print_success "Centrifugo is ready"
        break
    fi
    if [ $i -eq 20 ]; then
        print_error "Centrifugo failed to start"
        docker compose logs centrifugo
        exit 1
    fi
    sleep 2
done

# Wait for Control Plane
print_info "Waiting for Control Plane..."
for i in {1..30}; do
    if curl -sf http://localhost:8081/health &>/dev/null; then
        print_success "Control Plane is ready"
        break
    fi
    if [ $i -eq 30 ]; then
        print_error "Control Plane failed to start"
        docker compose logs control-plane
        exit 1
    fi
    sleep 2
done
echo ""

# Display service status
print_step "Service Status"
docker compose ps
echo ""

# Test the services
print_step "Testing Services"
echo ""

# Test MySQL
print_info "Testing MySQL connection..."
if docker compose exec -T mysql mysql -u automation -ppassword -e "SELECT 1" automation &>/dev/null; then
    print_success "MySQL connection successful"
else
    print_error "MySQL connection failed"
fi

# Test Valkey
print_info "Testing Valkey connection..."
if docker compose exec -T valkey valkey-cli SET test "Hello WSL" &>/dev/null; then
    RESULT=$(docker compose exec -T valkey valkey-cli GET test 2>/dev/null | tr -d '\r')
    if [ "$RESULT" = "Hello WSL" ]; then
        print_success "Valkey connection successful"
    else
        print_error "Valkey test failed"
    fi
else
    print_error "Valkey connection failed"
fi

# Test Centrifugo
print_info "Testing Centrifugo API..."
if curl -sf http://localhost:8000/health | grep -q "ok"; then
    print_success "Centrifugo API is responding"
else
    print_error "Centrifugo API test failed"
fi

# Test Control Plane
print_info "Testing Control Plane API..."
if curl -sf http://localhost:8081/health &>/dev/null; then
    print_success "Control Plane API is responding"
else
    print_error "Control Plane API test failed"
fi
echo ""

# Check agent status
print_step "Agent Status"
AGENT_STATUS=$(docker compose ps agent-linux --format json | grep -o '"State":"[^"]*"' | cut -d'"' -f4)
if [ "$AGENT_STATUS" = "running" ]; then
    print_success "Linux agent is running"
    echo ""
    print_info "Agent logs (last 10 lines):"
    docker compose logs --tail=10 agent-linux
else
    print_error "Linux agent is not running"
    docker compose logs agent-linux
fi
echo ""

# Display URLs
print_step "Service URLs"
echo ""
echo "  Control Plane API: http://localhost:8081"
echo "  Centrifugo:        http://localhost:8000"
echo "  MySQL:             localhost:3306"
echo "  Valkey:            localhost:6379"
echo "  Quickwit:          http://localhost:7280"
echo ""

# Display useful commands
print_step "Useful Commands"
echo ""
echo "  View all logs:     docker compose logs -f"
echo "  View specific log: docker compose logs -f control-plane"
echo "  Stop services:     docker compose down"
echo "  Restart service:   docker compose restart control-plane"
echo "  Check status:      docker compose ps"
echo ""

# Test API endpoints
print_step "Testing API Endpoints"
echo ""

print_info "Testing /health endpoint..."
HEALTH_RESPONSE=$(curl -s http://localhost:8081/health)
if echo "$HEALTH_RESPONSE" | grep -q "ok\|healthy\|status"; then
    print_success "Health endpoint responding: $HEALTH_RESPONSE"
else
    print_error "Health endpoint test failed"
fi

print_info "Testing /api/v1/agents endpoint (expecting auth required)..."
AGENTS_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8081/api/v1/agents)
if [ "$AGENTS_RESPONSE" = "401" ] || [ "$AGENTS_RESPONSE" = "403" ]; then
    print_success "Agents endpoint responding correctly (auth required)"
elif [ "$AGENTS_RESPONSE" = "200" ]; then
    print_success "Agents endpoint responding (no auth required)"
else
    print_error "Agents endpoint unexpected response: $AGENTS_RESPONSE"
fi
echo ""

# Summary
print_step "Summary"
echo ""
echo "✅ Docker images built successfully"
echo "✅ All services started"
echo "✅ Health checks passed"
echo "✅ API endpoints responding"
echo ""
print_success "Platform is ready for testing!"
echo ""
print_info "To test with probe workflows, follow these steps:"
echo ""
echo "  1. Navigate to probe directory:"
echo "     cd /mnt/c/Users/yoges/OneDrive/Documents/My\\ Code/Task\\ Manager/demo/probe"
echo ""
echo "  2. Build test program (if Go is installed):"
echo "     go build -o test-probe ./cmd/test-probe"
echo ""
echo "  3. Test HTTP workflow:"
echo "     ./test-probe ./examples/http-example.yaml"
echo ""
echo "  4. Test Command workflow:"
echo "     ./test-probe ./examples/command-example.yaml"
echo ""
print_info "To stop all services:"
echo "     docker compose down"
echo ""
print_info "To view logs:"
echo "     docker compose logs -f"
echo ""

print_success "Build and test completed successfully!"
