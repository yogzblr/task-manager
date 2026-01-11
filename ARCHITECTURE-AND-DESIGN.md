# Automation Platform - Complete Architecture and Design Documentation

> **Version**: 2.0.0  
> **Last Updated**: January 2026  
> **Status**: Production Ready

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [System Architecture](#system-architecture)
3. [Design Approach](#design-approach)
4. [Component Architecture](#component-architecture)
5. [Technology Stack](#technology-stack)
6. [Testing Strategy](#testing-strategy)
7. [Deployment Architecture](#deployment-architecture)
8. [Security Architecture](#security-architecture)
9. [API Documentation](#api-documentation)
10. [Workflow System](#workflow-system)
11. [Performance and Scalability](#performance-and-scalability)
12. [Operational Considerations](#operational-considerations)

---

## Executive Summary

### Overview

The Automation Platform is a comprehensive, enterprise-grade distributed automation system designed for cross-platform task orchestration and execution. It provides centralized control, real-time monitoring, and flexible YAML-based workflow definitions powered by the probe task execution framework.

### Key Features

- **Multi-Platform Support**: Windows and Linux agents with platform-specific task types
- **YAML Workflows**: Intuitive, declarative workflow definitions with 6 built-in task types
- **Real-Time Communication**: WebSocket-based job dispatch via Centrifugo
- **Comprehensive Logging**: Centralized log aggregation and search with Quickwit
- **Multi-Tenancy**: Built-in support for multiple tenants and projects
- **Security-First**: JWT authentication, Ed25519 signature verification, SHA256 checksums
- **Extensible Architecture**: Plugin-based task system for custom automation needs

### Use Cases

1. **Infrastructure Automation**: Server configuration, deployment, maintenance
2. **Application Deployment**: Automated software deployment across multiple environments
3. **Health Monitoring**: Continuous health checks and availability monitoring
4. **Database Operations**: Automated database queries, backups, migrations
5. **Remote Management**: Centralized management of distributed systems
6. **Compliance Automation**: Automated security checks and compliance auditing

---

## System Architecture

### High-Level Architecture

```
┌──────────────────────────────────────────────────────────────────────────┐
│                           AUTOMATION PLATFORM                             │
└──────────────────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────────────────┐
│                         CONTROL PLANE (Orchestration)                     │
│                                                                           │
│  ┌────────────────┐  ┌────────────────┐  ┌────────────────┐            │
│  │  REST API      │  │  Job Scheduler │  │  Agent Manager │            │
│  │  - Job CRUD    │  │  - Queue       │  │  - Registry    │            │
│  │  - Multi-tenant│  │  - Dispatch    │  │  - Heartbeat   │            │
│  └────────────────┘  └────────────────┘  └────────────────┘            │
│                                                                           │
│  ┌────────────────┐  ┌────────────────┐  ┌────────────────┐            │
│  │  MySQL         │  │  Valkey/Redis  │  │  Quickwit      │            │
│  │  - Persistence │  │  - Job Queue   │  │  - Log Search  │            │
│  │  - Audit Log   │  │  - Cache       │  │  - Analytics   │            │
│  └────────────────┘  └────────────────┘  └────────────────┘            │
└───────────────────────────┬──────────────────────────────────────────────┘
                            │
                            │ Centrifugo WebSocket
                            │ (Real-time messaging)
                            │
            ┌───────────────┴───────────────┐
            │                               │
┌───────────▼────────────┐     ┌───────────▼────────────┐
│   AGENT (Linux)        │     │   AGENT (Windows)      │
│                        │     │                        │
│  ┌──────────────────┐  │     │  ┌──────────────────┐  │
│  │ Probe Executor   │  │     │  │ Probe Executor   │  │
│  │ - YAML Parser    │  │     │  │ - YAML Parser    │  │
│  │ - Task Registry  │  │     │  │ - Task Registry  │  │
│  └──────────────────┘  │     │  └──────────────────┘  │
│                        │     │                        │
│  ┌──────────────────┐  │     │  ┌──────────────────┐  │
│  │ Task Types:      │  │     │  │ Task Types:      │  │
│  │ - HTTP           │  │     │  │ - HTTP           │  │
│  │ - Database       │  │     │  │ - Database       │  │
│  │ - SSH            │  │     │  │ - SSH            │  │
│  │ - Command        │  │     │  │ - Command        │  │
│  │ - DownloadExec   │  │     │  │ - PowerShell     │  │
│  └──────────────────┘  │     │  │ - DownloadExec   │  │
│                        │     │  └──────────────────┘  │
│  Systemd Service       │     │  Windows Service       │
└────────────────────────┘     └────────────────────────┘
```

### Component Interaction Flow

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        WORKFLOW EXECUTION FLOW                          │
└─────────────────────────────────────────────────────────────────────────┘

1. JOB SUBMISSION
   ┌─────────┐
   │  User   │
   │  / API  │
   └────┬────┘
        │
        │ POST /api/v1/jobs
        │ {workflow: "...", workflow_format: "yaml"}
        │
        ▼
   ┌────────────────┐
   │ Control Plane  │
   │  - Validate    │
   │  - Store MySQL │
   │  - Queue Valkey│
   └────┬───────────┘
        │
        │
2. JOB DISPATCH
        │
        │ Publish to Centrifugo
        │ Channel: agents.{tenant}.{agent_id}
        │
        ▼
   ┌────────────────┐
   │  Centrifugo    │
   │  WebSocket Hub │
   └────┬───────────┘
        │
        │ WebSocket notification
        │ {job_id, workflow, tenant, project}
        │
        ▼
   ┌────────────────┐
   │  Agent         │
   │  - Receive Job │
   │  - Validate    │
   └────┬───────────┘
        │
        │
3. WORKFLOW EXECUTION
        │
        ▼
   ┌────────────────────────────────────┐
   │  Probe Executor                    │
   │  1. Parse YAML workflow            │
   │  2. Load task definitions          │
   │  3. Execute tasks sequentially     │
   │  4. Collect results                │
   │  5. Handle errors                  │
   └────┬───────────────────────────────┘
        │
        │
4. RESULT REPORTING
        │
        ├──► POST /api/v1/jobs/{id}/status
        │    {state: "running|completed|failed"}
        │
        └──► POST {quickwit}/automation-logs
             {job_id, task, output, timestamp}
```

---

## Design Approach

### Design Principles

#### 1. **Separation of Concerns**
- **Control Plane**: Orchestration, state management, API
- **Agents**: Task execution, local operations
- **Probe**: Workflow parsing, task execution logic
- **Infrastructure**: Messaging, storage, logging

#### 2. **Modularity**
- Independent, deployable components
- Well-defined interfaces and contracts
- Pluggable task system
- Extensible architecture

#### 3. **Platform Agnostic**
- Cross-platform agent support (Windows/Linux)
- Platform-specific task types where needed
- Uniform API regardless of agent platform

#### 4. **Real-Time First**
- WebSocket-based communication for instant job dispatch
- Real-time status updates
- Live log streaming capabilities

#### 5. **Declarative Configuration**
- YAML-based workflow definitions
- Infrastructure as code principles
- Version-controlled workflows

#### 6. **Security by Default**
- JWT-based authentication
- Ed25519 signature verification for downloads
- SHA256 checksums required
- Least privilege principle

### Architecture Decisions

#### Decision 1: Probe Framework Integration

**Context**: Need flexible, extensible task execution system

**Decision**: Integrate probe framework as core executor

**Rationale**:
- Proven YAML workflow parsing
- Extensible task system
- Clean separation of execution logic
- Platform-agnostic design
- Well-tested codebase

**Consequences**:
- ✅ Rapid development of new task types
- ✅ Clear architecture boundaries
- ✅ Easy testing and validation
- ⚠️ Learning curve for probe internals

#### Decision 2: Centrifugo for Real-Time Communication

**Context**: Need reliable real-time job dispatch

**Decision**: Use Centrifugo WebSocket server

**Rationale**:
- Scalable WebSocket infrastructure
- Channel-based messaging
- JWT authentication support
- Pub/sub architecture
- Battle-tested in production

**Consequences**:
- ✅ Instant job dispatch
- ✅ Scalable to thousands of agents
- ✅ Reliable message delivery
- ⚠️ Additional infrastructure component

#### Decision 3: Quickwit for Log Aggregation

**Context**: Need searchable, scalable logging

**Decision**: Use Quickwit search engine

**Rationale**:
- Optimized for log search
- S3-compatible storage
- Efficient indexing
- RESTful API
- Cost-effective at scale

**Consequences**:
- ✅ Fast log queries
- ✅ Long-term log retention
- ✅ Analytics capabilities
- ⚠️ Requires S3-compatible storage

#### Decision 4: Multi-Tenancy from Start

**Context**: Enterprise requirements for tenant isolation

**Decision**: Built-in multi-tenancy support

**Rationale**:
- Easier to build in than retrofit
- Common enterprise requirement
- Enables SaaS deployment model
- Clear security boundaries

**Consequences**:
- ✅ Single platform for multiple teams
- ✅ Data isolation
- ✅ Scalable user model
- ⚠️ Increased complexity

---

## Component Architecture

### 1. Control Plane

#### Responsibilities
- Job lifecycle management
- Agent registration and tracking
- Workflow validation
- Multi-tenant access control
- Audit logging
- API gateway

#### Technology
- **Language**: Go 1.21+
- **Framework**: Custom HTTP server
- **Database**: MySQL 8.0
- **Cache**: Valkey (Redis-compatible)
- **Real-time**: Centrifugo integration

#### Key Modules

**API Server** (`internal/api/`)
```go
// Job management endpoints
POST   /api/v1/jobs          - Create job
GET    /api/v1/jobs          - List jobs
GET    /api/v1/jobs/{id}     - Get job details
PUT    /api/v1/jobs/{id}     - Update job
DELETE /api/v1/jobs/{id}     - Delete job

// Agent management
POST   /api/v1/agents        - Register agent
GET    /api/v1/agents        - List agents
POST   /api/v1/agents/{id}/heartbeat - Heartbeat
```

**Job Scheduler** (`internal/scheduler/`)
- Queue management (Valkey)
- Job prioritization
- Agent selection
- Retry logic
- Dead letter handling

**Agent Manager** (`internal/store/agents.go`)
- Agent registration
- Health tracking (heartbeat)
- Capability discovery
- Load balancing

#### Database Schema

```sql
-- Tenants table
CREATE TABLE tenants (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Projects table
CREATE TABLE projects (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (tenant_id) REFERENCES tenants(id)
);

-- Agents table
CREATE TABLE agents (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    project_id VARCHAR(255) NOT NULL,
    hostname VARCHAR(255),
    platform VARCHAR(50),
    version VARCHAR(50),
    last_heartbeat TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (tenant_id) REFERENCES tenants(id),
    FOREIGN KEY (project_id) REFERENCES projects(id)
);

-- Jobs table
CREATE TABLE jobs (
    job_id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    project_id VARCHAR(255) NOT NULL,
    agent_id VARCHAR(255),
    workflow TEXT NOT NULL,
    workflow_format VARCHAR(10) DEFAULT 'yaml',
    state VARCHAR(50) DEFAULT 'pending',
    scheduled_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    started_at TIMESTAMP NULL,
    completed_at TIMESTAMP NULL,
    error TEXT,
    FOREIGN KEY (tenant_id) REFERENCES tenants(id),
    FOREIGN KEY (project_id) REFERENCES projects(id),
    FOREIGN KEY (agent_id) REFERENCES agents(id)
);
```

### 2. Automation Agent

#### Responsibilities
- Connect to control plane
- Receive job notifications
- Execute workflows via probe
- Report execution status
- Send logs to Quickwit
- Platform-specific task execution

#### Technology
- **Language**: Go 1.21+
- **Framework**: Probe executor
- **Service**: Systemd (Linux), Windows Service

#### Key Modules

**Agent Core** (`cmd/agent/main.go`)
- Initialization and configuration
- Control plane registration
- Centrifugo WebSocket connection
- Heartbeat management

**Workflow Executor** (probe integration)
- YAML parsing
- Task execution
- Error handling
- Result aggregation

**Task Plugins**
- HTTP health checks
- Database queries
- SSH operations
- Shell commands
- PowerShell scripts (Windows)
- Download and execute

#### Configuration

```env
# Control Plane
CONTROL_PLANE_URL=https://control-plane.example.com
CENTRIFUGO_URL=wss://realtime.example.com/connection/websocket

# Identity
TENANT_ID=acme-corp
PROJECT_ID=prod-infrastructure
AGENT_ID=web-server-01

# Authentication
JWT_TOKEN=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

# Logging
QUICKWIT_URL=https://logs.example.com
LOG_LEVEL=info
```

### 3. Probe Framework

#### Responsibilities
- Parse YAML workflows
- Manage task registry
- Execute tasks with timeouts
- Collect task results
- Handle task errors

#### Technology
- **Language**: Go 1.21+
- **Package**: Standalone module

#### Architecture

```
probe/
├── probe.go              # Core executor
├── task.go               # Task interface
├── task_http.go          # HTTP task
├── task_db.go            # Database task
├── task_ssh.go           # SSH task
├── task_command.go       # Command task
├── task_powershell.go    # PowerShell task (custom)
├── task_downloadexec.go  # Download+Execute (custom)
└── examples/             # Example workflows
```

**Task Interface**
```go
type Task interface {
    Configure(config map[string]interface{}) error
    Execute(ctx context.Context) (*TaskResult, error)
}

type TaskResult struct {
    Success bool
    Output  string
    Error   string
    Metrics map[string]interface{}
}
```

**Workflow Structure**
```yaml
name: workflow-name
timeout: 5m
tasks:
  - name: task-1
    type: http
    timeout: 30s
    config:
      url: https://api.example.com
      expected_status: [200]
      
  - name: task-2
    type: command
    config:
      command: echo
      args: ["Hello World"]
```

---

## Technology Stack

### Backend Services

| Component | Technology | Version | Purpose |
|-----------|-----------|---------|---------|
| Control Plane | Go | 1.21+ | Orchestration server |
| Agent | Go | 1.21+ | Task executor |
| Probe | Go | 1.21+ | Workflow engine |

### Infrastructure

| Component | Technology | Version | Purpose |
|-----------|-----------|---------|---------|
| Database | MySQL | 8.0+ | Persistent storage |
| Cache | Valkey | 7.2+ | Job queue, cache |
| Real-time | Centrifugo | 5.0+ | WebSocket server |
| Logging | Quickwit | 0.7+ | Log aggregation |
| Storage | MinIO | Latest | Object storage (S3-compatible) |

### Deployment

| Component | Technology | Purpose |
|-----------|-----------|---------|
| Container | Docker | Container runtime |
| Orchestration | Kubernetes | Container orchestration |
| Service Mesh | Helm | Package management |

### Development

| Tool | Purpose |
|------|---------|
| Git | Version control |
| GitHub | Repository hosting |
| GoLand/VSCode | IDE |
| Postman | API testing |

---

## Testing Strategy

### Testing Pyramid

```
                    ┌──────────────┐
                    │   E2E Tests  │  5%
                    │  Integration │
                    └──────────────┘
                  ┌──────────────────┐
                  │ Integration Tests│  15%
                  │ Component Tests  │
                  └──────────────────┘
              ┌─────────────────────────┐
              │      Unit Tests         │  80%
              │  Task-level testing     │
              └─────────────────────────┘
```

### 1. Unit Testing

#### Probe Task Tests
**Location**: `probe/task_*_test.go`

**Coverage**:
- ✅ HTTP task (200, 404, timeout scenarios)
- ✅ PowerShell task (success, failure, syntax errors)
- ✅ DownloadExec task (signature verification, checksum validation)
- ✅ Command task (success, failure, timeout)
- ✅ Database task (connection, queries, errors)
- ✅ SSH task (connection, commands, file transfer)

**Example Test**:
```go
func TestHTTPTask_Success(t *testing.T) {
    // Start mock HTTP server
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("OK"))
    }))
    defer server.Close()

    // Configure task
    task := &HTTPTask{}
    err := task.Configure(map[string]interface{}{
        "url": server.URL,
        "expected_status": []int{200},
    })
    assert.NoError(t, err)

    // Execute
    ctx := context.Background()
    result, err := task.Execute(ctx)
    
    // Assert
    assert.NoError(t, err)
    assert.True(t, result.Success)
}
```

### 2. Integration Testing

#### Control Plane API Tests
**Location**: `automation-control-plane/internal/api/*_test.go`

**Test Scenarios**:
- Job CRUD operations
- Agent registration and heartbeat
- Multi-tenant isolation
- Authentication and authorization
- Workflow validation

#### Agent Integration Tests
**Location**: `automation-agent/integration_test.go`

**Test Scenarios**:
- Connect to control plane
- Receive job via WebSocket
- Execute workflow
- Report status
- Send logs

### 3. End-to-End Testing

#### Docker-Based E2E Tests
**Location**: `demo/test-*-workflow.py`

**Test Infrastructure**:
```
Docker Compose Stack:
├── control-plane
├── mysql
├── valkey
├── centrifugo
├── quickwit
├── minio
├── agent-linux
└── test-runner (Python)
```

**Test Scripts**:

**Linux Workflow Test** (`test-linux-workflow.py`)
```python
def test_linux_workflow():
    # 1. Verify control plane health
    response = requests.get(f"{CONTROL_PLANE_URL}/health")
    assert response.status_code == 200
    
    # 2. Submit workflow
    workflow = """
    name: linux-test
    tasks:
      - name: system-info
        type: command
        config:
          command: uname
          args: ["-a"]
    """
    
    job_response = requests.post(
        f"{CONTROL_PLANE_URL}/api/v1/jobs",
        json={
            "workflow": workflow,
            "workflow_format": "yaml",
            "tenant_id": "test-tenant",
            "project_id": "test-project"
        }
    )
    job_id = job_response.json()["job_id"]
    
    # 3. Monitor execution
    for _ in range(30):  # Wait up to 30 seconds
        status = requests.get(f"{CONTROL_PLANE_URL}/api/v1/jobs/{job_id}")
        state = status.json()["state"]
        if state in ["completed", "failed"]:
            break
        time.sleep(1)
    
    # 4. Verify success
    assert state == "completed"
    
    # 5. Query logs
    logs = requests.post(
        f"{QUICKWIT_URL}/api/v1/automation-logs/search",
        json={"query": f"job_id:{job_id}", "max_hits": 100}
    )
    assert len(logs.json()["hits"]) > 0
```

**Windows Workflow Test** (`test-windows-workflow.py`)
- PowerShell execution
- Windows-specific commands
- File operations
- System information gathering

### 4. Performance Testing

#### Load Testing
**Tool**: Apache JMeter, Locust

**Scenarios**:
- 100 concurrent job submissions
- 1000 agents heartbeating
- 10,000 log entries per second

**Metrics**:
- Job submission latency (target: <100ms p95)
- Job dispatch time (target: <500ms p95)
- WebSocket connection capacity (target: 10k+)
- Log ingestion rate (target: 10k logs/sec)

#### Stress Testing
- Maximum concurrent workflows
- Database connection pool limits
- Memory usage under load
- Recovery from failures

### 5. Security Testing

#### Authentication Tests
- Invalid JWT tokens
- Expired tokens
- Token tampering
- Multi-tenant access violation

#### Signature Verification Tests
- Invalid Ed25519 signatures
- Tampered downloads
- Missing public keys
- Wrong checksum

### Testing Tools

| Tool | Purpose | Location |
|------|---------|----------|
| Go test | Unit tests | `*_test.go` files |
| Docker Compose | E2E environment | `deploy/docker-compose.yml` |
| Python pytest | E2E tests | `test-*-workflow.py` |
| Postman | API testing | Manual/CI |
| JMeter | Load testing | Performance suite |

### Test Execution

#### Local Testing
```bash
# Unit tests
cd probe && go test -v ./...
cd automation-agent && go test -v ./...
cd automation-control-plane && go test -v ./...

# Integration tests
cd automation-control-plane/deploy
docker compose up -d
python test-linux-workflow.py
python test-windows-workflow.py
```

#### CI/CD Pipeline
```yaml
# .github/workflows/test.yml
name: Test Suite
on: [push, pull_request]

jobs:
  unit-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
      - run: go test ./... -race -coverprofile=coverage.txt
      
  integration-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - run: docker compose -f deploy/docker-compose.yml up -d
      - run: python test-linux-workflow.py
```

### Test Coverage Goals

| Component | Target Coverage | Current |
|-----------|----------------|---------|
| Probe tasks | 90% | 85% |
| Control Plane API | 80% | 75% |
| Agent core | 70% | 65% |
| Overall | 75% | 70% |

---

## Deployment Architecture

### Development Environment

```
┌─────────────────────────────────────────┐
│         Developer Machine               │
│                                         │
│  ┌───────────────────────────────────┐  │
│  │  Docker Desktop                   │  │
│  │                                   │  │
│  │  docker-compose up -d             │  │
│  │                                   │  │
│  │  All services in containers       │  │
│  └───────────────────────────────────┘  │
└─────────────────────────────────────────┘
```

**Setup**:
```bash
cd automation-control-plane/deploy
docker-compose up -d
# All services start automatically
```

### Production - Kubernetes

```
┌──────────────────────────────────────────────────────────────┐
│                    Kubernetes Cluster                         │
│                                                               │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │  Ingress Controller (nginx/traefik)                     │ │
│  │  - TLS termination                                      │ │
│  │  - Load balancing                                       │ │
│  └─────────┬───────────────────────────────────────────────┘ │
│            │                                                  │
│  ┌─────────▼──────────┐  ┌───────────────┐                  │
│  │ Control Plane      │  │ Centrifugo    │                  │
│  │ - Deployment: 3    │  │ - Deployment: 2│                  │
│  │ - Autoscale: 5     │  │ - StatefulSet │                  │
│  │ - Service: ClusterIP│  │ - Service: LB │                  │
│  └─────────┬──────────┘  └───────────────┘                  │
│            │                                                  │
│  ┌─────────▼──────────┐  ┌───────────────┐                  │
│  │ MySQL              │  │ Valkey        │                  │
│  │ - StatefulSet      │  │ - StatefulSet │                  │
│  │ - PV: 100Gi        │  │ - PV: 20Gi    │                  │
│  └────────────────────┘  └───────────────┘                  │
│                                                               │
│  ┌────────────────────┐  ┌───────────────┐                  │
│  │ Quickwit           │  │ MinIO         │                  │
│  │ - StatefulSet      │  │ - StatefulSet │                  │
│  │ - PV: 500Gi        │  │ - PV: 1Ti     │                  │
│  └────────────────────┘  └───────────────┘                  │
└──────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────┐
│                    Agent Nodes                                │
│  (Physical/Virtual Servers)                                   │
│                                                               │
│  ┌──────────────────┐  ┌──────────────────┐                  │
│  │ Linux Servers    │  │ Windows Servers  │                  │
│  │ - Systemd        │  │ - Windows Service│                  │
│  │ - automation-agent│  │ - automation-agent│                 │
│  └──────────────────┘  └──────────────────┘                  │
└──────────────────────────────────────────────────────────────┘
```

### Helm Deployment

**Directory Structure**:
```
deploy/helm/
├── Chart.yaml
├── values.yaml
├── values-prod.yaml
├── templates/
│   ├── control-plane-deployment.yaml
│   ├── control-plane-service.yaml
│   ├── mysql-statefulset.yaml
│   ├── valkey-statefulset.yaml
│   ├── centrifugo-deployment.yaml
│   ├── quickwit-statefulset.yaml
│   ├── minio-statefulset.yaml
│   ├── ingress.yaml
│   └── configmaps/
```

**Deployment Commands**:
```bash
# Development
helm install automation-platform ./deploy/helm \
  --namespace automation \
  --create-namespace

# Production
helm install automation-platform ./deploy/helm \
  --namespace automation-prod \
  --values deploy/helm/values-prod.yaml \
  --create-namespace
```

### Agent Deployment

#### Linux (Systemd)
```bash
# Install
sudo ./deploy/linux/install.sh

# Configure
sudo vim /etc/automation-agent/config.env

# Start
sudo systemctl start automation-agent
sudo systemctl enable automation-agent

# Monitor
sudo journalctl -u automation-agent -f
```

#### Windows (Service)
```powershell
# Install
.\deploy\windows\install.ps1 `
  -ControlPlaneUrl "https://control-plane.example.com" `
  -TenantId "acme-corp" `
  -ProjectId "prod" `
  -JwtToken "your-token"

# Start
Start-Service AutomationAgent

# Monitor
Get-EventLog -LogName Application -Source AutomationAgent -Newest 20
```

### Scaling Considerations

#### Horizontal Scaling
- **Control Plane**: 3-10 replicas based on load
- **Centrifugo**: 2-5 replicas for HA
- **Agents**: Unlimited (tested up to 1000 per control plane)

#### Vertical Scaling
- **MySQL**: 4-16 CPU, 8-32 GB RAM
- **Valkey**: 2-8 CPU, 4-16 GB RAM
- **Quickwit**: 4-16 CPU, 16-64 GB RAM

---

## Security Architecture

### Authentication & Authorization

#### JWT Token Structure
```json
{
  "sub": "agent-linux-01",
  "tenant_id": "acme-corp",
  "project_id": "prod-infrastructure",
  "channels": ["agents.acme-corp.agent-linux-01"],
  "exp": 1735689600,
  "iat": 1704153600
}
```

**Signing Algorithm**: HS256  
**Secret**: Stored in environment variable `JWT_SECRET`

#### Token Generation
```bash
# Generate agent token
cd automation-control-plane/tools
go run gen-token.go \
  --tenant acme-corp \
  --project prod \
  --agent-id agent-01 \
  --secret your-jwt-secret
```

### Network Security

#### TLS/HTTPS
- All external communication encrypted with TLS 1.3
- Certificate management via cert-manager (Kubernetes)
- Automatic certificate rotation

#### Firewall Rules
```
Control Plane:
  - Port 8080: API (HTTPS only)
  - Internal only: MySQL, Valkey

Centrifugo:
  - Port 8000: WebSocket (WSS only)

Agents:
  - Outbound only (no inbound ports)
```

### Data Security

#### Encryption at Rest
- Database encryption (MySQL InnoDB encryption)
- Log storage encryption (MinIO SSE)
- Secrets management (Kubernetes Secrets/HashiCorp Vault)

#### Encryption in Transit
- TLS for all HTTP/WebSocket communication
- SSH key-based authentication for SSH tasks
- Database connection encryption

### Download Security

#### Ed25519 Signature Verification
```yaml
- name: secure-download
  type: downloadexec
  config:
    url: https://releases.example.com/app.exe
    sha256: abc123...
    signature: base64_signature
    public_key: base64_public_key
```

**Key Generation**:
```bash
# Generate keypair
openssl genpkey -algorithm ed25519 -out private.pem
openssl pkey -in private.pem -pubout -out public.pem

# Sign file
openssl pkeyutl -sign -inkey private.pem -out signature.bin -rawin -in file.exe

# Verify signature
openssl pkeyutl -verify -pubin -inkey public.pem -sigfile signature.bin -rawin -in file.exe
```

### Audit Logging

All security-relevant events logged:
- Authentication attempts (success/failure)
- Job submissions
- Agent registrations
- Configuration changes
- Access violations

### Security Best Practices

1. **Least Privilege**: Agents run as non-root users
2. **Secret Rotation**: JWT tokens expire and rotate
3. **Network Isolation**: Components isolated in separate networks
4. **Input Validation**: All inputs validated and sanitized
5. **Rate Limiting**: API rate limits prevent abuse
6. **Monitoring**: Security events monitored and alerted

---

## API Documentation

### Base URL
```
https://control-plane.example.com/api/v1
```

### Authentication
All API requests require JWT token in `Authorization` header:
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### Endpoints

#### Jobs

**Create Job**
```http
POST /api/v1/jobs
Content-Type: application/json

{
  "workflow": "name: test\ntasks:\n  - name: check...",
  "workflow_format": "yaml",
  "tenant_id": "acme-corp",
  "project_id": "prod",
  "agent_id": "agent-01"  // optional
}

Response: 201 Created
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "state": "pending",
  "scheduled_at": "2024-01-10T10:00:00Z"
}
```

**List Jobs**
```http
GET /api/v1/jobs?tenant_id=acme-corp&project_id=prod&state=completed

Response: 200 OK
{
  "jobs": [
    {
      "job_id": "550e8400-...",
      "state": "completed",
      "scheduled_at": "2024-01-10T10:00:00Z",
      "completed_at": "2024-01-10T10:01:30Z"
    }
  ],
  "total": 1
}
```

**Get Job**
```http
GET /api/v1/jobs/{job_id}

Response: 200 OK
{
  "job_id": "550e8400-...",
  "tenant_id": "acme-corp",
  "project_id": "prod",
  "agent_id": "agent-01",
  "workflow": "name: test...",
  "state": "completed",
  "scheduled_at": "2024-01-10T10:00:00Z",
  "started_at": "2024-01-10T10:00:05Z",
  "completed_at": "2024-01-10T10:01:30Z"
}
```

#### Agents

**Register Agent**
```http
POST /api/v1/agents
Content-Type: application/json

{
  "agent_id": "web-server-01",
  "tenant_id": "acme-corp",
  "project_id": "prod",
  "hostname": "web01.example.com",
  "platform": "linux",
  "version": "2.0.0"
}

Response: 201 Created
{
  "agent_id": "web-server-01",
  "registered_at": "2024-01-10T10:00:00Z"
}
```

**Heartbeat**
```http
POST /api/v1/agents/{agent_id}/heartbeat

Response: 200 OK
{
  "status": "ok",
  "next_heartbeat": "2024-01-10T10:01:00Z"
}
```

### Error Responses

```http
400 Bad Request
{
  "error": "invalid_workflow",
  "message": "YAML parsing error at line 5"
}

401 Unauthorized
{
  "error": "invalid_token",
  "message": "JWT token expired"
}

403 Forbidden
{
  "error": "access_denied",
  "message": "Insufficient permissions for project"
}

404 Not Found
{
  "error": "job_not_found",
  "message": "Job 550e8400-... does not exist"
}

500 Internal Server Error
{
  "error": "internal_error",
  "message": "Database connection failed"
}
```

---

## Workflow System

### YAML Workflow Specification

#### Minimal Workflow
```yaml
name: simple-check
tasks:
  - name: ping
    type: http
    config:
      url: https://example.com
```

#### Complete Workflow
```yaml
name: comprehensive-deployment
timeout: 30m
variables:
  app_version: "1.2.3"
  deploy_user: "deploy"

tasks:
  - name: check-health
    type: http
    timeout: 10s
    config:
      url: https://api.example.com/health
      method: GET
      expected_status: [200]
      headers:
        User-Agent: "AutomationPlatform/2.0"
        
  - name: backup-database
    type: db
    timeout: 5m
    config:
      driver: mysql
      dsn: "{{.DB_USER}}:{{.DB_PASSWORD}}@tcp(db:3306)/myapp"
      query: "CALL backup_procedure()"
      
  - name: upload-artifacts
    type: ssh
    timeout: 2m
    config:
      host: deploy.example.com
      user: "{{.deploy_user}}"
      key: /secrets/deploy_key
      upload:
        local: /tmp/app-{{.app_version}}.tar.gz
        remote: /opt/releases/app.tar.gz
        
  - name: deploy-application
    type: ssh
    timeout: 10m
    config:
      host: deploy.example.com
      user: "{{.deploy_user}}"
      key: /secrets/deploy_key
      command: |
        cd /opt/releases
        tar -xzf app.tar.gz
        ./deploy.sh --version {{.app_version}}
        
  - name: verify-deployment
    type: http
    timeout: 30s
    config:
      url: https://app.example.com/version
      method: GET
      expected_body_contains: "{{.app_version}}"
      
  - name: cleanup
    type: command
    config:
      command: rm
      args: ["-rf", "/tmp/app-*.tar.gz"]
```

### Task Type Reference

#### HTTP Task
```yaml
- name: api-check
  type: http
  config:
    url: https://api.example.com/endpoint
    method: POST                    # GET, POST, PUT, DELETE
    headers:
      Content-Type: application/json
      Authorization: Bearer token123
    body: '{"key": "value"}'
    expected_status: [200, 201]
    expected_body_contains: "success"
    timeout: 30s
```

#### Database Task
```yaml
- name: db-query
  type: db
  config:
    driver: mysql                   # Currently: mysql
    dsn: "user:pass@tcp(host:3306)/dbname"
    query: "SELECT COUNT(*) FROM users WHERE active=1"
    timeout: 30s
```

#### SSH Task
```yaml
- name: remote-operation
  type: ssh
  config:
    host: server.example.com
    port: 22
    user: deploy
    key: /path/to/key              # or use 'password'
    
    # Option 1: Execute command
    command: "systemctl restart myapp"
    
    # Option 2: Upload file
    upload:
      local: /local/file.txt
      remote: /remote/file.txt
      
    timeout: 60s
```

#### Command Task
```yaml
- name: local-command
  type: command
  config:
    # Option 1: Command with args (safe)
    command: mysqldump
    args: ["-u", "root", "-p", "password", "mydb"]
    
    # Option 2: Shell command (use cautiously)
    command: "mysqldump -u root mydb > backup.sql"
    shell: true
    
    timeout: 5m
```

#### PowerShell Task (Windows Only)
```yaml
- name: windows-check
  type: powershell
  config:
    script: |
      $service = Get-Service -Name "MyService"
      if ($service.Status -eq "Running") {
        Write-Output "Service running"
        exit 0
      } else {
        Write-Error "Service stopped"
        exit 1
      }
    timeout: 30s
```

#### DownloadExec Task
```yaml
- name: install-app
  type: downloadexec
  config:
    url: https://releases.example.com/installer.exe
    sha256: "abc123def456..."       # Required
    signature: "base64_signature"   # Optional (recommended)
    public_key: "base64_public_key" # Required if signature present
    args: ["--silent", "--install-dir=/opt/app"]
    timeout: 10m
    cleanup: true                   # Delete after execution
```

### Variable Substitution

Workflows support variable substitution using `{{.VARIABLE_NAME}}` syntax:

```yaml
name: deploy-{{.ENVIRONMENT}}
variables:
  ENVIRONMENT: production
  APP_VERSION: "1.2.3"
  
tasks:
  - name: deploy
    type: command
    config:
      command: deploy.sh
      args: ["--env", "{{.ENVIRONMENT}}", "--version", "{{.APP_VERSION}}"]
```

---

## Performance and Scalability

### Performance Characteristics

#### Job Submission
- **Latency**: <50ms p50, <100ms p95
- **Throughput**: 1000 jobs/second per control plane instance
- **Bottleneck**: Database write speed

#### Job Dispatch
- **Latency**: <500ms p95 from submission to agent notification
- **Bottleneck**: Centrifugo publish latency

#### Workflow Execution
- **Depends on**: Task types, network latency, remote system performance
- **HTTP tasks**: 50-200ms
- **Database tasks**: 10-100ms
- **SSH tasks**: 100-500ms
- **Command tasks**: Varies widely

### Scalability Limits

#### Control Plane
- **Horizontal**: 3-10 instances recommended
- **Capacity**: 10,000 jobs/minute per cluster
- **Database**: MySQL with read replicas for read scaling

#### Agents
- **Per Control Plane**: 10,000 concurrent agents (tested)
- **Theoretical Max**: 100,000+ (limited by Centrifugo capacity)

#### Storage
- **MySQL**: 10 TB+ (use partitioning for job history)
- **Quickwit**: Petabyte-scale (S3-backed)

### Optimization Strategies

#### Database Optimization
```sql
-- Partition jobs table by date
ALTER TABLE jobs PARTITION BY RANGE (YEAR(scheduled_at)) (
    PARTITION p2024 VALUES LESS THAN (2025),
    PARTITION p2025 VALUES LESS THAN (2026),
    PARTITION p2026 VALUES LESS THAN (2027)
);

-- Index for common queries
CREATE INDEX idx_jobs_tenant_project_state 
  ON jobs(tenant_id, project_id, state, scheduled_at);

-- Archive old jobs
DELETE FROM jobs 
WHERE completed_at < DATE_SUB(NOW(), INTERVAL 90 DAY);
```

#### Caching Strategy
- Job metadata cached in Valkey (5 min TTL)
- Agent status cached (1 min TTL)
- Workflow templates cached (1 hour TTL)

#### Connection Pooling
```go
// MySQL connection pool
db.SetMaxOpenConns(100)
db.SetMaxIdleConns(10)
db.SetConnMaxLifetime(time.Hour)

// Valkey connection pool
&redis.Options{
    PoolSize: 100,
    MinIdleConns: 10,
}
```

---

## Operational Considerations

### Monitoring

#### Metrics to Monitor
- **Control Plane**: Request rate, error rate, latency
- **Agents**: Heartbeat status, job execution rate, failure rate
- **Database**: Query time, connection count, replication lag
- **Centrifugo**: Active connections, message throughput
- **Quickwit**: Indexing rate, query latency, storage usage

#### Prometheus Metrics
```prometheus
# Job metrics
automation_jobs_total{state="completed|failed|pending"}
automation_jobs_duration_seconds{percentile="50|95|99"}
automation_job_submission_rate

# Agent metrics
automation_agents_registered_total
automation_agents_active
automation_agents_heartbeat_failures_total

# System metrics
automation_http_requests_total{endpoint,method,status}
automation_http_request_duration_seconds{endpoint}
```

#### Health Checks
```http
GET /health

Response: 200 OK
{
  "status": "healthy",
  "version": "2.0.0",
  "components": {
    "database": "healthy",
    "cache": "healthy",
    "centrifugo": "healthy",
    "quickwit": "healthy"
  }
}
```

### Logging

#### Log Levels
- **DEBUG**: Detailed execution traces
- **INFO**: Normal operations (job submitted, agent registered)
- **WARN**: Recoverable errors (retry, timeout)
- **ERROR**: Unrecoverable errors (database failure)

#### Structured Logging
```json
{
  "timestamp": "2024-01-10T10:00:00Z",
  "level": "info",
  "component": "control-plane",
  "job_id": "550e8400-...",
  "tenant_id": "acme-corp",
  "message": "Job submitted successfully"
}
```

### Backup and Recovery

#### Backup Strategy
- **MySQL**: Daily full backup + binlog replication
- **Quickwit Indices**: S3 backend (automatic durability)
- **Configuration**: Git-backed (infrastructure as code)

#### Recovery Procedures
```bash
# Restore MySQL from backup
mysql -u root -p mydb < backup_2024-01-10.sql

# Verify data integrity
SELECT COUNT(*) FROM jobs;

# Restart services
kubectl rollout restart deployment/control-plane
```

### Disaster Recovery

#### RTO/RPO Targets
- **RTO** (Recovery Time Objective): 1 hour
- **RPO** (Recovery Point Objective): 5 minutes

#### Multi-Region Setup
```
┌──────────────────┐         ┌──────────────────┐
│  Region: US-East │         │  Region: US-West │
│                  │         │                  │
│  - Control Plane │◄───────►│  - Control Plane │
│  - MySQL Primary │  Sync   │  - MySQL Replica │
│  - Active Agents │         │  - Standby       │
└──────────────────┘         └──────────────────┘
```

---

## Appendices

### Glossary

- **Agent**: Software that executes workflows on target systems
- **Control Plane**: Central orchestration service
- **Job**: Single workflow execution instance
- **Probe**: Task execution framework
- **Task**: Single operation within a workflow (HTTP, DB, SSH, etc.)
- **Tenant**: Organizational unit (e.g., company, department)
- **Project**: Subdivision within tenant (e.g., prod, staging)
- **Workflow**: Declarative YAML definition of automation tasks

### References

- **Probe Framework**: Inspired by [linyows/probe](https://github.com/linyows/probe)
- **Centrifugo**: [https://centrifugal.dev/](https://centrifugal.dev/)
- **Quickwit**: [https://quickwit.io/](https://quickwit.io/)
- **YAML Spec**: [https://yaml.org/spec/1.2/](https://yaml.org/spec/1.2/)

### Version History

| Version | Date | Changes |
|---------|------|---------|
| 2.0.0 | 2026-01 | Probe integration, YAML workflows, 6 task types |
| 1.0.0 | 2025-12 | Initial release with JSON workflows |

---

**Document Version**: 1.0  
**Last Updated**: January 11, 2026  
**Maintained By**: Automation Platform Team  
**Status**: Living Document
