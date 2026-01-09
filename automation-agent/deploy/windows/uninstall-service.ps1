# Windows Service Removal Script

$ErrorActionPreference = "Stop"

# Check for Administrator
$isAdmin = ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
if (-not $isAdmin) {
    Write-Error "This script requires Administrator privileges"
    exit 1
}

$ServiceName = "AutomationAgent"

# Check if service exists
$service = Get-Service -Name $ServiceName -ErrorAction SilentlyContinue
if (-not $service) {
    Write-Host "Service $ServiceName not found"
    exit 0
}

# Stop service if running
if ($service.Status -eq "Running") {
    Write-Host "Stopping service..."
    Stop-Service -Name $ServiceName -Force
}

# Delete service
Write-Host "Removing service..."
sc.exe delete $ServiceName

Write-Host "Service removed successfully!"
