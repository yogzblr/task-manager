# Windows Agent Installation

## Prerequisites
- Windows 10/11 or Windows Server 2016+
- Administrator privileges
- Control plane running and accessible

## Installation Steps

### Option 1: Automated Installation (Recommended)

1. Open PowerShell
2. Navigate to this directory:
   ```powershell
   cd "C:\Users\yoges\OneDrive\Documents\My Code\demo\automation-agent\deploy\windows"
   ```

3. Run the installation script as Administrator:
   ```powershell
   .\install-as-admin.ps1
   ```

   This will:
   - Elevate to Administrator if needed
   - Install the agent binary to `C:\Program Files\AutomationAgent`
   - Create configuration in `C:\ProgramData\AutomationAgent`
   - Install and start the Windows service

### Option 2: Manual Installation

1. Open PowerShell as Administrator
2. Navigate to this directory
3. Run:
   ```powershell
   .\install.ps1 -ControlPlaneUrl "http://localhost:8081" `
                 -TenantId "test-tenant" `
                 -ProjectId "test-project" `
                 -AgentId "agent-windows-01" `
                 -JwtToken "YOUR_JWT_TOKEN_HERE"
   ```

## Verification

Check the service status:
```powershell
Get-Service -Name AutomationAgent
```

View service logs:
```powershell
Get-EventLog -LogName Application -Source "AutomationAgent" -Newest 10
```

## Uninstallation

Run:
```powershell
.\uninstall.ps1
```

## Troubleshooting

### Service won't start
- Check that the control plane URL is accessible
- Verify the JWT token is valid
- Check Windows Event Viewer for error messages

### Connection issues
- Ensure firewall allows outbound connections to control plane
- Verify proxy settings if behind a proxy
- Check network connectivity to control plane and Centrifugo

### Check agent logs
Logs are stored in: `C:\ProgramData\AutomationAgent\logs\`

## Configuration

The agent reads configuration from environment variables set in the wrapper script at:
`C:\Program Files\AutomationAgent\start-agent.cmd`

To modify configuration:
1. Stop the service: `Stop-Service AutomationAgent`
2. Edit the wrapper script
3. Start the service: `Start-Service AutomationAgent`
