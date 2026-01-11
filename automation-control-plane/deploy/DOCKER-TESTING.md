# Docker-Based Workflow Testing

This guide explains how to run workflow tests using the containerized test-runner service. This keeps your Windows machine clean and runs all tests within the Docker environment.

## Overview

The `test-runner` service is a Python 3.11 container that includes:
- ✅ All test scripts (Linux, Windows workflows)
- ✅ Required Python dependencies (requests)
- ✅ Pre-configured environment variables for Docker network
- ✅ Direct access to control plane and Quickwit services

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Docker Network                           │
│                                                             │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐ │
│  │ test-runner  │───▶│control-plane │───▶│    MySQL     │ │
│  │  (Python)    │    │   (Go API)   │    │              │ │
│  └──────────────┘    └──────────────┘    └──────────────┘ │
│         │                    │                             │
│         │                    ▼                             │
│         │            ┌──────────────┐    ┌──────────────┐ │
│         │            │ agent-linux  │    │  Centrifugo  │ │
│         │            │  (executes)  │    │              │ │
│         │            └──────────────┘    └──────────────┘ │
│         │                                                  │
│         └───────────────────▶┌──────────────┐             │
│                               │  Quickwit    │             │
│                               │  (logs API)  │             │
│                               └──────────────┘             │
└─────────────────────────────────────────────────────────────┘
```

## Prerequisites

Ensure all Docker services are running:

```bash
# In WSL Ubuntu
cd /mnt/c/Users/yoges/OneDrive/Documents/My\ Code/Task\ Manager/demo/automation-control-plane/deploy

# Start all services
docker compose up -d

# Verify services
docker compose ps
```

Expected services:
- ✅ mysql (healthy)
- ✅ valkey (running)
- ✅ centrifugo (healthy)
- ✅ quickwit (healthy)
- ✅ control-plane (healthy)
- ✅ agent-linux (running)

## Running Tests

### 1. Build the Test Runner Image

```bash
# In WSL Ubuntu
cd /mnt/c/Users/yoges/OneDrive/Documents/My\ Code/Task\ Manager/demo/automation-control-plane/deploy

# Build the test-runner service
docker compose build test-runner
```

### 2. Run Linux Shell Workflow Test

```bash
# Execute Linux workflow test
docker compose run --rm test-runner python test-linux-workflow.py
```

Expected output:
```
╔════════════════════════════════════════════════════════════╗
║     Linux Shell Workflow Test - Probe Integration         ║
╚════════════════════════════════════════════════════════════╝

✓ Control plane is accessible

============================================================
Submitting Linux Shell Workflow
============================================================
✓ Workflow submitted successfully!
  Job ID: 550e8400-e29b-41d4-a716-446655440000

============================================================
Monitoring Job Execution
============================================================
  Status: executing (2s elapsed)
✓ Job completed successfully!

============================================================
Searching Quickwit for Execution Logs
============================================================
✓ Found 15 log entries
```

### 3. Run Windows PowerShell Workflow Test

**Note**: This requires a Windows agent to be running (either natively on Windows or via Windows containers).

```bash
# Execute Windows workflow test (requires Windows agent)
docker compose run --rm test-runner python test-windows-workflow.py
```

### 4. Interactive Test Session

You can also start an interactive shell in the test-runner container:

```bash
# Start interactive session
docker compose run --rm test-runner /bin/bash

# Inside container:
python test-linux-workflow.py
python test-windows-workflow.py
ls -la  # View available files
exit
```

## Environment Variables

The test-runner container uses these environment variables (pre-configured in docker-compose.yml):

| Variable | Value | Description |
|----------|-------|-------------|
| `CONTROL_PLANE_URL` | `http://control-plane:8080` | Control plane API endpoint |
| `QUICKWIT_URL` | `http://quickwit:7280` | Quickwit search API endpoint |
| `TENANT_ID` | `test-tenant` | Default tenant ID |
| `PROJECT_ID` | `test-project` | Default project ID |

**Note**: Inside Docker network, services use internal DNS names (e.g., `control-plane:8080`) instead of `localhost:8081`.

## Service Configuration

### docker-compose.yml

```yaml
test-runner:
  build:
    context: ../..
    dockerfile: automation-control-plane/deploy/docker/Dockerfile.test-runner
  environment:
    - CONTROL_PLANE_URL=http://control-plane:8080
    - QUICKWIT_URL=http://quickwit:7280
    - TENANT_ID=test-tenant
    - PROJECT_ID=test-project
  depends_on:
    control-plane:
      condition: service_healthy
    quickwit:
      condition: service_healthy
  networks:
    - automation-network
  profiles:
    - test
```

