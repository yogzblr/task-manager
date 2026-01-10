# Running Docker Compose in WSL2 Ubuntu

## Prerequisites

1. **Enable Docker Desktop WSL Integration**:
   - Open Docker Desktop
   - Go to Settings → Resources → WSL Integration
   - Enable integration for your Ubuntu distribution
   - Click "Apply & Restart"

2. **Or Install Docker directly in Ubuntu WSL**:
   ```bash
   # Update package index
   sudo apt-get update
   
   # Install Docker
   curl -fsSL https://get.docker.com -o get-docker.sh
   sudo sh get-docker.sh
   
   # Add your user to docker group
   sudo usermod -aG docker $USER
   
   # Install docker-compose
   sudo apt-get install docker-compose-plugin
   ```

## Running Docker Compose

### Option 1: Using the provided script

```bash
# In Ubuntu WSL2 terminal
cd /mnt/c/Users/yoges/OneDrive/Documents/My\ Code/demo/automation-control-plane/deploy
chmod +x run-docker-compose.sh
./run-docker-compose.sh
```

### Option 2: Manual commands

```bash
# Navigate to deploy directory
cd /mnt/c/Users/yoges/OneDrive/Documents/My\ Code/demo/automation-control-plane/deploy

# Build all images
docker compose build

# Start all services
docker compose up -d

# View logs
docker compose logs -f

# View status
docker compose ps

# Stop services
docker compose down

# Stop and remove volumes
docker compose down -v
```

## Notes

- **Windows Agent**: The Windows agent container requires Windows containers, which won't work in WSL2 Ubuntu. You can comment out the `agent-windows` service in docker-compose.yml if needed.
- **Linux Agent**: The Linux agent with systemd requires `privileged: true`. If you encounter issues, you can modify the docker-compose.yml to run the agent directly without systemd.

## Troubleshooting

If you get permission errors:
```bash
sudo usermod -aG docker $USER
newgrp docker
```

If Docker daemon is not running:
```bash
# Check Docker Desktop is running on Windows
# Or start Docker service in Ubuntu:
sudo service docker start
```
