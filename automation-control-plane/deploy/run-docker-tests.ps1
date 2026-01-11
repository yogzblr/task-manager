# PowerShell script to run Docker-based workflow tests from Windows
# This keeps your Windows environment clean while testing the automation platform

$ErrorActionPreference = "Stop"

$DEPLOY_PATH = "/mnt/c/Users/yoges/OneDrive/Documents/My Code/Task Manager/demo/automation-control-plane/deploy"

Write-Host @"

╔══════════════════════════════════════════════════════════════╗
║     Docker-Based Workflow Testing - Run from Windows        ║
╚══════════════════════════════════════════════════════════════╝

"@ -ForegroundColor Cyan

# Function to run WSL commands
function Invoke-WslCommand {
    param(
        [string]$Command,
        [string]$Description
    )
    
    Write-Host "`n▶ $Description" -ForegroundColor Yellow
    Write-Host "  Running: $Command`n" -ForegroundColor Gray
    
    wsl -d Ubuntu-22.04 bash -c "cd '$DEPLOY_PATH' && $Command"
    
    if ($LASTEXITCODE -ne 0) {
        Write-Host "✗ Command failed with exit code: $LASTEXITCODE" -ForegroundColor Red
        return $false
    }
    return $true
}

# Main menu
Write-Host "Select an option:" -ForegroundColor Green
Write-Host "  1. Check Docker services status"
Write-Host "  2. Build test-runner image"
Write-Host "  3. Run Linux workflow test"
Write-Host "  4. Run Windows workflow test (requires Windows agent)"
Write-Host "  5. Interactive test-runner shell"
Write-Host "  6. View control plane logs"
Write-Host "  7. View Linux agent logs"
Write-Host "  8. Query Quickwit for recent logs"
Write-Host "  9. Start all services"
Write-Host "  0. Exit"
Write-Host ""

$choice = Read-Host "Enter choice (0-9)"

switch ($choice) {
    "1" {
        Invoke-WslCommand "docker compose ps" "Checking Docker service status"
    }
    
    "2" {
        Invoke-WslCommand "docker compose build test-runner" "Building test-runner image"
    }
    
    "3" {
        Write-Host "`n" -NoNewline
        Invoke-WslCommand "docker compose run --rm test-runner python test-linux-workflow.py" "Running Linux workflow test"
    }
    
    "4" {
        Write-Host "`n⚠ WARNING: This requires a Windows agent to be running!`n" -ForegroundColor Yellow
        $confirm = Read-Host "Is the Windows agent running? (y/n)"
        if ($confirm -eq "y") {
            Invoke-WslCommand "docker compose run --rm test-runner python test-windows-workflow.py" "Running Windows workflow test"
        } else {
            Write-Host "Cancelled. Start the Windows agent first." -ForegroundColor Red
        }
    }
    
    "5" {
        Write-Host "`nStarting interactive shell in test-runner container..." -ForegroundColor Green
        Write-Host "Commands available inside:" -ForegroundColor Gray
        Write-Host "  - python test-linux-workflow.py" -ForegroundColor Gray
        Write-Host "  - python test-windows-workflow.py" -ForegroundColor Gray
        Write-Host "  - ls -la" -ForegroundColor Gray
        Write-Host "  - exit (to leave container)`n" -ForegroundColor Gray
        
        wsl -d Ubuntu-22.04 bash -c "cd '$DEPLOY_PATH' && docker compose run --rm test-runner /bin/bash"
    }
    
    "6" {
        Invoke-WslCommand "docker compose logs --tail=50 control-plane" "Viewing control plane logs (last 50 lines)"
    }
    
    "7" {
        Invoke-WslCommand "docker compose logs --tail=50 agent-linux" "Viewing Linux agent logs (last 50 lines)"
    }
    
    "8" {
        Write-Host "`nQuerying Quickwit for recent automation logs..." -ForegroundColor Green
        $query = @'
{
  "query": "*",
  "max_hits": 20,
  "sort_by": "-timestamp"
}
'@
        
        wsl -d Ubuntu-22.04 bash -c @"
cd '$DEPLOY_PATH' && docker compose run --rm test-runner bash -c 'curl -X POST http://quickwit:7280/api/v1/automation-logs/search -H \"Content-Type: application/json\" -d ''$query'' | python -m json.tool'
"@
    }
    
    "9" {
        Invoke-WslCommand "docker compose up -d" "Starting all Docker services"
        Write-Host "`nWaiting for services to be ready..." -ForegroundColor Yellow
        Start-Sleep -Seconds 10
        Invoke-WslCommand "docker compose ps" "Checking service status"
    }
    
    "0" {
        Write-Host "`nExiting..." -ForegroundColor Green
        exit 0
    }
    
    default {
        Write-Host "`n✗ Invalid choice. Please run the script again." -ForegroundColor Red
        exit 1
    }
}

Write-Host "`n" -NoNewline
Write-Host "═══════════════════════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host "Test operation complete!" -ForegroundColor Green
Write-Host "═══════════════════════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host ""
