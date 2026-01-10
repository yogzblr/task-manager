# Enable Docker Desktop WSL Integration and Run Build Script
# Run this script as Administrator

Write-Host "=== Docker Desktop WSL Integration Setup ===" -ForegroundColor Cyan
Write-Host ""

# Check if Docker Desktop is running
$dockerProcess = Get-Process "Docker Desktop" -ErrorAction SilentlyContinue
if (-not $dockerProcess) {
    Write-Host "❌ Docker Desktop is not running!" -ForegroundColor Red
    Write-Host ""
    Write-Host "Please start Docker Desktop and wait for it to fully start." -ForegroundColor Yellow
    Write-Host "Then run this script again." -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Starting Docker Desktop for you..." -ForegroundColor Yellow
    
    # Try to start Docker Desktop
    $dockerPath = "C:\Program Files\Docker\Docker\Docker Desktop.exe"
    if (Test-Path $dockerPath) {
        Start-Process $dockerPath
        Write-Host "Waiting 30 seconds for Docker Desktop to start..." -ForegroundColor Yellow
        Start-Sleep -Seconds 30
    } else {
        Write-Host "Docker Desktop not found at: $dockerPath" -ForegroundColor Red
        Write-Host "Please install Docker Desktop from: https://www.docker.com/products/docker-desktop" -ForegroundColor Yellow
        exit 1
    }
}

Write-Host "✓ Docker Desktop is running" -ForegroundColor Green
Write-Host ""

# Instructions for enabling WSL integration
Write-Host "=== Enable WSL Integration ===" -ForegroundColor Cyan
Write-Host ""
Write-Host "Please follow these steps:" -ForegroundColor Yellow
Write-Host "1. Click on Docker Desktop icon in system tray" -ForegroundColor White
Write-Host "2. Click Settings (gear icon)" -ForegroundColor White
Write-Host "3. Go to: Resources → WSL Integration" -ForegroundColor White
Write-Host "4. Enable toggle for 'Ubuntu-22.04'" -ForegroundColor White
Write-Host "5. Click 'Apply & Restart'" -ForegroundColor White
Write-Host ""
Write-Host "Press Enter after enabling WSL integration..." -ForegroundColor Yellow
$null = Read-Host

# Verify Docker is available in WSL
Write-Host ""
Write-Host "=== Verifying Docker in WSL ===" -ForegroundColor Cyan
Write-Host ""

$dockerCheck = wsl -d Ubuntu-22.04 bash -c "docker --version 2>&1"
if ($LASTEXITCODE -eq 0 -and $dockerCheck -match "Docker version") {
    Write-Host "✓ Docker is available in WSL: $dockerCheck" -ForegroundColor Green
} else {
    Write-Host "❌ Docker is not available in WSL" -ForegroundColor Red
    Write-Host ""
    Write-Host "Output: $dockerCheck" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Please make sure:" -ForegroundColor Yellow
    Write-Host "1. Docker Desktop is running" -ForegroundColor White
    Write-Host "2. WSL integration is enabled for Ubuntu-22.04" -ForegroundColor White
    Write-Host "3. Docker Desktop has been restarted after enabling" -ForegroundColor White
    Write-Host ""
    Write-Host "Press Enter to try again or Ctrl+C to exit..." -ForegroundColor Yellow
    $null = Read-Host
    
    # Try again
    $dockerCheck = wsl -d Ubuntu-22.04 bash -c "docker --version 2>&1"
    if ($LASTEXITCODE -ne 0 -or $dockerCheck -notmatch "Docker version") {
        Write-Host "❌ Docker still not available. Please enable WSL integration manually." -ForegroundColor Red
        exit 1
    }
    Write-Host "✓ Docker is now available!" -ForegroundColor Green
}

Write-Host ""
Write-Host "=== Running Build and Test Script ===" -ForegroundColor Cyan
Write-Host ""

# Navigate to deploy directory and run the script
$deployPath = "/mnt/c/Users/yoges/OneDrive/Documents/My Code/Task Manager/demo/automation-control-plane/deploy"

Write-Host "Navigating to: $deployPath" -ForegroundColor Yellow
Write-Host "This may take 5-10 minutes on first run..." -ForegroundColor Yellow
Write-Host ""

# Run the build and test script
wsl -d Ubuntu-22.04 bash -c "cd '$deployPath' && chmod +x build-and-test-wsl.sh && ./build-and-test-wsl.sh"

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "=== Build and Test Completed Successfully! ===" -ForegroundColor Green
    Write-Host ""
    Write-Host "Services are now running:" -ForegroundColor Cyan
    Write-Host "  Control Plane: http://localhost:8081" -ForegroundColor White
    Write-Host "  Centrifugo:    http://localhost:8000" -ForegroundColor White
    Write-Host "  MySQL:         localhost:3306" -ForegroundColor White
    Write-Host "  Valkey:        localhost:6379" -ForegroundColor White
    Write-Host "  Quickwit:      http://localhost:7280" -ForegroundColor White
    Write-Host ""
    Write-Host "To view logs: wsl -d Ubuntu-22.04 bash -c `"cd '$deployPath' && docker compose logs -f`"" -ForegroundColor Yellow
    Write-Host "To stop:      wsl -d Ubuntu-22.04 bash -c `"cd '$deployPath' && docker compose down`"" -ForegroundColor Yellow
} else {
    Write-Host ""
    Write-Host "=== Build Failed ===" -ForegroundColor Red
    Write-Host ""
    Write-Host "Check the output above for errors." -ForegroundColor Yellow
    Write-Host "Common issues:" -ForegroundColor Yellow
    Write-Host "  - Ports already in use (8081, 3306, 6379, 8000)" -ForegroundColor White
    Write-Host "  - Docker daemon not accessible" -ForegroundColor White
    Write-Host "  - Insufficient disk space" -ForegroundColor White
    Write-Host ""
    Write-Host "To view detailed logs:" -ForegroundColor Yellow
    Write-Host "  wsl -d Ubuntu-22.04 bash -c `"cd '$deployPath' && docker compose logs`"" -ForegroundColor White
}
