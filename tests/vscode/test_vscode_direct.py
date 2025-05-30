#!/usr/bin/env python3
"""
VS Code MCP Integration Test
Tests if VS Code can connect to the nixai MCP server.
"""
import subprocess
import json
import time
import os

def test_vscode_mcp_connection():
    """Test MCP connection using the same method VS Code would use."""
    print("🧪 Testing VS Code MCP Connection")
    print("=" * 50)
    
    # Test 1: Direct script test
    print("\n📋 Test 1: Bridge Script Test")
    script_path = "/home/olafkfreund/Source/NIX/nix-ai-help/scripts/mcp-bridge.sh"
    
    if not os.path.exists(script_path):
        print("❌ Bridge script not found")
        return False
        
    try:
        # Test initialize request
        initialize_request = {
            "jsonrpc": "2.0",
            "id": 1,
            "method": "initialize",
            "params": {
                "protocolVersion": "2024-11-05",
                "capabilities": {},
                "clientInfo": {
                    "name": "vscode-test-client",
                    "version": "1.0.0"
                }
            }
        }
        
        # Run the bridge script
        process = subprocess.Popen(
            [script_path],
            stdin=subprocess.PIPE,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True
        )
        
        # Send request
        request_data = json.dumps(initialize_request) + "\n"
        stdout, stderr = process.communicate(input=request_data, timeout=10)
        
        if stdout:
            print("✅ Bridge script responding")
            try:
                response = json.loads(stdout.strip())
                if 'result' in response:
                    print("   ✅ Valid MCP response received")
                    print(f"   Server: {response['result'].get('serverInfo', {}).get('name', 'unknown')}")
                    return True
                else:
                    print("   ⚠️  Response without result")
            except json.JSONDecodeError:
                print("   ⚠️  Invalid JSON response")
        else:
            print("   ❌ No response from bridge script")
            if stderr:
                print(f"   Error: {stderr}")
                
    except subprocess.TimeoutExpired:
        print("   ❌ Bridge script timed out")
        process.kill()
    except Exception as e:
        print(f"   ❌ Error testing bridge script: {e}")
        
    return False

def test_socat_direct():
    """Test socat command directly as VS Code would use it."""
    print("\n📋 Test 2: Direct socat Test")
    
    try:
        # Test with socat command like VS Code uses
        cmd = ["socat", "STDIO", "UNIX-CONNECT:/tmp/nixai-mcp.sock"]
        
        initialize_request = {
            "jsonrpc": "2.0",
            "id": 1,
            "method": "initialize", 
            "params": {
                "protocolVersion": "2024-11-05",
                "capabilities": {},
                "clientInfo": {
                    "name": "vscode-direct-test",
                    "version": "1.0.0"
                }
            }
        }
        
        process = subprocess.Popen(
            cmd,
            stdin=subprocess.PIPE,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True
        )
        
        request_data = json.dumps(initialize_request) + "\n"
        stdout, stderr = process.communicate(input=request_data, timeout=10)
        
        if stdout:
            print("✅ Direct socat responding")
            try:
                response = json.loads(stdout.strip())
                if 'result' in response:
                    print("   ✅ Valid MCP response received")
                    return True
            except json.JSONDecodeError:
                print("   ⚠️  Invalid JSON response")
        else:
            print("   ❌ No response from socat")
            if stderr:
                print(f"   Error: {stderr}")
                
    except subprocess.TimeoutExpired:
        print("   ❌ socat timed out")
        process.kill()
    except Exception as e:
        print(f"   ❌ Error testing socat: {e}")
        
    return False

if __name__ == "__main__":
    success1 = test_vscode_mcp_connection()
    success2 = test_socat_direct()
    
    print(f"\n🎯 Test Results:")
    print(f"Bridge Script: {'✅ PASS' if success1 else '❌ FAIL'}")
    print(f"Direct socat: {'✅ PASS' if success2 else '❌ FAIL'}")
    
    if success1 and success2:
        print("\n🎉 MCP server is ready for VS Code integration!")
        print("VS Code extensions should be able to connect.")
    else:
        print("\n⚠️  There may be issues with VS Code MCP integration.")
