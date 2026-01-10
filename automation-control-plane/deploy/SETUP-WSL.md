# Setup and Run Docker Compose in WSL2 Ubuntu

## Prerequisites Setup

### Step 1: Enable Docker Desktop WSL Integration

1. **Start Docker Desktop** on Windows
2. Go to **Settings** (gear icon)
3. Navigate to **Resources** → **WSL Integration**
4. Enable the toggle for **Ubuntu-22.04**
5. Click **Apply & Restart**

### Step 2: Verify Docker Access

Open Ubuntu-22.04 WSL terminal and verify:

```bash
docker --version
docker compose version
```

If these commands work, proceed to Step 3.

## Running Docker Compose

### Quick Start

```bash
# Navigate to project directory
cd /mnt/c/Users/yoges/OneDrive/Documents/My\ Code/demo/automation-control-plane/deploy

# Make script executable
chmod +x run-in-ubuntu-wsl.sh

# Run the script
./run-in-ubuntu-wsl.sh
```

### Manual Commands

```bash
# Navigate to deploy directory
cd /mnt/c/Users/yoges/OneDrive/Documents/My\ Code/demo/automation-control-plane/deploy

# Build all images (excluding Windows agent for now)
docker compose build control-plane agent-linux

# Start services
docker compose up -d control-plane mysql valkey centrifugo quickwit agent-linux

# View logs
docker compose logs -f

# Check status
docker compose ps

# Stop services
docker compose down
```

## Note About Windows Agent

The `agent-windows` service requires Windows containers and cannot run in WSL2 Ubuntu. You have two options:

1. **Comment out the Windows agent** in docker-compose.yml (recommended for WSL2)
2. **Run it separately** on a Windows machine or Windows container host

To comment it out, edit `docker-compose.yml` and add `#` before the `agent-windows:` service section.

## Troubleshooting

### Docker command not found
- Make sure Docker Desktop is running
- Verify WSL integration is enabled for Ubuntu-22.04
- Restart Docker Desktop after enabling integration

### Permission denied
```bash
sudo usermod -aG docker $USER
newgrp docker
```

### Build errors
- Make sure you're in the correct directory
- Check that all source files are present
- Try building individual services: `docker compose build <service-name>`
