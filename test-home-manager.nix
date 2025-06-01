# Test Home Manager configuration
{
  config,
  pkgs,
  ...
}: {
  imports = [
    ./flake.nix.homeManagerModules.default
  ];

  # Enable nixai service for testing
  services.nixai = {
    enable = true;
    mcp.enable = true;
  };

  # Required for Home Manager
  home.username = "test";
  home.homeDirectory = "/home/test";
  home.stateVersion = "23.11";
}
