#!/usr/bin/env python3
"""
Test MCP protocol with exact JSON-RPC2 format
"""
import socket
import json

def test_mcp_simple():
    socket_path = "/tmp/nixai-mcp.sock"
    
    print(f"Connecting to {socket_path}")
    
    try:
        sock = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
        sock.settimeout(10)
        sock.connect(socket_path)
        print("Connected!")
        
        # Send initialize request in exact JSON-RPC2 format
        request = {
            "jsonrpc": "2.0",
            "id": 1,
            "method": "initialize",
            "params": {
                "protocolVersion": "2024-11-05",
                "capabilities": {},
                "clientInfo": {
                    "name": "test-client",
                    "version": "1.0.0"
                }
            }
        }
        
        message = json.dumps(request)
        print(f"Sending: {message}")
        
        sock.send(message.encode('utf-8'))
        print("Message sent, waiting for response...")
        
        # Wait for response
        response_data = sock.recv(4096)
        print(f"Received {len(response_data)} bytes")
        
        if response_data:
            response_str = response_data.decode('utf-8')
            print(f"Response: {response_str}")
        else:
            print("No response received")
        
        sock.close()
        print("Connection closed")
        
    except Exception as e:
        print(f"Error: {e}")
        import traceback
        traceback.print_exc()

if __name__ == "__main__":
    test_mcp_simple()
