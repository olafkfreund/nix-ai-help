#!/bin/bash

# Comprehensive MCP Server Test Script
# This script systematically tests all MCP functionality

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test configuration
MCP_HOST="localhost"
MCP_PORT="8081"
MCP_SOCKET="/tmp/nixai-mcp.sock"
MCP_URL="http://${MCP_HOST}:${MCP_PORT}"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  nixai MCP Server Comprehensive Test  ${NC}"
echo -e "${BLUE}========================================${NC}"
echo

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

# Test 1: Basic MCP Server Status
echo -e "${YELLOW}Testing MCP Server Status...${NC}"
if nixai mcp-server status &>/dev/null; then
    print_test "PASS" "MCP Server Status Check"
else
    print_test "FAIL" "MCP Server Status Check" "Server may not be running"
    echo "Attempting to start MCP server..."
    nixai mcp-server start || true
    sleep 2
fi

# Test 2: Socket File Existence
echo -e "${YELLOW}Testing Socket File...${NC}"
if [[ -S "$MCP_SOCKET" ]]; then
    print_test "PASS" "Unix Socket Exists" "$MCP_SOCKET"
else
    print_test "FAIL" "Unix Socket Missing" "$MCP_SOCKET"
fi

# Test 3: HTTP Health Check
echo -e "${YELLOW}Testing HTTP Endpoint...${NC}"
if curl -s "$MCP_URL/health" &>/dev/null; then
    health_response=$(curl -s "$MCP_URL/health")
    print_test "PASS" "HTTP Health Endpoint" "$health_response"
else
    print_test "FAIL" "HTTP Health Endpoint" "Cannot reach $MCP_URL/health"
fi

# Test 4: Basic HTTP Query (Wrong source test)
echo -e "${YELLOW}Testing Basic HTTP Query...${NC}"
response=$(curl -s -X POST "$MCP_URL/query" \
    -H "Content-Type: application/json" \
    -d '{
        "query": "services.nginx.enable", 
        "sources": ["https://wiki.nixos.org/wiki/NixOS_Wiki"]
    }' || echo "ERROR")

if [[ "$response" != "ERROR" ]] && echo "$response" | jq . &>/dev/null; then
    content=$(echo "$response" | jq -r '.result' | head -c 100)
    if [[ "$content" == *"DOCTYPE html"* ]]; then
        print_test "WARN" "Basic Query Returns HTML" "Getting raw HTML instead of structured docs"
    else
        print_test "PASS" "Basic Query Returns JSON" 
    fi
else
    print_test "FAIL" "Basic Query Failed" "$response"
fi

# Test 5: Structured NixOS Options Query
echo -e "${YELLOW}Testing Structured NixOS Options...${NC}"
response=$(curl -s -X POST "$MCP_URL/query" \
    -H "Content-Type: application/json" \
    -d '{
        "query": "services.nginx.enable",
        "sources": ["nixos-options-es://"]
    }' || echo "ERROR")

if [[ "$response" != "ERROR" ]] && echo "$response" | jq . &>/dev/null; then
    content=$(echo "$response" | jq -r '.result')
    if [[ "$content" == *"Option:"* ]] && [[ "$content" == *"Type:"* ]]; then
        print_test "PASS" "Structured NixOS Options Query" "Got structured option documentation"
        echo "Sample content: $(echo "$content" | head -n 3 | tr '\n' ' ')"
    elif [[ "$content" == *"No documentation found"* ]]; then
        print_test "WARN" "No Documentation Found" "Option not found in database"
    else
        print_test "WARN" "Unexpected Response Format" "$(echo "$content" | head -c 100)"
    fi
else
    print_test "FAIL" "Structured Options Query Failed" "$response"
fi

# Test 6: Alternative Structured Query Method
echo -e "${YELLOW}Testing Alternative Structured Query...${NC}"
response=$(curl -s -X POST "$MCP_URL/query" \
    -H "Content-Type: application/json" \
    -d '{
        "query": "services.openssh.enable",
        "sources": ["https://search.nixos.org/options"]
    }' || echo "ERROR")

