# Run installation as Administrator
# This script will elevate and run the installer

$scriptPath = Split-Path -Parent $MyInvocation.MyCommand.Path
$installScript = Join-Path $scriptPath "install.ps1"

# Configuration
$ControlPlaneUrl = "http://localhost:8081"
$TenantId = "test-tenant"
$ProjectId = "test-project"
$AgentId = "agent-windows-01"
$JwtToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhZ2VudF9pZCI6ImFnZW50LXdpbmRvd3MtMDEiLCJ0ZW5hbnRfaWQiOiJ0ZXN0LXRlbmFudCIsInByb2plY3RfaWQiOiJ0ZXN0LXByb2plY3QiLCJleHAiOjE3OTk0OTUwMDksImlhdCI6MTc2Nzk1OTAwOX0.JKQXv4YeRRA46gPU-cJpyV83FC2ZFXxWrR_M1zkuQO0"

# Check if running as Administrator
$isAdmin = ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)

if (-not $isAdmin) {
    Write-Host "Elevating to Administrator..."
    $arguments = "-ExecutionPolicy Bypass -File `"$installScript`" -ControlPlaneUrl `"$ControlPlaneUrl`" -TenantId `"$TenantId`" -ProjectId `"$ProjectId`" -AgentId `"$AgentId`" -JwtToken `"$JwtToken`""
    Start-Process powershell.exe -Verb RunAs -ArgumentList $arguments -Wait
    exit
}

# If already admin, run directly
& $installScript -ControlPlaneUrl $ControlPlaneUrl -TenantId $TenantId -ProjectId $ProjectId -AgentId $AgentId -JwtToken $JwtToken
