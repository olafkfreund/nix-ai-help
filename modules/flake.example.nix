# Example of how to add nixai modules to your flake.nix
{
  description = "Flake for integrating nixai modules";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    nixai.url = "github:your-username/nix-ai-help"; # Replace with actual repo URL
    home-manager = {
      url = "github:nix-community/home-manager";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = {
    self,
    nixpkgs,
    nixai,
    home-manager,
    ...
  }: let
    system = "x86_64-linux"; # Adjust for your system
    pkgs = nixpkgs.legacyPackages.${system};
  in {
    # NixOS system configuration
    nixosConfigurations.example = nixpkgs.lib.nixosSystem {
      inherit system;
      modules = [
        # Import nixai NixOS module
        nixai.nixosModules.default

        # Your configuration
        {
          services.nixai = {
            enable = true;
            mcp = {
              enable = true;
              # Customize options as needed
              socketPath = "/run/nixai/mcp.sock";
            };
          };
        }
      ];
    };

    # Home Manager configuration
    homeConfigurations.example = home-manager.lib.homeManagerConfiguration {
      inherit pkgs;
      modules = [
        # Import nixai Home Manager module
        nixai.homeManagerModules.default

        # Your configuration
        {
          services.nixai = {
            enable = true;
            mcp = {
              enable = true;
              # Customize options as needed
              socketPath = "$HOME/.local/share/nixai/mcp.sock";
            };
            # Enable VS Code integration
            vscodeIntegration = true;
          };

          # Enable Home Manager's VS Code module
          programs.vscode.enable = true;
        }
      ];
    };
  };
}
