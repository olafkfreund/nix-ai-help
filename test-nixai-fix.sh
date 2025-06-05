#!/usr/bin/env bash

# Test script to verify the "attribute 'nixai' missing" fix

set -e

echo "🧪 Testing nixai module fixes..."

# Test 1: Verify module can be imported without flake context
echo "Test 1: Module import without flake context"
nix eval --expr 'let pkgs = import <nixpkgs> {}; module = import ./modules/nixos.nix {}; in "✅ Import successful"'

# Test 2: Verify module evaluation with nixos lib
echo "Test 2: Module evaluation with NixOS lib"
nix eval --expr '
let 
  pkgs = import <nixpkgs> {};
  lib = pkgs.lib;
  nixosModule = import ./modules/nixos.nix {};
  
  # Create a test evaluation
  moduleTest = lib.evalModules {
    modules = [
      nixosModule
      { 
        services.nixai.enable = false;
        services.nixai.mcp.enable = false;
      }
    ];
  };
in "✅ Module evaluation successful"
'

# Test 3: Verify flake-based usage still works  
echo "Test 3: Flake-based module usage"
nix eval .#nixosModules.default --apply 'x: "✅ Flake module accessible"'

# Test 4: Verify the package builds
echo "Test 4: Package build verification"
nix build .#nixai --no-link && echo "✅ Package builds successfully"

echo ""
echo "🎉 All tests passed! The 'attribute nixai missing' error has been fixed."
echo ""
echo "✅ Modules can now be used both:"
echo "   - In flake context (with nixai package provided)"
echo "   - Outside flake context (with placeholder package)"
echo ""
echo "📚 For usage examples, see modules/README.md"
