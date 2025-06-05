# Test configuration to verify nixai module works without documentation errors
{pkgs, ...}: {
  imports = [./modules/nixos.nix];

  services.nixai = {
    enable = true;
    mcp = {
      enable = true;
      aiProvider = "ollama";
      aiModel = "llama3";
    };
  };

  # Minimal config to make this buildable
  system.stateVersion = "25.11";
  boot.loader.grub.enable = false;
  fileSystems."/" = {
    device = "tmpfs";
    fsType = "tmpfs";
  };
}
