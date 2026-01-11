# Workflow Testing Guide

This directory contains Python test scripts for testing probe workflows on both Linux and Windows agents.

## Test Scripts

### 1. `test-linux-workflow.py`
Tests Linux shell commands and system information gathering.

**Features:**
- Submits shell workflow to control plane
- Monitors job execution
- Queries Quickwit for logs
- Based on [probe flat-outputs example](https://github.com/linyows/probe/blob/main/examples/flat-outputs.yml)

**Workflow Tasks:**
- Authentication setup simulation
- Environment information (hostname, user, date)
- Output variable testing
- System information (memory, disk usage)

### 2. `test-windows-workflow.py`
Tests Windows PowerShell commands and system operations.

**Features:**
- Submits PowerShell workflow to control plane
- Monitors job execution
- Queries Quickwit for logs
- Tests Windows-specific features

**Workflow Tasks:**
- PowerShell environment check
- System information via Get-ComputerInfo
- Process listing by CPU usage
- File operations testing
- Variable output testing

## Prerequisites

### Required Services
All services must be running:
- ✅ Control Plane (http://localhost:8081)
- ✅ MySQL (localhost:3306)
- ✅ Valkey (localhost:6379)
- ✅ Centrifugo (localhost:8000)
- ✅ Quickwit (localhost:7280)
- ✅ Linux Agent (Docker)
- ⏳ Windows Agent (native binary) - for Windows tests

### Start Docker Services
```bash
# In WSL Ubuntu
cd /mnt/c/Users/yoges/OneDrive/Documents/My\ Code/Task\ Manager/demo/automation-control-plane/deploy
docker compose up -d
docker compose ps  # Verify all services are running
```

### Install Python Dependencies
```bash
pip install requests
```

## Running the Tests

### Linux Shell Test
```bash
# Test Linux agent with shell commands
python test-linux-workflow.py
```

Expected output:
```
✓ Control plane is accessible
✓ Workflow submitted successfully!
  Job ID: <uuid>
✓ Job completed successfully!
✓ Found X log entries
```

### Windows PowerShell Test
```bash
# IMPORTANT: Start Windows agent first!
# See: automation-agent/WINDOWS-AGENT-SUCCESS.md

# Then run the test
python test-windows-workflow.py
```

**Note**: Windows agent must be running with valid JWT token before running this test.

## Quickwit Log Search

Both scripts automatically search Quickwit for execution logs. The scripts query:
- Job execution logs
- Agent activity logs
- Task output logs

### Manual Quickwit Queries

You can also query Quickwit directly:

```bash
# Search for specific job
curl -X POST http://localhost:7280/api/v1/automation-logs/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "job_id:YOUR_JOB_ID",
    "max_hits": 100
  }'

# Search for agent logs
curl -X POST http://localhost:7280/api/v1/automation-logs/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "agent_id:agent-linux-01 OR agent_id:agent-windows-01",
    "max_hits": 50
  }'

# Search by time range
curl -X POST http://localhost:7280/api/v1/automation-logs/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "*",
    "start_timestamp": 1704909600,
    "end_timestamp": 1704996000,
    "max_hits": 100
  }'
```

## Test Workflow Examples

### Linux Shell Workflow (test-linux-workflow.py)
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

### Windows PowerShell Workflow (test-windows-workflow.py)
```yaml
name: Windows PowerShell Test

tasks:
  - name: PowerShell environment check
    type: powershell
    script: |
      Write-Host "PowerShell Version: $($PSVersionTable.PSVersion)"
      Write-Host "Computer: $env:COMPUTERNAME"
      
  - name: System information
    type: powershell
    script: |
      Get-ComputerInfo | Select-Object CsName, WindowsVersion
```

## Troubleshooting

### Control Plane Not Accessible
```bash
# Check if Docker services are running
wsl -d Ubuntu-22.04 bash -c "cd /mnt/c/Users/yoges/OneDrive/Documents/My\ Code/Task\ Manager/demo/automation-control-plane/deploy && docker compose ps"

# Check control plane logs
wsl -d Ubuntu-22.04 bash -c "cd /mnt/c/Users/yoges/OneDrive/Documents/My\ Code/Task\ Manager/demo/automation-control-plane/deploy && docker compose logs control-plane"
```

### Job Not Executing
```bash
# Check Linux agent logs
wsl -d Ubuntu-22.04 bash -c "cd /mnt/c/Users/yoges/OneDrive/Documents/My\ Code/Task\ Manager/demo/automation-control-plane/deploy && docker compose logs agent-linux"

# Check database for jobs
wsl -d Ubuntu-22.04 bash -c "cd /mnt/c/Users/yoges/OneDrive/Documents/My\ Code/Task\ Manager/demo/automation-control-plane/deploy && docker compose exec mysql mysql -u automation -ppassword automation -e 'SELECT job_id, state FROM jobs ORDER BY created_at DESC LIMIT 5;'"
```

### Quickwit Not Returning Results
```bash
# Check if Quickwit is running
curl http://localhost:7280/health

# Check Quickwit logs
wsl -d Ubuntu-22.04 bash -c "cd /mnt/c/Users/yoges/OneDrive/Documents/My\ Code/Task\ Manager/demo/automation-control-plane/deploy && docker compose logs quickwit"

# Note: Quickwit may need additional configuration for log indexing
```

### Windows Agent Not Running
```bash
# Check if process is running
Get-Process | Where-Object {$_.ProcessName -like '*automation-agent*'}

# Start Windows agent
cd "C:\Users\yoges\OneDrive\Documents\My Code\Task Manager\demo\automation-agent"
# Set environment variables and run
.\automation-agent-windows.exe
```

## Expected Results

### Successful Linux Test
```
✓ Control plane is accessible
✓ Workflow submitted successfully!
  Job ID: 550e8400-e29b-41d4-a716-446655440000
  Status: pending (0s elapsed)
  Status: executing (2s elapsed)
  Status: completed (5s elapsed)
✓ Job completed successfully!
✓ Found 15 log entries
```

### Successful Windows Test
```
✓ Control plane is accessible
Is the Windows agent running? (y/n): y
✓ Workflow submitted successfully!
  Job ID: 660e8400-e29b-41d4-a716-446655440001
  Status: executing (2s elapsed)
  Status: completed (8s elapsed)
✓ Job completed successfully!
✓ Found 22 log entries
```

## Additional Resources

- [Probe Examples](https://github.com/linyows/probe/tree/main/examples)
- [Control Plane API Documentation](../automation-control-plane/openapi.yaml)
- [Quickwit Search API](https://quickwit.io/docs/reference/rest-api)

## Next Steps

1. **Run Linux test** to verify basic workflow execution
2. **Start Windows agent** with proper JWT token
3. **Run Windows test** to verify PowerShell execution
4. **Query Quickwit** to view detailed execution logs
5. **Create custom workflows** for your specific use cases
