#!/usr/bin/env python3
"""
VS Code MCP Integration Diagnostic Tool
Checks if the nixai MCP server is properly configured and
provides guidance for VS Code integration.
"""

import os
import sys
import json
import socket
import subprocess
import time
from pathlib import Path

RESET = "\033[0m"
BOLD = "\033[1m"
GREEN = "\033[32m"
YELLOW = "\033[33m"
RED = "\033[31m"
BLUE = "\033[34m"

def print_success(msg):
    print(f"{GREEN}✓ {msg}{RESET}")

def print_warning(msg):
    print(f"{YELLOW}⚠ {msg}{RESET}")

def print_error(msg):
    print(f"{RED}✗ {msg}{RESET}")

def print_info(msg):
    print(f"{BLUE}ℹ {msg}{RESET}")

def print_header(msg):
    print(f"\n{BOLD}{msg}{RESET}")
    print("=" * len(msg))

def check_mcp_server_running():
    """Check if the MCP server process is running."""
    print_header("Checking MCP Server Status")
    
    running = False
    try:
        result = subprocess.run(
            ["pgrep", "-f", "nixai mcp-server"],
            capture_output=True,
            text=True
        )
        if result.returncode == 0 and result.stdout.strip():
            print_success("MCP server is running")
            running = True
        else:
            print_error("MCP server is not running")
            print_info("Start it with: nixai mcp-server start")
            running = False
    except Exception as e:
        print_error(f"Error checking server status: {e}")
        running = False
    
    return running

def check_unix_socket():
    """Check if the Unix socket exists and has correct permissions."""
    print_header("Checking Unix Socket")
    
    socket_path = "/tmp/nixai-mcp.sock"
    if os.path.exists(socket_path):
        if os.path.getsize(socket_path) == 0:
            print_warning(f"Socket file exists but has zero size: {socket_path}")
            return False
            
        try:
            mode = os.stat(socket_path).st_mode
            if mode & 0o777 != 0o755:  # srwxr-xr-x
                print_warning(f"Socket has unusual permissions: {oct(mode & 0o777)}")
            
            print_success(f"Unix socket exists: {socket_path}")
            return True
        except Exception as e:
            print_error(f"Error checking socket: {e}")
            return False
    else:
        print_error(f"Unix socket not found: {socket_path}")
        return False

def test_mcp_protocol():
    """Test if the MCP protocol is working correctly."""
    print_header("Testing MCP Protocol")
    
    socket_path = "/tmp/nixai-mcp.sock"
    if not os.path.exists(socket_path):
        print_error(f"Socket not found: {socket_path}")
        return False
    
    try:
        # Initialize request
        sock = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
        sock.settimeout(5)
        sock.connect(socket_path)
        
        initialize_request = {
            "jsonrpc": "2.0",
            "id": 1,
            "method": "initialize",
            "params": {
                "protocolVersion": "2024-11-05",
                "capabilities": {},
                "clientInfo": {
                    "name": "mcp-diagnostic-tool",
                    "version": "1.0.0"
                }
            }
        }
        
        request_data = json.dumps(initialize_request) + "\n"
        sock.sendall(request_data.encode('utf-8'))
        
        response_data = sock.recv(4096).decode('utf-8')
        if not response_data:
            print_error("No response received")
            sock.close()
            return False
        
        try:
            response = json.loads(response_data)
            if 'result' in response:
                print_success("MCP protocol working correctly")
                print_info(f"Server name: {response['result']['serverInfo']['name']}")
                print_info(f"Protocol version: {response['result']['protocolVersion']}")
                sock.close()
                
                # Test tools/list
                sock = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
                sock.settimeout(5)
                sock.connect(socket_path)
                
                list_request = {
                    "jsonrpc": "2.0",
                    "id": 2,
                    "method": "tools/list",
                    "params": {}
                }
                
                request_data = json.dumps(list_request) + "\n"
                sock.sendall(request_data.encode('utf-8'))
                
                response_data = sock.recv(4096).decode('utf-8')
                response = json.loads(response_data)
                tools = response['result']['tools']
                print_success(f"Found {len(tools)} tools:")
                for tool in tools:
                    print_info(f"  - {tool['name']}: {tool['description']}")
                
                return True
            else:
                print_error(f"Invalid response: {response}")
                sock.close()
                return False
        except json.JSONDecodeError:
            print_error(f"Invalid JSON response: {response_data}")
            sock.close()
            return False
    except Exception as e:
        print_error(f"Error testing MCP protocol: {e}")
        return False

