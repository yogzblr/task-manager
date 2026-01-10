# Build and Test Instructions

## Prerequisites

### Install Go

1. Download Go from: https://go.dev/dl/
2. Install Go 1.21 or later
3. Verify installation: `go version`

### Install Git (if not already installed)

1. Download from: https://git-scm.com/download/win
2. Install with default options

## Build the Probe Module

```powershell
# Navigate to probe directory
cd "C:\Users\yoges\OneDrive\Documents\My Code\Task Manager\demo\probe"

# Download dependencies
go mod download

# Tidy up dependencies
go mod tidy

# Run tests
go test ./... -v

# Build test program
go build -o test-probe.exe ./cmd/test-probe
```

## Test with Example Workflows

```powershell
# Test HTTP example
.\test-probe.exe .\examples\http-example.yaml

# Test Command example  
.\test-probe.exe .\examples\command-example.yaml

# Test with automation-agent examples
.\test-probe.exe ..\automation-agent\examples\workflows\simple-health-check.yaml
.\test-probe.exe ..\automation-agent\examples\workflows\command-execution.yaml
```

## Build the Automation Agent

```powershell
# Navigate to agent directory
cd "C:\Users\yoges\OneDrive\Documents\My Code\Task Manager\demo\automation-agent"

# Build agent
go build -o automation-agent.exe ./cmd/agent
```

## Quick Test Script

After installing Go, run:

```powershell
cd "C:\Users\yoges\OneDrive\Documents\My Code\Task Manager\demo"
.\build-and-test.ps1
```

## Manual Testing Steps

If Go is installed, you can test step by step:

### 1. Test Probe Unit Tests

```powershell
cd probe
go test -v -run TestProbeHTTPTask
go test -v -run TestProbeCommandTask
go test -v -run TestPowerShellTask  # Windows only
go test -v -run TestDownloadExecTask
```

### 2. Test Example Workflows

```powershell
# Build test tool
go build -o test-probe.exe ./cmd/test-probe

# Test HTTP workflow
./test-probe.exe ./examples/http-example.yaml

# Test Command workflow
./test-probe.exe ./examples/command-example.yaml
```

### 3. Integration Test with Agent

```powershell
cd ../automation-agent

# Build agent
go build -o automation-agent.exe ./cmd/agent

# Test with example workflow (requires control plane)
# Set environment variables first:
$env:CONTROL_PLANE_URL="http://localhost:8080"
$env:CENTRIFUGO_URL="ws://localhost:8000/connection/websocket"
$env:TENANT_ID="test"
$env:PROJECT_ID="test"
$env:JWT_TOKEN="your-token"

# Run agent
./automation-agent.exe
```

## Troubleshooting

### Go Not Found

```powershell
# Check PATH
$env:PATH -split ';' | Select-String "Go"

# If not found, add Go to PATH:
[Environment]::SetEnvironmentVariable("Path", $env:Path + ";C:\Program Files\Go\bin", [EnvironmentVariableTarget]::User)

# Restart PowerShell
```

### Module Dependencies

```powershell
cd probe
go mod download
go mod verify
```

### Build Errors

```powershell
# Clean module cache
go clean -modcache

# Re-download dependencies
go mod download

# Try building again
go build ./...
```

## Expected Test Results

### HTTP Example

```
=== Testing Workflow: http-example.yaml ===

Workflow: http-example
Overall Success: true

Task 1: check-httpbin (http)
  ✓ SUCCESS
  Output: map[body:... headers:... status_code:200]

Task 2: check-httpbin-404 (http)
  ✓ SUCCESS
  Output: map[body:... headers:... status_code:404]

Task 3: check-github-api (http)
  ✓ SUCCESS
  Output: map[body:... headers:... status_code:200]

=== All tasks completed successfully ===
```

### Command Example

```
=== Testing Workflow: command-example.yaml ===

Workflow: command-example
Overall Success: true

Task 1: echo-test (command)
  ✓ SUCCESS
  Output: map[exit_code:0 output:Hello from probe!]

Task 2: list-current-directory (command)
  ✓ SUCCESS
  Output: map[exit_code:0 output:<directory listing>]

Task 3: date-command (command)
  ✓ SUCCESS
  Output: map[exit_code:0 output:<current date>]

=== All tasks completed successfully ===
```

## Next Steps After Testing

1. ✅ Verify all tests pass
2. ✅ Test example workflows
3. ✅ Build agent successfully
4. 📤 Push to GitHub (see PUSH-TO-GITHUB.md)
5. 🚀 Deploy to environment
