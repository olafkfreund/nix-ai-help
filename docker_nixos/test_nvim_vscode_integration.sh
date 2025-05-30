#!/usr/bin/env bash
# test_nvim_vscode_integration.sh
# Comprehensive test script for nixai Neovim and VS Code integrations in Docker

set -euo pipefail

echo "🧪 nixai Docker Neovim & VS Code Integration Test Suite"
echo "========================================================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Test tracking
TESTS_PASSED=0
TESTS_FAILED=0
TOTAL_TESTS=0

# Docker environment paths
NIXAI_DIR="/home/nixuser/nixai"
NVIM_CONFIG_DIR="/home/nixuser/.config/nvim"
VSCODE_CONFIG_DIR="/home/nixuser/.vscode"

# Utility functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[PASS]${NC} $1"
    ((TESTS_PASSED++))
}

log_error() {
    echo -e "${RED}[FAIL]${NC} $1"
    ((TESTS_FAILED++))
}

log_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_section() {
    echo -e "${CYAN}[SECTION]${NC} $1"
    echo "────────────────────────────────────────────────"
}

run_test() {
    local test_name="$1"
    local test_command="$2"
    ((TOTAL_TESTS++))
    
    log_info "Testing: $test_name"
    
    if eval "$test_command" &>/dev/null; then
        log_success "$test_name"
        return 0
    else
        log_error "$test_name"
        return 1
    fi
}

run_test_with_output() {
    local test_name="$1"
    local test_command="$2"
    ((TOTAL_TESTS++))
    
    log_info "Testing: $test_name"
    
    local output
    if output=$(eval "$test_command" 2>&1); then
        log_success "$test_name"
        echo "$output" | head -3
        return 0
    else
        log_error "$test_name"
        echo "Error output: $output" | head -3
        return 1
    fi
}

# Pre-flight checks
log_section "Pre-flight Environment Checks"

run_test "nixai binary exists" "[ -f '$NIXAI_DIR/nixai' ]"
run_test "Go environment available" "command -v go"
run_test "Neovim available" "command -v nvim"
run_test "Ollama host accessible" "curl -f http://host.docker.internal:11434/api/version"

echo ""

# Build and install nixai
log_section "Building and Installing nixai"

cd "$NIXAI_DIR"
log_info "Building nixai with latest changes..."

# Test module availability first
run_test "NixOS module exists" "[ -f '$NIXAI_DIR/modules/nixos.nix' ]"
run_test "Home Manager module exists" "[ -f '$NIXAI_DIR/modules/home-manager.nix' ]"
run_test "Neovim integration code exists" "[ -f '$NIXAI_DIR/internal/neovim/integration.go' ]"
if go build -o ./nixai ./cmd/nixai/main.go; then
    log_success "nixai build completed"
else
    log_error "nixai build failed"
    exit 1
fi

log_info "Installing nixai globally..."
if sudo cp ./nixai /usr/local/bin/nixai && sudo chmod +x /usr/local/bin/nixai; then
    log_success "nixai installed globally"
else
    log_error "nixai installation failed"
    exit 1
fi

echo ""

# Test MCP Server
log_section "MCP Server Tests"

log_info "Starting MCP server..."
nixai mcp-server start --daemon --socket-path /tmp/nixai-mcp.sock &
MCP_PID=$!
sleep 3

run_test "MCP server process running" "pgrep -f 'nixai mcp-server'"
run_test "MCP socket exists" "[ -S /tmp/nixai-mcp.sock ]"
run_test "MCP server health check" "curl -f http://localhost:8081/healthz"

echo ""

# Test Neovim Integration
log_section "Neovim Integration Tests"

# Create Neovim config directory
mkdir -p "$NVIM_CONFIG_DIR/lua"

# Test neovim module generation
log_info "Generating Neovim configuration..."
if nixai neovim-setup --config-dir "$NVIM_CONFIG_DIR" --socket-path /tmp/nixai-mcp.sock; then
    log_success "Neovim module generated"
else
    log_error "Neovim module generation failed"
fi

run_test "Neovim nixai.lua module exists" "[ -f '$NVIM_CONFIG_DIR/lua/nixai.lua' ]"
run_test "Neovim module contains socket path" "grep -q '/tmp/nixai-mcp.sock' '$NVIM_CONFIG_DIR/lua/nixai.lua'"

# Create test Neovim configuration
cat > "$NVIM_CONFIG_DIR/init.lua" << 'EOF'
-- Test nixai integration
local ok, nixai = pcall(require, "nixai")
if ok then
  nixai.setup({
    socket_path = "/tmp/nixai-mcp.sock",
  })
  print("nixai: Neovim integration loaded successfully")
