# Test to verify the "attribute 'nixai' missing" error is fixed
# This test simulates using the module outside of flake context where pkgs.nixai doesn't exist
let
  # Simulate a minimal nixpkgs without nixai
  pkgs = import <nixpkgs> {};

  # Import the nixos module without providing nixaiPackage (simulates old usage)
  nixosModule = import ./modules/nixos.nix {};

  # Create a minimal config to test the module
  config = {
    services.nixai.enable = false; # We just need to evaluate the options
  };

  lib = pkgs.lib;

  # Evaluate the module to check if it works
  moduleResult = lib.evalModules {
    modules = [
      nixosModule
      {inherit config;}
    ];
  };
in {
  # This should not throw "attribute 'nixai' missing" error
  success = moduleResult.config.services.nixai.enable == false;

  # Check that the default package is available (should be the placeholder)
  hasDefaultPackage = moduleResult.config.services.nixai.mcp.package != null;

  # Check that the placeholder package has the expected structure
  packageName = moduleResult.config.services.nixai.mcp.package.pname or "unknown";
}
