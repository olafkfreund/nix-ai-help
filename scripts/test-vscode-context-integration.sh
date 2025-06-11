#!/bin/bash
# VS Code MCP Context Integration Test Script
# Tests the new context-aware features for VS Code integration

set -euo pipefail

echo "🧪 Testing VS Code MCP Context Integration"
echo "=========================================="
echo

# Configuration
SOCKET_PATH="/tmp/nixai-mcp.sock"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counter
TESTS_PASSED=0
TESTS_FAILED=0

# Function to run a test
run_test() {
    local test_name="$1"
    local test_command="$2"
    
    echo -n "Testing $test_name... "
    
    if eval "$test_command" >/dev/null 2>&1; then
        echo -e "${GREEN}✅ PASS${NC}"
        ((TESTS_PASSED++))
    else
        echo -e "${RED}❌ FAIL${NC}"
        ((TESTS_FAILED++))
    fi
}

# Function to test MCP tool via JSON-RPC
test_mcp_tool() {
    local tool_name="$1"
    local args="$2"
    
    local request="{\"jsonrpc\": \"2.0\", \"id\": 1, \"method\": \"tools/call\", \"params\": {\"name\": \"$tool_name\", \"arguments\": $args}}"
    
    echo "$request" | socat - UNIX-CONNECT:"$SOCKET_PATH" | grep -q '"result"'
}

echo "📋 Prerequisites Check"
echo "---------------------"

# Check if nixai is available
if ! command -v "$PROJECT_ROOT/nixai" &> /dev/null; then
    echo -e "${RED}❌ nixai binary not found${NC}"
    exit 1
fi
echo -e "${GREEN}✅ nixai binary found${NC}"

# Check if socat is available
if ! command -v socat &> /dev/null; then
    echo -e "${RED}❌ socat not found (required for VS Code integration)${NC}"
    exit 1
fi
echo -e "${GREEN}✅ socat found${NC}"

echo

echo "🚀 Starting MCP Server"
echo "----------------------"

# Start MCP server
cd "$PROJECT_ROOT"
./nixai mcp-server start -d

# Wait for server to start
sleep 2

# Check if server is running
if [[ ! -S "$SOCKET_PATH" ]]; then
    echo -e "${RED}❌ MCP server failed to start${NC}"
    exit 1
fi
echo -e "${GREEN}✅ MCP server started successfully${NC}"

echo

echo "🧪 Testing Context MCP Tools"
echo "----------------------------"

# Test 1: get_nixos_context
run_test "get_nixos_context tool" "test_mcp_tool 'get_nixos_context' '{\"format\": \"text\", \"detailed\": false}'"

# Test 2: detect_nixos_context
run_test "detect_nixos_context tool" "test_mcp_tool 'detect_nixos_context' '{\"verbose\": false}'"

# Test 3: context_status
run_test "context_status tool" "test_mcp_tool 'context_status' '{\"includeMetrics\": false}'"

# Test 4: reset_nixos_context
run_test "reset_nixos_context tool" "test_mcp_tool 'reset_nixos_context' '{\"confirm\": true}'"

echo

echo "🔧 Testing VS Code Integration Components"
echo "----------------------------------------"

# Test bridge script
run_test "MCP bridge script" "timeout 2s '$SCRIPT_DIR/mcp-bridge.sh' < /dev/null"

# Test configuration generation
if [[ -f "$PROJECT_ROOT/modules/home-manager.nix" ]]; then
    run_test "Home Manager module syntax" "nix-instantiate --parse '$PROJECT_ROOT/modules/home-manager.nix'"
fi

echo

echo "📊 Testing Real-World Context Scenarios"
echo "--------------------------------------"

# Test context-aware query simulation
test_context_aware_query() {
    local context_request='{"jsonrpc": "2.0", "id": 1, "method": "tools/call", "params": {"name": "get_nixos_context", "arguments": {"format": "json", "detailed": true}}}'
    local context_response=$(echo "$context_request" | socat - UNIX-CONNECT:"$SOCKET_PATH")
    
    # Check if we got a valid JSON response with context data
    echo "$context_response" | grep -q '"systemType"' && 
    echo "$context_response" | grep -q '"usesFlakes"' &&
    echo "$context_response" | grep -q '"homeManagerType"'
}

run_test "Context-aware query simulation" "test_context_aware_query"

# Test context refresh scenario
test_context_refresh() {
    # Force a context refresh
    local refresh_request='{"jsonrpc": "2.0", "id": 1, "method": "tools/call", "params": {"name": "detect_nixos_context", "arguments": {"verbose": true}}}'
    echo "$refresh_request" | socat - UNIX-CONNECT:"$SOCKET_PATH" | grep -q '"content"'
}

run_test "Context refresh scenario" "test_context_refresh"

echo

echo "🎯 VS Code Extension Compatibility"
echo "----------------------------------"

# Simulate VS Code extension interaction patterns
test_vscode_extension_pattern() {
    # Simulate how VS Code extensions would interact with context tools
    local init_request='{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {"tools": {"listChanged": false}}, "clientInfo": {"name": "vscode-nixai", "version": "1.0.0"}}}'
    echo "$init_request" | socat - UNIX-CONNECT:"$SOCKET_PATH" | grep -q '"result"'
}

run_test "VS Code extension pattern" "test_vscode_extension_pattern"

# Test tools list includes context tools
test_tools_list() {
    local tools_request='{"jsonrpc": "2.0", "id": 1, "method": "tools/list", "params": {}}'
    local tools_response=$(echo "$tools_request" | socat - UNIX-CONNECT:"$SOCKET_PATH")
    
    echo "$tools_response" | grep -q 'get_nixos_context' &&
    echo "$tools_response" | grep -q 'detect_nixos_context' &&
    echo "$tools_response" | grep -q 'context_status' &&
    echo "$tools_response" | grep -q 'reset_nixos_context'
}

run_test "Context tools in tools list" "test_tools_list"

echo

echo "🧹 Cleanup"
echo "----------"

# Stop MCP server
./nixai mcp-server stop
echo -e "${GREEN}✅ MCP server stopped${NC}"

echo

echo "📈 Test Results"
echo "==============="
echo -e "Tests Passed: ${GREEN}$TESTS_PASSED${NC}"
echo -e "Tests Failed: ${RED}$TESTS_FAILED${NC}"
echo -e "Total Tests: $((TESTS_PASSED + TESTS_FAILED))"

if [[ $TESTS_FAILED -eq 0 ]]; then
    echo -e "\n${GREEN}🎉 All tests passed! VS Code context integration is ready.${NC}"
    exit 0
else
    echo -e "\n${RED}❌ Some tests failed. Please check the issues above.${NC}"
    exit 1
fi
