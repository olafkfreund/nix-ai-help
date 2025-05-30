#!/usr/bin/env python3
"""
Test script for nixai MCP protocol functionality.
Tests JSON-RPC2 communication over Unix socket.
"""
import json
import socket
import time

def test_mcp_protocol():
    """Test MCP protocol over Unix socket."""
    socket_path = "/tmp/nixai-mcp.sock"
    
    print("🧪 Testing MCP Protocol over Unix Socket")
    print("=" * 50)
    
    try:
        # Connect to Unix socket
        sock = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
        sock.connect(socket_path)
        print("✅ Connected to Unix socket")
        
        # Test 1: Initialize request
        print("\n📋 Test 1: Initialize Request")
        initialize_request = {
            "jsonrpc": "2.0",
            "id": 1,
            "method": "initialize",
            "params": {
                "protocolVersion": "2024-11-05",
                "capabilities": {
                    "experimental": {},
                    "sampling": {}
                },
                "clientInfo": {
                    "name": "test-client",
                    "version": "1.0.0"
                }
            }
        }
        
        request_data = json.dumps(initialize_request) + "\n"
        sock.sendall(request_data.encode('utf-8'))
        
        # Read response
        response_data = sock.recv(4096).decode('utf-8')
        if response_data:
            print("✅ Received initialize response")
            try:
                response = json.loads(response_data.strip())
                print(f"   Response ID: {response.get('id')}")
                if 'result' in response:
                    print("   ✅ Initialize successful")
                else:
                    print("   ⚠️  Initialize response without result")
            except json.JSONDecodeError:
                print("   ⚠️  Response not valid JSON")
        else:
            print("   ❌ No response received")
        
        # Test 2: Tools list request
        print("\n📋 Test 2: Tools List Request")
        tools_request = {
            "jsonrpc": "2.0",
            "id": 2,
            "method": "tools/list",
            "params": {}
        }
        
        request_data = json.dumps(tools_request) + "\n"
        sock.sendall(request_data.encode('utf-8'))
        
        # Read response
        time.sleep(0.1)  # Small delay for response
        response_data = sock.recv(4096).decode('utf-8')
        if response_data:
            print("✅ Received tools list response")
            try:
                response = json.loads(response_data.strip())
                if 'result' in response and 'tools' in response['result']:
                    tools = response['result']['tools']
                    print(f"   Found {len(tools)} tools:")
                    for tool in tools:
                        print(f"     - {tool.get('name', 'unknown')}")
                else:
                    print("   ⚠️  No tools in response")
            except json.JSONDecodeError:
                print("   ⚠️  Response not valid JSON")
        else:
            print("   ❌ No response received")
        
        # Test 3: Tool call request
        print("\n📋 Test 3: Tool Call Request")
        tool_call_request = {
            "jsonrpc": "2.0",
            "id": 3,
            "method": "tools/call",
            "params": {
                "name": "query_nixos_docs",
                "arguments": {
                    "query": "services.nginx.enable"
                }
            }
        }
        
        request_data = json.dumps(tool_call_request) + "\n"
        sock.sendall(request_data.encode('utf-8'))
        
        # Read response
        time.sleep(0.2)  # Longer delay for query processing
        response_data = sock.recv(8192).decode('utf-8')
        if response_data:
            print("✅ Received tool call response")
            try:
                response = json.loads(response_data.strip())
                if 'result' in response:
                    result = response['result']
                    if isinstance(result, dict) and 'content' in result:
                        content = result['content']
                        if isinstance(content, list) and len(content) > 0:
                            text_content = content[0].get('text', '')
                            print(f"   Result: {text_content[:100]}...")
                        else:
                            print("   ⚠️  Unexpected content format")
                    else:
                        print(f"   Result: {str(result)[:100]}...")
                else:
                    print("   ⚠️  No result in response")
            except json.JSONDecodeError as e:
                print(f"   ⚠️  Response not valid JSON: {e}")
        else:
            print("   ❌ No response received")
        
        sock.close()
        print("\n🎉 MCP Protocol Tests Complete!")
        
    except Exception as e:
        print(f"❌ Error testing MCP protocol: {e}")

if __name__ == "__main__":
    test_mcp_protocol()
