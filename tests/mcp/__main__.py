#!/usr/bin/env python3
"""
Main MCP test runner for NixAI project

This script runs all MCP-related tests in the appropriate order
"""
import sys
import os
import importlib
import subprocess
import time

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
        test_func_name = module_name.split('.')[-1]  # e.g., 'test_mcp_protocol' from 'tests.mcp.test_mcp_protocol'
        
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
    print_header("Running All MCP Tests")
    
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
    
    # Basic socket connectivity test
    run_test("Simple Socket Test", run_python_test, "tests.mcp.test_simple_socket")
    
    # MCP protocol tests
    run_test("Raw Socket Test", run_python_test, "tests.mcp.test_raw_socket")
    run_test("MCP Simple Test", run_python_test, "tests.mcp.test_mcp_simple")
    run_test("MCP Protocol Test", run_python_test, "tests.mcp.test_mcp_protocol")
    
    # MCP server test
    script_path = os.path.join(os.path.dirname(__file__), "test-mcp-server.sh")
    run_test("MCP Server Test", run_shell_test, script_path)
    
    # Print summary
    print_header("Test Summary")
    print(f"Total tests: {passed + failed}")
    print(f"Passed: {passed}")
    print(f"Failed: {failed}")
    
    return failed == 0

if __name__ == "__main__":
    success = main()
    sys.exit(0 if success else 1)
