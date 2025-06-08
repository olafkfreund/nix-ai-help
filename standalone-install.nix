# Standalone installation for nixai from local source
# Usage: nix-build standalone-install.nix && result/bin/nixai --help
let
  pkgs = import <nixpkgs> {};
in
  pkgs.callPackage ./package.nix {
    # Use current directory as source
    srcOverride = ./.;
    version = "latest";
    buildDate = "2025-06-08T00:00:00Z";
  }
