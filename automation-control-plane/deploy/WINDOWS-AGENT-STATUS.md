# Windows Agent - Build and Test Summary

## Current Status: ⚠️ Requires Go Installation

### Why Windows Agent Can't Be Built in Docker Right Now

**Docker Container Limitation**: Docker Desktop can only run either Linux OR Windows containers at a time, not both simultaneously. Since the control plane and all infrastructure services are running as Linux containers in WSL2, we cannot run a Windows container alongside them.

### ✅ Solution: Build and Run Natively on Windows

The Windows agent is designed to work as a **native Windows application**, not requiring Docker. This is actually the preferred approach for production as well.

---

## 🎯 How to Build and Run the Windows Agent

### Prerequisites

1. **Go installed on Windows** (required for building)
   - Download from: https://go.dev/dl/
   - Install and add to PATH
   - Verify: `go version`

2. **Control Plane running in WSL** (already done ✅)
   - Services: MySQL, Valkey, Centrifugo, Control Plane, Linux Agent
   - All running and healthy in Docker Compose

### Step 1: Build the Windows Agent

```powershell
# Option A: Build on Windows (if Go is installed)
cd "C:\Users\yoges\OneDrive\Documents\My Code\Task Manager\demo\automation-agent"
go build -o automation-agent-windows.exe ./cmd/agent

# Option B: Cross-compile in WSL (requires Go in WSL)
wsl -d Ubuntu-22.04 bash -c "cd /mnt/c/Users/yoges/OneDrive/Documents/My\ Code/Task\ Manager/demo/automation-agent && CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o automation-agent-windows.exe ./cmd/agent"
```

### Step 2: Run the Windows Agent

```powershell
# Use the provided script
cd "C:\Users\yoges\OneDrive\Documents\My Code\Task Manager\demo\automation-agent"
.\run-windows-agent.ps1
```

**OR manually:**

```powershell
# Set environment variables
$env:CONTROL_PLANE_URL = "http://localhost:8081"
$env:CENTRIFUGO_URL = "ws://localhost:8000/connection/websocket"
$env:TENANT_ID = "test-tenant"
$env:PROJECT_ID = "test-project"
$env:AGENT_ID = "agent-windows-01"
$env:JWT_TOKEN = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhZ2VudF9pZCI6ImFnZW50LXdpbmRvd3MtMDEiLCJ0ZW5hbnRfaWQiOiJ0ZXN0LXRlbmFudCIsInByb2plY3RfaWQiOiJ0ZXN0LXByb2plY3QiLCJleHAiOjE3OTk0OTUwMDksImlhdCI6MTc2Nzk1OTAwOX0.JKQXv4YeRRA46gPU-cJpyV83FC2ZFXxWrR_M1zkuQO0"
$env:LOG_LEVEL = "debug"

# Run the agent
.\automation-agent-windows.exe
```

### Step 3: Verify Registration

```powershell
# Check in WSL database
wsl -d Ubuntu-22.04 bash -c "cd /mnt/c/Users/yoges/OneDrive/Documents/My\ Code/Task\ Manager/demo/automation-control-plane/deploy && docker compose exec -T mysql mysql -u automation -ppassword automation -e 'SELECT agent_id, os FROM agents;'"
```

You should see:
- `agent-linux-01` (os: linux) - Running in Docker
- `agent-windows-01` (os: windows) - Running natively on Windows

---

## 🧪 Testing Windows-Specific Features

Once the Windows agent is running, you can test Windows-specific tasks:

### 1. PowerShell Task

```yaml
# test-powershell.yaml
tasks:
  - name: powershell-test
    type: powershell
    script: |
      Write-Host "Testing PowerShell on Windows"
      Get-ComputerInfo | Select-Object OsName, OsVersion, WindowsVersion
      Get-Process | Select-Object -First 5 ProcessName, CPU, WorkingSet
    timeout: 30s
```

### 2. Windows Command Task

