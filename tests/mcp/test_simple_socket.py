#!/usr/bin/env python3
"""
Simple test to check if we can connect to the Unix socket
Part of the NixAI MCP test suite
"""
import socket
import sys

def test_simple_socket():
    """Test simple socket connection to MCP server"""
    socket_path = "/tmp/nixai-mcp.sock"
    
    print("🧪 Testing MCP Socket Connection")
    print("=" * 40)
    print(f"Socket path: {socket_path}")
    
    try:
        sock = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
        print("✅ Socket created successfully")
        
        sock.settimeout(5)  # 5 second timeout
        sock.connect(socket_path)
        print("✅ Connected successfully!")
        
        sock.close()
        print("✅ Connection closed properly")
        return True
        
    except socket.timeout:
        print("❌ Connection timed out")
        return False
    except Exception as e:
        print(f"❌ Error: {e}")
        return False

# Entry point for direct execution
if __name__ == "__main__":
    success = test_simple_socket()
    sys.exit(0 if success else 1)
