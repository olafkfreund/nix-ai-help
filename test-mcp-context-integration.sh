#!/bin/bash
# Comprehensive MCP Context Integration Test Suite
# Tests all context functionality end-to-end including editor integrations

echo "🧪 MCP Context Integration - Comprehensive Test Suite"
echo "======================================================"
echo

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test counters
TESTS_TOTAL=0
TESTS_PASSED=0
TESTS_FAILED=0

# Helper functions
test_start() {
    TESTS_TOTAL=$((TESTS_TOTAL + 1))
    echo -e "${BLUE}[TEST $TESTS_TOTAL]${NC} $1"
}

test_pass() {
    TESTS_PASSED=$((TESTS_PASSED + 1))
    echo -e "${GREEN}✅ PASS${NC}: $1"
    echo
}

test_fail() {
    TESTS_FAILED=$((TESTS_FAILED + 1))
    echo -e "${RED}❌ FAIL${NC}: $1"
    echo
}

test_info() {
    echo -e "${YELLOW}ℹ️  INFO${NC}: $1"
}

# Test 1: Basic nixai context system
test_start "Basic Context System Functionality"
echo "Testing core nixai context commands..."

# Test context status
./nixai context status --format json > /tmp/context_status.json 2>/dev/null
if [ $? -eq 0 ] && [ -s /tmp/context_status.json ]; then
    test_pass "Context status command working"
else
    test_fail "Context status command failed"
fi

# Test context show
./nixai context show --format json > /tmp/context_show.json 2>/dev/null
if [ $? -eq 0 ] && [ -s /tmp/context_show.json ]; then
    test_pass "Context show command working"
else
    test_fail "Context show command failed"
fi

# Test 2: MCP server health
test_start "MCP Server Health Check"
echo "Checking MCP server status..."

./nixai mcp-server status > /tmp/mcp_status.txt 2>/dev/null
if [ $? -eq 0 ]; then
    if grep -q "Running" /tmp/mcp_status.txt; then
        test_pass "MCP server is running"
    else
        test_info "Starting MCP server..."
        ./nixai mcp-server start --daemon > /dev/null 2>&1
        sleep 2
        ./nixai mcp-server status > /tmp/mcp_status2.txt 2>/dev/null
        if grep -q "Running" /tmp/mcp_status2.txt; then
            test_pass "MCP server started successfully"
        else
            test_fail "MCP server failed to start"
        fi
    fi
else
    test_fail "MCP server status check failed"
fi

# Test 3: Unix socket availability
test_start "Unix Socket Connectivity"
if [ -S "/tmp/nixai-mcp.sock" ]; then
    test_pass "Unix socket exists at /tmp/nixai-mcp.sock"
    
    # Test socket connectivity
    echo '{"jsonrpc": "2.0", "id": 1, "method": "tools/list"}' | socat - UNIX-CONNECT:/tmp/nixai-mcp.sock > /tmp/socket_test.json 2>/dev/null
    if [ $? -eq 0 ] && grep -q "tools" /tmp/socket_test.json; then
        test_pass "Socket communication working"
    else
        test_fail "Socket communication failed"
    fi
else
    test_fail "Unix socket not found"
fi

# Test 4: All 5 context MCP tools
test_start "MCP Context Tools Functionality"
echo "Testing all 5 context MCP tools..."

# Test get_nixos_context
echo '{"jsonrpc": "2.0", "id": 1, "method": "tools/call", "params": {"name": "get_nixos_context", "arguments": {"format": "text", "detailed": false}}}' | socat - UNIX-CONNECT:/tmp/nixai-mcp.sock > /tmp/get_context.json 2>/dev/null
if [ $? -eq 0 ] && grep -q "result" /tmp/get_context.json; then
    test_pass "get_nixos_context MCP tool working"
else
    test_fail "get_nixos_context MCP tool failed"
fi

# Test detect_nixos_context
echo '{"jsonrpc": "2.0", "id": 2, "method": "tools/call", "params": {"name": "detect_nixos_context", "arguments": {"verbose": false}}}' | socat - UNIX-CONNECT:/tmp/nixai-mcp.sock > /tmp/detect_context.json 2>/dev/null
if [ $? -eq 0 ] && grep -q "result" /tmp/detect_context.json; then
    test_pass "detect_nixos_context MCP tool working"
