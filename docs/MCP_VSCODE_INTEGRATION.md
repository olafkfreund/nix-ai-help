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

### 5. `get_nixos_context` ✨ NEW
Get current NixOS system context information.
- **Arguments**: `format` (string, optional) - Response format ("text" or "json"), `detailed` (boolean, optional) - Include detailed information
- **Usage**: Get context-aware information about your NixOS system setup (flakes, Home Manager, services, etc.)

### 6. `detect_nixos_context` ✨ NEW
Force re-detection of NixOS system context.
- **Arguments**: `verbose` (boolean, optional) - Show detection process details
- **Usage**: Refresh system context detection when configuration changes

### 7. `reset_nixos_context` ✨ NEW
Clear cached context and force refresh.
- **Arguments**: `confirm` (boolean, optional) - Confirm reset operation
- **Usage**: Clear context cache and perform fresh system detection

### 8. `context_status` ✨ NEW
Show context detection system status and health.
- **Arguments**: `includeMetrics` (boolean, optional) - Include performance metrics
- **Usage**: Check if context system is working properly and get health information

## Usage Examples

### Context-Aware NixOS Assistance ✨ NEW

The new context tools enable AI assistants to understand your specific NixOS configuration:

#### In VS Code with Claude Dev (Cline)

1. **Get System Context**:
   - Ask: "Using the context tools, show me my current NixOS configuration setup"
   - Claude will use `get_nixos_context` to provide tailored information

2. **Context-Aware Configuration Help**:
   - Ask: "How should I configure nginx based on my current system setup?"
   - Claude will first get your context, then provide recommendations specific to your configuration (flakes vs channels, Home Manager type, etc.)

3. **System Health Check**:
   - Ask: "Check my NixOS context system health"
   - Claude will use `context_status` to verify everything is working correctly

#### In VS Code with GitHub Copilot Chat

1. **Context-Aware Suggestions**:
   ```
   @nixai Using my system context, suggest the best way to configure PostgreSQL
   ```

2. **System Detection**:
   ```
   @nixai Get my current NixOS context and explain my configuration type
   ```

3. **Troubleshooting**:
   ```
   @nixai Check context status and refresh if needed
   ```

### Traditional Documentation Queries

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

## AI Prompt Templates for Context-Aware Assistance ✨ NEW

Use these prompt templates to get the most out of the context tools:

### Context-Aware Configuration Prompts

1. **Smart Configuration Suggestions**:
   ```
   Using get_nixos_context, analyze my system setup and suggest the optimal configuration for [SERVICE/FEATURE]. Consider my current flakes usage, Home Manager setup, and enabled services.
   ```

2. **Troubleshooting with Context**:
   ```
   First check my context status, then help me debug [ISSUE]. Use my system context to provide specific solutions for my configuration type.
   ```

3. **Migration Assistance**:
   ```
   Based on my current NixOS context, guide me through migrating to [TARGET] (e.g., flakes, different Home Manager setup). Show step-by-step instructions tailored to my current setup.
   ```

4. **Service Configuration**:
   ```
   Using my NixOS context, show me how to configure [SERVICE] in the way that best fits my current system architecture (flakes/channels, Home Manager type, etc.).
   ```

### Advanced Context Workflows

1. **Full System Analysis**:
   ```
   1. Get my current context with detailed information
   2. Check context system health
   3. Analyze my configuration and suggest improvements
   4. Provide specific next steps based on my setup
   ```

2. **Configuration Validation**:
   ```
   Using context tools, validate that my current NixOS setup is optimal and suggest any improvements based on best practices for my configuration type.
   ```

3. **Automated Health Check**:
   ```
   Run a complete system health check including context status, and provide a summary of my NixOS system state with recommendations.
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
  
  # Context system configuration ✨ NEW
  context:
    cache_ttl: 3600  # Cache context for 1 hour
    auto_detect: true  # Auto-detect context changes
    detailed_detection: false  # Set to true for verbose context
```

### Enhanced VS Code Settings ✨ NEW

For optimal context-aware assistance, update your VS Code settings:

```json
{
  "mcp.servers": {
    "nixai": {
      "command": "bash",
      "args": ["-c", "socat STDIO UNIX-CONNECT:/tmp/nixai-mcp.sock"],
      "env": {},
      "capabilities": {
        "context": true,
        "system_detection": true
      }
    }
  },
  "copilot.mcp.servers": {
    "nixai": {
      "command": "bash", 
      "args": ["-c", "socat STDIO UNIX-CONNECT:/tmp/nixai-mcp.sock"],
      "env": {},
      "contextAware": true
    }
  },
  "claude-dev.mcpServers": {
    "nixai": {
      "command": "bash",
      "args": ["-c", "socat STDIO UNIX-CONNECT:/tmp/nixai-mcp.sock"],
      "env": {},
      "useContext": true
    }
  },
  "mcp.enableDebug": true,
  "claude-dev.enableMcp": true,
  "automata.mcp.enabled": true,
  "zebradev.mcp.enabled": true,
  
  // Context-aware AI settings ✨ NEW
  "nixai.contextIntegration": {
    "autoRefresh": true,
    "contextTimeout": 5000,
    "enableDetailedContext": false
  }
}
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