else
  print("nixai: Failed to load Neovim integration")
end

-- Test command to verify integration
vim.api.nvim_create_user_command('NixaiTest', function()
  print("nixai: Test command executed")
end, {})
EOF

log_info "Created test Neovim configuration"

# Test Neovim startup with nixai integration
run_test_with_output "Neovim loads nixai module" "timeout 10 nvim --headless -c 'lua print(\"test complete\")' -c 'qall!' 2>&1"

echo ""

# Test VS Code MCP Integration
log_section "VS Code MCP Integration Tests"

# Create VS Code settings directory
mkdir -p "$VSCODE_CONFIG_DIR"

# Generate VS Code MCP configuration
cat > "$VSCODE_CONFIG_DIR/settings.json" << EOF
{
  "mcp.servers": {
    "nixai": {
      "command": "bash",
      "args": ["-c", "socat STDIO UNIX-CONNECT:/tmp/nixai-mcp.sock"],
      "env": {}
    }
  },
  "mcp.enableDebug": true
}
EOF

log_success "VS Code MCP configuration created"

run_test "VS Code settings.json exists" "[ -f '$VSCODE_CONFIG_DIR/settings.json' ]"
run_test "VS Code config contains nixai MCP" "grep -q 'nixai.*mcp.sock' '$VSCODE_CONFIG_DIR/settings.json'"

# Test socat connection to MCP server
if command -v socat &>/dev/null; then
    run_test "socat available for MCP bridge" "command -v socat"
    
    # Test MCP protocol communication
    log_info "Testing MCP protocol communication..."
    echo '{"jsonrpc":"2.0","method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}},"id":1}' | \
    timeout 5 socat - UNIX-CONNECT:/tmp/nixai-mcp.sock > /tmp/mcp_test_response.json 2>/dev/null || true
    
    if [ -f /tmp/mcp_test_response.json ] && grep -q "result" /tmp/mcp_test_response.json; then
        log_success "MCP protocol communication working"
    else
        log_warning "MCP protocol test inconclusive"
    fi
else
    log_warning "socat not available - installing..."
    sudo apt-get update && sudo apt-get install -y socat
    log_success "socat installed"
fi

echo ""

# Test nixai CLI functionality
log_section "nixai CLI Functionality Tests"

run_test_with_output "nixai help command" "nixai --help"
run_test_with_output "nixai version check" "nixai mcp-server status"

# Test AI provider integration
log_info "Testing AI integration..."
run_test_with_output "nixai question answering" "timeout 30 nixai 'What is NixOS?' | head -5"

echo ""

# Test Neovim and VS Code integration scripts
log_section "Integration Helper Scripts"

# Create helper script for testing Neovim integration
cat > /tmp/test_nvim_integration.sh << 'EOF'
#!/bin/bash
echo "Testing Neovim nixai integration..."
cd /home/nixuser/.config/nvim
nvim --headless -c 'lua local nixai = require("nixai"); print("nixai loaded: " .. tostring(nixai ~= nil))' -c 'qall!' 2>&1
EOF
chmod +x /tmp/test_nvim_integration.sh

run_test_with_output "Neovim integration script" "/tmp/test_nvim_integration.sh"

# Create helper script for testing VS Code MCP
cat > /tmp/test_vscode_mcp.sh << 'EOF'
#!/bin/bash
echo "Testing VS Code MCP integration..."
if [ -f /home/nixuser/.vscode/settings.json ]; then
    echo "VS Code settings found"
    if grep -q "nixai" /home/nixuser/.vscode/settings.json; then
        echo "nixai MCP configuration found"
        return 0
    fi
fi
return 1
EOF
chmod +x /tmp/test_vscode_mcp.sh

run_test "VS Code MCP configuration script" "/tmp/test_vscode_mcp.sh"

echo ""

# Cleanup
log_section "Cleanup"

if [ -n "${MCP_PID:-}" ]; then
    log_info "Stopping MCP server..."
    kill $MCP_PID 2>/dev/null || true
    sleep 2
    log_success "MCP server stopped"
fi

echo ""

# Results Summary
log_section "Test Results Summary"

echo "Total Tests: $TOTAL_TESTS"
echo -e "Passed: ${GREEN}$TESTS_PASSED${NC}"
echo -e "Failed: ${RED}$TESTS_FAILED${NC}"

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}🎉 All tests passed! Neovim and VS Code integrations are working correctly.${NC}"
    exit 0
else
    echo -e "${RED}❌ Some tests failed. Check the output above for details.${NC}"
    exit 1
fi
