Organize test files into structured test directory

This commit reorganizes all test files into a proper directory structure:

1. Created test directories:
   - tests/mcp/ - For MCP protocol and server tests
   - tests/vscode/ - For VS Code integration tests
   - tests/providers/ - For AI provider tests
   - tests/integration/ - For future integration tests

2. Standardized naming conventions:
   - Python files now use snake_case (test_mcp_protocol.py)
   - Shell scripts maintain kebab-case (test-mcp-server.sh)

3. Created runner scripts:
   - tests/run_all.sh - Runs all test suites
   - tests/run_mcp.sh - Runs only MCP tests
   - tests/run_vscode.sh - Runs only VS Code tests
   - tests/run_providers.sh - Runs only provider tests

4. Added justfile targets for the new test structure:
   - test-all - Runs all tests
   - test-mcp - Runs only MCP tests
   - test-vscode - Runs only VS Code tests
   - test-providers - Runs only provider tests

5. Improved path handling for tests:
   - Tests find repository root automatically
   - Tests can be run from any directory

6. Added test environment validation with tests/check-compatibility.sh

7. Added tests/README.md with comprehensive documentation

8. Removed original test files from the root directory

This restructuring makes the test suite easier to maintain, enhances organization,
and provides clearer test grouping and execution paths.
