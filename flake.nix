{
  description = "NixAI: A console-based application for diagnosing and configuring NixOS using AI models.";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = {
    self,
    nixpkgs,
    flake-utils,
    ...
  }: let
    # System-dependent NixOS modules (using eachDefaultSystemPassThrough)
    nixosModules = flake-utils.lib.eachDefaultSystemPassThrough (system: {
      default = import ./modules/nixos.nix;
    });
    nixosModule = nixosModules.default;

    # System-dependent Home Manager modules (using eachDefaultSystemPassThrough)
    homeManagerModules = flake-utils.lib.eachDefaultSystemPassThrough (system: {
      default = import ./modules/home-manager.nix;
    });
    homeManagerModule = homeManagerModules.default;
  in
    flake-utils.lib.eachDefaultSystem (system: let
      pkgs = import nixpkgs {inherit system;};
      nixai = pkgs.callPackage ./package.nix { inherit (pkgs) lib buildGoModule; inherit (self) rev; };
    in {
      packages = { inherit nixai; };
      defaultPackage = self.packages.${system}.nixai;
      overlays.default = final: prev: { inherit nixai; };
      apps.nixai = {
        type = "app";
        program = "${self.packages.${system}.nixai}/bin/nixai";
        meta = {
          description = "Run nixai from the command line";
        };
      };
      defaultApp = self.apps.${system}.nixai;
      devShells.default = pkgs.mkShell {
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
      devShells.docker = pkgs.mkShell {
        name = "nixai-docker-devshell";
        buildInputs = with pkgs; [
          go
          just
          neovim
          git
          curl
          python3
          nodejs
          alejandra
          nixos-install-tools
          jq
          htop
          tree
        ];
        shellHook = ''
          echo "🐳 [nixai] Docker isolated environment ready!"
          echo "📁 Working with cloned repository (no host mounting)"
          echo "🔧 Available tools: go $(go version | cut -d' ' -f3), just $(just --version)"
          if [ -z "$OLLAMA_HOST" ]; then
            export OLLAMA_HOST="http://host.docker.internal:11434"
            echo "🤖 Ollama host set to: $OLLAMA_HOST"
          fi
          if [ -d "/home/nixuser/nixai" ]; then
            cd /home/nixuser/nixai
            echo "📂 Changed to cloned nixai directory: $(pwd)"
          fi
          echo ""
          echo "🚀 Available Docker commands:"
          echo "  just build-docker     - Build nixai in container"
          echo "  just run-docker       - Run built nixai"
          echo "  just install-docker   - Install nixai globally"
          echo "  just help            - Show all available commands"
          echo ""
        '';
      };
      formatter = pkgs.alejandra;
      checks.lint =
        pkgs.runCommand "golangci-lint" {
          buildInputs = [pkgs.golangci-lint pkgs.go];
        } ''
          cd $PWD
          golangci-lint run ./...
          touch $out
        '';
    })
    // {
      nixosModules = nixosModules;
      nixosModule = nixosModule;
      homeManagerModules = homeManagerModules;
      homeManagerModule = homeManagerModule;
      # Flake-level metadata
      flakeMetadata = {
        maintainers = ["olafkfreund"];
        homepage = "https://github.com/olafkfreund/nix-ai-help";
        license = "MIT";
        platforms = ["x86_64-linux" "aarch64-linux"];
      };
    };
}
