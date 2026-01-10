# Automation Agent Windows Installation Script
# Requires Administrator privileges

param(
    [string]$ControlPlaneUrl = "",
    [string]$TenantId = "",
    [string]$ProjectId = "",
    [string]$AgentId = "",
    [string]$JwtToken = ""
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
$InstallDir = "C:\Program Files\AutomationAgent"
$ConfigDir = "C:\ProgramData\AutomationAgent"
$LogDir = "C:\ProgramData\AutomationAgent\logs"
$BinaryName = "automation-agent.exe"

Write-Host "Installing Automation Agent..."

# Create directories
Write-Host "Creating directories..."
New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
New-Item -ItemType Directory -Force -Path $ConfigDir | Out-Null
New-Item -ItemType Directory -Force -Path $LogDir | Out-Null

# Copy binary
if (Test-Path ".\$BinaryName") {
    Write-Host "Installing binary..."
    Copy-Item ".\$BinaryName" "$InstallDir\$BinaryName" -Force
} else {
    Write-Error "Binary not found: .\$BinaryName"
    exit 1
}

# Create environment file for service
$EnvFile = "$ConfigDir\service-env.txt"
Write-Host "Creating environment configuration..."
$envConfig = @"
CONTROL_PLANE_URL=$ControlPlaneUrl
TENANT_ID=$TenantId
PROJECT_ID=$ProjectId
AGENT_ID=$AgentId
JWT_TOKEN=$JwtToken
LOG_LEVEL=info
"@
Set-Content -Path $EnvFile -Value $envConfig

# Install Windows service
Write-Host "Installing Windows service..."
$service = Get-Service -Name $ServiceName -ErrorAction SilentlyContinue
if ($service) {
    Stop-Service -Name $ServiceName -Force -ErrorAction SilentlyContinue
    sc.exe delete $ServiceName
    Start-Sleep -Seconds 2
}

# Set up environment variables for the service
$env:CONTROL_PLANE_URL = $ControlPlaneUrl
$env:TENANT_ID = $TenantId
$env:PROJECT_ID = $ProjectId
$env:AGENT_ID = $AgentId
$env:JWT_TOKEN = $JwtToken
$env:LOG_LEVEL = "info"

# Create service with environment variables embedded in the binary path
# Windows services don't directly support environment files, so we'll use a wrapper approach
$wrapperScript = @"
@echo off
setlocal
set CONTROL_PLANE_URL=$ControlPlaneUrl
set TENANT_ID=$TenantId
set PROJECT_ID=$ProjectId
set AGENT_ID=$AgentId
set JWT_TOKEN=$JwtToken
set LOG_LEVEL=info
"%~dp0$BinaryName"
"@
$wrapperPath = "$InstallDir\start-agent.cmd"
Set-Content -Path $wrapperPath -Value $wrapperScript

$binPath = $wrapperPath
New-Service -Name $ServiceName `
    -DisplayName $DisplayName `
    -BinaryPathName $binPath `
    -StartupType Automatic `
    -Description "Automation Agent Service" | Out-Null

# Configure service recovery
sc.exe failure $ServiceName reset= 86400 actions= restart/60000/restart/60000/restart/60000

# Start service
Write-Host "Starting service..."
Start-Service -Name $ServiceName

Write-Host "Installation complete!"
Get-Service -Name $ServiceName
