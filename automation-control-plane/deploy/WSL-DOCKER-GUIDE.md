# WSL Docker Compose - Complete Setup and Test Guide

## Quick Start (Step by Step)

### Step 1: Start Docker Desktop (Windows)

1. Start **Docker Desktop** on Windows
2. Wait for it to fully start (icon in system tray should be stable)
3. Go to **Settings** (gear icon)
4. Navigate to **Resources** → **WSL Integration**
5. Enable the toggle for **Ubuntu-22.04**
6. Click **Apply & Restart**

### Step 2: Start Ubuntu WSL

Open PowerShell and run:

```powershell
wsl -d Ubuntu-22.04
```

This will start Ubuntu in WSL2.

### Step 3: Run the Build and Test Script

Once in Ubuntu WSL terminal:

```bash
# Navigate to deploy directory
cd /mnt/c/Users/yoges/OneDrive/Documents/My\ Code/Task\ Manager/demo/automation-control-plane/deploy

# Make script executable
chmod +x build-and-test-wsl.sh

# Run the script
./build-and-test-wsl.sh
```

The script will:
- ✅ Verify Docker is available
- ✅ Clean up previous runs
- ✅ Build control plane image
- ✅ Build Linux agent image
- ✅ Start all services (MySQL, Valkey, Centrifugo, Quickwit, Control Plane, Agent)
- ✅ Wait for services to be healthy
- ✅ Run health checks
- ✅ Test API endpoints
- ✅ Display service URLs and useful commands

### Step 4: Verify Services

After the script completes, you can access:

- **Control Plane API**: http://localhost:8081
- **Centrifugo**: http://localhost:8000
- **MySQL**: localhost:3306
- **Valkey**: localhost:6379
- **Quickwit**: http://localhost:7280

### Step 5: Test with Probe Workflows

If Go is installed in WSL:

```bash
cd /mnt/c/Users/yoges/OneDrive/Documents/My\ Code/Task\ Manager/demo/probe

# Build test program
go build -o test-probe ./cmd/test-probe

# Test HTTP workflow
./test-probe ./examples/http-example.yaml

# Test Command workflow
./test-probe ./examples/command-example.yaml
```

## Alternative: Manual Commands

If you prefer to run commands manually:

```bash
# Navigate to deploy directory
cd /mnt/c/Users/yoges/OneDrive/Documents/My\ Code/Task\ Manager/demo/automation-control-plane/deploy

# Clean up previous runs
docker compose down -v

# Build images
docker compose build control-plane agent-linux

# Start services (excluding Windows agent)
docker compose up -d mysql valkey centrifugo quickwit control-plane agent-linux

# Check status
docker compose ps

# View logs
docker compose logs -f

# Test health endpoints
curl http://localhost:8081/health
curl http://localhost:8000/health
```

## Useful Commands

### View Logs

```bash
# All services
docker compose logs -f

# Specific service
docker compose logs -f control-plane
docker compose logs -f agent-linux
docker compose logs -f mysql
```

### Restart Services

```bash
# Restart specific service
docker compose restart control-plane

# Restart all services
docker compose restart
```

### Stop Services

```bash
# Stop all services
docker compose down

# Stop and remove volumes
docker compose down -v
```

### Execute Commands in Containers

```bash
# MySQL
docker compose exec mysql mysql -u automation -ppassword automation

# Valkey
docker compose exec valkey valkey-cli

# Control Plane shell
docker compose exec control-plane /bin/sh
```

## Testing API Endpoints

### Health Check

```bash
curl http://localhost:8081/health
```

Expected response: `{"status":"ok"}` or similar

### List Agents

```bash
# Without auth (should return 401/403)
curl http://localhost:8081/api/v1/agents

# With JWT token
curl -H "Authorization: Bearer YOUR_TOKEN" http://localhost:8081/api/v1/agents
```

### Create Job (Example)

```bash
curl -X POST http://localhost:8081/api/v1/jobs \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "project_id": "test-project",
    "workflow": "name: test\ntasks:\n  - name: echo\n    type: command\n    config:\n      command: echo\n      args: [\"Hello\"]",
    "workflow_format": "yaml",
    "priority": 1
  }'
```