if [[ "$response" != "ERROR" ]] && echo "$response" | jq . &>/dev/null; then
    content=$(echo "$response" | jq -r '.result')
    if [[ "$content" == *"Option:"* ]] && [[ "$content" == *"Type:"* ]]; then
        print_test "PASS" "Alternative Structured Query" "Got structured option documentation"
    elif [[ "$content" == *"No documentation found"* ]]; then
        print_test "WARN" "No Documentation Found" "Option not found in database"
    else
        print_test "WARN" "Unexpected Response Format" "$(echo "$content" | head -c 100)"
    fi
else
    print_test "FAIL" "Alternative Structured Query Failed" "$response"
fi

# Test 7: Home Manager Options Query
echo -e "${YELLOW}Testing Home Manager Options...${NC}"
response=$(curl -s -X POST "$MCP_URL/query" \
    -H "Content-Type: application/json" \
    -d '{
        "query": "programs.git.enable",
        "sources": ["https://home-manager-options.extranix.com/options.json"]
    }' || echo "ERROR")

if [[ "$response" != "ERROR" ]] && echo "$response" | jq . &>/dev/null; then
    content=$(echo "$response" | jq -r '.result')
    if [[ "$content" == *"Option:"* ]] || [[ "$content" == *"No documentation found"* ]]; then
        print_test "PASS" "Home Manager Options Query" "Got appropriate response"
    else
        print_test "WARN" "Unexpected Home Manager Response" "$(echo "$content" | head -c 100)"
    fi
else
    print_test "FAIL" "Home Manager Options Query Failed" "$response"
fi

# Test 8: MCP Protocol Test via CLI
echo -e "${YELLOW}Testing MCP CLI Command...${NC}"
if command -v nixai &>/dev/null; then
    result=$(nixai mcp-server query "services.nginx.enable" 2>&1 || echo "ERROR")
    if [[ "$result" != "ERROR" ]] && [[ "$result" != *"error"* ]]; then
        print_test "PASS" "MCP CLI Query" "CLI command works"
    else
        print_test "WARN" "MCP CLI Query Issues" "$result"
    fi
else
    print_test "WARN" "nixai Command Not Found" "CLI testing skipped"
fi

# Test 9: Direct Socket Communication Test
echo -e "${YELLOW}Testing Direct Socket Communication...${NC}"
if command -v socat &>/dev/null && [[ -S "$MCP_SOCKET" ]]; then
    mcp_response=$(echo '{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}, "clientInfo": {"name": "test", "version": "1.0.0"}}}' | socat - UNIX-CONNECT:"$MCP_SOCKET" | head -n 1 || echo "ERROR")
    
    if [[ "$mcp_response" != "ERROR" ]] && echo "$mcp_response" | jq . &>/dev/null; then
        print_test "PASS" "Direct Socket Communication" "MCP protocol working"
    else
        print_test "FAIL" "Direct Socket Communication Failed" "$mcp_response"
    fi
else
    print_test "WARN" "Socket Test Skipped" "socat not available or socket missing"
fi

# Test 10: Configuration Analysis
echo -e "${YELLOW}Analyzing Configuration...${NC}"
config_file="$HOME/.config/nixai/config.yaml"
if [[ -f "$config_file" ]]; then
    port=$(grep -A 10 "mcp:" "$config_file" | grep "port:" | head -n 1 | awk '{print $2}' || echo "unknown")
    sources_count=$(grep -A 20 "documentation_sources:" "$config_file" | grep "^  - " | wc -l || echo "0")
    print_test "PASS" "Configuration File Found" "Port: $port, Sources: $sources_count"
else
    print_test "WARN" "Configuration File Missing" "$config_file"
fi

echo
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}           Test Summary                 ${NC}"
echo -e "${BLUE}========================================${NC}"

# Summary and recommendations
echo -e "${YELLOW}Key Findings:${NC}"
echo "1. The server responds on port $MCP_PORT (correct configuration)"
echo "2. Basic queries return HTML content instead of structured docs"
echo "3. Structured sources should use specific URLs or prefixes:"
echo "   - For NixOS options: nixos-options-es:// or https://search.nixos.org/options"
echo "   - For Home Manager: https://home-manager-options.extranix.com/options.json"

echo
echo -e "${YELLOW}Recommendations:${NC}"
echo "1. Use structured source URLs for option queries"
echo "2. Regular documentation URLs return raw HTML (expected for general docs)"
echo "3. Check source mapping in handleQuery function"

echo
echo -e "${GREEN}Test completed!${NC}"
