# NixAI Tests

This directory contains all tests for the NixAI project, organized by type and functionality.

## Directory Structure

- `mcp/` - Tests for MCP protocol and server functionality
- `vscode/` - Tests for VS Code integration and connectivity
- `providers/` - Tests for AI provider integration (Ollama, Gemini, OpenAI)
- `integration/` - General integration tests

## How to Run Tests

### All Tests
```bash
# Run all tests
./tests/run_all.sh

# Alternative using just
just test-all
```

### Specific Test Groups
```bash
# MCP tests only
./tests/run_mcp.sh
just test-mcp

# VS Code integration tests
./tests/run_vscode.sh
just test-vscode

# AI Provider tests
./tests/run_providers.sh
just test-providers
```

### Go Unit Tests
Go unit tests remain in their respective package directories and follow Go conventions:
```bash
# Run all Go unit tests
go test ./...

# Run tests for a specific package
go test ./internal/mcp
```

## Test Naming Conventions

- Python files: Use snake_case (`test_mcp_protocol.py`)
- Shell scripts: Use kebab-case (`test-mcp-server.sh`)
- Prefixes: 
  - `unit_` - Unit tests
  - `integration_` - Integration tests
  - `e2e_` - End-to-end tests

## Adding New Tests

1. Place your test in the appropriate directory
2. Follow naming conventions
3. Ensure it's executable (`chmod +x your_test_file`)
4. Update run scripts if necessary
5. Consider adding a `just` target for your test

## Expected Outputs

All tests should clearly indicate:
- Test name/description
- Pass/fail status (preferably with ✅/❌ symbols)
- Summary of results

## Test Dependencies

Python-based tests require:
- Python 3.8+
- socket module
- json module
- subprocess module

Shell-based tests require:
- bash/zsh
- curl
- socat (for Unix socket tests)
- grep, awk, and other standard Unix tools
