# Docker Compose Build and Test - Complete Instructions

## 🚀 Quick Start (Copy and Run)

### Step 1: Enable Docker in WSL

1. **Start Docker Desktop on Windows**
   - Find Docker Desktop icon and start it
   - Wait until it shows "Docker Desktop is running" in system tray

2. **Enable WSL Integration**
   ```
   Docker Desktop → Settings → Resources → WSL Integration
   → Enable "Ubuntu-22.04"
   → Click "Apply & Restart"
   ```

3. **Verify Docker is enabled** (in PowerShell):
   ```powershell
   wsl -d Ubuntu-22.04 docker --version
   ```
   
   If you see version info, you're ready! If not, restart Docker Desktop.

### Step 2: Run Build and Test

Copy and paste these commands into PowerShell:

```powershell
# Start WSL Ubuntu
wsl -d Ubuntu-22.04

# Once in Ubuntu terminal, run:
cd /mnt/c/Users/yoges/OneDrive/Documents/My\ Code/Task\ Manager/demo/automation-control-plane/deploy && chmod +x build-and-test-wsl.sh && ./build-and-test-wsl.sh
```

**OR** run in one command from PowerShell:

```powershell
wsl -d Ubuntu-22.04 bash -c "cd /mnt/c/Users/yoges/OneDrive/Documents/My\ Code/Task\ Manager/demo/automation-control-plane/deploy && chmod +x build-and-test-wsl.sh && ./build-and-test-wsl.sh"
```

That's it! The script will:
- ✅ Build all Docker images
- ✅ Start all services
- ✅ Run health checks
- ✅ Test API endpoints
- ✅ Display service URLs

## 📊 What Gets Built and Started

### Services Started:

1. **MySQL** (Port 3306) - Database
2. **Valkey** (Port 6379) - Cache/Queue  
3. **Centrifugo** (Port 8000) - WebSocket messaging
4. **Quickwit** (Port 7280) - Log search
5. **Control Plane** (Port 8081) - Main API
6. **Linux Agent** - Automation agent

### Build Time:
- First build: 5-10 minutes (downloads base images)
- Subsequent builds: 1-2 minutes (uses cache)

## 🧪 Testing After Build

### Test Control Plane API

From PowerShell or WSL:

```bash
# Health check
curl http://localhost:8081/health

# List agents (requires auth)
curl http://localhost:8081/api/v1/agents
```

### Test with Probe Workflows

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

## 🔍 Viewing Logs

From WSL Ubuntu:

```bash
cd /mnt/c/Users/yoges/OneDrive/Documents/My\ Code/Task\ Manager/demo/automation-control-plane/deploy

# All logs
docker compose logs -f

# Specific service
docker compose logs -f control-plane
docker compose logs -f agent-linux

# Last 50 lines
docker compose logs --tail=50 control-plane
```

## 🛑 Stopping Services

```bash
# Stop all services
docker compose down

# Stop and remove volumes (fresh start)
docker compose down -v
```

## 🔧 Troubleshooting

### "Docker command not found"

**Solution:**
1. Make sure Docker Desktop is running on Windows
2. Go to Docker Desktop → Settings → Resources → WSL Integration
3. Enable "Ubuntu-22.04"
4. Click "Apply & Restart"
5. Close and reopen WSL terminal

### Build fails with "Cannot connect to Docker daemon"

**Solution:**
```bash
# Check Docker is running
docker ps

# If error, restart Docker Desktop on Windows
# Then try again
```

### Port already in use (8081, 3306, etc.)

**Solution:**
```bash
# Stop conflicting services
docker compose down

# Or check what's using the port (in PowerShell)
netstat -ano | findstr :8081

# Kill the process using the port
taskkill /PID <process_id> /F
```

### Services not starting

**Solution:**
```bash
# Check logs
docker compose logs <service-name>

# Rebuild from scratch
docker compose down -v
docker compose build --no-cache control-plane
docker compose up -d
```

## 📦 Service URLs

Access from Windows browser or curl:

| Service | URL | Purpose |
|---------|-----|---------|
| Control Plane | http://localhost:8081 | REST API |
| Centrifugo | http://localhost:8000 | WebSocket |
| Quickwit | http://localhost:7280 | Logs UI |

Database connections:
- MySQL: `localhost:3306` (user: automation, pass: password)
- Valkey: `localhost:6379`

## 🎯 Success Checklist

After running the build script, verify:

- [ ] `docker compose ps` shows all services "Up"
- [ ] `curl http://localhost:8081/health` returns success
- [ ] `curl http://localhost:8000/health` returns success  
- [ ] Agent logs show "connected" or "running"
- [ ] No error messages in `docker compose logs`

## 📝 Useful Commands Reference

```bash
# Navigate to deploy directory
cd /mnt/c/Users/yoges/OneDrive/Documents/My\ Code/Task\ Manager/demo/automation-control-plane/deploy

# Build images
docker compose build

# Start services
docker compose up -d

# Stop services
docker compose down

# View status
docker compose ps

# View logs (follow)
docker compose logs -f

# View logs (specific service)
docker compose logs -f control-plane

# Restart service
docker compose restart control-plane

# Execute command in container
docker compose exec control-plane /bin/sh
docker compose exec mysql mysql -u automation -ppassword automation

# Remove everything (fresh start)
docker compose down -v
docker system prune -a
```

## 🚀 Next Steps

After services are running:

1. **Test API endpoints** with curl
2. **Create a job** via API
3. **Watch agent logs** to see job execution
4. **Test probe workflows** with the test program
5. **Build and test the Windows agent** (separate setup)

## 📚 Additional Documentation

- Full setup: `WSL-DOCKER-GUIDE.md`
- Docker Compose: `docker-compose.yml`
- Build script: `build-and-test-wsl.sh`
- WSL setup: `SETUP-WSL.md`

---

## ⚡ TL;DR - One Command

```powershell
# From PowerShell (after enabling Docker Desktop WSL integration):
wsl -d Ubuntu-22.04 bash -c "cd /mnt/c/Users/yoges/OneDrive/Documents/My\ Code/Task\ Manager/demo/automation-control-plane/deploy && chmod +x build-and-test-wsl.sh && ./build-and-test-wsl.sh"
```

This single command will build, start, and test everything!
