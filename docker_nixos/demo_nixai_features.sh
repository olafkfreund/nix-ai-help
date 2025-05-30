#!/usr/bin/env bash
# demo_nixai_features.sh
# Interactive demo of nixai features in Docker environment
# Run this inside the Docker container to see nixai in action

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Demo functions
demo_header() {
    echo -e "\n${CYAN}========================================${NC}"
    echo -e "${CYAN}$1${NC}"
    echo -e "${CYAN}========================================${NC}\n"
}

demo_step() {
    echo -e "${BLUE}[DEMO]${NC} $1"
    echo -e "${YELLOW}Running:${NC} $2"
    echo ""
}

wait_for_user() {
    echo -e "${GREEN}Press Enter to continue...${NC}"
    read -r
}

# Check if we're inside Docker
if [ ! -f /.dockerenv ]; then
    echo -e "${RED}❌ This demo should be run inside the nixai Docker container${NC}"
    echo "   Start the container with: ./docker_nixos/build_and_run_docker.sh"
    echo "   Then run: cd /home/nixuser/nixai && ./docker_nixos/demo_nixai_features.sh"
    exit 1
fi

echo -e "${GREEN}🐳 nixai Docker Environment Demo${NC}"
echo -e "${GREEN}==================================${NC}"
echo ""
echo "This demo will showcase all major nixai features in the Docker environment."
echo "Each step will show you the command and then execute it."
echo ""

wait_for_user

# Demo 1: Basic nixai functionality
demo_header "1. Basic nixai Functionality"

demo_step "Show nixai help" "nixai --help"
nixai --help
echo ""

demo_step "Show nixai version" "nixai --version"
nixai --version
echo ""

wait_for_user

# Demo 2: NixOS Option Explanation
demo_header "2. NixOS Option Explanation"

demo_step "Explain programs.git option" "nixai explain-option programs.git"
echo -e "${YELLOW}Note: This may take a moment as it queries documentation...${NC}"
timeout 30 nixai explain-option programs.git || echo -e "${YELLOW}(Timed out - this feature requires MCP server)${NC}"
echo ""

wait_for_user

# Demo 3: Home Manager Integration
demo_header "3. Home Manager Integration"

demo_step "Explain Home Manager Neovim option" "nixai explain-home-option programs.neovim"
echo -e "${YELLOW}Note: This may take a moment as it queries documentation...${NC}"
timeout 30 nixai explain-home-option programs.neovim || echo -e "${YELLOW}(Timed out - this feature requires MCP server)${NC}"
echo ""

wait_for_user

# Demo 4: Direct Question Answering
demo_header "4. Direct Question Answering"

demo_step "Ask a direct question about NixOS" 'nixai "How do I install packages in NixOS?"'
echo -e "${YELLOW}Note: This requires AI provider and may take a moment...${NC}"
timeout 45 nixai "How do I install packages in NixOS?" || echo -e "${YELLOW}(Timed out or requires AI provider configuration)${NC}"
echo ""

wait_for_user

# Demo 5: Development Environment
demo_header "5. Development Environment"

demo_step "Enter Nix development shell" "nix develop .#docker --command bash -c 'echo Inside Nix dev shell && which go && which just'"
cd /home/nixuser/nixai
nix develop .#docker --command bash -c 'echo "✅ Inside Nix dev shell" && echo "Go version: $(go version)" && echo "Just version: $(just --version)"'
echo ""

wait_for_user

# Demo 6: Building nixai
demo_header "6. Building nixai from Source"

demo_step "Build nixai for Docker" "just build-docker"
cd /home/nixuser/nixai
nix develop .#docker --command just build-docker
echo ""

demo_step "Test the built binary" "/tmp/nixai --help"
/tmp/nixai --help | head -10
echo "..."
echo ""

wait_for_user

# Demo 7: Testing
demo_header "7. Testing Framework"

demo_step "Run basic Go tests" "go test ./pkg/utils"
cd /home/nixuser/nixai
nix develop .#docker --command go test -v ./pkg/utils
echo ""

wait_for_user

# Demo 8: MCP Server
demo_header "8. MCP Server Integration"

demo_step "Show MCP server help" "nixai mcp-server --help"
nixai mcp-server --help
echo ""

echo -e "${BLUE}[DEMO]${NC} MCP server provides documentation integration for:"
echo "  • NixOS Wiki"
echo "  • Nix Manual"
echo "  • Nixpkgs Manual"
echo "  • Home Manager Documentation"
echo ""

wait_for_user

# Demo 9: Nix Build
demo_header "9. Nix Build System"

demo_step "Build with Nix" "nix build"
cd /home/nixuser/nixai
echo -e "${YELLOW}Note: This may take several minutes on first run...${NC}"
if timeout 300 nix build --no-link; then
    echo -e "${GREEN}✅ Nix build successful${NC}"
else
    echo -e "${YELLOW}(Nix build timed out - this is normal for first run)${NC}"
fi
echo ""

wait_for_user

# Demo 10: Integration Examples
demo_header "10. Real-world Usage Examples"

echo -e "${BLUE}[DEMO]${NC} Here are some real-world nixai usage examples:"
echo ""

echo -e "${CYAN}Configuration Help:${NC}"
echo "  nixai 'Generate NixOS config for Nginx with SSL'"
echo "  nixai 'How do I configure SSH keys in NixOS?'"
echo ""

echo -e "${CYAN}Package Management:${NC}"
echo "  nixai 'What is the difference between nix-env and nix profile?'"
echo "  nixai 'How do I update my NixOS system?'"
echo ""

echo -e "${CYAN}Troubleshooting:${NC}"
echo "  nixai 'My NixOS rebuild failed, what should I check?'"
echo "  nixai 'How do I fix permission issues in Nix?'"
echo ""

echo -e "${CYAN}Option Explanations:${NC}"
echo "  nixai explain-option services.openssh"
echo "  nixai explain-option boot.loader.grub"
echo "  nixai explain-home-option programs.git"
echo ""

wait_for_user

# Demo Summary
demo_header "Demo Complete!"

echo -e "${GREEN}🎉 You've seen all major nixai features!${NC}"
echo ""
echo -e "${CYAN}Quick Reference:${NC}"
echo "  nixai --help                          # Show help"
echo "  nixai 'your question'                # Ask any question"
echo "  nixai explain-option <option>        # Explain NixOS option"
echo "  nixai explain-home-option <option>   # Explain Home Manager option"
echo "  nixai --interactive                  # Interactive mode"
echo "  nixai mcp-server start               # Start MCP server"
echo ""
echo -e "${CYAN}Development:${NC}"
echo "  cd /home/nixuser/nixai && nix develop .#docker # Enter dev shell"
echo "  just build-docker                            # Build for Docker"
echo "  just test                                     # Run tests"
echo "  just help                                     # Show all commands"
echo ""
echo -e "${CYAN}Testing:${NC}"
echo "  cd /home/nixuser/nixai && ./docker_nixos/test_docker_nixai.sh  # Run comprehensive tests"
echo ""

echo -e "${GREEN}The nixai Docker environment is ready for development and testing!${NC}"
