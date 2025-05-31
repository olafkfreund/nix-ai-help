#!/usr/bin/env zsh
# nixai_all_commands_demo.sh
# Runs all major nixai commands as shown in the manual, with MCP server startup and pauses for review.

set -euo pipefail

print_section() {
  echo "\n\033[1;36m========== $1 =========\033[0m\n"
}

pause() {
  echo "\033[1;33m[Press enter to continue]\033[0m"
  read _
}

# 1. Start MCP server
print_section "Start MCP Server"
./nixai mcp-server start -d &
sleep 3
./nixai mcp-server status
pause

# 2. Direct Question Assistant
print_section "Direct Question Assistant"
./nixai "how do I enable SSH in NixOS?"
pause
./nixai --ask "how do I update packages in NixOS?"
pause

# 3. Diagnosing NixOS Issues
print_section "Diagnosing NixOS Issues"
# ./nixai diagnose --log-file /var/log/nixos/nixos-rebuild.log || true
# pause
./nixai diagnose --nix-log /nix/store/vqbq2gvsxbw0g959znxvf1li9szz36c-bash52-031.drv || true
pause
echo 'services.nginx.enable = true;' | ./nixai diagnose || true
pause

# 4. Explaining NixOS and Home Manager Options
print_section "Explaining NixOS and Home Manager Options"
./nixai explain-option services.nginx.enable || true
pause
./nixai explain-home-option programs.git.enable || true
pause
./nixai explain-option "how to enable SSH access" || true
pause

# 5. Searching for Packages and Services
print_section "Searching for Packages and Services"
./nixai search pkg nginx || true
pause
./nixai search service postgresql || true
pause

# 6. AI-Powered Package Repository Analysis
print_section "AI-Powered Package Repository Analysis"
./nixai package-repo . --local || true
pause
./nixai package-repo https://github.com/psf/requests || true
pause
./nixai package-repo https://github.com/expressjs/express --output ./nixpkgs || true
pause
./nixai package-repo . --analyze-only || true
pause
# ./nixai package-repo https://github.com/user/rust-app --output ./derivations --name my-rust-app || true
# pause
# ./nixai package-repo https://github.com/user/monorepo || true
# pause
# ./nixai package-repo git@github.com:yourorg/private-repo.git --ssh-key ~/.ssh/id_ed25519 || true
# pause
./nixai package-repo . --output-format json || true
pause

# 7. System Health Checks
print_section "System Health Checks"
./nixai health || true
pause
./nixai health --nixos-path ~/.config/nixos || true
pause
./nixai health --log-level debug || true
pause

# 8. Development Environment (devenv) Feature
print_section "Development Environment (devenv) Feature"
./nixai devenv list || true
pause
./nixai devenv create python demo-py --framework fastapi --with-poetry --services postgres,redis || true
pause
./nixai devenv create golang demo-go --framework gin --with-grpc || true
pause
./nixai devenv create nodejs demo-node --with-typescript --services mongodb || true
pause
./nixai devenv create rust demo-rust --with-wasm || true
pause
./nixai devenv suggest "web app with database and REST API" || true
pause

# 9. Interactive Mode
print_section "Interactive Mode"
echo "exit" | ./nixai interactive || true
pause

# 10. Neovim Integration (with Home Manager reference)
print_section "Neovim Integration"
echo "See docs/neovim-integration.md for Home Manager config and troubleshooting."
./nixai neovim-setup || true
pause
./nixai neovim-setup --socket-path=/tmp/nixai-mcp.sock || true
pause
./nixai neovim-setup --config-dir=~/.config/nvim || true
pause

# 11. Advanced Usage
print_section "Advanced Usage"
./nixai search --nixos-path /etc/nixos pkg nginx || true
pause
# ./nixai diagnose --provider openai --log-file  || true
# pause
./nixai service-examples nginx || true
pause
./nixai flake explain --flake /etc/nixos/flake.nix || true
pause
journalctl -xe | ./nixai diagnose || true
pause
cat /etc/nixos/configuration.nix | ./nixai explain-option || true
pause
# ./nixai diagnose --provider ollama --model llama3 --temperature 0.2 --log-file  || true
# pause
# ./nixai package-repo git@github.com:yourorg/private-repo.git --ssh-key ~/.ssh/id_ed25519 || true
# pause
./nixai package-repo . --output-format json || true
pause
./nixai health --output-format json > health_report.json || true
pause
./nixai package-repo . --analyze-only --output-format json > analysis.json || true
pause

print_section "All nixai commands demo complete!"
