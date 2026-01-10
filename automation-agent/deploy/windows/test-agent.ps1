# Test the agent binary before installing as a service
# This runs the agent in the foreground so you can see the output

$scriptPath = Split-Path -Parent $MyInvocation.MyCommand.Path

# Set environment variables
$env:CONTROL_PLANE_URL = "http://localhost:8081"
$env:CENTRIFUGO_URL = "ws://localhost:8000/connection/websocket"
$env:TENANT_ID = "test-tenant"
$env:PROJECT_ID = "test-project"
$env:AGENT_ID = "agent-windows-01"
$env:JWT_TOKEN = "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJhZ2VudF9pZCI6ImFnZW50LXdpbmRvd3MtMDEiLCJ0ZW5hbnRfaWQiOiJ0ZXN0LXRlbmFudCIsInByb2plY3RfaWQiOiJ0ZXN0LXByb2plY3QiLCJleHAiOjE3OTk1MTU4MTUsImlhdCI6MTc2Nzk3OTgxNX0.qBn7er3PDk3oF--bK-oHfg8bWLnAChA0zpNaV1A19s8"
$env:LOG_LEVEL = "debug"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Testing Automation Agent" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Configuration:" -ForegroundColor Yellow
Write-Host "  Control Plane: $($env:CONTROL_PLANE_URL)" -ForegroundColor Gray
Write-Host "  Centrifugo:    $($env:CENTRIFUGO_URL)" -ForegroundColor Gray
Write-Host "  Tenant ID:     $($env:TENANT_ID)" -ForegroundColor Gray
Write-Host "  Project ID:    $($env:PROJECT_ID)" -ForegroundColor Gray
Write-Host "  Agent ID:      $($env:AGENT_ID)" -ForegroundColor Gray
Write-Host ""
Write-Host "Starting agent... (Press Ctrl+C to stop)" -ForegroundColor Green
Write-Host ""

# Run the agent
& "$scriptPath\automation-agent.exe"
