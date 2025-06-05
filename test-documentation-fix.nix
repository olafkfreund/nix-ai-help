{
  description = "Test configuration to verify nixai module documentation fix";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    nixai.url = "path:."; # Use local flake
  };

  outputs = {
    self,
    nixpkgs,
    nixai,
  }: {
    nixosConfigurations.test = nixpkgs.lib.nixosSystem {
      system = "x86_64-linux";
      modules = [
        nixai.nixosModules.default
        {
          services.nixai = {
            enable = true;
            mcp = {
              enable = true;
              aiProvider = "ollama";
              aiModel = "llama3";
            };
          };

          # Minimal bootable config
          system.stateVersion = "25.11";
          boot.loader.systemd-boot.enable = true;
          boot.loader.efi.canTouchEfiVariables = true;
          fileSystems."/" = {
            device = "/dev/disk/by-uuid/dummy";
            fsType = "ext4";
          };

          # Disable problematic services for test
          documentation.enable = false;
          documentation.man.enable = false;
          documentation.nixos.enable = false;
        }
      ];
    };
  };
}
