{
  description = "NixAI: A console-based application for diagnosing and configuring NixOS using AI models.";
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };
  outputs = {
    self,
    nixpkgs,
  }: let
    system = "x86_64-linux";
    pkgs = import nixpkgs {inherit system;};
  in {
    packages.${system}.nixai = pkgs.buildGoModule {
      pname = "nixai";
      version = "0.1.0";
      src = ./.;
      vendorHash = "sha256-abbfa/rHLiGcA88anY9cLlFH8fGw/hcSmUOw+uN9kQ0=";
      doCheck = false; # Disable tests in Nix build due to network/sandbox restrictions
      meta = {
        description = "A tool for diagnosing and configuring NixOS using AI.";
        license = pkgs.lib.licenses.mit;
        maintainers = [];
      };
    };
    apps.${system}.nixai = {
      type = "app";
      program = "${self.packages.${system}.nixai}/bin/nixai";
    };
    defaultPackage.${system} = self.packages.${system}.nixai;
    defaultApp.${system} = self.apps.${system}.nixai;

    # NixOS module
    nixosModules.default = import ./modules/nixos.nix;

    # Home Manager module
    homeManagerModules.default = import ./modules/home-manager.nix;

    # Legacy names for backward compatibility
    nixosModule = self.nixosModules.default;
    homeManagerModule = self.homeManagerModules.default;

    devShells.${system}.default = pkgs.mkShell {
      buildInputs = with pkgs; [
        go
        just
        golangci-lint
        git
        curl
        nix
      ];
      shellHook = ''
        export GOPATH=$(pwd)/go
        export PATH=$GOPATH/bin:$PATH
        echo "🚀 Nix development environment ready!"
        echo "Available tools: go $(go version | cut -d' ' -f3), just $(just --version)"
      '';
    };
  };
}
