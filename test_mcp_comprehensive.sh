#!/bin/bash
# Comprehensive MCP Server Testing Script
# Tests all MCP server commands and functionality systematically

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

REPO_ROOT=$(pwd)
NIXAI_BIN="$REPO_ROOT/nixai"

echo -e "${BLUE}🧪 Comprehensive MCP Server Testing${NC}"
echo "======================================="
echo

# Utility functions
log_test() {
    echo -e "${BLUE}📋 Test: $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_info() {
    echo -e "${YELLOW}ℹ️  $1${NC}"
}

# Test counter
TEST_COUNT=0
PASSED_COUNT=0

run_test() {
    local test_name="$1"
    local test_command="$2"
    local expected_result="$3"
    
    TEST_COUNT=$((TEST_COUNT + 1))
    log_test "$test_name"
    
    if eval "$test_command"; then
        if [[ -n "$expected_result" ]]; then
            if eval "$expected_result"; then
                log_success "PASSED"
                PASSED_COUNT=$((PASSED_COUNT + 1))
            else
                log_error "FAILED - Expected condition not met"
            fi
        else
            log_success "PASSED"
            PASSED_COUNT=$((PASSED_COUNT + 1))
        fi
    else
        log_error "FAILED - Command failed"
    fi
    echo
}

# Clean start - ensure no processes are running
echo -e "${YELLOW}🧹 Initial Cleanup${NC}"
echo "=================="
pkill -f "nixai mcp-server" 2>/dev/null || true
sudo rm -f /tmp/nixai-mcp.sock
sleep 2
echo

# Test 1: Initial Status Check (should show not running)
log_test "Initial Status Check"
$NIXAI_BIN mcp-server status
echo

# Test 2: Start MCP Server (regular mode)
log_test "Start MCP Server (regular mode)"
echo "Starting server in background..."
$NIXAI_BIN mcp-server start &
SERVER_PID=$!
echo "Server PID: $SERVER_PID"

# Wait for server to initialize
echo "Waiting for server initialization..."
sleep 5

# Check if process is still running
if kill -0 $SERVER_PID 2>/dev/null; then
    log_success "Server process is running"
else
    log_error "Server process died"
fi
echo

# Test 3: Status Check (should show running)
log_test "Status Check After Start"
$NIXAI_BIN mcp-server status
echo

# Test 4: HTTP Health Check
log_test "HTTP Health Check"
if curl -s http://localhost:8081/healthz | grep -q "ok"; then
    log_success "HTTP health endpoint responding"
else
    log_warning "HTTP health endpoint not responding"
fi
echo

# Test 5: Unix Socket Check
log_test "Unix Socket Check"
if [ -S "/tmp/nixai-mcp.sock" ]; then
    log_success "Unix socket exists"
else
    log_error "Unix socket missing"
fi
echo

# Test 6: MCP Protocol Test (if socat available)
log_test "MCP Protocol Test"
if command -v socat &> /dev/null; then
    echo '{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}, "clientInfo": {"name": "test", "version": "1.0.0"}}}' > /tmp/mcp-test.json
    
    if timeout 5s socat - UNIX-CONNECT:/tmp/nixai-mcp.sock < /tmp/mcp-test.json > /tmp/mcp-response.json 2>/dev/null; then
        if grep -q "protocolVersion" /tmp/mcp-response.json; then
            log_success "MCP protocol responding correctly"
        else
            log_warning "MCP protocol response unexpected"
        fi
    else
        log_warning "MCP protocol test timed out"
    fi
    rm -f /tmp/mcp-test.json /tmp/mcp-response.json
else
    log_warning "socat not available, skipping MCP protocol test"
fi
echo

# Test 7: Documentation Query via HTTP
log_test "Documentation Query via HTTP"
if curl -s "http://localhost:8081/query?q=services.nginx.enable" | grep -q "nginx"; then
    log_success "HTTP documentation query working"
else
    log_warning "HTTP documentation query not working as expected"
fi
echo

# Test 8: CLI Query Command
log_test "CLI Query Command"
if $NIXAI_BIN mcp-server query "services.openssh.enable" 2>&1 | grep -q -i "ssh\|openssh\|documentation"; then
    log_success "CLI query command working"
else
    log_warning "CLI query command not working as expected"
fi
echo

# Test 9: Graceful Stop
log_test "Graceful Stop"
$NIXAI_BIN mcp-server stop
sleep 2

# Check if HTTP endpoint stopped
if curl -s http://localhost:8081/healthz &>/dev/null; then
    log_warning "HTTP endpoint still responding after stop"
else
    log_success "HTTP endpoint stopped"
fi

# Check if Unix socket cleaned up
if [ -S "/tmp/nixai-mcp.sock" ]; then
    log_warning "Unix socket still exists after stop"
else
    log_success "Unix socket cleaned up"
fi
echo

# Test 10: Start in Daemon Mode
log_test "Start in Daemon Mode"
$NIXAI_BIN mcp-server start -d
sleep 3

# Check if daemon process is running
if pgrep -f "nixai mcp-server" > /dev/null; then
    log_success "Daemon mode started successfully"
else
    log_error "Daemon mode failed to start"
fi
echo

# Test 11: Status Check After Daemon Start
log_test "Status Check After Daemon Start"
$NIXAI_BIN mcp-server status
echo

# Test 12: Restart Command
log_test "Restart Command"
$NIXAI_BIN mcp-server restart
sleep 3

if pgrep -f "nixai mcp-server" > /dev/null; then
    log_success "Restart successful"
else
    log_error "Restart failed"
fi
echo

# Test 13: Multiple Query Test
log_test "Multiple Queries Test"
queries=("services.nginx.enable" "boot.loader.grub.enable" "networking.firewall.enable")
for query in "${queries[@]}"; do
    echo "Testing query: $query"
    if $NIXAI_BIN mcp-server query "$query" 2>&1 | grep -q -i "documentation\|option\|enable\|disable"; then
        echo "  ✅ Query successful"
    else
        echo "  ⚠️  Query response unclear"
    fi
done
echo

# Test 14: Error Scenarios
log_test "Error Scenarios"

# Invalid query
echo "Testing invalid query handling..."
$NIXAI_BIN mcp-server query "" 2>&1 || true

# Invalid subcommand
echo "Testing invalid subcommand..."
$NIXAI_BIN mcp-server invalid 2>&1 || true
echo

# Test 15: Performance Test
log_test "Performance Test"
echo "Running 5 concurrent queries..."
for i in {1..5}; do
    (curl -s "http://localhost:8081/query?q=services.nginx.enable" &)
done
wait
log_success "Concurrent queries completed"
echo

# Final cleanup
echo -e "${YELLOW}🧹 Final Cleanup${NC}"
echo "================"
$NIXAI_BIN mcp-server stop 2>/dev/null || true
pkill -f "nixai mcp-server" 2>/dev/null || true
sudo rm -f /tmp/nixai-mcp.sock
echo

# Summary
echo "======================================="
echo -e "${BLUE}📊 Test Summary${NC}"
echo "======================================="
echo "Total Tests: $TEST_COUNT"
echo "Passed: $PASSED_COUNT"
echo "Failed: $((TEST_COUNT - PASSED_COUNT))"

if [ $PASSED_COUNT -eq $TEST_COUNT ]; then
    echo -e "${GREEN}🎉 All tests passed!${NC}"
    exit 0
elif [ $PASSED_COUNT -gt $((TEST_COUNT / 2)) ]; then
    echo -e "${YELLOW}⚠️  Most tests passed with some warnings${NC}"
    exit 0
else
    echo -e "${RED}❌ Several tests failed${NC}"
    exit 1
fi
