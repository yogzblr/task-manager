# 🚀 Complete Docker WSL2 Setup and Execution Guide

## Current Status

✅ Ubuntu-22.04 WSL2 is installed
❌ Docker Desktop WSL integration needs to be enabled
📦 Build scripts are ready
📝 All documentation is complete

## Step-by-Step Instructions

### Option 1: Automated Setup (Recommended)

Run this PowerShell script as Administrator:

```powershell
cd "C:\Users\yoges\OneDrive\Documents\My Code\Task Manager\demo\automation-control-plane\deploy"
.\setup-and-run-wsl.ps1
```

This script will:
1. Check if Docker Desktop is running
2. Guide you through enabling WSL integration
3. Verify Docker is accessible in WSL
4. Automatically run the build and test script

### Option 2: Manual Setup

#### Step 1: Start Docker Desktop

1. Start **Docker Desktop** on Windows
2. Wait for it to show "Docker Desktop is running" in system tray

#### Step 2: Enable WSL Integration

1. Click Docker Desktop icon in system tray
2. Click **Settings** (gear icon)
3. Navigate to: **Resources** → **WSL Integration**
4. **Enable the toggle** for **Ubuntu-22.04**
5. Click **Apply & Restart**
6. Wait for Docker Desktop to restart (~30 seconds)

#### Step 3: Verify Docker in WSL

In PowerShell:

```powershell
wsl -d Ubuntu-22.04 docker --version
```

Expected output: `Docker version XX.XX.XX...`

If you see "command not found", Docker integration isn't enabled yet. Repeat Step 2.

#### Step 4: Run Build and Test

**ONE-LINER** (copy and paste into PowerShell):

```powershell
wsl -d Ubuntu-22.04 bash -c "cd /mnt/c/Users/yoges/OneDrive/Documents/My\ Code/Task\ Manager/demo/automation-control-plane/deploy && chmod +x build-and-test-wsl.sh && ./build-and-test-wsl.sh"
```

**OR** run in steps:

```powershell
# Start WSL
wsl -d Ubuntu-22.04

# Inside WSL terminal, run:
cd /mnt/c/Users/yoges/OneDrive/Documents/My\ Code/Task\ Manager/demo/automation-control-plane/deploy
chmod +x build-and-test-wsl.sh
./build-and-test-wsl.sh
```

## What the Build Script Does

The `build-and-test-wsl.sh` script will:

1. ✅ **Verify** Docker is available
2. ✅ **Clean up** previous containers/volumes
3. ✅ **Build** Control Plane Docker image (~3-5 minutes)
4. ✅ **Build** Linux Agent Docker image (~2-3 minutes)
5. ✅ **Start** all services:
   - MySQL (database)
   - Valkey (cache/queue)
   - Centrifugo (WebSocket)
   - Quickwit (logs)
   - Control Plane (API)
   - Linux Agent
6. ✅ **Wait** for services to be healthy
7. ✅ **Test** all health endpoints
8. ✅ **Verify** API responses
9. ✅ **Display** service URLs and commands

**Total time**: ~5-10 minutes on first run (downloads base images)

## Expected Output

