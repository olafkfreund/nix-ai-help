# NixAI MCP Server - VS Code Integration

## Overview
The nixai MCP (Model Context Protocol) server enables VS Code and other IDEs to access NixOS documentation and configuration help directly within the editor environment.

## Installation & Setup

### 1. Start the MCP Server
```bash
# Start the server in background mode
nixai mcp-server start -d

# Check server status
nixai mcp-server status
```

### 2. VS Code MCP Extension Setup
1. Install the MCP extension for VS Code (if available)
2. Add the following configuration to your VS Code settings:

```json
{
  "mcp.servers": {
    "nixai": {
      "command": "socat",
      "args": ["STDIO", "UNIX-CONNECT:/tmp/nixai-mcp.sock"],
      "env": {}
    }
  }
}
```

### 3. Alternative: Direct MCP Client Configuration
For MCP-compatible editors, use this configuration:

```json
{
  "mcpServers": {
    "nixai": {
      "command": "socat",
      "args": ["STDIO", "UNIX-CONNECT:/tmp/nixai-mcp.sock"],
      "env": {}
    }
  }
}
```

## Available MCP Tools

The nixai MCP server provides the following tools:

### 1. `query_nixos_docs`
Query NixOS documentation from multiple sources.
- **Arguments**: `query` (string) - The search query
- **Usage**: Search for NixOS configuration examples, options, and documentation

### 2. `explain_nixos_option`
Get detailed explanations of NixOS configuration options.
- **Arguments**: `option` (string) - The NixOS option name (e.g., "services.nginx.enable")
- **Usage**: Understand what an option does, its type, default value, and examples

### 3. `explain_home_manager_option`
Get explanations for Home Manager configuration options.
- **Arguments**: `option` (string) - The Home Manager option name
- **Usage**: Learn about Home Manager-specific configuration options

### 4. `search_nixos_packages`
Search for NixOS packages.
- **Arguments**: `query` (string) - Package search terms
- **Usage**: Find available packages in the NixOS package collection

## Usage Examples

### In VS Code with MCP Extension
1. Open Command Palette (`Ctrl+Shift+P`)
2. Search for "MCP: Query NixOS Documentation"
3. Enter your query (e.g., "nginx configuration")
4. View results directly in VS Code

### Direct MCP Protocol Test
```bash
# Test the MCP server directly
echo '{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}, "clientInfo": {"name": "test", "version": "1.0.0"}}}' | socat - UNIX-CONNECT:/tmp/nixai-mcp.sock
```

## Configuration

### Server Configuration
Located in `configs/default.yaml`:
```yaml
mcp_server:
  host: localhost
  port: 8081
  socket_path: /tmp/nixai-mcp.sock
  auto_start: false
  documentation_sources:
    - nixos-options-es://options
    - https://wiki.nixos.org/wiki/NixOS_Wiki
    - https://nix.dev/manual/nix
    - https://nixos.org/manual/nixpkgs/stable/
    - https://nix.dev/manual/nix/2.28/language/
    - https://nix-community.github.io/home-manager/
```

### Auto-Start
To automatically start the MCP server with nixai commands, set `auto_start: true` in the configuration.

## Troubleshooting

### Server Not Starting
```bash
# Check if port is in use
lsof -i :8081

# Stop existing server
nixai mcp-server stop

# Restart server
nixai mcp-server start -d
```

### Socket Permission Issues
```bash
# Check socket permissions
ls -la /tmp/nixai-mcp.sock

# If needed, manually set permissions
chmod 666 /tmp/nixai-mcp.sock
```

### VS Code Not Connecting
1. Ensure `socat` is installed: `which socat`
2. Verify socket exists: `ls -la /tmp/nixai-mcp.sock`
3. Test connection manually: `echo "test" | socat - UNIX-CONNECT:/tmp/nixai-mcp.sock`
4. Restart VS Code after configuration changes

## Development

### Testing the MCP Server
```bash
# Run comprehensive tests
./test-mcp-server.sh

# Test specific functionality
./nixai explain-option services.nginx.enable
curl "http://localhost:8081/query?q=services.nginx.enable"
```

### Building from Source
```bash
# Build nixai with MCP support
go build -o nixai cmd/nixai/main.go

# Run tests
go test ./internal/mcp/...
```

## Integration Benefits

Using the MCP server with VS Code provides:

1. **Contextual Help**: Get NixOS help without leaving your editor
2. **Intelligent Suggestions**: AI-powered configuration assistance
3. **Documentation Access**: Direct access to NixOS documentation
4. **Package Discovery**: Search and explore NixOS packages
5. **Option Explanations**: Understand configuration options in context

## Security Considerations

- The MCP server runs locally and does not expose external ports by default
- Unix socket communication provides secure local IPC
- No sensitive data is transmitted over network connections
- All documentation queries are read-only operations

## Future Enhancements

- Integration with VS Code IntelliSense for Nix files
- Real-time syntax validation for NixOS configurations
- Automated configuration generation based on user intent
- Integration with nixfmt for code formatting
- Support for flake.nix assistance and validation
