#!/usr/bin/env python3
"""
Main VS Code integration test runner for NixAI project

This script runs all VS Code integration tests in the appropriate order
"""
import sys
import os
import importlib
import subprocess

def print_header(message):
    """Print a formatted header"""
    print("\n" + "=" * 60)
    print(f"🧪 {message}")
    print("=" * 60)

def run_python_test(module_name):
    """Run a Python test module by importing and executing it"""
    try:
        # Import the module dynamically
        module = importlib.import_module(module_name)
        
        # Find and run the main test function
        # Convention: test modules should have a main function with the same name as the file
        test_func_name = module_name.split('.')[-1]  # e.g., 'test_vscode_direct' from 'tests.vscode.test_vscode_direct'
        
        if hasattr(module, test_func_name):
            test_func = getattr(module, test_func_name)
            return test_func()
        elif hasattr(module, 'main'):
            return module.main()
        else:
            print(f"❌ Error: No test function found in {module_name}")
            return False
    except Exception as e:
        print(f"❌ Error running {module_name}: {e}")
        return False

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
    print_header("Running All VS Code Integration Tests")
    
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
    
    # Direct VS Code tests
    run_test("VS Code Direct Test", run_python_test, "tests.vscode.test_vscode_direct")
    
    # VS Code MCP integration tests
    run_test("VS Code MCP Integration Test", run_python_test, "tests.vscode.test_vscode_mcp_integration")
    
    # VS Code shell script tests
    script_path_mcp = os.path.join(os.path.dirname(__file__), "test-vscode-mcp.sh")
    script_path_live = os.path.join(os.path.dirname(__file__), "test-vscode-live.sh")
    
    run_test("VS Code MCP Shell Test", run_shell_test, script_path_mcp)
    run_test("VS Code Live Test", run_shell_test, script_path_live)
    
    # Print summary
    print_header("Test Summary")
    print(f"Total tests: {passed + failed}")
    print(f"Passed: {passed}")
    print(f"Failed: {failed}")
    
    return failed == 0

if __name__ == "__main__":
    success = main()
    sys.exit(0 if success else 1)
