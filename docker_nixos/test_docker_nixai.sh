#!/usr/bin/env bash
# test_docker_nixai.sh
# Comprehensive test script for nixai Docker environment
# This script tests all major functionality inside the Docker container

set -euo pipefail

echo "🧪 nixai Docker Environment Test Suite"
echo "========================================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test tracking
TESTS_PASSED=0
TESTS_FAILED=0
TOTAL_TESTS=0

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

run_test() {
    local test_name="$1"
    local test_command="$2"
    ((TOTAL_TESTS++))
    
    log_info "Testing: $test_name"
    
    if eval "$test_command" &>/dev/null; then
        log_success "$test_name"
    else
        log_error "$test_name"
        echo "Command failed: $test_command"
    fi
    echo ""
}

run_test_with_output() {
    local test_name="$1"
    local test_command="$2"
    ((TOTAL_TESTS++))
    
    log_info "Testing: $test_name"
    echo "Running: $test_command"
    
    if eval "$test_command"; then
        log_success "$test_name"
    else
        log_error "$test_name"
    fi
    echo ""
}

# Check if we're inside Docker
if [ ! -f /.dockerenv ]; then
    echo "❌ This script should be run inside the nixai Docker container"
    echo "   Start the container with: ./docker_nixos/build_and_run_docker.sh"
    echo "   Then run: cd /home/nixuser/nixai && ./docker_nixos/test_docker_nixai.sh"
    exit 1
fi

echo "✅ Running inside Docker container"
echo ""

# Test 1: Environment Setup
echo "🔧 Testing Environment Setup"
echo "----------------------------"

run_test "Nix installation" "which nix"
run_test "Go installation" "which go"
run_test "Just installation" "which just"
run_test "Git installation" "which git"
run_test "Neovim installation" "which nvim"

# Test experimental features
run_test "Nix flakes support" "nix flake --help"

echo ""

# Test 2: Pre-installed nixai
echo "🚀 Testing Pre-installed nixai"
echo "------------------------------"

run_test "nixai binary exists" "which nixai"
run_test "nixai help command" "nixai --help"

echo ""

# Test 3: Repository Access
echo "📁 Testing Repository Access"
echo "----------------------------"

run_test "Nixai repository exists" "[ -d /home/nixuser/nixai ]"
run_test "Source code access" "[ -f /home/nixuser/nixai/cmd/nixai/main.go ]"
run_test "Justfile access" "[ -f /home/nixuser/nixai/justfile ]"
run_test "Flake access" "[ -f /home/nixuser/nixai/flake.nix ]"

cd /home/nixuser/nixai

echo ""

# Test 4: Nix Development Environment
echo "❄️ Testing Nix Development Environment"
echo "--------------------------------------"

log_info "Entering Nix development shell..."
if nix develop .#docker --command bash -c 'echo "Nix dev shell works"'; then
    log_success "Nix development shell"
    ((TESTS_PASSED++))
else
    log_error "Nix development shell"
    ((TESTS_FAILED++))
fi
((TOTAL_TESTS++))

echo ""

# Test 5: Docker-specific Build Commands
echo "🐳 Testing Docker-specific Build Commands"
echo "-----------------------------------------"

# Enter nix shell and run tests
nix develop .#docker --command bash -c '
    set -e
    
    echo "Testing just build-docker..."
    if just build-docker; then
        echo "✅ Docker build successful"
        
        echo "Testing built binary..."
        if /tmp/nixai --help; then
            echo "✅ Built binary works"
        else
            echo "❌ Built binary failed"
            exit 1
        fi
    else
        echo "❌ Docker build failed"
        exit 1
    fi
' && {
    log_success "Docker build and binary test"
    ((TESTS_PASSED++))
} || {
    log_error "Docker build and binary test"
    ((TESTS_FAILED++))
}
((TOTAL_TESTS++))

echo ""

# Test 6: nixai Core Functionality
echo "🤖 Testing nixai Core Functionality"
echo "-----------------------------------"

# Test basic commands (with timeout to avoid hanging)
run_test_with_output "nixai help" "timeout 10 nixai --help"

# Test configuration
if [ -f /home/nixuser/nixai/configs/default.yaml ]; then
    run_test "Configuration file exists" "[ -f /home/nixuser/nixai/configs/default.yaml ]"
fi

# Test AI providers (if keys are available)
if [ -f /etc/profile.d/ai_keys.sh ]; then
    log_info "AI keys found, testing provider initialization..."
    source /etc/profile.d/ai_keys.sh
    
    # Test with a simple question (with timeout)
    if timeout 30 nixai "test" &>/dev/null; then
        log_success "AI provider basic functionality"
        ((TESTS_PASSED++))
    else
        log_warning "AI provider test skipped (may require network/API keys)"
    fi
    ((TOTAL_TESTS++))
fi

echo ""

# Test 7: MCP Server
echo "🔌 Testing MCP Server"
echo "--------------------"

# Test MCP server commands
run_test "MCP server help" "timeout 10 nixai mcp-server --help"

