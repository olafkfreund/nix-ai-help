#!/bin/bash

# Test script to verify that the MCP server properly uses custom sources from POST requests
# This tests the fix for the bug where the server ignored the 'sources' field

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

MCP_URL="http://localhost:8081"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}   Testing MCP Sources Fix             ${NC}"
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

# Test 1: Use NixOS ElasticSearch API directly
echo -e "${YELLOW}Test 1: Direct NixOS ElasticSearch Query...${NC}"
response=$(curl -s -X POST "$MCP_URL/query" \
    -H "Content-Type: application/json" \
    -d '{
        "query": "services.nginx.enable",
        "sources": ["nixos-options-es://options"]
    }' || echo "ERROR")

if [[ "$response" != "ERROR" ]] && echo "$response" | jq . &>/dev/null; then
    content=$(echo "$response" | jq -r '.result')
    if [[ "$content" == *"Option:"* ]] && [[ "$content" == *"services.nginx.enable"* ]]; then
        print_test "PASS" "ElasticSearch API Query" "Got structured NixOS option data"
        echo "    Content preview: $(echo "$content" | head -n 2 | tr '\n' ' ')"
    elif [[ "$content" == *"No documentation found"* ]]; then
        print_test "WARN" "ElasticSearch API Query" "Option not found in database (expected for some options)"
    else
        print_test "FAIL" "ElasticSearch API Query" "Unexpected response format"
        echo "    Content: $(echo "$content" | head -c 150)..."
    fi
else
    print_test "FAIL" "ElasticSearch API Query" "Request failed or invalid JSON response"
fi

echo

# Test 2: Use alternative structured source
echo -e "${YELLOW}Test 2: Alternative Structured Source Query...${NC}"
response=$(curl -s -X POST "$MCP_URL/query" \
    -H "Content-Type: application/json" \
    -d '{
        "query": "services.openssh.enable",
        "sources": ["https://search.nixos.org/options"]
    }' || echo "ERROR")

if [[ "$response" != "ERROR" ]] && echo "$response" | jq . &>/dev/null; then
    content=$(echo "$response" | jq -r '.result')
    if [[ "$content" == *"Option:"* ]] && [[ "$content" == *"openssh"* ]]; then
        print_test "PASS" "Alternative Structured Query" "Got structured option data"
    elif [[ "$content" == *"No documentation found"* ]]; then
        print_test "WARN" "Alternative Structured Query" "Option not found"
    else
        print_test "FAIL" "Alternative Structured Query" "Unexpected response"
        echo "    Content: $(echo "$content" | head -c 150)..."
    fi
else
    print_test "FAIL" "Alternative Structured Query" "Request failed"
fi

echo

# Test 3: Verify that wrong sources are NOT used (this tests the fix)
echo -e "${YELLOW}Test 3: Custom Sources Override Test...${NC}"
response=$(curl -s -X POST "$MCP_URL/query" \
    -H "Content-Type: application/json" \
    -d '{
        "query": "services.nginx.enable",
        "sources": ["https://example.com/nonexistent"]
    }' || echo "ERROR")

if [[ "$response" != "ERROR" ]] && echo "$response" | jq . &>/dev/null; then
    content=$(echo "$response" | jq -r '.result')
    if [[ "$content" == *"No relevant documentation found"* ]] || [[ "$content" == *"failed to fetch"* ]]; then
        print_test "PASS" "Custom Sources Override" "Server correctly used custom source (which failed as expected)"
        echo "    This confirms the server is NOT falling back to default sources"
    elif [[ "$content" == *"wiki.nixos.org"* ]] || [[ "$content" == *"nix.dev"* ]]; then
        print_test "FAIL" "Custom Sources Override" "Server ignored custom sources and used defaults"
        echo "    This indicates the bug is NOT fixed"
    else
        print_test "WARN" "Custom Sources Override" "Unexpected response, need manual verification"
        echo "    Content: $(echo "$content" | head -c 150)..."
    fi
else
    print_test "FAIL" "Custom Sources Override" "Request failed"
fi

echo

# Test 4: Verify default sources are used when no sources specified
echo -e "${YELLOW}Test 4: Default Sources Fallback Test...${NC}"
response=$(curl -s -X POST "$MCP_URL/query" \
    -H "Content-Type: application/json" \
    -d '{
        "query": "services.nginx.enable"
    }' || echo "ERROR")

if [[ "$response" != "ERROR" ]] && echo "$response" | jq . &>/dev/null; then
    content=$(echo "$response" | jq -r '.result')
    if [[ "$content" == *"wiki.nixos.org"* ]] || [[ "$content" == *"nix.dev"* ]] || [[ "$content" == *"Option:"* ]]; then
        print_test "PASS" "Default Sources Fallback" "Server used default sources when none specified"
    else
        print_test "WARN" "Default Sources Fallback" "Unexpected response format"
        echo "    Content: $(echo "$content" | head -c 150)..."
    fi
else
    print_test "FAIL" "Default Sources Fallback" "Request failed"
fi

echo

# Test 5: Test cache behavior with different sources
echo -e "${YELLOW}Test 5: Cache Behavior with Different Sources...${NC}"

# First request with specific source
response1=$(curl -s -X POST "$MCP_URL/query" \
    -H "Content-Type: application/json" \
    -d '{
        "query": "services.ssh.enable",
        "sources": ["nixos-options-es://options"]
    }' || echo "ERROR")

# Second request with different source for same query
response2=$(curl -s -X POST "$MCP_URL/query" \
    -H "Content-Type: application/json" \
    -d '{
        "query": "services.ssh.enable",
        "sources": ["https://wiki.nixos.org/wiki/NixOS_Wiki"]
    }' || echo "ERROR")

if [[ "$response1" != "ERROR" ]] && [[ "$response2" != "ERROR" ]]; then
    content1=$(echo "$response1" | jq -r '.result' 2>/dev/null || echo "$response1")
    content2=$(echo "$response2" | jq -r '.result' 2>/dev/null || echo "$response2")
    
    if [[ "$content1" != "$content2" ]]; then
        print_test "PASS" "Cache Behavior Test" "Different sources produce different results (cache keys work correctly)"
    else
        print_test "WARN" "Cache Behavior Test" "Same results from different sources (may indicate caching issue)"
    fi
else
    print_test "FAIL" "Cache Behavior Test" "One or both requests failed"
fi

echo
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}           Test Summary                 ${NC}"
echo -e "${BLUE}========================================${NC}"
echo -e "${YELLOW}Key Findings:${NC}"
echo "1. The server should now use custom sources specified in POST requests"
echo "2. Cache keys should include both query and sources to avoid conflicts"
echo "3. Default sources should only be used when no custom sources are provided"
echo "4. ElasticSearch API should provide structured option documentation"

echo
echo -e "${GREEN}Sources fix test completed!${NC}"