```
=== Automation Platform - WSL Build and Test ===
✓ Running in WSL Ubuntu
✓ Docker is available: Docker version 24.0.x...
✓ Docker Compose is available: Docker Compose version v2.x.x
✓ Working directory: /mnt/c/Users/yoges/.../deploy

=== Cleaning Up Previous Runs ===
✓ Previous containers and volumes removed

=== Building Docker Images ===
Building control-plane image...
✓ Control plane image built

Building agent-linux image...
✓ Linux agent image built

=== Starting Services ===
✓ All services started

=== Waiting for Services to be Healthy ===
✓ MySQL is ready
✓ Valkey is ready
✓ Centrifugo is ready
✓ Control Plane is ready

=== Service Status ===
NAME               STATUS        PORTS
mysql              Up (healthy)  0.0.0.0:3306->3306/tcp
valkey             Up (healthy)  0.0.0.0:6379->6379/tcp
centrifugo         Up (healthy)  0.0.0.0:8000->8000/tcp
quickwit           Up (healthy)  0.0.0.0:7280->7280/tcp
control-plane      Up (healthy)  0.0.0.0:8081->8080/tcp
agent-linux        Up            

=== Testing Services ===
✓ MySQL connection successful
✓ Valkey connection successful
✓ Centrifugo API is responding
✓ Control Plane API is responding

=== Agent Status ===
✓ Linux agent is running

=== Service URLs ===
  Control Plane API: http://localhost:8081
  Centrifugo:        http://localhost:8000
  MySQL:             localhost:3306
  Valkey:            localhost:6379
  Quickwit:          http://localhost:7280

=== Build and test completed successfully! ===
```

## After Build Completes

### Access Services

Open in your Windows browser:
- **Control Plane API**: http://localhost:8081/health
- **Centrifugo**: http://localhost:8000/health
- **Quickwit UI**: http://localhost:7280

### View Logs

```powershell
# All services
wsl -d Ubuntu-22.04 bash -c "cd /mnt/c/Users/yoges/OneDrive/Documents/My\ Code/Task\ Manager/demo/automation-control-plane/deploy && docker compose logs -f"

# Specific service
wsl -d Ubuntu-22.04 bash -c "cd /mnt/c/Users/yoges/OneDrive/Documents/My\ Code/Task\ Manager/demo/automation-control-plane/deploy && docker compose logs -f control-plane"
```

### Check Status

```powershell
wsl -d Ubuntu-22.04 bash -c "cd /mnt/c/Users/yoges/OneDrive/Documents/My\ Code/Task\ Manager/demo/automation-control-plane/deploy && docker compose ps"
```

### Stop Services

```powershell
wsl -d Ubuntu-22.04 bash -c "cd /mnt/c/Users/yoges/OneDrive/Documents/My\ Code/Task\ Manager/demo/automation-control-plane/deploy && docker compose down"
```

### Restart Services

```powershell
wsl -d Ubuntu-22.04 bash -c "cd /mnt/c/Users/yoges/OneDrive/Documents/My\ Code/Task\ Manager/demo/automation-control-plane/deploy && docker compose restart"
```

## Test with Probe Workflows

Once services are running, test with probe workflows:

### If Go is installed in WSL:

```bash
cd /mnt/c/Users/yoges/OneDrive/Documents/My\ Code/Task\ Manager/demo/probe

# Build test program
go build -o test-probe ./cmd/test-probe

# Test HTTP workflow (tests against httpbin.org and GitHub API)
./test-probe ./examples/http-example.yaml

# Test Command workflow
./test-probe ./examples/command-example.yaml
```

### Test Control Plane API:

```powershell
# Health check
curl http://localhost:8081/health

# List agents (will require JWT token)
curl http://localhost:8081/api/v1/agents
```

## Troubleshooting

### "Docker command not found"

**Cause**: Docker Desktop WSL integration not enabled

**Solution**:
1. Open Docker Desktop
2. Go to Settings → Resources → WSL Integration
3. Enable Ubuntu-22.04
4. Click Apply & Restart
5. Wait 30 seconds
6. Try again

### "Cannot connect to Docker daemon"

**Cause**: Docker Desktop not running

**Solution**:
1. Start Docker Desktop
2. Wait for "Docker Desktop is running"
3. Try again

### Ports already in use

**Cause**: Services already running on ports 8081, 3306, 6379, 8000, 7280

**Solution**:
```powershell
# Check what's using the port
netstat -ano | findstr :8081

# Stop existing containers
wsl -d Ubuntu-22.04 bash -c "cd /mnt/c/Users/yoges/OneDrive/Documents/My\ Code/Task\ Manager/demo/automation-control-plane/deploy && docker compose down"
```

