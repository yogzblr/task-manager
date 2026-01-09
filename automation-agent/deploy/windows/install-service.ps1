# Windows Service Installation Script

param(
    [string]$BinaryPath = "C:\Program Files\AutomationAgent\automation-agent.exe"
)

$ErrorActionPreference = "Stop"

# Check for Administrator
$isAdmin = ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
if (-not $isAdmin) {
    Write-Error "This script requires Administrator privileges"
    exit 1
}

$ServiceName = "AutomationAgent"
$DisplayName = "Automation Agent"

# Check if service exists
$service = Get-Service -Name $ServiceName -ErrorAction SilentlyContinue
if ($service) {
    if ($service.Status -eq "Running") {
        Stop-Service -Name $ServiceName -Force
    }
    sc.exe delete $ServiceName
    Start-Sleep -Seconds 2
}

# Create service
Write-Host "Installing service..."
New-Service -Name $ServiceName `
    -DisplayName $DisplayName `
    -BinaryPathName $BinaryPath `
    -StartupType Automatic `
    -Description "Automation Agent Service" | Out-Null

# Configure service recovery
sc.exe failure $ServiceName reset= 86400 actions= restart/60000/restart/60000/restart/60000

# Start service
Write-Host "Starting service..."
Start-Service -Name $ServiceName

Write-Host "Service installed and started successfully!"