def check_vscode_config():
    """Check if VS Code is properly configured for MCP."""
    print_header("Checking VS Code Configuration")
    
    workspace_dir = os.getcwd()
    settings_path = os.path.join(workspace_dir, '.vscode', 'settings.json')
    
    if not os.path.exists(settings_path):
        print_error(f"VS Code settings not found: {settings_path}")
        return False
    
    try:
        with open(settings_path, 'r') as f:
            settings = json.load(f)
        
        servers_found = []
        if 'mcp.servers' in settings and 'nixai' in settings['mcp.servers']:
            servers_found.append('mcp.servers')
            
        if 'copilot.mcp.servers' in settings and 'nixai' in settings['copilot.mcp.servers']:
            servers_found.append('copilot.mcp.servers')
            
        if 'claude-dev.mcpServers' in settings and 'nixai' in settings['claude-dev.mcpServers']:
            servers_found.append('claude-dev.mcpServers')
            
        if 'mcpServers' in settings and 'nixai' in settings['mcpServers']:
            servers_found.append('mcpServers')
        
        if servers_found:
            print_success(f"Found MCP server configurations: {', '.join(servers_found)}")
            
            # Check command format
            for server_key in servers_found:
                nixai_config = settings[server_key]['nixai']
                command = nixai_config.get('command', '')
                args = nixai_config.get('args', [])
                
                if command == 'socat' and len(args) >= 2 and 'UNIX-CONNECT:/tmp/nixai-mcp.sock' in args:
                    print_success(f"{server_key} has correct socat command")
                elif command == 'bash' and '-c' in args and 'socat' in ' '.join(args) and 'UNIX-CONNECT:/tmp/nixai-mcp.sock' in ' '.join(args):
                    print_success(f"{server_key} has correct bash+socat command")
                else:
                    print_warning(f"{server_key} may have incorrect command format: {command} {args}")
            
            return True
        else:
            print_error("No MCP server configurations found")
            return False
    except Exception as e:
        print_error(f"Error checking VS Code config: {e}")
        return False

def check_vscode_extensions():
    """Check if required VS Code extensions are installed."""
    print_header("Checking VS Code Extensions")
    
    try:
        result = subprocess.run(
            ["code", "--list-extensions"],
            capture_output=True,
            text=True
        )
        
        if result.returncode != 0:
            print_error("Failed to list VS Code extensions")
            return False
        
        extensions = result.stdout.splitlines()
        
        required_extensions = {
            'automatalabs.copilot-mcp': 'Copilot MCP',
            'zebradev.mcp-server-runner': 'MCP Server Runner',
            'saoudrizwan.claude-dev': 'Claude Dev (Cline)'
        }
        
        found_extensions = []
        missing_extensions = []
        
        for ext_id, ext_name in required_extensions.items():
            if any(e.lower() == ext_id.lower() for e in extensions):
                found_extensions.append(ext_name)
            else:
                missing_extensions.append(f"{ext_name} ({ext_id})")
        
        if found_extensions:
            print_success(f"Found required extensions: {', '.join(found_extensions)}")
        
        if missing_extensions:
            print_warning(f"Missing recommended extensions: {', '.join(missing_extensions)}")
            print_info("Install them from VS Code Extensions marketplace")
        
        return len(found_extensions) > 0
    except Exception as e:
        print_error(f"Error checking VS Code extensions: {e}")
        return False

