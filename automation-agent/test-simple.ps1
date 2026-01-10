# Simple test to verify agent binary works
$env:CONTROL_PLANE_URL = "http://localhost:8081"
$env:CENTRIFUGO_URL = "ws://localhost:8000/connection/websocket"
$env:TENANT_ID = "test-tenant"
$env:PROJECT_ID = "test-project"
$env:AGENT_ID = "agent-windows-test"
$env:JWT_TOKEN = "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJhZ2VudF9pZCI6ImFnZW50LXdpbmRvd3MtMDEiLCJ0ZW5hbnRfaWQiOiJ0ZXN0LXRlbmFudCIsInByb2plY3RfaWQiOiJ0ZXN0LXByb2plY3QiLCJleHAiOjE3OTk1MTU4MTUsImlhdCI6MTc2Nzk3OTgxNX0.qBn7er3PDk3oF--bK-oHfg8bWLnAChA0zpNaV1A19s8"
$env:LOG_LEVEL = "debug"

Write-Host "Environment variables set" -ForegroundColor Green
Write-Host "Control Plane: $env:CONTROL_PLANE_URL"
Write-Host "Agent ID: $env:AGENT_ID"
Write-Host ""
Write-Host "Running agent for 10 seconds..." -ForegroundColor Yellow

# Run agent in background
$job = Start-Process -FilePath "C:\Users\yoges\OneDrive\Documents\My Code\demo\automation-agent\deploy\windows\automation-agent.exe" -PassThru

# Wait 10 seconds
Start-Sleep -Seconds 10

# Stop agent
if (!$job.HasExited) {
    $job | Stop-Process -Force
    Write-Host "`nAgent stopped" -ForegroundColor Yellow
} else {
    Write-Host "`nAgent exited with code: $($job.ExitCode)" -ForegroundColor Red
}