```yaml
# test-windows-cmd.yaml
tasks:
  - name: cmd-test
    type: command
    command: cmd.exe
    args:
      - /c
      - systeminfo | findstr /C:"OS Name" /C:"OS Version"
    timeout: 30s
```

### 3. DownloadExec Task

```yaml
# test-downloadexec.yaml
tasks:
  - name: download-test
    type: downloadexec
    url: https://example.com/tool.exe
    checksum: "actual-sha256-hash-here"
    args:
      - "--version"
    timeout: 60s
```

---

## 📊 Current System Status

### ✅ What's Working

| Component | Status | Platform | Notes |
|-----------|--------|----------|-------|
| Control Plane | ✅ Running | Linux (Docker) | Port 8081 |
| MySQL | ✅ Running | Linux (Docker) | Port 3306 |
| Valkey | ✅ Running | Linux (Docker) | Port 6379 |
| Centrifugo | ✅ Running | Linux (Docker) | Port 8000 |
| Quickwit | ✅ Running | Linux (Docker) | Port 7280 |
| Linux Agent | ✅ Running | Linux (Docker) | Registered |
| Windows Agent | ⏳ Pending | Windows (Native) | Needs Go to build |

### 📝 Files Created

1. **run-windows-agent.ps1** - PowerShell script to run Windows agent
2. **WINDOWS-AGENT-BUILD.md** - Detailed Windows agent build guide  
3. **Dockerfile.agent-windows** - Updated for proper cross-compilation
4. **WSL-BUILD-SUCCESS.md** - Linux agent build success documentation

---

## 🎯 Next Steps for You

### Immediate Actions:

1. **Install Go on Windows** (if not already installed)
   - Download: https://go.dev/dl/
   - Run installer
   - Verify: `go version` in PowerShell

2. **Build Windows Agent**
   ```powershell
   cd "C:\Users\yoges\OneDrive\Documents\My Code\Task Manager\demo\automation-agent"
   go build -o automation-agent-windows.exe ./cmd/agent
   ```

3. **Run Windows Agent**
   ```powershell
   .\run-windows-agent.ps1
   ```

4. **Verify Both Agents**
   ```bash
   # Check in database
   wsl -d Ubuntu-22.04 bash -c "cd /mnt/c/Users/yoges/OneDrive/Documents/My\ Code/Task\ Manager/demo/automation-control-plane/deploy && docker compose exec -T mysql mysql -u automation -ppassword automation -e 'SELECT agent_id, os, created_at FROM agents;'"
   ```

### Optional: Install Go in WSL

If you want to cross-compile from WSL:

```bash
# In WSL Ubuntu
sudo apt update
sudo apt install golang-go

# Or install latest version
wget https://go.dev/dl/go1.21.6.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.21.6.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
go version
```

---

## 🎊 Summary

**The automation platform is fully operational with:**

- ✅ Control plane and all infrastructure services running in WSL Docker
- ✅ Linux agent running and registered
- ✅ Windows agent code complete and ready to build
- ✅ All probe tasks integrated (HTTP, DB, SSH, Command, PowerShell, DownloadExec)
- ✅ Database migrations applied
- ✅ Test data loaded
- ✅ System ready for workflow execution

**To complete the Windows agent testing:**
- Install Go on Windows
- Build the Windows agent binary
- Run it with the provided script
- Both agents will then be operational simultaneously!

---

## 📚 Documentation

- **Linux Agent**: See `WSL-BUILD-SUCCESS.md` for complete Linux agent build details
- **Windows Agent**: See `WINDOWS-AGENT-BUILD.md` for comprehensive Windows agent guide
- **Running in WSL**: See `RUN-DOCKER-WSL.md` for WSL Docker Compose documentation
- **Quick Start**: See `QUICK-START-WSL.md` for simplified commands
- **Docker Guide**: See `WSL-DOCKER-GUIDE.md` for detailed Docker setup

All documentation is in: `C:\Users\yoges\OneDrive\Documents\My Code\Task Manager\demo\automation-control-plane\deploy\`
