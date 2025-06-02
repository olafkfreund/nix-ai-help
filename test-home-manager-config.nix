# Test Home Manager configuration using nixai flake
{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    home-manager = {
      url = "github:nix-community/home-manager";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    nixai.url = "path:./.";
  };

  outputs = {
    self,
    nixpkgs,
    home-manager,
    nixai,
  }: let
    system = "x86_64-linux";
    pkgs = nixpkgs.legacyPackages.${system};
  in {
    homeConfigurations."test" = home-manager.lib.homeManagerConfiguration {
      inherit pkgs;
      modules = [
        nixai.homeManagerModules.default
        {
          home.username = "test";
          home.homeDirectory = "/home/test";
          home.stateVersion = "25.05";

          # Enable nixai service
          services.nixai = {
            enable = true;
            mcp = {
              enable = true;
              aiProvider = "ollama";
              aiModel = "llama3";
            };
          };
        }
      ];
    };
  };
}
