# Nixvim + Home Manager + nixai Neovim Integration Example Module
# Save as modules/nixvim-nixai-example.nix
{
  config,
  pkgs,
  ...
}: let
  nixai-flake = builtins.getFlake "github:olafkfreund/nix-ai-help";
in {
  imports = [nixai-flake.homeManagerModules.default];

  # Enable nixai and Neovim integration
  services.nixai = {
    enable = true;
    mcp.enable = true;
    neovimIntegration = {
      enable = true;
      useNixVim = true;
      keybindings = {
        askNixai = "<leader>na";
        askNixaiVisual = "<leader>na";
        startMcpServer = "<leader>ns";
      };
      autoStartMcp = true;
    };
  };

  # Nixvim configuration (minimal example)
  programs.nixvim = {
    enable = true;
    extraConfigVim = ''
      set number
      set relativenumber
      set expandtab
      set shiftwidth=2
      set tabstop=2
    '';
    # Optionally add plugins, LSP, etc.
  };
}