### Dockerfile.test-runner

```dockerfile
FROM python:3.11-slim

WORKDIR /tests

RUN pip install --no-cache-dir requests

COPY test-linux-workflow.py /tests/
COPY test-windows-workflow.py /tests/
COPY run-all-tests.py /tests/
COPY WORKFLOW-TESTING.md /tests/

ENV CONTROL_PLANE_URL=http://control-plane:8080
ENV QUICKWIT_URL=http://quickwit:7280
ENV TENANT_ID=test-tenant
ENV PROJECT_ID=test-project
```

## Advantages of Docker-Based Testing

✅ **Clean Windows Environment**: No Python installation required on Windows
✅ **Network Isolation**: Tests run within Docker network with direct service access
✅ **Reproducible**: Same environment every time
✅ **Version Control**: Test container definition is versioned with code
✅ **CI/CD Ready**: Easy to integrate into automated pipelines
✅ **No Port Conflicts**: Uses internal Docker DNS (no localhost port mapping issues)

## Troubleshooting

### Test Runner Cannot Connect to Control Plane

```bash
# Check if control plane is healthy
docker compose ps control-plane

# View control plane logs
docker compose logs control-plane

# Check network connectivity from test-runner
docker compose run --rm test-runner ping -c 3 control-plane
```

### Services Not Ready

```bash
# Wait for all services to be healthy
docker compose up -d
sleep 10  # Give services time to start

# Check health status
docker compose ps
```

### Rebuild Test Runner After Script Changes

```bash
# Rebuild when you modify test scripts
docker compose build test-runner

# Or force rebuild without cache
docker compose build --no-cache test-runner
```

### View Test Runner Logs

```bash
# Run with verbose output
docker compose run --rm test-runner python -u test-linux-workflow.py
```

## Running Tests from Windows PowerShell

You can also run the tests from Windows PowerShell using WSL:

```powershell
# Run Linux test
wsl -d Ubuntu-22.04 bash -c "cd '/mnt/c/Users/yoges/OneDrive/Documents/My Code/Task Manager/demo/automation-control-plane/deploy' && docker compose run --rm test-runner python test-linux-workflow.py"

# Or create an alias in PowerShell profile
function Run-WorkflowTest {
    param([string]$TestScript)
    wsl -d Ubuntu-22.04 bash -c "cd '/mnt/c/Users/yoges/OneDrive/Documents/My Code/Task Manager/demo/automation-control-plane/deploy' && docker compose run --rm test-runner python $TestScript"
}

# Usage:
Run-WorkflowTest "test-linux-workflow.py"
```

## Manual Quickwit Queries from Test Runner

```bash
# Start interactive session
docker compose run --rm test-runner /bin/bash

# Inside container, query Quickwit directly
curl -X POST http://quickwit:7280/api/v1/automation-logs/search \
  -H "Content-Type: application/json" \
  -d '{"query": "agent_id:agent-linux-01", "max_hits": 50}'
```

## Test Workflow Examples

### Linux Shell Test

The test submits this workflow to the Linux agent:

```yaml
name: Linux Shell Test - Flat Outputs

tasks:
  - name: Step with outputs
    type: command
    command: echo
    args: ["Setting up authentication"]
    
  - name: Test environment info
    type: command
    command: bash
    args:
      - -c
      - |
        echo "Hostname: $(hostname)"
        echo "User: $(whoami)"
        echo "Platform: Linux"
```

### Windows PowerShell Test

The test submits this workflow to the Windows agent:

```yaml
name: Windows PowerShell Test

tasks:
  - name: PowerShell environment check
    type: powershell
    script: |
      Write-Host "PowerShell Version: $($PSVersionTable.PSVersion)"
      Write-Host "Computer: $env:COMPUTERNAME"
```

## Next Steps

1. ✅ Run Linux workflow test to verify Linux agent
2. 🔲 Start Windows agent (if needed for Windows tests)
3. 🔲 Query Quickwit to view detailed execution logs
4. 🔲 Create custom workflows for your use cases

## Cleanup

```bash
# Stop test-runner (if running)
docker compose down test-runner

# Or stop all services
docker compose down

# Remove test-runner image
docker rmi automation-control-plane-deploy-test-runner
```

## References

- [Main Testing Guide](../../WORKFLOW-TESTING.md)
- [Probe Examples](https://github.com/linyows/probe/tree/main/examples)
- [Docker Compose Documentation](https://docs.docker.com/compose/)
