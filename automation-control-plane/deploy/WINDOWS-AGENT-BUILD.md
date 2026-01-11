# Windows Agent Build and Test Guide

## ⚠️ Important Limitation

**Windows containers cannot run alongside Linux containers in the same Docker environment.**

Docker Desktop can run either:
- Linux containers (default, what we're using for the control plane)
- Windows containers (requires switching modes)

Since the control plane and other services are running as Linux containers, we cannot run the Windows agent container simultaneously.

## 🎯 Alternative Options

### Option 1: Build Windows Binary Directly on Windows Host

Instead of using Docker, build and run the Windows agent natively:

```powershell
# Navigate to agent directory
cd C:\Users\yoges\OneDrive\Documents\My Code\Task Manager\demo\automation-agent

# Build Windows binary
go build -o automation-agent.exe ./cmd/agent

# Set environment variables
$env:CONTROL_PLANE_URL = "http://localhost:8081"
$env:CENTRIFUGO_URL = "ws://localhost:8000/connection/websocket"
$env:TENANT_ID = "test-tenant"
$env:PROJECT_ID = "test-project"
$env:AGENT_ID = "agent-windows-01"
$env:JWT_TOKEN = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhZ2VudF9pZCI6ImFnZW50LWxpbnV4LTAxIiwidGVuYW50X2lkIjoidGVzdC10ZW5hbnQiLCJwcm9qZWN0X2lkIjoidGVzdC1wcm9qZWN0IiwiZXhwIjoxNzk5NDk1MDA5LCJpYXQiOjE3Njc5NTkwMDl9.JKQXv4YeRRA46gPU-cJpyV83FC2ZFXxWrR_M1zkuQO0"
$env:LOG_LEVEL = "debug"

# Run agent
.\automation-agent.exe
```

### Option 2: Cross-Compile in Linux and Run on Windows

We can build the Windows binary in WSL and then run it on Windows:

```powershell
# Build in WSL (cross-compile)
wsl -d Ubuntu-22.04 bash -c "cd /mnt/c/Users/yoges/OneDrive/Documents/My\ Code/Task\ Manager/demo/automation-agent && CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o automation-agent.exe ./cmd/agent"

# Then follow Option 1 to run it
```

### Option 3: Docker Windows Container Mode (Not Recommended)

To run Windows containers, you would need to:

1. Switch Docker Desktop to Windows container mode
2. Stop all Linux containers (including control plane)
3. Build and run Windows agent
4. Switch back to Linux mode for control plane

This is impractical for testing both agents simultaneously.

## ✅ Recommended Approach: Native Windows Build

**Steps:**

1. **Install Go on Windows** (if not already installed):
   - Download from https://go.dev/dl/
   - Run installer
   - Verify: `go version`

2. **Build the Windows agent**:
   ```powershell
   cd "C:\Users\yoges\OneDrive\Documents\My Code\Task Manager\demo\automation-agent"
   go build -o automation-agent.exe ./cmd/agent
   ```

3. **Run the agent** (with control plane running in WSL):
   ```powershell
   # Set environment
   $env:CONTROL_PLANE_URL = "http://localhost:8081"
   $env:CENTRIFUGO_URL = "ws://localhost:8000/connection/websocket"
   $env:TENANT_ID = "test-tenant"
   $env:PROJECT_ID = "test-project"
   $env:AGENT_ID = "agent-windows-01"
   $env:JWT_TOKEN = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhZ2VudF9pZCI6ImFnZW50LXdpbmRvd3MtMDEiLCJ0ZW5hbnRfaWQiOiJ0ZXN0LXRlbmFudCIsInByb2plY3RfaWQiOiJ0ZXN0LXByb2plY3QiLCJleHAiOjE3OTk0OTUwMDksImlhdCI6MTc2Nzk1OTAwOX0.JKQXv4YeRRA46gPU-cJpyV83FC2ZFXxWrR_M1zkuQO0"
   $env:LOG_LEVEL = "debug"
   
   # Run
   .\automation-agent.exe
   ```

4. **Verify registration**:
   ```bash
   # In WSL, check database
   wsl -d Ubuntu-22.04 bash -c "cd /mnt/c/Users/yoges/OneDrive/Documents/My\ Code/Task\ Manager/demo/automation-control-plane/deploy && docker compose exec -T mysql mysql -u automation -ppassword automation -e 'SELECT * FROM agents;'"
   ```

## 🧪 Testing Windows-Specific Features

Once the Windows agent is running, you can test:

### PowerShell Task
```yaml
tasks:
  - name: windows-ps-test
    type: powershell
    script: |
      Write-Host "Testing PowerShell task"
      Get-Process | Select-Object -First 5
    timeout: 30s
```

### DownloadExec Task
```yaml
tasks:
  - name: download-test
    type: downloadexec
    url: https://example.com/tool.exe
    checksum: "sha256-hash-here"
    args:
      - "--version"
    timeout: 60s
```

## 📊 Expected Behavior

When running natively on Windows:

- ✅ Agent registers with control plane
- ✅ Connects to Centrifugo WebSocket
- ✅ Can execute all task types:
  - HTTP tasks
  - Command tasks (cmd.exe)
  - PowerShell tasks
  - DownloadExec tasks
  - Database tasks (if MySQL client installed)
  - SSH tasks (if SSH client available)

## 🔍 Troubleshooting

### Agent Won't Connect
- Check control plane is accessible: `curl http://localhost:8081/health`
- Check Centrifugo is running: `curl http://localhost:8000`
- Verify network connectivity between Windows and WSL

### PowerShell Tasks Fail
- Ensure PowerShell is in PATH
- Check execution policy: `Get-ExecutionPolicy`
- May need: `Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser`

### Database Connection Issues
- MySQL client may not be available on Windows
- Consider using HTTP tasks for remote operations instead

---

## 📝 Summary

**The Windows agent works best when built and run natively on Windows**, not in a Docker container. This allows:
- Simultaneous operation with Linux control plane
- Native Windows features (PowerShell, Windows services, etc.)
- Simpler deployment and testing
- No Docker mode switching required

The updated `Dockerfile.agent-windows` can still be used for containerized deployments in production Windows container environments, but for development and testing, **native execution is recommended**.
