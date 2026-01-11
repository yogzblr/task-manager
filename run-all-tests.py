#!/usr/bin/env python3
"""
Quick Test Runner - Runs both Linux and Windows workflow tests
"""

import subprocess
import sys
import os

def run_test(script_name, description):
    """Run a test script and report results"""
    print("\n" + "=" * 70)
    print(f"  {description}")
    print("=" * 70)
    
    try:
        result = subprocess.run(
            [sys.executable, script_name],
            capture_output=False,
            text=True
        )
        return result.returncode == 0
    except Exception as e:
        print(f"Error running {script_name}: {e}")
        return False

def main():
    print("""
╔══════════════════════════════════════════════════════════════════╗
║          Workflow Test Suite - Linux & Windows Agents           ║
╚══════════════════════════════════════════════════════════════════╝
    """)
    
    # Check script location
    demo_dir = os.path.dirname(os.path.abspath(__file__))
    os.chdir(demo_dir)
    
    results = {}
    
    # Test 1: Linux workflow
    print("\n[1/2] Running Linux Shell Workflow Test...")
    results['linux'] = run_test('test-linux-workflow.py', 'Linux Shell Test')
    
    # Test 2: Windows workflow
    print("\n[2/2] Running Windows PowerShell Workflow Test...")
    print("Note: This requires Windows agent to be running")
    
    response = input("\nRun Windows test? (y/n): ")
    if response.lower() == 'y':
        results['windows'] = run_test('test-windows-workflow.py', 'Windows PowerShell Test')
    else:
        print("Skipping Windows test")
        results['windows'] = None
    
    # Summary
    print("\n" + "=" * 70)
    print("  TEST SUMMARY")
    print("=" * 70)
    print(f"  Linux Test:   {'✓ PASSED' if results['linux'] else '✗ FAILED'}")
    if results['windows'] is not None:
        print(f"  Windows Test: {'✓ PASSED' if results['windows'] else '✗ FAILED'}")
    else:
        print(f"  Windows Test: ⊘ SKIPPED")
    print("=" * 70)
    
    # Exit code
    if results['linux'] and (results['windows'] is None or results['windows']):
        print("\n✓ All tests passed!")
        return 0
    else:
        print("\n✗ Some tests failed")
        return 1

if __name__ == "__main__":
    sys.exit(main())
