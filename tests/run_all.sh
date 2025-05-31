#!/bin/bash
#
# Run all test suites for NixAI
#

set -e

# Determine path to repository root (handles running from any directory)
REPO_ROOT=$(git rev-parse --show-toplevel 2>/dev/null || echo "$HOME/Source/NIX/nix-ai-help")
NIXAI_BIN="$REPO_ROOT/nixai"

echo "🧪 Running All NixAI Tests"
echo "=========================="

# First check for test environment compatibility
echo "Checking test environment compatibility..."
"$REPO_ROOT/tests/check-compatibility.sh"

# Define color codes for output
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Track overall status
OVERALL_STATUS=0

# Helper function to run a test script and track status
run_test() {
    local test_script=$1
    local test_name=$2
    
    echo -e "\n${BLUE}🔬 Running $test_name tests...${NC}"
    
    if [ -x "$test_script" ]; then
        if $test_script; then
            echo -e "${GREEN}✅ $test_name tests PASSED${NC}"
        else
            echo -e "${RED}❌ $test_name tests FAILED${NC}"
            OVERALL_STATUS=1
        fi
    else
        echo -e "${RED}❌ $test_script not found or not executable${NC}"
        OVERALL_STATUS=1
    fi
}

# MCP Tests
run_test "$REPO_ROOT/tests/run_mcp.sh" "MCP"

# VS Code Integration Tests
run_test "$REPO_ROOT/tests/run_vscode.sh" "VS Code"

# Provider Tests
run_test "$REPO_ROOT/tests/run_providers.sh" "AI Provider"

# Go Unit Tests
echo -e "\n${BLUE}🔬 Running Go unit tests...${NC}"
if go test ./...; then
    echo -e "${GREEN}✅ Go unit tests PASSED${NC}"
else
    echo -e "${RED}❌ Go unit tests FAILED${NC}"
    OVERALL_STATUS=1
fi

echo ""
echo "=========================="
if [ $OVERALL_STATUS -eq 0 ]; then
    echo -e "${GREEN}✅ ALL TESTS PASSED${NC}"
else
    echo -e "${RED}❌ SOME TESTS FAILED${NC}"
fi

exit $OVERALL_STATUS
