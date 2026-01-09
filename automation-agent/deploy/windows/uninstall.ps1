# Automation Agent Windows Uninstallation Script
# Requires Administrator privileges

$ErrorActionPreference = "Stop"

# Check for Administrator
$isAdmin = ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
if (-not $isAdmin) {
    Write-Error "This script requires Administrator privileges"
    exit 1
}

$ServiceName = "AutomationAgent"
$InstallDir = "C:\Program Files\AutomationAgent"
$ConfigDir = "C:\ProgramData\AutomationAgent"

Write-Host "Uninstalling Automation Agent..."

# Stop and remove service
$service = Get-Service -Name $ServiceName -ErrorAction SilentlyContinue
if ($service) {
    if ($service.Status -eq "Running") {
        Write-Host "Stopping service..."
        Stop-Service -Name $ServiceName -Force
    }
    
    Write-Host "Removing service..."
    sc.exe delete $ServiceName
    Start-Sleep -Seconds 2
}

# Remove installation files
if (Test-Path $InstallDir) {
    Write-Host "Removing installation directory..."
    Remove-Item -Path $InstallDir -Recurse -Force
}

# Remove configuration (optional)
$removeConfig = Read-Host "Remove configuration files? (y/N)"
if ($removeConfig -eq "y" -or $removeConfig -eq "Y") {
    if (Test-Path $ConfigDir) {
        Remove-Item -Path $ConfigDir -Recurse -Force
    }
}

Write-Host "Uninstallation complete!"
