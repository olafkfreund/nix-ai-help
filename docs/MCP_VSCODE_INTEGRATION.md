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
1. Install the required MCP extensions:
   - `automatalabs.copilot-mcp` - Copilot MCP extension
   - `zebradev.mcp-server-runner` - MCP Server Runner
   - `saoudrizwan.claude-dev` - Claude Dev (Cline)

2. Add the following configuration to your VS Code settings (`.vscode/settings.json`):

```json
{
  "mcp.servers": {
    "nixai": {
      "command": "bash",
      "args": ["-c", "socat STDIO UNIX-CONNECT:/tmp/nixai-mcp.sock"],
      "env": {}
    }
  },
  "copilot.mcp.servers": {
    "nixai": {
      "command": "bash",
      "args": ["-c", "socat STDIO UNIX-CONNECT:/tmp/nixai-mcp.sock"],
      "env": {}
    }
  },
  "claude-dev.mcpServers": {
    "nixai": {
      "command": "bash",
      "args": ["-c", "socat STDIO UNIX-CONNECT:/tmp/nixai-mcp.sock"],
      "env": {}
    }
  },
  "mcp.enableDebug": true,
  "claude-dev.enableMcp": true,
  "automata.mcp.enabled": true,
  "zebradev.mcp.enabled": true
}
```

### 3. Verify MCP Integration
Use our diagnostic tool to verify the integration is working:

```bash
# Run the diagnostic tool
./vscode-mcp-diagnostic.py
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

### In VS Code with Claude Dev (Cline)
1. Open Command Palette (`Ctrl+Shift+P`)
2. Search for "Claude Dev: Open Claude Dev"
3. In the chat, ask: "What does services.nginx.enable do in NixOS?"
4. Claude should query the MCP server and provide an answer

### In VS Code with GitHub Copilot Chat
1. Press `Ctrl+I` to open Copilot Chat
2. Ask: "Using the MCP server, explain the services.nginx.enable option in NixOS"
3. Or use the direct syntax: "@nixai explain services.nginx.enable"

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
5. Run our diagnostic tool: `./vscode-mcp-diagnostic.py`

### Manual VS Code Activation
Some extensions require manual activation of MCP features:

1. For Claude Dev (Cline):
   - Open VS Code Settings
   - Search for "claude-dev.enableMcp"
   - Ensure it's checked/enabled
   - Restart VS Code

2. For GitHub Copilot:
   - Open VS Code Settings
   - Search for "copilot.mcp"
   - Ensure MCP features are enabled
   - Restart VS Code

## Development

### Testing the MCP Server
```bash
# Run comprehensive tests
./test-mcp-server.sh

# Run VS Code integration diagnostics
./vscode-mcp-diagnostic.py

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

## VS Code Integration Testing ✅

### Current Status (May 30, 2025)

The MCP server is **fully functional** and ready for VS Code integration:

- ✅ **MCP Protocol**: Working correctly with JSON-RPC2 over Unix socket
- ✅ **Server Architecture**: Fixed and stable (Start method handles connections properly)
- ✅ **Socket Management**: Unix socket listener properly managed with mutex protection
- ✅ **All Tools Working**: query_nixos_docs, explain_nixos_option, explain_home_manager_option, search_nixos_packages
- ✅ **Extensions Installed**: Copilot MCP, Claude Dev, MCP Server Runner
- ✅ **VS Code Configuration**: Proper MCP server configuration in settings.json

### Testing Steps

1. **Start the MCP Server**:
   ```bash
   nixai mcp-server start -d
   ```

2. **Run the Diagnostic Tool**:
   ```bash
   ./vscode-mcp-diagnostic.py
   ```

3. **Open VS Code and Test**:
   - Open test-mcp-integration.nix
   - Ask AI assistants about NixOS options
   - Check if MCP tools are accessible

For detailed instructions on manual testing, refer to the diagnostic tool output.

## Integration Benefits

Using the MCP server with VS Code provides:

1. **Contextual Help**: Get NixOS help without leaving your editor
2. **Intelligent Suggestions**: AI-powered configuration assistance
3. **Documentation Access**: Direct access to NixOS documentation
4. **Package Discovery**: Search and explore NixOS packages
5. **Option Explanations**: Understand configuration options in context
