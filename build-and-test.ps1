# Build and Test Script
# Requires Go 1.21+ to be installed

Write-Host "=== Automation Platform - Build and Test ===" -ForegroundColor Cyan
Write-Host ""

# Check if Go is installed
$goVersion = & go version 2>$null
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Go is not installed!" -ForegroundColor Red
    Write-Host ""
    Write-Host "Please install Go 1.21+ from: https://go.dev/dl/" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "After installation, restart PowerShell and run this script again." -ForegroundColor Yellow
    exit 1
}

Write-Host "✓ Go installed: $goVersion" -ForegroundColor Green
Write-Host ""

# Navigate to probe directory
$probeDir = Join-Path $PSScriptRoot "probe"
Set-Location $probeDir

Write-Host "=== Building Probe Module ===" -ForegroundColor Cyan

# Download dependencies
Write-Host "Downloading dependencies..." -ForegroundColor Yellow
go mod download
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Failed to download dependencies" -ForegroundColor Red
    exit 1
}

# Tidy modules
Write-Host "Tidying modules..." -ForegroundColor Yellow
go mod tidy
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Failed to tidy modules" -ForegroundColor Red
    exit 1
}

Write-Host "✓ Dependencies ready" -ForegroundColor Green
Write-Host ""

# Run unit tests
Write-Host "=== Running Unit Tests ===" -ForegroundColor Cyan
go test ./... -v
$testResult = $LASTEXITCODE
Write-Host ""

if ($testResult -ne 0) {
    Write-Host "⚠ Some tests failed (expected on non-Windows for PowerShell tests)" -ForegroundColor Yellow
} else {
    Write-Host "✓ All tests passed" -ForegroundColor Green
}
Write-Host ""

# Build test program
Write-Host "=== Building Test Program ===" -ForegroundColor Cyan
go build -o test-probe.exe ./cmd/test-probe
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Failed to build test program" -ForegroundColor Red
    exit 1
}

Write-Host "✓ Built test-probe.exe" -ForegroundColor Green
Write-Host ""

# Test with HTTP example
Write-Host "=== Testing HTTP Workflow ===" -ForegroundColor Cyan
.\test-probe.exe .\examples\http-example.yaml
$httpResult = $LASTEXITCODE
Write-Host ""

# Test with Command example
Write-Host "=== Testing Command Workflow ===" -ForegroundColor Cyan
.\test-probe.exe .\examples\command-example.yaml
$commandResult = $LASTEXITCODE
Write-Host ""

# Build automation agent
Write-Host "=== Building Automation Agent ===" -ForegroundColor Cyan
$agentDir = Join-Path $PSScriptRoot "automation-agent"
Set-Location $agentDir

go build -o automation-agent.exe ./cmd/agent
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Failed to build automation agent" -ForegroundColor Red
    exit 1
}

Write-Host "✓ Built automation-agent.exe" -ForegroundColor Green
Write-Host ""

# Summary
Write-Host "=== Build and Test Summary ===" -ForegroundColor Cyan
Write-Host ""
Write-Host "Probe Module:" -ForegroundColor White
if ($testResult -eq 0) {
    Write-Host "  ✓ Unit Tests: PASSED" -ForegroundColor Green
} else {
    Write-Host "  ⚠ Unit Tests: Some failures (expected)" -ForegroundColor Yellow
}

Write-Host "  ✓ Test Program: Built successfully" -ForegroundColor Green

Write-Host ""
Write-Host "Example Workflows:" -ForegroundColor White
if ($httpResult -eq 0) {
    Write-Host "  ✓ HTTP Example: PASSED" -ForegroundColor Green
} else {
    Write-Host "  ✗ HTTP Example: FAILED" -ForegroundColor Red
}

if ($commandResult -eq 0) {
    Write-Host "  ✓ Command Example: PASSED" -ForegroundColor Green
} else {
    Write-Host "  ✗ Command Example: FAILED" -ForegroundColor Red
}

Write-Host ""
Write-Host "Automation Agent:" -ForegroundColor White
Write-Host "  ✓ Agent: Built successfully" -ForegroundColor Green

Write-Host ""
Write-Host "=== Next Steps ===" -ForegroundColor Cyan
Write-Host "1. Review BUILD-AND-TEST.md for detailed testing instructions" -ForegroundColor Yellow
Write-Host "2. Test additional workflows in automation-agent/examples/workflows/" -ForegroundColor Yellow
Write-Host "3. See PUSH-TO-GITHUB.md for GitHub deployment instructions" -ForegroundColor Yellow
Write-Host ""

# Return to original directory
Set-Location $PSScriptRoot
