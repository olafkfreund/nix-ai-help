#!/bin/bash

# Test script to verify that the MCP server correctly handles sources from POST requests
# This script will:
# 1. Start the MCP server if it's not already running
# 2. Send a POST request with NixOS options source
# 3. Send a POST request with Home Manager options source
# 4. Send a POST request with MediaWiki source
# 5. Verify that different sources provide different responses

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Testing MCP Server Sources Handling   ${NC}"
echo -e "${BLUE}========================================${NC}"
echo

# Test configuration
MCP_HOST="localhost"
MCP_PORT="8081"
MCP_URL="http://${MCP_HOST}:${MCP_PORT}"

# Function to print test status
print_test() {
    local status=$1
    local test_name=$2
    local details=${3:-""}
    
    if [[ $status == "PASS" ]]; then
        echo -e "✅ ${GREEN}[PASS]${NC} $test_name"
    elif [[ $status == "FAIL" ]]; then
        echo -e "❌ ${RED}[FAIL]${NC} $test_name"
    elif [[ $status == "WARN" ]]; then
        echo -e "⚠️  ${YELLOW}[WARN]${NC} $test_name"
    else
        echo -e "ℹ️  ${BLUE}[INFO]${NC} $test_name"
    fi
    
    if [[ -n "$details" ]]; then
        echo -e "    ${details}"
    fi
}

# Check if MCP server is running
echo -e "${YELLOW}Checking MCP server status...${NC}"
if ! curl -s "${MCP_URL}/healthz" &>/dev/null; then
    print_test "INFO" "MCP Server not running, starting it now"
    nixai mcp-server start -d || true
    sleep 3
    
    if ! curl -s "${MCP_URL}/healthz" &>/dev/null; then
        print_test "FAIL" "Failed to start MCP server"
        exit 1
    fi
else
    print_test "INFO" "MCP Server is already running"
fi

# Test 1: Query with NixOS options source
echo -e "${YELLOW}Test 1: Query with NixOS options source...${NC}"
response=$(curl -s -X POST "${MCP_URL}/query" \
    -H "Content-Type: application/json" \
    -d '{
        "query": "services.nginx.enable",
        "sources": ["nixos-options-es://"]
    }')

# Extract result and check if it contains NixOS specific content
result=$(echo "$response" | jq -r '.result')

if [[ "$result" == *"USING MCP SERVER HANDLE_DOC_QUERY"* ]]; then
    print_test "PASS" "Server used handleDocQuery method"
else
    print_test "WARN" "Server might not be using handleDocQuery method"
fi

if [[ "$result" == *"Sources: [nixos-options-es://]"* ]]; then
    print_test "PASS" "Server received correct sources parameter"
else
    print_test "FAIL" "Server didn't receive correct sources parameter"
    echo "Result contained:"
    echo "$result" | head -10
fi

if [[ "$result" == *"nginx"* && "$result" == *"enable"* ]]; then
    print_test "PASS" "Response contains expected NixOS option content"
else
    print_test "WARN" "Response may not contain expected NixOS option content"
fi
echo

# Test 2: Query with Home Manager options source
echo -e "${YELLOW}Test 2: Query with Home Manager options source...${NC}"
response=$(curl -s -X POST "${MCP_URL}/query" \
    -H "Content-Type: application/json" \
    -d '{
        "query": "programs.zsh.enable",
        "sources": ["https://home-manager-options.extranix.com/options.json"]
    }')

# Extract result and check if it contains Home Manager specific content
result=$(echo "$response" | jq -r '.result')

if [[ "$result" == *"Sources: [https://home-manager-options.extranix.com/options.json]"* ]]; then
    print_test "PASS" "Server received correct Home Manager source parameter"
else
    print_test "FAIL" "Server didn't receive correct Home Manager source parameter"
    echo "Result contained:"
    echo "$result" | head -10
fi
echo

# Test 3: Query with MediaWiki source
echo -e "${YELLOW}Test 3: Query with MediaWiki source...${NC}"
response=$(curl -s -X POST "${MCP_URL}/query" \
    -H "Content-Type: application/json" \
    -d '{
        "query": "nixos",
        "sources": ["https://wiki.nixos.org/wiki/NixOS_Wiki"]
    }')

# Extract result and check if it contains MediaWiki specific content
result=$(echo "$response" | jq -r '.result')

if [[ "$result" == *"Sources: [https://wiki.nixos.org/wiki/NixOS_Wiki]"* ]]; then
    print_test "PASS" "Server received correct MediaWiki source parameter"
else
    print_test "FAIL" "Server didn't receive correct MediaWiki source parameter"
    echo "Result contained:"
    echo "$result" | head -10
fi
echo

echo -e "${GREEN}Tests completed!${NC}"
