# Automation Agent

Cross-platform automation agent that extends the `probe` task runner with control plane integration.

## Features

- Extends probe task runner
- Proxy-aware networking
- Workflow DSL execution
- Artifact signature verification
- Atomic binary upgrades
- Systemd/Windows service support

## Installation

### Linux

```bash
sudo ./deploy/linux/install.sh
```

### Windows

```powershell
.\deploy\windows\install.ps1 -ControlPlaneUrl "https://cp.example.com" -TenantId "tenant" -ProjectId "project" -JwtToken "token"
```

## Configuration

Set environment variables or use configuration file:

- `CONTROL_PLANE_URL` - Control plane API URL
- `TENANT_ID` - Tenant ID
- `PROJECT_ID` - Project ID
- `AGENT_ID` - Agent ID (optional, auto-generated if not set)
- `JWT_TOKEN` - JWT authentication token

## License

Proprietary
