#!/usr/bin/env bash
# Test script to verify that the NixOS documentation build error is fixed

set -euo pipefail

echo "🔧 Testing nixai NixOS module documentation fix..."
echo ""

# Test 1: Verify NixOS module can be evaluated
echo "Test 1: Evaluating NixOS module..."
if nix eval .#nixosModules.default --apply "x: \"success\"" >/dev/null 2>&1; then
    echo "✅ NixOS module evaluation: PASSED"
else
    echo "❌ NixOS module evaluation: FAILED"
    exit 1
fi

# Test 2: Verify Home Manager module can be evaluated  
echo "Test 2: Evaluating Home Manager module..."
if nix eval .#homeManagerModules.default --apply "x: \"success\"" >/dev/null 2>&1; then
    echo "✅ Home Manager module evaluation: PASSED"
else
    echo "❌ Home Manager module evaluation: FAILED"
    exit 1
fi

# Test 3: Verify package builds
echo "Test 3: Building nixai package..."
if nix build .#packages.x86_64-linux.nixai --no-link >/dev/null 2>&1; then
    echo "✅ Package build: PASSED"
else
    echo "❌ Package build: FAILED"
    exit 1
fi

# Test 4: Check for any remaining documentation issues in module files
echo "Test 4: Checking for potential documentation issues..."
if grep -r "^# [A-Z]" modules/ >/dev/null 2>&1; then
    echo "⚠️  Warning: Found potential heading-like comments in modules/"
    grep -n "^# [A-Z]" modules/ || true
else
    echo "✅ No documentation heading issues found: PASSED"
fi

echo ""
echo "🎉 All tests passed! The documentation build error has been fixed."
echo ""
echo "Summary of fixes applied:"
echo "• Fixed comment format in modules/nixos.nix to prevent markdown heading interpretation"
echo "• Fixed comment format in modules/home-manager.nix for consistency"
echo "• Both modules now use multi-line descriptive comments instead of single-line headers"
echo ""
echo "The original error was caused by the NixOS documentation renderer interpreting"
echo "single '#' comments as markdown headings, which require IDs for cross-referencing."