### Build fails

**Solution**:
```powershell
# Clean everything and rebuild
wsl -d Ubuntu-22.04 bash -c "cd /mnt/c/Users/yoges/OneDrive/Documents/My\ Code/Task\ Manager/demo/automation-control-plane/deploy && docker compose down -v && docker system prune -af && docker compose build --no-cache"
```

### Services not healthy

**Solution**:
```powershell
# Check logs for specific service
wsl -d Ubuntu-22.04 bash -c "cd /mnt/c/Users/yoges/OneDrive/Documents/My\ Code/Task\ Manager/demo/automation-control-plane/deploy && docker compose logs control-plane"
```

## Quick Reference Commands

All commands assume you're in PowerShell. Replace `SERVICE` with actual service name.

```powershell
# Navigate to deploy directory in WSL
$DEPLOY = "cd /mnt/c/Users/yoges/OneDrive/Documents/My\ Code/Task\ Manager/demo/automation-control-plane/deploy"

# Build images
wsl -d Ubuntu-22.04 bash -c "$DEPLOY && docker compose build"

# Start services
wsl -d Ubuntu-22.04 bash -c "$DEPLOY && docker compose up -d"

# Stop services
wsl -d Ubuntu-22.04 bash -c "$DEPLOY && docker compose down"

# View logs
wsl -d Ubuntu-22.04 bash -c "$DEPLOY && docker compose logs -f"

# View logs for specific service
wsl -d Ubuntu-22.04 bash -c "$DEPLOY && docker compose logs -f SERVICE"

# Check status
wsl -d Ubuntu-22.04 bash -c "$DEPLOY && docker compose ps"

# Restart service
wsl -d Ubuntu-22.04 bash -c "$DEPLOY && docker compose restart SERVICE"

# Execute command in container
wsl -d Ubuntu-22.04 bash -c "$DEPLOY && docker compose exec SERVICE /bin/sh"

# Remove everything
wsl -d Ubuntu-22.04 bash -c "$DEPLOY && docker compose down -v"
```

## Success Checklist

After running the build script, you should see:

- [x] Docker available in WSL
- [x] All images built successfully
- [x] All services started
- [x] Health checks passing
- [x] MySQL ready
- [x] Valkey ready  
- [x] Centrifugo ready
- [x] Control Plane ready
- [x] Agent running
- [x] APIs responding
- [x] No error logs

## Next Steps

1. ✅ **Services Running** - You're here after build completes
2. 🧪 **Test APIs** - Use curl or browser
3. 🔬 **Test Workflows** - Run probe examples
4. 📊 **Monitor** - Watch logs
5. 🚀 **Deploy** - Move to production

## Documentation

- **This Guide**: `RUN-DOCKER-WSL.md`
- **Quick Start**: `QUICK-START-WSL.md`
- **Full Guide**: `WSL-DOCKER-GUIDE.md`
- **Setup**: `SETUP-WSL.md`
- **Build Script**: `build-and-test-wsl.sh`
- **Setup Script**: `setup-and-run-wsl.ps1`

---

## 🎯 TL;DR

```powershell
# 1. Enable Docker Desktop WSL integration (one-time setup)
#    Docker Desktop → Settings → Resources → WSL Integration → Enable Ubuntu-22.04

# 2. Run this command:
cd "C:\Users\yoges\OneDrive\Documents\My Code\Task Manager\demo\automation-control-plane\deploy"
.\setup-and-run-wsl.ps1

# OR run manually:
wsl -d Ubuntu-22.04 bash -c "cd /mnt/c/Users/yoges/OneDrive/Documents/My\ Code/Task\ Manager/demo/automation-control-plane/deploy && chmod +x build-and-test-wsl.sh && ./build-and-test-wsl.sh"
```

**That's it!** Services will be built and running in ~5-10 minutes.
