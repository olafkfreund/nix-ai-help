#!/bin/bash
#
# Test environment compatibility checker for NixAI tests
#

echo "🧪 NixAI Test Environment Checker"
echo "================================="

# Define color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Check Python version (must be 3.8+)
echo -n "Checking Python version: "
if command -v python3 &> /dev/null; then
    python_version=$(python3 --version | awk '{print $2}')
    python_major=$(echo $python_version | cut -d. -f1)
    python_minor=$(echo $python_version | cut -d. -f2)
    
    if [ $python_major -ge 3 ] && [ $python_minor -ge 8 ]; then
        echo -e "${GREEN}✅ Python $python_version${NC}"
    else
        echo -e "${RED}❌ Python $python_version (required: 3.8+)${NC}"
    fi
else
    echo -e "${RED}❌ Python 3 not found${NC}"
fi

# Check for socat (needed for Unix socket tests)
echo -n "Checking for socat: "
if command -v socat &> /dev/null; then
    socat_version=$(socat -V 2>&1 | head -n 1)
    echo -e "${GREEN}✅ $socat_version${NC}"
else
    echo -e "${RED}❌ socat not found (required for MCP socket tests)${NC}"
    echo -e "  ${YELLOW}Install with: sudo apt install socat${NC} (Ubuntu/Debian)"
    echo -e "  ${YELLOW}Install with: sudo pacman -S socat${NC} (Arch)"
    echo -e "  ${YELLOW}Install with: nix-env -i socat${NC} (NixOS)"
fi

# Check for curl (needed for HTTP tests)
echo -n "Checking for curl: "
if command -v curl &> /dev/null; then
    curl_version=$(curl --version | head -n 1 | awk '{print $1" "$2}')
    echo -e "${GREEN}✅ $curl_version${NC}"
else
    echo -e "${RED}❌ curl not found (required for HTTP tests)${NC}"
fi

# Check for VS Code (optional, needed for full VS Code integration tests)
echo -n "Checking for VS Code: "
if command -v code &> /dev/null; then
    code_version=$(code --version | head -n 1)
    echo -e "${GREEN}✅ VS Code $code_version${NC}"
else
    echo -e "${YELLOW}⚠️ VS Code not found (optional, needed for full VS Code integration)${NC}"
fi

# Check if NixAI binary exists
echo -n "Checking for nixai binary: "
if [ -x "./nixai" ]; then
    nixai_version=$(./nixai --version 2>/dev/null || echo "version unknown")
    echo -e "${GREEN}✅ nixai $nixai_version${NC}"
else
    echo -e "${RED}❌ nixai binary not found (run 'just build' first)${NC}"
fi

echo "================================="
echo "Run ./tests/run_all.sh to execute all tests"
echo "Or use 'just test-all' to run via justfile"