## Troubleshooting

### Docker command not found in WSL

**Solution:**
1. Make sure Docker Desktop is running on Windows
2. Enable WSL integration in Docker Desktop settings
3. Restart Docker Desktop
4. Close and reopen WSL terminal

### Permission denied errors

```bash
sudo usermod -aG docker $USER
newgrp docker
```

### Services not starting

```bash
# View logs for specific service
docker compose logs control-plane

# Check if ports are in use
sudo netstat -tulpn | grep -E '8081|3306|6379|8000'
```

### Build failures

```bash
# Clean build cache
docker system prune -af

# Rebuild specific service
docker compose build --no-cache control-plane
```

### Can't access from Windows browser

- Make sure services are bound to `0.0.0.0` not `localhost`
- Check Windows Firewall settings
- Try accessing from WSL first: `curl http://localhost:8081/health`

## Windows Agent Note

The `agent-windows` service requires Windows containers and **cannot run in WSL2 Ubuntu**. 

Options:
1. **Skip it** - The build script already excludes it
2. **Run separately** - Build and run on a Windows host with Docker Desktop in Windows container mode

## Next Steps After Services Are Running

1. **Test probe workflows** using the test program
2. **Create jobs** via API
3. **Monitor agent** execution in logs
4. **Test workflow execution** end-to-end

## Quick Reference: Service Ports

| Service | Port | Purpose |
|---------|------|---------|
| Control Plane | 8081 | REST API |
| Centrifugo | 8000 | WebSocket messaging |
| MySQL | 3306 | Database |
| Valkey | 6379 | Cache/Queue |
| Quickwit | 7280 | Log search |

## Full Stack Architecture

```
┌─────────────────────────────────────────────────────────┐
│                     Docker Network                       │
│                                                          │
│  ┌──────────────┐    ┌──────────────┐   ┌──────────┐  │
│  │  MySQL:3306  │◄───┤Control Plane │───►│Valkey    │  │
│  │              │    │    :8081     │   │:6379     │  │
│  └──────────────┘    └──────┬───────┘   └──────────┘  │
│                              │                          │
│  ┌──────────────┐    ┌──────▼───────┐   ┌──────────┐  │
│  │ Quickwit     │◄───┤  Centrifugo  │◄──┤Agent     │  │
│  │   :7280      │    │    :8000     │   │(Linux)   │  │
│  └──────────────┘    └──────────────┘   └──────────┘  │
│                                                          │
└─────────────────────────────────────────────────────────┘
         ▲                      ▲
         │                      │
    Accessible from        Accessible from
    Windows Browser        Windows Browser
```

## Environment Variables

Default values in docker-compose.yml:

```
MYSQL_DSN=automation:password@tcp(mysql:3306)/automation
VALKEY_ADDR=valkey:6379
CENTRIFUGO_URL=http://centrifugo:8000
QUICKWIT_URL=http://quickwit:7280
JWT_SECRET=change-me-in-production
```

To override, create a `.env` file in the deploy directory.

## Success Indicators

When everything is working:

1. ✅ All containers show "Up" status in `docker compose ps`
2. ✅ Health endpoints respond with 200 OK
3. ✅ Agent shows "running" status
4. ✅ Control plane API returns responses
5. ✅ No error logs in `docker compose logs`

## Getting Help

- Check logs: `docker compose logs -f <service-name>`
- Verify network: `docker network inspect deploy_automation-network`
- Check container: `docker compose exec <service-name> /bin/sh`
- Restart service: `docker compose restart <service-name>`

---

**Ready to start? Run these commands:**

```bash
# In PowerShell
wsl -d Ubuntu-22.04

# In WSL Ubuntu
cd /mnt/c/Users/yoges/OneDrive/Documents/My\ Code/Task\ Manager/demo/automation-control-plane/deploy
chmod +x build-and-test-wsl.sh
./build-and-test-wsl.sh
```
