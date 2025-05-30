#!/usr/bin/env python3
"""
Simple test to check if we can connect to the Unix socket
"""
import socket
import sys

def test_socket_connection():
    socket_path = "/tmp/nixai-mcp.sock"
    
    print(f"Testing connection to {socket_path}")
    
    try:
        sock = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
        print("Socket created")
        
        sock.settimeout(5)  # 5 second timeout
        sock.connect(socket_path)
        print("Connected successfully!")
        
        sock.close()
        print("Connection closed")
        
    except socket.timeout:
        print("Connection timed out")
    except Exception as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    test_socket_connection()
