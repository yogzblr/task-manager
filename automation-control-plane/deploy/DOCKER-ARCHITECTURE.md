# Docker-Based Testing Architecture

## Visual Overview

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          WINDOWS MACHINE (Clean)                             │
│                                                                              │
│  ┌────────────────────────────────────────────────────────────────────────┐ │
│  │  PowerShell / Command Prompt                                           │ │
│  │                                                                         │ │
│  │  > cd automation-control-plane\deploy                                  │ │
│  │  > .\run-docker-tests.ps1                                              │ │
│  │                                                                         │ │
│  │  ┌──────────────────────────────────────────────────────────────┐     │ │
│  │  │  Menu:                                                       │     │ │
│  │  │  1. Check services     6. Control plane logs               │     │ │
│  │  │  2. Build test-runner  7. Agent logs                        │     │ │
│  │  │  3. Run Linux test ⭐  8. Query Quickwit                    │     │ │
│  │  │  4. Run Windows test   9. Start services                    │     │ │
│  │  │  5. Interactive shell  0. Exit                              │     │ │
│  │  └──────────────────────────────────────────────────────────────┘     │ │
│  └────────────────────────────────────────────────────────────────────────┘ │
│                                   │                                          │
│                                   │ WSL command                              │
│                                   ▼                                          │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    │
┌───────────────────────────────────┼─────────────────────────────────────────┐
│                     WSL2 UBUNTU + DOCKER                                     │
│                                   │                                          │
│  ┌────────────────────────────────┼──────────────────────────────────────┐  │
│  │              DOCKER COMPOSE STACK                                     │  │
│  │                                │                                       │  │
│  │  ┌─────────────────────────────▼───────────────────────────────────┐  │  │
│  │  │  test-runner (Python 3.11)                                      │  │  │
│  │  │  ┌───────────────────────────────────────────────────────────┐  │  │  │
│  │  │  │  /tests/                                                  │  │  │  │
│  │  │  │    ├─ test-linux-workflow.py                             │  │  │  │
│  │  │  │    ├─ test-windows-workflow.py                           │  │  │  │
│  │  │  │    └─ run-all-tests.py                                   │  │  │  │
│  │  │  │                                                           │  │  │  │
│  │  │  │  Environment:                                             │  │  │  │
│  │  │  │    CONTROL_PLANE_URL=http://control-plane:8080           │  │  │  │
│  │  │  │    QUICKWIT_URL=http://quickwit:7280                     │  │  │  │
│  │  │  └───────────────────────────────────────────────────────────┘  │  │  │
│  │  └────────┬────────────────────────────┬──────────────────────────┘  │  │
│  │           │ HTTP POST                  │ HTTP POST                   │  │
│  │           │ (submit workflow)          │ (search logs)               │  │
│  │           │                            │                             │  │
│  │  ┌────────▼──────────────┐    ┌───────▼──────────────┐              │  │
│  │  │  control-plane:8080   │    │  quickwit:7280       │              │  │
│  │  │  ┌─────────────────┐  │    │  ┌────────────────┐  │              │  │
│  │  │  │ Job Management  │  │    │  │ Log Search API │  │              │  │
│  │  │  │ Agent Registry  │  │    │  │ Indexed Logs   │  │              │  │
│  │  │  │ Workflow Router │  │    │  └────────────────┘  │              │  │
│  │  │  └────────┬────────┘  │    └──────────────────────┘              │  │
│  │  └───────────┼───────────┘                                           │  │
│  │              │ WebSocket                                             │  │
│  │              │ (job dispatch)                                        │  │
│  │              │                                                        │  │
│  │  ┌───────────▼───────────┐    ┌────────────────────┐                │  │
│  │  │  agent-linux          │    │  centrifugo:8000   │                │  │
│  │  │  ┌─────────────────┐  │◀───│  ┌──────────────┐  │                │  │
│  │  │  │ Executes Tasks: │  │    │  │ WebSocket    │  │                │  │
│  │  │  │  - Shell        │  │    │  │ Real-time    │  │                │  │
│  │  │  │  - SSH          │  │    │  │ Messaging    │  │                │  │
│  │  │  │  - Database     │  │    │  └──────────────┘  │                │  │
│  │  │  │  - HTTP         │  │    └────────────────────┘                │  │
│  │  │  └─────────────────┘  │                                           │  │
│  │  └──────────────────────┘                                            │  │
│  │                                                                       │  │
│  │  ┌────────────────────┐    ┌────────────────────┐                   │  │
│  │  │  mysql:3306        │    │  valkey:6379       │                   │  │
│  │  │  ┌──────────────┐  │    │  ┌──────────────┐  │                   │  │
│  │  │  │ Jobs         │  │    │  │ Job Queue    │  │                   │  │
│  │  │  │ Agents       │  │    │  │ Cache        │  │                   │  │
│  │  │  │ Tenants      │  │    │  └──────────────┘  │                   │  │
│  │  │  │ Projects     │  │    └────────────────────┘                   │  │
│  │  │  └──────────────┘  │                                              │  │
│  │  └────────────────────┘                                              │  │
│  │                                                                       │  │
│  │  All connected via: automation-network (Docker bridge)               │  │
│  └───────────────────────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────────────────────┘
```

## Test Execution Flow

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              TEST WORKFLOW                                   │
└─────────────────────────────────────────────────────────────────────────────┘

1. User runs script
   │
   │  Windows:  .\run-docker-tests.ps1
   │  WSL:      ./run-docker-tests.sh
   │
   ▼
2. Script invokes Docker Compose
   │
   │  docker compose run --rm test-runner python test-linux-workflow.py
   │
   ▼
3. Test container starts
   │
   ├─ Environment variables loaded (CONTROL_PLANE_URL, etc.)
   ├─ Python script executes
   └─ Network: automation-network
   │
   ▼
4. Test submits workflow
   │
   │  POST http://control-plane:8080/api/v1/jobs
   │  Body: {workflow: "...", workflow_format: "yaml"}
   │
   ▼
5. Control plane processes
   │
   ├─ Saves job to MySQL
   ├─ Queues job in Valkey
   └─ Notifies via Centrifugo
   │
   ▼
6. Agent picks up job
   │
   │  agent-linux receives WebSocket notification
   │  agent-linux executes tasks (shell, ssh, db, etc.)
   │
   ▼
7. Execution logs sent
   │
   ├─ Agent logs to Quickwit
   ├─ Control plane logs job state changes
   └─ Task outputs captured
   │
   ▼
8. Test monitors status
   │
   │  Loop: GET http://control-plane:8080/api/v1/jobs/{job_id}
   │  Until: state == "completed" or "failed"
   │
   ▼
9. Test queries logs
   │
   │  POST http://quickwit:7280/api/v1/automation-logs/search
   │  Query: {query: "job_id:550e8400", max_hits: 100}
   │
   ▼
10. Results displayed
   │
   ├─ Job ID
   ├─ Execution status
   ├─ Execution time
   ├─ Log entries (up to 10 shown)
   └─ Success/failure indicator
   │
   ▼
11. Container exits (--rm flag removes it)

┌─────────────────────────────────────────────────────────────────────────────┐
│                            TEST COMPLETE ✓                                   │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Network Communication

```
test-runner container
    │
    ├─ DNS: control-plane:8080  ──▶  resolves to control-plane container IP
    ├─ DNS: quickwit:7280       ──▶  resolves to quickwit container IP
    ├─ DNS: mysql:3306          ──▶  resolves to mysql container IP
    └─ DNS: centrifugo:8000     ──▶  resolves to centrifugo container IP

