# 🧪 Workflow Testing with Docker

> **Run workflow tests in Docker containers - keep your Windows machine clean!**

## 🚀 Quick Start (30 seconds)

```powershell
# From Windows PowerShell
cd "C:\Users\yoges\OneDrive\Documents\My Code\Task Manager\demo\automation-control-plane\deploy"
.\run-docker-tests.ps1
```

**Choose from the menu:**
1. ✅ Build test-runner (first time only)
2. ✅ Run Linux workflow test
3. ✅ View results

Done! Your Windows machine stays clean. 🎉

## 🎯 What This Does

This Docker-based testing solution:

- ✅ **Runs Python tests in containers** - No Python installation on Windows
- ✅ **Tests workflows end-to-end** - Linux shell & Windows PowerShell
- ✅ **Queries Quickwit logs** - Verifies execution and logging
- ✅ **Uses internal networking** - No port conflicts, automatic DNS
- ✅ **Cleans up automatically** - Containers removed after tests

## 📚 Documentation

| Document | Purpose |
|----------|---------|
| **[QUICKSTART-DOCKER-TESTS.md](QUICKSTART-DOCKER-TESTS.md)** | ⭐ Start here! Quick 5-minute guide |
| **[DOCKER-TESTING.md](DOCKER-TESTING.md)** | 📖 Complete reference guide |
| **[DOCKER-ARCHITECTURE.md](DOCKER-ARCHITECTURE.md)** | 🏗️ Architecture diagrams & details |
| **[DOCKER-TESTS-SUMMARY.md](DOCKER-TESTS-SUMMARY.md)** | 📊 Implementation summary |

## 🛠️ What You Need

### Prerequisites
- ✅ WSL2 Ubuntu with Docker
- ✅ Docker Compose services running
- ✅ Control plane + agents deployed

### Check Services
```bash
# In WSL
cd /mnt/c/Users/yoges/OneDrive/Documents/My\ Code/Task\ Manager/demo/automation-control-plane/deploy
docker compose ps
```

All services should be healthy or running.

## 🎮 Usage Methods

### Method 1: Interactive Menu (Easiest)

**Windows PowerShell:**
```powershell
.\run-docker-tests.ps1
```

**WSL Bash:**
```bash
./run-docker-tests.sh
```

### Method 2: Direct Commands

```bash
# Build test-runner image (first time)
docker compose build test-runner

# Run Linux workflow test
docker compose run --rm test-runner python test-linux-workflow.py

# Run Windows workflow test
docker compose run --rm test-runner python test-windows-workflow.py

# Interactive shell
docker compose run --rm test-runner /bin/bash
```

## 📊 What Gets Tested

### Linux Shell Workflow
Based on [probe flat-outputs.yml](https://github.com/linyows/probe/blob/main/examples/flat-outputs.yml):

```yaml
✅ Shell command execution (echo, bash)
✅ Environment variables (hostname, user, date)
✅ System information (memory, disk usage)
✅ Output variable handling
✅ Task execution on Linux agent
```

### Windows PowerShell Workflow
Custom Windows-specific tests:

```yaml
✅ PowerShell script execution
✅ System information (Get-ComputerInfo)
✅ Process listing (Get-Process)
✅ File operations (create, read, delete)
✅ Task execution on Windows agent
```

### Quickwit Integration
```yaml
✅ Log ingestion from agents
✅ Search API queries
✅ Job ID correlation
✅ Agent ID filtering
✅ Timestamp sorting
```

## 🎨 Architecture

```
Windows (Clean)
    │
    └─── run-docker-tests.ps1 (launcher)
         │
         ▼
WSL2 + Docker
    │
    ├─── test-runner (Python 3.11) ─┐
    │                                │
    ├─── control-plane (Go API)   ◀─┼─ HTTP: Submit workflow
    │                                │
    ├─── agent-linux (Executor)   ◀─┘   Executes tasks
    │
    ├─── quickwit (Logs) ◀────────────── HTTP: Query logs
    │
    └─── mysql, valkey, centrifugo (Infrastructure)
```

All containers communicate via internal Docker DNS - no localhost configuration needed!

## 🎯 Example Test Run

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

  ... (more logs) ...

============================================================
TEST COMPLETE ✓
============================================================
```

## 🔧 Troubleshooting

### Services Not Running
```bash
docker compose up -d
docker compose ps  # Check all are healthy
```

### Rebuild After Changes
```bash
docker compose build test-runner
```

### View Logs
```bash
docker compose logs control-plane
docker compose logs agent-linux
```

### Network Issues
```bash
# Test connectivity from test-runner
docker compose run --rm test-runner ping -c 3 control-plane
docker compose run --rm test-runner curl http://control-plane:8080/health
```

## 📦 Files in This Directory

```
deploy/
├── docker-compose.yml                  (✅ test-runner service added)
├── run-docker-tests.ps1                (PowerShell launcher)
├── run-docker-tests.sh                 (Bash launcher)
│
├── QUICKSTART-DOCKER-TESTS.md          (⭐ Start here!)
├── DOCKER-TESTING.md                   (Complete guide)
├── DOCKER-ARCHITECTURE.md              (Architecture details)
├── DOCKER-TESTS-SUMMARY.md             (Implementation notes)
│
└── docker/
    └── Dockerfile.test-runner          (Python test container)
```

## ✨ Key Features

| Feature | Description |
|---------|-------------|
| **Clean Windows** | No Python installation needed |
| **Auto DNS** | Uses Docker DNS (control-plane:8080) |
| **Auto Cleanup** | Containers removed with `--rm` flag |
| **Reproducible** | Same environment every time |
| **Fast** | Tests complete in 10-15 seconds |
| **CI/CD Ready** | Easy to automate |

## 🔐 Security

- Uses test credentials only (`test-tenant`, `test-project`)
- Network isolated (Docker bridge)
- No external access from containers
- JWT tokens for testing only (long expiry)

## 📈 Performance

```
Operation                 Time
─────────────────────────────────
Build test-runner         ~30s (first time)
Run Linux test           10-15s
Run Windows test         15-20s
Query Quickwit           ~500ms
Container cleanup        <1s (automatic)
```

## 🎓 Learning Resources

- [Probe Examples](https://github.com/linyows/probe/tree/main/examples) - Official workflow examples
- [Docker Compose Docs](https://docs.docker.com/compose/) - Docker reference
- [Quickwit Search API](https://quickwit.io/docs/reference/rest-api) - Log queries

## 🚦 Next Steps

1. ✅ Run the menu script: `.\run-docker-tests.ps1`
2. ✅ Build test-runner (option 2)
3. ✅ Run Linux test (option 3)
4. ✅ Query Quickwit (option 8) for detailed logs
5. 🔲 Create custom workflows for your use cases

## 🆘 Support

If you encounter issues:

1. Check service health: `docker compose ps`
2. View logs: `docker compose logs <service>`
3. Restart services: `docker compose restart`
4. Rebuild test-runner: `docker compose build test-runner`

## 🎉 Success Criteria

Your test is successful when you see:

```
✓ Control plane is accessible
✓ Workflow submitted successfully!
✓ Job completed successfully!
✓ Found X log entries
```

All happening inside Docker containers with zero impact on your Windows environment! 🚀

---

**Ready to test?**

```powershell
.\run-docker-tests.ps1
```

**Select option 2 (build), then option 3 (test). Done!** ✨
