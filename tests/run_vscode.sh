#!/bin/bash
#
# Run VS Code integration tests for NixAI
#

set -e

echo "🧪 Running VS Code Integration Tests"
echo "==================================="

# Define color codes for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Track overall status
OVERALL_STATUS=0

# Determine path to nixai executable (handles running from any directory)
REPO_ROOT=$(git rev-parse --show-toplevel 2>/dev/null || echo "$HOME/Source/NIX/nix-ai-help")
NIXAI_BIN="$REPO_ROOT/nixai"

# Check if MCP server is running (required for VS Code tests)
echo "Checking MCP server status..."
if pgrep -f "nixai mcp-server" > /dev/null; then
    echo -e "${GREEN}✅ MCP server is running${NC}"
else
    echo -e "${RED}⚠️  MCP server not running, starting...${NC}"
    $NIXAI_BIN mcp-server start -d
    sleep 3
    
    if pgrep -f "nixai mcp-server" > /dev/null; then
        echo -e "${GREEN}✅ MCP server started successfully${NC}"
    else
        echo -e "${RED}❌ Failed to start MCP server${NC}"
        exit 1
    fi
fi

echo ""

# Run all VS Code tests
for test_script in tests/vscode/*.{py,sh}; do
    if [ -x "$test_script" ]; then
        echo "Running $test_script..."
        if $test_script; then
            echo -e "${GREEN}✅ $test_script PASSED${NC}"
        else
            echo -e "${RED}❌ $test_script FAILED${NC}"
            OVERALL_STATUS=1
        fi
        echo ""
    fi
done

echo "==================================="
if [ $OVERALL_STATUS -eq 0 ]; then
    echo -e "${GREEN}✅ ALL VS CODE TESTS PASSED${NC}"
else
    echo -e "${RED}❌ SOME VS CODE TESTS FAILED${NC}"
fi

exit $OVERALL_STATUS
