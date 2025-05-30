#!/usr/bin/env python3
"""
Test raw socket communication to see what's being sent/received
"""
import socket
import json

def test_raw_communication():
    socket_path = "/tmp/nixai-mcp.sock"
    
    print(f"Connecting to {socket_path}")
    
    try:
        sock = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
        sock.settimeout(5)  # 5 second timeout
        sock.connect(socket_path)
        print("Connected successfully!")
        
        # Send a simple JSON-RPC2 initialize request
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
        
        # Convert to JSON and add newline (some JSON-RPC implementations expect this)
        message = json.dumps(request) + "\n"
        print(f"Sending: {message.strip()}")
        
        sock.send(message.encode('utf-8'))
        print("Message sent, waiting for response...")
        
        # Try to receive response
        sock.settimeout(10)  # Give more time for response
        response_data = sock.recv(4096)
        print(f"Received raw bytes: {response_data}")
        
        if response_data:
            try:
                response_str = response_data.decode('utf-8')
                print(f"Decoded response: {response_str}")
                response_json = json.loads(response_str)
                print(f"Parsed response: {json.dumps(response_json, indent=2)}")
            except Exception as e:
                print(f"Error parsing response: {e}")
        else:
            print("No response received")
        
        sock.close()
        print("Connection closed")
        
    except socket.timeout:
        print("Connection timed out waiting for response")
    except Exception as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    test_raw_communication()
