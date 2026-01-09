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

# Create configuration file
$ConfigFile = "$ConfigDir\config.yaml"
if (-not (Test-Path $ConfigFile)) {
    Write-Host "Creating configuration file..."
    $config = @"
control_plane_url: $ControlPlaneUrl
tenant_id: $TenantId
project_id: $ProjectId
agent_id: $AgentId
jwt_token: $JwtToken
log_level: info
"@
    Set-Content -Path $ConfigFile -Value $config
}

# Install Windows service
Write-Host "Installing Windows service..."
$service = Get-Service -Name $ServiceName -ErrorAction SilentlyContinue
if ($service) {
    Stop-Service -Name $ServiceName -Force -ErrorAction SilentlyContinue
    sc.exe delete $ServiceName
    Start-Sleep -Seconds 2
}

$binPath = "$InstallDir\$BinaryName"
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
