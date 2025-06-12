#!/usr/bin/env bash

# Enhanced Nix Command Validation Demo
# This script demonstrates the local Nix commands available for validation

echo "🔧 Enhanced Nix Command Validation Demo"
echo "========================================"
echo

# Function to run command and show output
run_demo() {
    local title="$1"
    local cmd="$2"
    local description="$3"
    
    echo "📋 $title"
    echo "   Command: $cmd"
    echo "   Purpose: $description"
    echo "   Output:"
    echo "   -------"
    
    if eval "$cmd" 2>/dev/null; then
        echo "   ✅ Success"
    else
        echo "   ❌ Failed (this is expected if package/option doesn't exist)"
    fi
    echo
}

echo "🔍 PACKAGE VERIFICATION COMMANDS"
echo "---------------------------------"

run_demo "Package Search (JSON)" \
    "nix search nixpkgs firefox --json | head -5" \
    "Search for packages with detailed JSON output"

run_demo "Package Attribute Query" \
    "nix-env -qaP firefox 2>/dev/null | head -3" \
    "Query package with attribute path information"

run_demo "Package Metadata" \
    "nix eval nixpkgs#firefox.meta.description --raw 2>/dev/null || echo 'Package not found'" \
    "Get package description via nix eval"

run_demo "Package Version" \
    "nix eval nixpkgs#firefox.version --raw 2>/dev/null || echo 'Version not found'" \
    "Get package version information"

echo "⚙️  OPTION VERIFICATION COMMANDS"
echo "--------------------------------"

run_demo "NixOS Option Query" \
    "nixos-option services.nginx.enable 2>/dev/null | head -5 || echo 'Option query failed'" \
    "Query NixOS option details"

run_demo "Option Type via nix-instantiate" \
    "nix-instantiate --eval -E '(import <nixpkgs/nixos> {}).options.services.nginx.enable.type' 2>/dev/null || echo 'Type query failed'" \
    "Get option type using nix-instantiate"

run_demo "Option Default Value" \
    "nix-instantiate --eval -E '(import <nixpkgs/nixos> {}).options.services.nginx.enable.default or null' 2>/dev/null || echo 'Default query failed'" \
    "Get option default value"

echo "🧪 SYNTAX VALIDATION COMMANDS"
echo "-----------------------------"

# Create a test Nix expression
test_expr='{ pkgs, ... }: { environment.systemPackages = with pkgs; [ firefox ]; }'
echo "$test_expr" > /tmp/test-config.nix

run_demo "Syntax Check" \
    "nix-instantiate --parse /tmp/test-config.nix >/dev/null && echo 'Valid syntax' || echo 'Invalid syntax'" \
    "Check Nix expression syntax"

run_demo "Configuration Validation" \
    "nix-instantiate --eval -E '(import <nixpkgs/nixos> { configuration = /tmp/test-config.nix; }).config.system.build.toplevel' >/dev/null 2>&1 && echo 'Valid NixOS config' || echo 'Invalid config'" \
    "Validate as NixOS configuration"

echo "🎯 FLAKE VALIDATION COMMANDS"
echo "----------------------------"

# Create a test flake
mkdir -p /tmp/test-flake
cat > /tmp/test-flake/flake.nix << 'EOF'
{
  description = "Test flake";
  inputs.nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
  outputs = { nixpkgs, ... }: {
    devShells.x86_64-linux.default = nixpkgs.legacyPackages.x86_64-linux.mkShell {
      buildInputs = with nixpkgs.legacyPackages.x86_64-linux; [ hello ];
    };
  };
}
EOF

run_demo "Flake Syntax Check" \
    "cd /tmp/test-flake && nix flake check --no-build >/dev/null 2>&1 && echo 'Valid flake' || echo 'Invalid flake'" \
    "Validate flake syntax and structure"

run_demo "Flake Show Outputs" \
    "cd /tmp/test-flake && nix flake show 2>/dev/null | head -5 || echo 'Flake show failed'" \
    "Show flake outputs structure"

run_demo "Flake Metadata" \
    "cd /tmp/test-flake && nix flake metadata --json 2>/dev/null | head -3 || echo 'Metadata failed'" \
    "Get flake metadata"

echo "🖥️  INTERACTIVE REPL COMMANDS"
echo "----------------------------"

run_demo "REPL Package Check" \
    "echo 'with import <nixpkgs> {}; firefox.meta.available or false' | timeout 5 nix repl 2>/dev/null | tail -1 || echo 'REPL check failed'" \
    "Use nix repl to check package availability"

run_demo "REPL Option Check" \
    "echo 'with import <nixpkgs/nixos> {}; options.services.nginx.enable.type' | timeout 5 nix repl 2>/dev/null | tail -1 || echo 'REPL option check failed'" \
    "Use nix repl to check option types"

echo "🏠 HOME MANAGER VALIDATION"
echo "-------------------------"

run_demo "Home Manager Check" \
    "which home-manager >/dev/null && echo 'Home Manager available' || echo 'Home Manager not installed'" \
    "Check if Home Manager is available"

run_demo "HM Dry Run" \
    "which home-manager >/dev/null && home-manager build --dry-run 2>/dev/null && echo 'HM config valid' || echo 'HM not available or config invalid'" \
    "Test Home Manager configuration"

echo "🔧 COMMAND AVAILABILITY CHECKS"
echo "------------------------------"

commands=("nix" "nixos-rebuild" "nix-env" "nixos-option" "systemctl" "which" "home-manager")

for cmd in "${commands[@]}"; do
    if which "$cmd" >/dev/null 2>&1; then
        echo "   ✅ $cmd - Available"
    else
        echo "   ❌ $cmd - Not available"
    fi
done

echo
echo "🚀 VALIDATION SCORING POTENTIAL"
echo "==============================="
echo
echo "These commands can be used to create sophisticated scoring algorithms:"
echo
echo "📊 Quality Scoring Factors:"
echo "   • Package existence verification (nix search, nix-env -qa)"
echo "   • Option validity checking (nixos-option, nix-instantiate)"  
echo "   • Syntax correctness (nix-instantiate --parse)"
echo "   • Configuration validity (nixos-rebuild dry-build)"
echo "   • Flake structure validation (nix flake check)"
echo "   • Command availability (which, type)"
echo "   • Interactive verification (nix repl)"
echo
echo "🎯 Automated Scoring Weights:"
echo "   • Syntax Valid: +30 points"
echo "   • All packages exist: +25 points"
echo "   • All options valid: +25 points"
echo "   • Commands available: +10 points"
echo "   • Flake structure valid: +10 points"
echo "   Total: 100 points maximum"
echo
echo "✨ This creates a comprehensive, automated quality assessment system!"

# Cleanup
rm -f /tmp/test-config.nix
rm -rf /tmp/test-flake
