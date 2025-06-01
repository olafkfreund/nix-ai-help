# Minimal test to verify Home Manager module works
let
  flake = builtins.getFlake (toString ./.);
  system = "x86_64-linux";
  pkgs = import <nixpkgs> {inherit system;};

  homeModule = flake.homeManagerModules.default;

  # Create a minimal config that enables nixai
  config = {
    services.nixai = {
      enable = true;
      mcp.enable = false; # Start simple
    };
  };

  # Evaluate the module with the config
  eval = pkgs.lib.evalModules {
    modules = [homeModule config];
    specialArgs = {inherit pkgs;};
  };
in {
  # Check if the config evaluation works and what packages are defined
  nixaiEnabled = eval.config.services.nixai.enable;
  mcpPackage = eval.config.services.nixai.mcp.package;
  packageName = eval.config.services.nixai.mcp.package.pname or "unknown";
}
