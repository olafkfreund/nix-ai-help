# Test configuration to verify Home Manager module works
{
  config,
  pkgs,
  lib,
  ...
}: let
  nixai-flake = builtins.getFlake (toString ./flake.nix);
in {
  imports = [
    nixai-flake.homeManagerModules.default
  ];

  # Enable nixai service for testing
  services.nixai = {
    enable = true;
    mcp.enable = true;
  };

  # Required for Home Manager
  home.username = "test";
  home.homeDirectory = "/home/test";
  home.stateVersion = "25.05";
}
