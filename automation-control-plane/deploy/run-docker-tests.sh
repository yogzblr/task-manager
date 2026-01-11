#!/bin/bash
# Bash script to run Docker-based workflow tests from WSL
# This keeps your Windows environment clean while testing the automation platform

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${CYAN}"
cat << "EOF"
╔══════════════════════════════════════════════════════════════╗
║     Docker-Based Workflow Testing - Run from WSL            ║
╚══════════════════════════════════════════════════════════════╝
EOF
echo -e "${NC}"

# Function to run docker compose commands
run_command() {
    local description="$1"
    shift
    echo -e "\n${YELLOW}▶ ${description}${NC}"
    echo -e "${CYAN}  Running: $@${NC}\n"
    "$@"
}

# Main menu
echo -e "${GREEN}Select an option:${NC}"
echo "  1. Check Docker services status"
echo "  2. Build test-runner image"
echo "  3. Run Linux workflow test"
echo "  4. Run Windows workflow test (requires Windows agent)"
echo "  5. Interactive test-runner shell"
echo "  6. View control plane logs"
echo "  7. View Linux agent logs"
echo "  8. Query Quickwit for recent logs"
echo "  9. Start all services"
echo "  0. Exit"
echo ""

read -p "Enter choice (0-9): " choice

case $choice in
    1)
        run_command "Checking Docker service status" docker compose ps
        ;;
    
    2)
        run_command "Building test-runner image" docker compose build test-runner
        ;;
    
    3)
        echo ""
        run_command "Running Linux workflow test" \
            docker compose run --rm test-runner python test-linux-workflow.py
        ;;
    
    4)
        echo -e "\n${YELLOW}⚠ WARNING: This requires a Windows agent to be running!${NC}\n"
        read -p "Is the Windows agent running? (y/n): " confirm
        if [ "$confirm" = "y" ]; then
            run_command "Running Windows workflow test" \
                docker compose run --rm test-runner python test-windows-workflow.py
        else
            echo -e "${RED}Cancelled. Start the Windows agent first.${NC}"
        fi
        ;;
    
    5)
        echo -e "\n${GREEN}Starting interactive shell in test-runner container...${NC}"
        echo -e "${CYAN}Commands available inside:${NC}"
        echo "  - python test-linux-workflow.py"
        echo "  - python test-windows-workflow.py"
        echo "  - ls -la"
        echo -e "  - exit (to leave container)\n"
        
        docker compose run --rm test-runner /bin/bash
        ;;
    
    6)
        run_command "Viewing control plane logs (last 50 lines)" \
            docker compose logs --tail=50 control-plane
        ;;
    
    7)
        run_command "Viewing Linux agent logs (last 50 lines)" \
            docker compose logs --tail=50 agent-linux
        ;;
    
    8)
        echo -e "\n${GREEN}Querying Quickwit for recent automation logs...${NC}\n"
        docker compose run --rm test-runner bash -c \
            'curl -X POST http://quickwit:7280/api/v1/automation-logs/search \
            -H "Content-Type: application/json" \
            -d "{\"query\": \"*\", \"max_hits\": 20, \"sort_by\": \"-timestamp\"}" | python -m json.tool'
        ;;
    
    9)
        run_command "Starting all Docker services" docker compose up -d
        echo -e "\n${YELLOW}Waiting for services to be ready...${NC}"
        sleep 10
        run_command "Checking service status" docker compose ps
        ;;
    
    0)
        echo -e "\n${GREEN}Exiting...${NC}"
        exit 0
        ;;
    
    *)
        echo -e "\n${RED}✗ Invalid choice. Please run the script again.${NC}"
        exit 1
        ;;
esac

echo -e "\n${CYAN}═══════════════════════════════════════════════════════════════${NC}"
echo -e "${GREEN}Test operation complete!${NC}"
echo -e "${CYAN}═══════════════════════════════════════════════════════════════${NC}\n"