else
    test_fail "detect_nixos_context MCP tool failed"
fi

# Test reset_nixos_context
echo '{"jsonrpc": "2.0", "id": 3, "method": "tools/call", "params": {"name": "reset_nixos_context", "arguments": {"confirm": true}}}' | socat - UNIX-CONNECT:/tmp/nixai-mcp.sock > /tmp/reset_context.json 2>/dev/null
if [ $? -eq 0 ] && grep -q "result" /tmp/reset_context.json; then
    test_pass "reset_nixos_context MCP tool working"
else
    test_fail "reset_nixos_context MCP tool failed"
fi

# Test context_status
echo '{"jsonrpc": "2.0", "id": 4, "method": "tools/call", "params": {"name": "context_status", "arguments": {"includeMetrics": true}}}' | socat - UNIX-CONNECT:/tmp/nixai-mcp.sock > /tmp/context_status_mcp.json 2>/dev/null
if [ $? -eq 0 ] && grep -q "result" /tmp/context_status_mcp.json; then
    test_pass "context_status MCP tool working"
else
    test_fail "context_status MCP tool failed"
fi

# Test context_diff
echo '{"jsonrpc": "2.0", "id": 5, "method": "tools/call", "params": {"name": "context_diff", "arguments": {}}}' | socat - UNIX-CONNECT:/tmp/nixai-mcp.sock > /tmp/context_diff.json 2>/dev/null
if [ $? -eq 0 ] && grep -q "result" /tmp/context_diff.json; then
    test_pass "context_diff MCP tool working"
else
    test_fail "context_diff MCP tool failed"
fi

# Test 5: VS Code integration files
test_start "VS Code Integration Configuration"
echo "Checking VS Code integration setup..."

if [ -f ".vscode/settings.json" ]; then
    if grep -q "nixai" .vscode/settings.json; then
        test_pass "VS Code settings.json contains nixai configuration"
    else
        test_fail "VS Code settings.json missing nixai configuration"
    fi
else
    test_info "VS Code settings.json not found (optional for testing)"
fi

if [ -f "scripts/mcp-bridge.sh" ]; then
    if [ -x "scripts/mcp-bridge.sh" ]; then
        test_pass "MCP bridge script exists and is executable"
    else
        test_fail "MCP bridge script not executable"
    fi
else
    test_fail "MCP bridge script missing"
fi

# Test 6: Neovim integration
test_start "Neovim Integration Files"
echo "Checking Neovim integration components..."

if [ -f "internal/neovim/integration.go" ]; then
    if grep -q "show_context_aware_suggestion" internal/neovim/integration.go; then
        test_pass "Neovim integration contains context-aware functionality"
    else
        test_fail "Neovim integration missing context-aware features"
    fi
else
    test_fail "Neovim integration file missing"
fi

if [ -f "modules/nixvim-nixai-example.nix" ]; then
    if grep -q "contextAware" modules/nixvim-nixai-example.nix; then
        test_pass "NixVim example contains context-aware configuration"
    else
        test_fail "NixVim example missing context-aware configuration"
    fi
else
    test_fail "NixVim example configuration missing"
fi

# Test 7: Documentation completeness
test_start "Documentation Completeness"
echo "Checking documentation coverage..."

if [ -f "docs/MCP_VSCODE_INTEGRATION.md" ]; then
    if grep -q "context-aware" docs/MCP_VSCODE_INTEGRATION.md; then
        test_pass "VS Code integration documentation includes context features"
    else
        test_fail "VS Code integration documentation missing context features"
    fi
else
    test_fail "VS Code integration documentation missing"
fi

if [ -f "docs/neovim-integration.md" ]; then
    if grep -q "Context-Aware" docs/neovim-integration.md; then
        test_pass "Neovim integration documentation includes context features"
    else
        test_fail "Neovim integration documentation missing context features"
    fi
else
    test_fail "Neovim integration documentation missing"
fi

# Test 8: Context data validation
test_start "Context Data Validation"
echo "Validating context data structure and content..."

