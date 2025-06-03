#!/usr/bin/env bash
# Test script to ensure interactive commands work properly

echo "Running interactive mode tests..."

# Basic test function
test_command() {
  local cmd=$1
  local input="$cmd"
  local expected=$2
  
  echo -e "\nTesting: $cmd"
  result=$(echo -e "$input\nexit" | ./nixai interactive 2>/dev/null)
  
  if echo "$result" | grep -q "$expected"; then
    echo "✅ Command works: $cmd"
  else
    echo "❌ Command failed: $cmd"
    echo "Expected to contain: \"$expected\""
  fi
}

# Test key commands that were previously stub commands
test_command "community" "NixOS Community Resources"
test_command "community forums" "NixOS Community Forums"
test_command "configure" "NixOS Configuration Assistant"
test_command "configure wizard" "NixOS Configuration Wizard"
test_command "diagnose" "NixOS System Diagnostics"
test_command "diagnose system" "System Health Check"
test_command "doctor" "NixOS Health Checks"
test_command "doctor quick" "Quick Health Check"
test_command "flake" "Nix Flake Utilities"
test_command "learn" "NixOS Learning Resources"
test_command "logs" "NixOS Log Analysis"
test_command "mcp-server" "MCP Server Management"
test_command "neovim-setup" "Neovim Integration Setup"
test_command "package-repo" "Git Repository Analysis"

echo -e "\nInteractive mode tests completed."