def create_test_file():
    """Create a test Nix file for testing with VS Code."""
    print_header("Creating Test File")
    
    test_file = "test-mcp-integration.nix"
    
    content = """# Test file for MCP integration
# Try asking an AI assistant (Claude Dev or GitHub Copilot) about these options

{
  # Nginx web server configuration
  services.nginx = {
    enable = true;  # What does this option do?
    
    # What are recommended virtual host settings?
    virtualHosts."example.com" = {
      root = "/var/www/example";
      locations."/" = {
        index = "index.html";
      };
    };
  };
  
  # What does this Home Manager option do?
  programs.git = {
    enable = true;
    userName = "Example User";
    userEmail = "user@example.com";
  };
}
"""
    
    try:
        with open(test_file, 'w') as f:
            f.write(content)
        print_success(f"Created test file: {test_file}")
        print_info("You can use this file to test VS Code MCP integration")
        print_info("Open it and ask an AI assistant about the commented options")
        return True
    except Exception as e:
        print_error(f"Error creating test file: {e}")
        return False

def provide_manual_activation_steps():
    """Provide steps to manually activate MCP integration in VS Code."""
    print_header("VS Code Manual Activation Steps")
    
    print(f"{BOLD}Testing with Claude Dev (Cline):{RESET}")
    print("1. Open VS Code")
    print("2. Open Command Palette (Ctrl+Shift+P)")
    print("3. Type 'Claude Dev: Open Claude Dev' and press Enter")
    print("4. In the chat, ask: 'What does services.nginx.enable do in NixOS?'")
    print("5. If MCP integration works, Claude should query the MCP server")
    
    print(f"\n{BOLD}Testing with GitHub Copilot:{RESET}")
    print("1. Open VS Code")
    print("2. Press Ctrl+I to open Copilot Chat")
    print("3. Try asking: 'What does the services.nginx.enable option do in NixOS?'")
    print("4. Try this syntax: '@nixai explain services.nginx.enable'")
    
    print(f"\n{BOLD}Testing via MCP Server Runner:{RESET}")
    print("1. Open VS Code")
    print("2. Open Command Palette (Ctrl+Shift+P)")
    print("3. Type 'MCP: List Servers' and press Enter")
    print("4. Check if 'nixai' appears in the list")
    print("5. Type 'MCP: Connect to Server' and select 'nixai'")
    
    print(f"\n{BOLD}Checking VS Code Developer Tools:{RESET}")
    print("1. Open VS Code")
    print("2. Select Help > Toggle Developer Tools")
    print("3. Look for any MCP-related errors or messages")

def diagnostic_summary(results):
    """Provide a summary of diagnostic results."""
    print_header("Diagnostic Summary")
    
    success_count = sum(1 for r in results.values() if r)
    total_count = len(results)
    
    print(f"Results: {success_count}/{total_count} checks passed")
    
    if results['server_running'] and results['socket_exists'] and results['protocol_working']:
        print_success("MCP server is working correctly!")
    else:
        print_error("MCP server has issues - see diagnostics above")
    
    if results['vscode_config'] and results['vscode_extensions']:
        print_success("VS Code appears to be properly configured")
    else:
        print_warning("VS Code configuration needs attention")
    
    if all(results.values()):
        print(f"\n{GREEN}{BOLD}All checks passed! VS Code MCP integration should work.{RESET}")
        print("If you're still having issues, follow the manual activation steps above.")
    else:
        print(f"\n{YELLOW}{BOLD}Some checks failed. Follow recommendations above to fix issues.{RESET}")

def main():
    """Run all diagnostic tests and provide recommendations."""
    print(f"{BOLD}{BLUE}VS Code MCP Integration Diagnostic Tool{RESET}")
    print("=======================================")
    
    results = {}
    
    results['server_running'] = check_mcp_server_running()
    results['socket_exists'] = check_unix_socket()
    results['protocol_working'] = test_mcp_protocol()
    results['vscode_config'] = check_vscode_config()
    results['vscode_extensions'] = check_vscode_extensions()
    results['test_file_created'] = create_test_file()
    
    provide_manual_activation_steps()
    diagnostic_summary(results)

if __name__ == "__main__":
    main()