# Extract context data and validate
./nixai context show --format json > /tmp/context_validation.json 2>/dev/null
if [ $? -eq 0 ] && [ -s /tmp/context_validation.json ]; then
    # Check for required fields
    required_fields=("SystemType" "UsesFlakes" "HomeManagerType" "NixOSVersion")
    all_fields_present=true
    
    for field in "${required_fields[@]}"; do
        if ! grep -q "\"$field\"" /tmp/context_validation.json; then
            all_fields_present=false
            test_info "Missing field: $field"
        fi
    done
    
    if [ "$all_fields_present" = true ]; then
        test_pass "Context data contains all required fields"
    else
        test_fail "Context data missing required fields"
    fi
else
    test_fail "Context data validation failed"
fi

# Test 9: Performance validation
test_start "Performance Validation"
echo "Testing context system performance..."

start_time=$(date +%s%N)
./nixai context show > /dev/null 2>&1
end_time=$(date +%s%N)
duration=$(( (end_time - start_time) / 1000000 )) # Convert to milliseconds

if [ $duration -lt 2000 ]; then # Less than 2 seconds
    test_pass "Context retrieval performance acceptable ($duration ms)"
else
    test_fail "Context retrieval too slow ($duration ms)"
fi

# Test 10: Integration workflow test
test_start "End-to-End Integration Workflow"
echo "Testing complete workflow: context detection → MCP access → editor integration..."

# Simulate a complete workflow
workflow_success=true

# Step 1: Detect context
./nixai context detect > /tmp/workflow_detect.txt 2>&1
if [ $? -ne 0 ]; then
    workflow_success=false
    test_info "Context detection failed in workflow"
fi

# Step 2: Access via MCP
echo '{"jsonrpc": "2.0", "id": 1, "method": "tools/call", "params": {"name": "get_nixos_context", "arguments": {"format": "text"}}}' | socat - UNIX-CONNECT:/tmp/nixai-mcp.sock > /tmp/workflow_mcp.json 2>/dev/null
if [ $? -ne 0 ] || ! grep -q "result" /tmp/workflow_mcp.json; then
    workflow_success=false
    test_info "MCP access failed in workflow"
fi

# Step 3: Check integration files
if [ ! -f "internal/neovim/integration.go" ] || [ ! -f "docs/MCP_VSCODE_INTEGRATION.md" ]; then
    workflow_success=false
    test_info "Integration files missing in workflow"
fi

if [ "$workflow_success" = true ]; then
    test_pass "End-to-end integration workflow successful"
else
    test_fail "End-to-end integration workflow failed"
fi

# Clean up temporary files
echo "🧹 Cleaning up temporary test files..."
rm -f /tmp/context_*.json /tmp/mcp_*.txt /tmp/socket_test.json /tmp/get_context.json
rm -f /tmp/detect_context.json /tmp/reset_context.json /tmp/context_diff.json
rm -f /tmp/workflow_*.txt /tmp/workflow_*.json /tmp/context_validation.json

# Final results
echo "📊 Test Results Summary"
echo "======================"
echo -e "Total Tests: ${BLUE}$TESTS_TOTAL${NC}"
echo -e "Passed: ${GREEN}$TESTS_PASSED${NC}"
echo -e "Failed: ${RED}$TESTS_FAILED${NC}"
echo

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}🎉 ALL TESTS PASSED!${NC}"
    echo -e "${GREEN}✅ MCP Context Integration is working perfectly${NC}"
    echo
    echo "🚀 Ready for production deployment!"
    echo "   • VS Code users can access context via MCP tools"
    echo "   • Neovim users can use context-aware keymaps"  
    echo "   • Context system provides intelligent, system-specific assistance"
    exit 0
else
    echo -e "${RED}❌ SOME TESTS FAILED${NC}"
    echo -e "${YELLOW}⚠️  Please review the failed tests above${NC}"
    echo
    echo "🔧 Common fixes:"
    echo "   • Start MCP server: ./nixai mcp-server start --daemon"
    echo "   • Check socket permissions: ls -la /tmp/nixai-mcp.sock"
    echo "   • Verify nixai installation: ./nixai --version"
    exit 1
fi
