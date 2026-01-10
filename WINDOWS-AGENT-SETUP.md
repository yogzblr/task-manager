# Windows Agent Setup Guide

## Overview
The Windows automation agent has been successfully built and is ready for installation. This guide will walk you through testing and installing the agent as a Windows service.

## What Was Done

### 1. Built Windows Agent Binary
- Cross-compiled from Linux container (13MB executable)
- Located at: `automation-agent/deploy/windows/automation-agent.exe`
- Compatible with Windows 10/11 and Windows Server 2016+

### 2. Created Installation Scripts
- **install-as-admin.ps1**: Auto-elevates and installs (recommended)
- **install.ps1**: Main installer (requires admin privileges)
- **test-agent.ps1**: Test agent before installing as service
- **uninstall.ps1**: Removes the agent and service

### 3. Configuration
The agent is pre-configured to connect to:
- **Control Plane**: http://localhost:8081
- **Centrifugo**: ws://localhost:8000/connection/websocket
- **Tenant ID**: test-tenant
- **Project ID**: test-project
- **Agent ID**: agent-windows-01
- **JWT Token**: Valid token (expires in 1 year)

## Installation Options

### Option 1: Test First (Recommended)

Before installing as a service, test the agent to ensure it connects properly:

```powershell
cd "C:\Users\yoges\OneDrive\Documents\My Code\demo\automation-agent\deploy\windows"
.\test-agent.ps1
```

You should see:
- Connection to control plane
- Registration with the server
- Centrifugo WebSocket connection
- Agent status updates

Press `Ctrl+C` to stop the test.

### Option 2: Install as Windows Service

Once you've verified the agent works, install it as a service:

```powershell
cd "C:\Users\yoges\OneDrive\Documents\My Code\demo\automation-agent\deploy\windows"
.\install-as-admin.ps1
```

This will:
1. Request Administrator elevation
2. Copy binary to `C:\Program Files\AutomationAgent\`
3. Create configuration in `C:\ProgramData\AutomationAgent\`
4. Create a service wrapper script
5. Install and start the Windows service named "AutomationAgent"

### Option 3: Manual Installation with Custom Config

```powershell
# Run PowerShell as Administrator
cd "C:\Users\yoges\OneDrive\Documents\My Code\demo\automation-agent\deploy\windows"

.\install.ps1 `
    -ControlPlaneUrl "http://your-control-plane:8081" `
    -TenantId "your-tenant-id" `
    -ProjectId "your-project-id" `
    -AgentId "your-agent-id" `
    -JwtToken "your-jwt-token"
```

## Verification

### Check Service Status
```powershell
Get-Service -Name AutomationAgent
```

Expected output:
```
Status   Name               DisplayName
------   ----               -----------
Running  AutomationAgent    Automation Agent
```

### View Service Details
```powershell
Get-Service -Name AutomationAgent | Format-List *
```

### Check Windows Event Logs
```powershell
Get-EventLog -LogName Application -Source "AutomationAgent" -Newest 10
```

### Control the Service
```powershell
# Stop the service
Stop-Service -Name AutomationAgent

# Start the service
Start-Service -Name AutomationAgent

# Restart the service
Restart-Service -Name AutomationAgent
```

## File Locations

- **Binary**: `C:\Program Files\AutomationAgent\automation-agent.exe`
- **Wrapper Script**: `C:\Program Files\AutomationAgent\start-agent.cmd`
- **Configuration**: `C:\ProgramData\AutomationAgent\service-env.txt`
- **Logs**: `C:\ProgramData\AutomationAgent\logs\`

## Troubleshooting

### Service Won't Start
1. Check control plane is running:
   ```powershell
   Test-NetConnection -ComputerName localhost -Port 8081
   ```

2. Check Centrifugo is running:
   ```powershell
   Test-NetConnection -ComputerName localhost -Port 8000
   ```

3. View Windows Event Viewer:
   - Open Event Viewer (eventvwr)
   - Navigate to Windows Logs > Application
   - Look for AutomationAgent entries

### Connection Issues
- Verify JWT token is valid
- Check firewall isn't blocking connections
- Ensure tenant and project exist in database

### Update Configuration
1. Stop the service:
   ```powershell
   Stop-Service -Name AutomationAgent
   ```

2. Edit the wrapper script:
   ```powershell
   notepad "C:\Program Files\AutomationAgent\start-agent.cmd"
   ```

3. Start the service:
   ```powershell
   Start-Service -Name AutomationAgent
   ```

## Uninstallation

```powershell
cd "C:\Users\yoges\OneDrive\Documents\My Code\demo\automation-agent\deploy\windows"
.\uninstall.ps1
```

This will:
- Stop the service
- Remove the service registration
- Delete installed files
- Clean up configuration

## Next Steps

1. **Test the Agent**: Run `test-agent.ps1` to verify connectivity
2. **Install as Service**: Run `install-as-admin.ps1` for production deployment
3. **Monitor Logs**: Check Windows Event Viewer for agent activity
4. **Create Jobs**: Use the control plane API to send jobs to the agent

## Architecture Notes

### Windows Service Implementation
- The agent runs as a Windows service under LocalSystem account
- A wrapper script (`start-agent.cmd`) sets environment variables
- Configuration is passed via environment variables (no config files)
- Service automatically restarts on failure (3 retries)

### Security
- JWT token authentication with control plane
- All connections are outbound-only (agent-initiated)
- Supports proxy-aware networking
- Artifact verification with SHA256 and Ed25519 signatures

### Agent Capabilities
- Execute shell commands and PowerShell scripts
- Download and execute binaries with signature verification
- Run workflows with multiple steps
- Auto-upgrade capability
- Lease-based job execution (exactly-once semantics)

## Files in deploy/windows Directory

```
automation-agent.exe      - Windows agent binary (13MB)
install-as-admin.ps1      - Auto-elevating installer
install.ps1               - Main installation script
test-agent.ps1            - Test runner for debugging
uninstall.ps1             - Uninstaller
install-service.ps1       - Service-only installer
uninstall-service.ps1     - Service-only uninstaller
automation-agent.wxs      - WiX installer definition
automation-agent.nsi      - NSIS installer definition
README.md                 - Installation documentation
```

## Support

For issues or questions:
1. Check the troubleshooting section above
2. Review Windows Event Viewer logs
3. Check agent logs in `C:\ProgramData\AutomationAgent\logs\`
4. Verify control plane and Centrifugo are accessible