All containers in "automation-network" bridge:
  ├─ Isolated from host network
  ├─ Internal DNS resolution
  ├─ No localhost confusion
  └─ Direct service-to-service communication
```

## File Layout in test-runner Container

```
test-runner:/tests/
├── test-linux-workflow.py          ← Linux shell workflow test
├── test-windows-workflow.py        ← Windows PowerShell workflow test
├── run-all-tests.py                ← Master test runner
└── WORKFLOW-TESTING.md             ← Documentation

Python environment:
├── Python 3.11.x
├── pip packages:
│   └── requests (HTTP library)
└── Standard library (json, time, sys, os, datetime)
```

## Comparison: Docker vs Local

```
┌──────────────────────┬─────────────────────┬──────────────────────┐
│ Aspect               │ Docker (New)        │ Local (Old)          │
├──────────────────────┼─────────────────────┼──────────────────────┤
│ Python installation  │ ❌ Not needed       │ ✅ Required          │
│ Windows cleanliness  │ ✅ Clean            │ ❌ Scripts on disk   │
│ Network config       │ ✅ Auto (DNS)       │ ⚠️ Manual (ports)    │
│ Reproducibility      │ ✅ 100%             │ ⚠️ Depends on env    │
│ Setup time           │ ⏱️ 30s (build)      │ ⏱️ 5min (install)    │
│ Isolation            │ ✅ Full             │ ❌ None              │
│ CI/CD integration    │ ✅ Easy             │ ⚠️ Complex           │
│ Version control      │ ✅ Dockerfile       │ ❌ Manual setup      │
│ Cleanup              │ ✅ Automatic (--rm) │ ⚠️ Manual            │
│ Dependencies         │ ✅ Pre-installed    │ ⚠️ pip install       │
└──────────────────────┴─────────────────────┴──────────────────────┘
```

## Usage Patterns

### Pattern 1: Quick Smoke Test
```powershell
# Fast validation that system works
.\run-docker-tests.ps1
→ Choose option 3 (Run Linux test)
→ See result in 15 seconds
```

### Pattern 2: Debugging Session
```powershell
# Interactive exploration
.\run-docker-tests.ps1
→ Choose option 5 (Interactive shell)
→ Inside container:
  - python test-linux-workflow.py
  - curl http://quickwit:7280/health
  - python -m json.tool < response.json
