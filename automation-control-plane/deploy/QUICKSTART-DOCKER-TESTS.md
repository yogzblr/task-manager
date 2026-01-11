# Quick Start: Docker-Based Workflow Testing

This guide shows you how to run workflow tests using Docker containers, keeping your Windows machine clean.

## What You Get

✅ **No Python Installation Needed**: Tests run in a Python 3.11 container
✅ **No Port Conflicts**: Uses internal Docker networking
✅ **Clean Windows Environment**: No test files on your Windows machine
✅ **Reproducible Tests**: Same environment every time
✅ **Easy to Use**: Simple menu-driven scripts

## Prerequisites

1. Docker services must be running in WSL Ubuntu
2. All services should be healthy (control-plane, mysql, quickwit, etc.)

## Method 1: PowerShell (Recommended for Windows)

From your Windows PowerShell:

```powershell
cd "C:\Users\yoges\OneDrive\Documents\My Code\Task Manager\demo\automation-control-plane\deploy"
.\run-docker-tests.ps1
```

**Menu Options:**
```
1. Check Docker services status
2. Build test-runner image
3. Run Linux workflow test ⭐
4. Run Windows workflow test (requires Windows agent)
5. Interactive test-runner shell
6. View control plane logs
7. View Linux agent logs
8. Query Quickwit for recent logs
9. Start all services
0. Exit
```

## Method 2: Bash Script (WSL)

From WSL Ubuntu terminal:

```bash
cd /mnt/c/Users/yoges/OneDrive/Documents/My\ Code/Task\ Manager/demo/automation-control-plane/deploy
./run-docker-tests.sh
```

Same menu as PowerShell version.

## Method 3: Direct Docker Compose Commands

### Build Test Runner
```bash
# In WSL
cd /mnt/c/Users/yoges/OneDrive/Documents/My\ Code/Task\ Manager/demo/automation-control-plane/deploy
docker compose build test-runner
```

### Run Linux Test
```bash
docker compose run --rm test-runner python test-linux-workflow.py
```

### Run Windows Test
```bash
docker compose run --rm test-runner python test-windows-workflow.py
```

### Interactive Shell
```bash
docker compose run --rm test-runner /bin/bash
# Inside container:
python test-linux-workflow.py
ls -la
exit
```

## First Time Setup

### Step 1: Start Services
```bash
# In WSL
cd /mnt/c/Users/yoges/OneDrive/Documents/My\ Code/Task\ Manager/demo/automation-control-plane/deploy
docker compose up -d
```

### Step 2: Wait for Services
```bash
# Check status (all should be healthy or running)
docker compose ps
```

Expected output:
```
NAME                          STATUS
mysql                         Up (healthy)
valkey                        Up
centrifugo                    Up (healthy)
quickwit                      Up (healthy)
control-plane                 Up (healthy)
agent-linux                   Up
```

### Step 3: Build Test Runner
```bash
docker compose build test-runner
```

### Step 4: Run Your First Test
```bash
docker compose run --rm test-runner python test-linux-workflow.py
```

## Expected Test Output

```
╔════════════════════════════════════════════════════════════╗
║     Linux Shell Workflow Test - Probe Integration         ║
╚════════════════════════════════════════════════════════════╝

This script will:
1. Submit a Linux shell workflow to the control plane
2. Monitor job execution status
3. Query Quickwit for execution logs
4. Display results

Prerequisites:
- Control plane running at: http://control-plane:8080
- Linux agent running (Docker)
- Quickwit running at: http://quickwit:7280

✓ Control plane is accessible

============================================================
Submitting Linux Shell Workflow
============================================================
✓ Workflow submitted successfully!
  Job ID: 550e8400-e29b-41d4-a716-446655440000

============================================================
Monitoring Job Execution
============================================================
  Status: pending (0s elapsed)
  Status: executing (2s elapsed)
  Status: completed (5s elapsed)
✓ Job completed successfully!

============================================================
Searching Quickwit for Execution Logs
============================================================
✓ Found 15 log entries

  [1] 2026-01-10T10:30:00Z [INFO]
      Job 550e8400 started on agent-linux-01...
  [2] 2026-01-10T10:30:01Z [INFO]
      Task 'Step with outputs' executing...
  ...

============================================================
TEST COMPLETE
============================================================
```

## Troubleshooting

### Services Not Running
```bash
# Check service status
docker compose ps

# Start services
docker compose up -d

# View logs
docker compose logs control-plane
```

### Test Runner Build Failed
```bash
# Rebuild without cache
docker compose build --no-cache test-runner

# Check if test files exist
ls -la ../../test-*.py
```

### Cannot Connect to Control Plane
```bash
# Check if control plane is healthy
docker compose ps control-plane

# Test network connectivity
docker compose run --rm test-runner ping -c 3 control-plane
```

### Rebuild After Script Changes
```bash
# When you modify test scripts, rebuild:
docker compose build test-runner
```

## What Gets Tested

### Linux Workflow Test
- ✅ Shell command execution (`echo`, `bash`)
- ✅ Environment variables (hostname, user, date)
- ✅ System information (memory, disk usage)
- ✅ Output variable handling
- ✅ Based on [probe flat-outputs.yml](https://github.com/linyows/probe/blob/main/examples/flat-outputs.yml)

### Windows Workflow Test
- ✅ PowerShell execution
- ✅ Get-ComputerInfo
- ✅ Process listing
- ✅ File operations
- ⚠️ Requires Windows agent running

## Architecture

```
Your Windows Machine
├─ No Python needed
├─ No test scripts needed
└─ Just Docker Desktop + WSL2

WSL2 Ubuntu (Docker)
├─ test-runner container (Python 3.11)
│  ├─ test-linux-workflow.py
│  ├─ test-windows-workflow.py
│  └─ requests library
├─ control-plane container
├─ agent-linux container
├─ quickwit container
└─ All services networked together
```

## Advantages

| Feature | Docker-Based | Local Python |
|---------|--------------|--------------|
| Windows Python needed | ❌ No | ✅ Yes |
| Port configuration | ❌ Auto | ⚠️ Manual |
| Network isolation | ✅ Yes | ❌ No |
| Clean Windows | ✅ Yes | ❌ No |
| CI/CD ready | ✅ Yes | ⚠️ Maybe |
| Reproducible | ✅ Always | ⚠️ Sometimes |

## Next Steps

1. ✅ Run `run-docker-tests.ps1` (PowerShell) or `run-docker-tests.sh` (WSL)
2. ✅ Choose option `2` to build test-runner
3. ✅ Choose option `3` to run Linux workflow test
4. ✅ View results in Quickwit (option `8`)
5. 🔲 Create custom workflows for your use cases

## Documentation

- 📖 [Detailed Docker Testing Guide](DOCKER-TESTING.md)
- 📖 [Main Workflow Testing Guide](../../WORKFLOW-TESTING.md)
- 📖 [WSL Build Success](WSL-BUILD-SUCCESS.md)

## Clean Up

```bash
# Stop test-runner (it auto-removes with --rm flag)
# No cleanup needed for test runs

# To remove test-runner image:
docker rmi automation-control-plane-deploy-test-runner

# To stop all services:
docker compose down
```

---

**Ready to test?** Run the script now! 🚀

```powershell
.\run-docker-tests.ps1
```
