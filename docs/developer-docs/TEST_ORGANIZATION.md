# Test Organization for NixAI Project

This document describes the organization of tests in the NixAI project after the cleanup and standardization effort.

## 📁 Directory Structure

```
tests/
├── README.md                   # Test documentation
├── check-compatibility.sh      # Test environment compatibility checker
├── requirements.txt            # Python test dependencies
├── run_all.sh                  # Script to run all tests
├── run_mcp.sh                  # Script to run MCP tests
├── run_providers.sh            # Script to run provider tests
├── run_vscode.sh               # Script to run VS Code tests
├── mcp/                        # MCP protocol and server tests
│   ├── test_mcp_protocol.py    # Tests JSON-RPC2 communication over Unix socket
│   ├── test-mcp-server.sh      # Tests MCP server functionality
│   ├── test_mcp_simple.py      # Tests MCP protocol with exact JSON-RPC2 format
│   ├── test_raw_socket.py      # Tests raw socket communication
│   └── test_simple_socket.py   # Simple test for Unix socket connectivity
├── vscode/                     # VS Code integration tests
│   ├── test_vscode_direct.py   # Tests VS Code direct connection
│   ├── test_vscode_mcp_integration.py # Tests VS Code MCP integration
│   ├── test-vscode-mcp.sh      # Basic VS Code MCP test
│   ├── test-vscode-live.sh     # Full VS Code MCP integration live test
│   └── vscode_mcp_diagnostic.py # VS Code MCP diagnostic tool
└── providers/                  # AI provider tests
    └── test-all-providers.sh   # Tests all AI providers
```

## 🏷️ Naming Conventions

- **Python files**: Use snake_case (`test_mcp_protocol.py`)
- **Shell scripts**: Use kebab-case (`test-mcp-server.sh`)

## 🧪 Test Categories

### MCP Tests (tests/mcp/)

Tests for the Model Context Protocol server and client components:
- Socket connectivity
- Protocol message format
- JSON-RPC2 communication
- Server functionality

### VS Code Tests (tests/vscode/)

Tests for VS Code MCP integration:
- VS Code extension configuration
- MCP connectivity from VS Code
- Protocol validation
- Interactive diagnostics

### Provider Tests (tests/providers/)

Tests for AI provider integration:
- Ollama provider
- Gemini provider
- OpenAI provider
- Provider switching

### Go Unit Tests (internal/*/)

Go unit tests remain in their respective package directories following Go conventions.

## 🛠️ Running Tests

### Using just commands:

```bash
# Run all tests
just test-all

# Run specific test groups
just test-mcp
just test-vscode
just test-providers

# Run Go unit tests only
just test
```

### Using shell scripts directly:

```bash
# Run all tests
./tests/run_all.sh

# Run specific test groups
./tests/run_mcp.sh
./tests/run_vscode.sh
./tests/run_providers.sh

# Check compatibility
./tests/check-compatibility.sh
```

## 🚀 Next Steps

1. Add more specific tests for individual AI providers
2. Create integration tests for CLI commands
3. Implement end-to-end tests for complex workflows
4. Add CI integration for automated testing