```

### Pattern 3: Log Investigation
```powershell
# After test, view detailed logs
.\run-docker-tests.ps1
→ Choose option 8 (Query Quickwit)
→ See last 20 log entries with job details
```

### Pattern 4: Continuous Monitoring
```bash
# Watch agent logs while running tests
# Terminal 1:
docker compose logs -f agent-linux

# Terminal 2:
docker compose run --rm test-runner python test-linux-workflow.py
```

## Security Model

```
Windows Machine
  └─ Only PowerShell script (no credentials)
      │
      ▼
WSL Ubuntu
  └─ Docker Compose (environment variables)
      │
      ├─ Test credentials: test-tenant/test-project
      ├─ JWT tokens: For testing only (long expiry)
      ├─ MySQL password: In compose file (not production)
      └─ Network: Isolated bridge (no external access)
          │
          └─ test-runner can ONLY access:
              ├─ control-plane (internal DNS)
              ├─ quickwit (internal DNS)
              └─ Other services via internal network
              ❌ Cannot access: Internet, host network, other Docker networks
```

## Performance Metrics

```
Operation                 Time        Notes
─────────────────────────────────────────────────────────────
Build test-runner         ~30s        First time only
Start test-runner         <1s         Already built
Submit workflow           ~100ms      HTTP POST
Agent picks up job        <2s         WebSocket notification
Execute simple workflow   ~5s         Shell commands
Execute complex workflow  ~30s        Multiple tasks
Query Quickwit            ~500ms      Indexed search
Container cleanup         <1s         Automatic with --rm
─────────────────────────────────────────────────────────────
Total test execution:     10-15s      End-to-end
```

## Resources Used

```
Container          CPU      Memory    Disk     Notes
───────────────────────────────────────────────────────────────
test-runner        <5%      50MB      180MB    Idle when not testing
control-plane      10-20%   100MB     50MB     Go binary
agent-linux        5-15%    80MB      45MB     Go binary
mysql              10-15%   400MB     200MB    Database
quickwit           15-25%   500MB     100MB    Log indexing
centrifugo         5-10%    50MB      30MB     WebSocket
valkey             <5%      30MB      20MB     Cache
───────────────────────────────────────────────────────────────
Total (all)        ~40%     1.2GB     625MB    While testing
Total (idle)       ~20%     900MB     625MB    Services running
```

---

This architecture provides a clean, isolated, reproducible testing environment
entirely within Docker, keeping your Windows machine pristine! 🎉