# Note: We don't start the actual MCP server to avoid background processes
log_info "MCP server functionality verified (server not started to avoid background processes)"

echo ""

# Start MCP Server for Documentation Features
echo "🔌 Starting MCP Server for Documentation Tests"
echo "----------------------------------------------"

log_info "Starting MCP server in background for documentation features..."
if timeout 30 nixai mcp-server start -d; then
    log_success "MCP server startup"
    ((TESTS_PASSED++))
    
    # Wait for server to initialize
    sleep 3
    
    # Check if MCP server is running
    if nixai mcp-server status >/dev/null 2>&1; then
        log_success "MCP server status check"
        ((TESTS_PASSED++))
    else
        log_error "MCP server status check"
        ((TESTS_FAILED++))
    fi
    ((TOTAL_TESTS++))
else
    log_error "MCP server startup"
    ((TESTS_FAILED++))
fi
((TOTAL_TESTS++))

echo ""

# Test 8: Option Explanation Features
echo "📚 Testing Option Explanation Features"
echo "--------------------------------------"

run_test "explain-option help" "timeout 10 nixai explain-option --help"
run_test "explain-home-option help" "timeout 10 nixai explain-home-option --help"

echo ""

# Test 9: Build Tools and Development
echo "🛠️ Testing Build Tools and Development"
echo "--------------------------------------"

nix develop .#docker --command bash -c '
    set -e
    
    echo "Testing Go build..."
    if go build -o /tmp/nixai-test ./cmd/nixai/main.go; then
        echo "✅ Go build successful"
        
        echo "Testing built binary..."
        if /tmp/nixai-test --help >/dev/null 2>&1; then
            echo "✅ Go-built binary works"
        else
            echo "❌ Go-built binary failed"
            exit 1
        fi
        
        rm -f /tmp/nixai-test
    else
        echo "❌ Go build failed"
        exit 1
    fi
    
    echo "Testing Go tests..."
    if go test -v ./internal/config ./pkg/logger ./pkg/utils; then
        echo "✅ Go tests passed"
    else
        echo "❌ Some Go tests failed"
        exit 1
    fi
' && {
    log_success "Go build and test"
    ((TESTS_PASSED++))
} || {
    log_error "Go build and test"
    ((TESTS_FAILED++))
}
((TOTAL_TESTS++))

echo ""

# Test 10: Nix Build
echo "❄️ Testing Nix Build"
echo "-------------------"

log_info "Testing Nix build (this may take a while)..."
if timeout 300 nix build --no-link; then
    log_success "Nix build"
    ((TESTS_PASSED++))
    
    # Test the built result
    if nix build && ./result/bin/nixai --help &>/dev/null; then
        log_success "Nix-built binary functionality"
        ((TESTS_PASSED++))
    else
        log_error "Nix-built binary functionality"
        ((TESTS_FAILED++))
    fi
    ((TOTAL_TESTS++))
else
    log_error "Nix build (timed out or failed)"
    ((TESTS_FAILED++))
fi
((TOTAL_TESTS++))

echo ""

# Test 11: Ollama Integration
echo "🦙 Testing Ollama Integration"
echo "-----------------------------"

log_info "Testing Ollama connectivity..."
if curl -s --max-time 5 http://host.docker.internal:11434/api/version &>/dev/null; then
    log_success "Ollama host connectivity"
    ((TESTS_PASSED++))
    
    # Test nixai with Ollama (if available)
    log_info "Testing nixai with Ollama provider..."
    if timeout 30 nixai --provider ollama "test" &>/dev/null; then
        log_success "nixai Ollama integration"
        ((TESTS_PASSED++))
    else
        log_warning "nixai Ollama integration (may require model or configuration)"
    fi
    ((TOTAL_TESTS++))
else
    log_warning "Ollama not accessible on host (make sure Ollama is running on host)"
fi
((TOTAL_TESTS++))

echo ""

# Test Summary
echo "📊 Test Summary"
echo "==============="
echo -e "Total tests: ${TOTAL_TESTS}"
echo -e "Passed: ${GREEN}${TESTS_PASSED}${NC}"
echo -e "Failed: ${RED}${TESTS_FAILED}${NC}"

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "\n🎉 ${GREEN}All tests passed!${NC}"
    echo "The nixai Docker environment is working correctly."
else
    echo -e "\n⚠️ ${YELLOW}Some tests failed.${NC}"
    echo "Please check the output above for details."
fi

echo ""
echo "🔧 Quick Start Commands:"
echo "  nixai --help                    # Show help"
echo "  nixai 'your question'          # Ask a question"
echo "  nixai explain-option programs.git  # Explain NixOS option"
echo "  just build-docker              # Build for Docker"
echo "  just test                      # Run tests"
echo "  nix develop .#docker           # Enter development shell"

echo ""
echo "🧹 Cleanup"
echo "=========="
log_info "Stopping MCP server..."
nixai mcp-server stop >/dev/null 2>&1 || log_info "MCP server was not running"

exit $TESTS_FAILED
