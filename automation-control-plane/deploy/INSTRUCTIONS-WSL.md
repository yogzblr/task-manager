# Instructions: Running Docker Compose in WSL2 Ubuntu

## Step 1: Open Ubuntu WSL2 Terminal

Open your Ubuntu WSL2 distribution terminal (not the docker-desktop one).

## Step 2: Navigate to the Project Directory

```bash
cd /mnt/c/Users/yoges/OneDrive/Documents/My\ Code/demo/automation-control-plane/deploy
```

## Step 3: Make Scripts Executable

```bash
chmod +x run-docker-compose.sh
chmod +x run-in-ubuntu-wsl.sh
```

## Step 4: Run Docker Compose

### Option A: Using the script
```bash
./run-in-ubuntu-wsl.sh
```

### Option B: Manual commands
```bash
# Build all images
docker compose build

# Start services (Linux services only - Windows agent won't work in WSL)
docker compose up -d control-plane mysql valkey centrifugo quickwit agent-linux

# View logs
docker compose logs -f

# Check status
docker compose ps
```

## Important Notes

1. **Windows Agent**: The `agent-windows` service requires Windows containers and won't work in WSL2 Ubuntu. You can either:
   - Comment it out in docker-compose.yml
   - Or ignore the error when starting services

2. **Docker Desktop Integration**: Make sure Docker Desktop WSL integration is enabled:
   - Docker Desktop → Settings → Resources → WSL Integration
   - Enable for your Ubuntu distribution

3. **Systemd in Docker**: The Linux agent uses systemd which requires `privileged: true`. If you encounter issues, you can modify the agent-linux service to run directly without systemd.

## Troubleshooting

If you get "permission denied" errors:
```bash
sudo usermod -aG docker $USER
newgrp docker
```

If Docker daemon is not accessible:
- Make sure Docker Desktop is running on Windows
- Or install Docker directly in Ubuntu WSL
