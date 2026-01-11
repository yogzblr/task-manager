"""
Windows PowerShell Workflow Test Script
Tests probe workflow execution on Windows agent and searches Quickwit for results
"""

import requests
import json
import time
import sys
import os
from datetime import datetime

# Configuration (from environment or defaults)
CONTROL_PLANE_URL = os.getenv("CONTROL_PLANE_URL", "http://localhost:8081")
QUICKWIT_URL = os.getenv("QUICKWIT_URL", "http://localhost:7280")
TENANT_ID = os.getenv("TENANT_ID", "test-tenant")
PROJECT_ID = os.getenv("PROJECT_ID", "test-project")
JWT_TOKEN = os.getenv("JWT_TOKEN", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhZ2VudF9pZCI6ImFnZW50LWxpbnV4LTAxIiwidGVuYW50X2lkIjoidGVzdC10ZW5hbnQiLCJwcm9qZWN0X2lkIjoidGVzdC1wcm9qZWN0IiwiZXhwIjoxNzk5NDk1MDA5LCJpYXQiOjE3Njc5NTkwMDl9.JKQXv4YeRRA46gPU-cJpyV83FC2ZFXxWrR_M1zkuQO0")

# Windows PowerShell workflow
WINDOWS_WORKFLOW = """
name: Windows PowerShell Test

tasks:
  - name: PowerShell environment check
    type: powershell
    script: |
      Write-Host "=== PowerShell Environment ==="
      Write-Host "PowerShell Version: $($PSVersionTable.PSVersion)"
      Write-Host "Computer Name: $env:COMPUTERNAME"
      Write-Host "User Name: $env:USERNAME"
      Write-Host "OS: $env:OS"
      Write-Host "Processor: $env:PROCESSOR_ARCHITECTURE"
    timeout: 30s

  - name: Test system information
    type: powershell
    script: |
      Write-Host "=== System Information ==="
      Get-ComputerInfo | Select-Object CsName, WindowsVersion, OsArchitecture | Format-List
      Write-Host "`n=== Memory Information ==="
      $memory = Get-CimInstance Win32_OperatingSystem
      Write-Host "Total Memory: $([math]::Round($memory.TotalVisibleMemorySize/1MB, 2)) GB"
      Write-Host "Free Memory: $([math]::Round($memory.FreePhysicalMemory/1MB, 2)) GB"
    timeout: 30s

  - name: Test process listing
    type: powershell
    script: |
      Write-Host "=== Top 10 Processes by CPU ==="
      Get-Process | Sort-Object CPU -Descending | Select-Object -First 10 ProcessName, CPU, WorkingSet | Format-Table -AutoSize
    timeout: 30s

  - name: Test file operations
    type: powershell
    script: |
      Write-Host "=== Testing File Operations ==="
      $testFile = "$env:TEMP\\probe-test-$(Get-Date -Format 'yyyyMMdd-HHmmss').txt"
      "Test content from probe workflow" | Out-File -FilePath $testFile
      Write-Host "Created test file: $testFile"
      Write-Host "File content:"
      Get-Content $testFile
      Remove-Item $testFile
      Write-Host "Test file cleaned up"
    timeout: 30s

  - name: Test output variables
    type: powershell
    script: |
      Write-Host "=== Testing Variables ==="
      $token = "secret-token-123"
      $userId = "windows-user-456"
      Write-Host "Token: $token"
      Write-Host "User ID: $userId"
      Write-Host "Test completed at: $(Get-Date)"
    timeout: 30s
"""


class WindowsWorkflowTester:
    def __init__(self):
        self.session = requests.Session()
        self.job_id = None

    def submit_workflow(self):
        """Submit workflow to control plane"""
        print("=" * 60)
        print("Submitting Windows PowerShell Workflow")
        print("=" * 60)
        
        payload = {
            "tenant_id": TENANT_ID,
            "project_id": PROJECT_ID,
            "workflow": WINDOWS_WORKFLOW,
            "workflow_format": "yaml"
        }
        
        try:
            headers = {
                "Authorization": f"Bearer {JWT_TOKEN}",
                "Content-Type": "application/json"
            }
            response = self.session.post(
                f"{CONTROL_PLANE_URL}/api/jobs",
                json=payload,
                headers=headers,
                timeout=10
            )
            
            if response.status_code == 201:
                job_data = response.json()
                self.job_id = job_data.get("job_id")
                print(f"✓ Workflow submitted successfully!")
                print(f"  Job ID: {self.job_id}")
                return True
            else:
                print(f"✗ Failed to submit workflow: {response.status_code}")
                print(f"  Response: {response.text}")
                return False
                
        except Exception as e:
            print(f"✗ Error submitting workflow: {e}")
            return False

    def check_job_status(self, timeout=60):
        """Poll job status until completion"""
        print("\n" + "=" * 60)
        print("Monitoring Job Execution")
        print("=" * 60)
        
        if not self.job_id:
            print("✗ No job ID available")
            return False
        
        start_time = time.time()
        
        while time.time() - start_time < timeout:
            try:
                headers = {"Authorization": f"Bearer {JWT_TOKEN}"}
                response = self.session.get(
                    f"{CONTROL_PLANE_URL}/api/jobs/{self.job_id}",
                    params={"tenant_id": TENANT_ID},
                    headers=headers,
                    timeout=5
                )
                
                if response.status_code == 200:
                    job = response.json()
                    state = job.get("state", "unknown")
                    
                    print(f"  Status: {state} ({int(time.time() - start_time)}s elapsed)")
                    
                    if state == "completed":
                        print("✓ Job completed successfully!")
                        return True
                    elif state == "failed":
                        print("✗ Job failed")
                        return False
                    
                    time.sleep(2)
                else:
                    print(f"  Warning: Status check returned {response.status_code}")
                    time.sleep(2)
                    
            except Exception as e:
                print(f"  Error checking status: {e}")
                time.sleep(2)
        
        print("✗ Job timeout reached")
        return False

    def search_quickwit(self):
        """Search Quickwit for job execution logs"""
        print("\n" + "=" * 60)
        print("Searching Quickwit for Execution Logs")
        print("=" * 60)
        
        if not self.job_id:
            print("✗ No job ID to search for")
            return
        
        # Wait for logs to be indexed
        time.sleep(3)
        
        # Build Quickwit query
        query = {
            "query": f"job_id:{self.job_id} OR agent_id:agent-windows-01",
            "max_hits": 100,
            "sort_by": "timestamp"
        }
        
        try:
            response = self.session.post(
                f"{QUICKWIT_URL}/api/v1/automation-logs/search",
                json=query,
                timeout=10
            )
            
            if response.status_code == 200:
                results = response.json()
                hits = results.get("hits", [])
                
                print(f"✓ Found {len(hits)} log entries")
                print()
                
                for i, hit in enumerate(hits[:10], 1):
                    source = hit.get("_source", {})
                    timestamp = source.get("timestamp", "N/A")
                    level = source.get("level", "INFO")
                    message = source.get("message", "")
                    
                    print(f"  [{i}] {timestamp} [{level}]")
                    print(f"      {message[:100]}...")
                    print()
                
                if len(hits) > 10:
                    print(f"  ... and {len(hits) - 10} more entries")
                
            else:
                print(f"⚠ Quickwit query returned {response.status_code}")
                print("  Note: Logs may not be indexed yet or Quickwit may not be fully configured")
                print(f"  Response: {response.text[:200]}")
                
        except requests.exceptions.ConnectionError:
            print("⚠ Could not connect to Quickwit")
            print("  Quickwit may not be running or not yet configured")
        except Exception as e:
            print(f"⚠ Error querying Quickwit: {e}")

    def check_windows_agent_status(self):
        """Check if Windows agent is running"""
        print("\n" + "=" * 60)
        print("Windows Agent Status")
        print("=" * 60)
        
        try:
            # Check database for Windows agent
            print("Checking agent registration in database...")
            print("  (This would query the control plane API)")
            print()
            print("To verify Windows agent is running:")
            print("  1. Check Task Manager for 'automation-agent-windows.exe'")
            print("  2. Or run: Get-Process | Where-Object {$_.ProcessName -like '*automation-agent*'}")
            
        except Exception as e:
            print(f"  Error: {e}")

    def run_test(self):
        """Run complete test"""
        print("\n")
        print("=" * 60)
        print("WINDOWS POWERSHELL WORKFLOW TEST")
        print(f"Started at: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
        print("=" * 60)
        
        # Check if Windows agent is available
        print("\n⚠ IMPORTANT:")
        print("  This test requires the Windows agent to be running!")
        print("  Location: C:\\Users\\yoges\\OneDrive\\Documents\\My Code\\Task Manager\\demo\\automation-agent")
        print("  Run: .\\automation-agent-windows.exe (with proper JWT token)")
        print()
        
        response = input("Is the Windows agent running? (y/n): ")
        if response.lower() != 'y':
            print("\nPlease start the Windows agent first.")
            print("See: WINDOWS-AGENT-SUCCESS.md for instructions")
            return False
        
        # Step 1: Submit workflow
        if not self.submit_workflow():
            print("\n✗ Test failed at workflow submission")
            return False
        
        # Step 2: Monitor execution
        time.sleep(2)
        if not self.check_job_status(timeout=90):
            print("\n✗ Test failed during execution")
        
        # Step 3: Search logs
        self.search_quickwit()
        
        # Step 4: Check agent status
        self.check_windows_agent_status()
        
        print("\n" + "=" * 60)
        print("TEST COMPLETE")
        print("=" * 60)
        return True


if __name__ == "__main__":
    print("""
╔════════════════════════════════════════════════════════════╗
║   Windows PowerShell Workflow Test - Probe Integration    ║
╚════════════════════════════════════════════════════════════╝

This script will:
1. Submit a Windows PowerShell workflow to the control plane
2. Monitor job execution status
3. Query Quickwit for execution logs
4. Display results

Prerequisites:
- Control plane running at: http://localhost:8081
- Windows agent running (native binary)
- Quickwit running at: http://localhost:7280

""")
    
    # Check if control plane is accessible
    try:
        response = requests.get(f"{CONTROL_PLANE_URL}/health", timeout=5)
        print("✓ Control plane is accessible\n")
    except:
        print("✗ Control plane is not accessible!")
        print(f"  URL: {CONTROL_PLANE_URL}")
        print("  Make sure Docker services are running in WSL")
        sys.exit(1)
    
    # Run test
    tester = WindowsWorkflowTester()
    success = tester.run_test()
    
    sys.exit(0 if success else 1)
