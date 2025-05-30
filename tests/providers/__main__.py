#!/usr/bin/env python3
"""
Main AI provider integration test runner for NixAI project

This script runs all AI provider integration tests
"""
import sys
import os
import subprocess

def print_header(message):
    """Print a formatted header"""
    print("\n" + "=" * 60)
    print(f"🧪 {message}")
    print("=" * 60)

def run_shell_test(script_path):
    """Run a shell test script"""
    try:
        result = subprocess.run([script_path], check=False)
        return result.returncode == 0
    except Exception as e:
        print(f"❌ Error running {script_path}: {e}")
        return False

def main():
    """Main test runner function"""
    print_header("Running All AI Provider Tests")
    
    # Track test results
    passed = 0
    failed = 0
    
    # Helper function to run a test and track results
    def run_test(name, test_func, *args):
        nonlocal passed, failed
        print(f"\n📌 Running test: {name}")
        
        if test_func(*args):
            print(f"✅ {name} PASSED")
            passed += 1
        else:
            print(f"❌ {name} FAILED")
            failed += 1
        
        print("-" * 40)
    
    # Run the main all providers test
    script_path = os.path.join(os.path.dirname(__file__), "test-all-providers.sh")
    run_test("All Providers Test", run_shell_test, script_path)
    
    # Print summary
    print_header("Test Summary")
    print(f"Total tests: {passed + failed}")
    print(f"Passed: {passed}")
    print(f"Failed: {failed}")
    
    return failed == 0

if __name__ == "__main__":
    success = main()
    sys.exit(0 if success else 1)
