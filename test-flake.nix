{
  description = "Test flake for nixai module";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs = {
    self,
    nixpkgs,
  }: {
    nixosConfigurations.test = nixpkgs.lib.nixosSystem {
      system = "x86_64-linux";
      modules = [
        ./modules/nixos.nix
        {
          services.nixai = {
            enable = true;
            mcp = {
              enable = true;
              package = nixpkgs.legacyPackages.x86_64-linux.hello; # dummy package for test
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
      ];
    };
  };
}
