# Standalone package.nix build
# This version can be built directly with nix-build without callPackage
with import <nixpkgs> {};
  callPackage ./package.nix {
    version = "0.1.0";
    src = ./.;
    gitCommit = "unknown";
  }
