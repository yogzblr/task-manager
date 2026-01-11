# Run Windows Agent (Native)
# This script runs the Windows agent binary on your Windows host
# while the control plane runs in WSL Docker containers

Write-Host "=== Starting Windows Automation Agent ===" -ForegroundColor Cyan
Write-Host ""

# Set working directory
$agentDir = "C:\Users\yoges\OneDrive\Documents\My Code\Task Manager\demo\automation-agent"
Set-Location $agentDir

# Check if binary exists
if (-not (Test-Path ".\automation-agent-windows.exe")) {
    Write-Host "ERROR: automation-agent-windows.exe not found!" -ForegroundColor Red
    Write-Host "Please build it first:" -ForegroundColor Yellow
    Write-Host "  cd $agentDir" -ForegroundColor Yellow
    Write-Host "  go build -o automation-agent-windows.exe ./cmd/agent" -ForegroundColor Yellow
    exit 1
}

# Set environment variables
$env:CONTROL_PLANE_URL = "http://localhost:8081"
$env:CENTRIFUGO_URL = "ws://localhost:8000/connection/websocket"
$env:TENANT_ID = "test-tenant"
$env:PROJECT_ID = "test-project"
$env:AGENT_ID = "agent-windows-01"
$env:JWT_TOKEN = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhZ2VudF9pZCI6ImFnZW50LXdpbmRvd3MtMDEiLCJ0ZW5hbnRfaWQiOiJ0ZXN0LXRlbmFudCIsInByb2plY3RfaWQiOiJ0ZXN0LXByb2plY3QiLCJleHAiOjE3OTk0OTUwMDksImlhdCI6MTc2Nzk1OTAwOX0.JKQXv4YeRRA46gPU-cJpyV83FC2ZFXxWrR_M1zkuQO0"
$env:LOG_LEVEL = "debug"

Write-Host "Configuration:" -ForegroundColor Green
Write-Host "  Control Plane: $env:CONTROL_PLANE_URL"
Write-Host "  Centrifugo:    $env:CENTRIFUGO_URL"
Write-Host "  Tenant ID:     $env:TENANT_ID"
Write-Host "  Project ID:    $env:PROJECT_ID"
Write-Host "  Agent ID:      $env:AGENT_ID"
Write-Host "  Log Level:     $env:LOG_LEVEL"
Write-Host ""

# Check if control plane is accessible
Write-Host "Checking control plane connectivity..." -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest -Uri "http://localhost:8081/health" -TimeoutSec 5 -ErrorAction Stop
    Write-Host "  Control plane is accessible!" -ForegroundColor Green
} catch {
    Write-Host "  WARNING: Cannot reach control plane at http://localhost:8081/health" -ForegroundColor Red
    Write-Host "  Make sure Docker Compose services are running in WSL" -ForegroundColor Yellow
    Write-Host "  Command: wsl -d Ubuntu-22.04 bash -c 'cd /mnt/c/Users/yoges/OneDrive/Documents/My\ Code/Task\ Manager/demo/automation-control-plane/deploy && docker compose ps'" -ForegroundColor Yellow
    Write-Host ""
    $continue = Read-Host "Continue anyway? (y/N)"
    if ($continue -ne 'y') {
        exit 1
    }
}

Write-Host ""
Write-Host "Starting Windows agent..." -ForegroundColor Cyan
Write-Host "Press Ctrl+C to stop" -ForegroundColor Yellow
Write-Host ""

# Run the agent
.\automation-agent-windows.exe
