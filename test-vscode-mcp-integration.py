#!/usr/bin/env python3
"""
VS Code MCP Integration Test
Test script to validate MCP server integration with VS Code extensions
"""

import json
import socket
import subprocess
import time
import os
import sys

def test_socket_connection():
    """Test basic socket connectivity"""
    print("🔌 Testing Unix socket connection...")
    try:
        sock = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
        sock.connect('/tmp/nixai-mcp.sock')
        print("✅ Unix socket connection successful")
        sock.close()
        return True
    except Exception as e:
        print(f"❌ Socket connection failed: {e}")
        return False

def test_mcp_protocol():
    """Test MCP protocol communication"""
    print("🧪 Testing MCP protocol...")
    try:
        sock = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
        sock.connect('/tmp/nixai-mcp.sock')
        
        # Send initialize request
        init_request = {
            "jsonrpc": "2.0",
            "id": 1,
            "method": "initialize",
            "params": {
                "protocolVersion": "2024-11-05",
                "capabilities": {
                    "roots": {"listChanged": True},
                    "sampling": {}
                },
                "clientInfo": {
                    "name": "vscode-test",
                    "version": "1.0.0"
                }
            }
        }
        
        message = json.dumps(init_request) + '\n'
        sock.send(message.encode())
        
        response = sock.recv(4096).decode().strip()
        data = json.loads(response)
        
        if 'result' in data:
            print("✅ MCP protocol initialization successful")
            sock.close()
            return True
        else:
            print(f"❌ MCP protocol failed: {data}")
            sock.close()
            return False
            
    except Exception as e:
        print(f"❌ MCP protocol test failed: {e}")
        return False

def check_vscode_settings():
    """Check VS Code MCP configuration files"""
    print("⚙️  Checking VS Code MCP configuration...")
    
    settings_files = [
        '/home/olafkfreund/Source/NIX/nix-ai-help/.vscode/settings.json',
        '/home/olafkfreund/Source/NIX/nix-ai-help/.vscode/mcp-settings.json',
        '/home/olafkfreund/.config/Code/User/mcp-settings.json'
    ]
    
    for settings_file in settings_files:
        if os.path.exists(settings_file):
            print(f"✅ Found: {settings_file}")
            try:
                with open(settings_file, 'r') as f:
                    config = json.load(f)
                    if 'mcpServers' in config or 'mcp.servers' in config:
                        print(f"✅ MCP configuration found in {settings_file}")
                    else:
                        print(f"⚠️  No MCP configuration in {settings_file}")
            except Exception as e:
                print(f"❌ Error reading {settings_file}: {e}")
        else:
            print(f"❌ Missing: {settings_file}")

def check_extensions():
    """Check installed VS Code extensions"""
    print("🧩 Checking VS Code extensions...")
    try:
        result = subprocess.run(['code', '--list-extensions'], 
                              capture_output=True, text=True, check=True)
        extensions = result.stdout.strip().split('\n')
        
        mcp_extensions = [
            'automatalabs.copilot-mcp',
            'saoudrizwan.claude-dev', 
            'zebradev.mcp-server-runner'
        ]
        
        for ext in mcp_extensions:
            if ext in extensions:
                print(f"✅ Installed: {ext}")
            else:
                print(f"❌ Missing: {ext}")
                
    except Exception as e:
        print(f"❌ Error checking extensions: {e}")

def test_vs_code_mcp_integration():
    """Main test function"""
    print("🚀 VS Code MCP Integration Test")
    print("=" * 50)
    
    # Test 1: Socket connection
    socket_ok = test_socket_connection()
    print()
    
    # Test 2: MCP protocol
    protocol_ok = test_mcp_protocol()
    print()
    
    # Test 3: VS Code settings
    check_vscode_settings()
    print()
    
    # Test 4: Extensions
    check_extensions()
    print()
    
    # Summary
    print("📋 Test Summary")
    print("=" * 20)
    if socket_ok and protocol_ok:
        print("✅ MCP Server: READY")
        print("✅ Protocol: WORKING")
        print("📝 Next steps:")
        print("   1. Open VS Code in this workspace")
        print("   2. Reload window (Ctrl+Shift+P -> Developer: Reload Window)")
        print("   3. Check MCP extensions for nixai server")
        print("   4. Test MCP tools in VS Code chat/copilot")
    else:
        print("❌ MCP Server: ISSUES DETECTED")
        print("🔧 Please fix server issues before VS Code testing")

if __name__ == "__main__":
    test_vs_code_mcp_integration()
