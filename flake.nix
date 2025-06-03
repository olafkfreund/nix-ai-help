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
  }:
    flake-utils.lib.eachDefaultSystem (system: let
      pkgs = import nixpkgs {inherit system;};
    in {
      packages.nixai = pkgs.buildGoModule {
        pname = "nixai";
        version = "0.1.0";
        src = ./.;
        vendorHash = "sha256-KZ2U8ErZIvuSwPL7hD3roDI+v1xIX/zssf0lHueTZV4=";
        modVendor = true;
        proxyVendor = true;
        doCheck = false;
        subPackages = ["cmd/nixai"];
        ldflags = let
          version =
            if (self ? rev)
            then self.rev
            else "dirty";
          gitCommit =
            if (self ? rev)
            then builtins.substring 0 7 self.rev
            else "unknown";
          buildDate = "1970-01-01T00:00:00Z";
        in [
          "-X nix-ai-help/pkg/version.Version=${version}"
          "-X nix-ai-help/pkg/version.GitCommit=${gitCommit}"
          "-X nix-ai-help/pkg/version.BuildDate=${buildDate}"
        ];
        meta = {
          description = "A tool for diagnosing and configuring NixOS using AI.";
          license = pkgs.lib.licenses.mit;
          maintainers = ["olafkfreund"];
        };
      };
      defaultPackage = self.packages.${system}.nixai;
      apps.nixai = {
        type = "app";
        program = "${self.packages.${system}.nixai}/bin/nixai";
        meta = {
          description = "Run nixai from the command line";
        };
      };
      defaultApp = self.apps.${system}.nixai;
      nixosModules.default = import ./modules/nixos.nix;
      homeManagerModules.default = import ./modules/home-manager.nix {
        nixaiPackage = self.packages.${system}.nixai;
      };
      nixosModule = self.nixosModules.default;
      homeManagerModule = self.homeManagerModules.default;
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
      # Add a formatter output for Nix code
      formatter = pkgs.alejandra;
      # Add a basic check (lint) for Go code
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
      # Flake-level metadata
      flakeMetadata = {
        maintainers = ["olafkfreund"];
        homepage = "https://github.com/olafkfreund/nix-ai-help";
        license = "MIT";
        platforms = ["x86_64-linux" "aarch64-linux"];
      };
    };
}
