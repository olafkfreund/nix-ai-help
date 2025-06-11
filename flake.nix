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
      packages.default = self.packages.${system}.nixai;
      packages.nixai = pkgs.callPackage ./package.nix {
        version = "1.0.0";
        src = ./.; # Use src instead of srcOverride for direct source
        rev = self.rev or null;
        gitCommit =
          if (self ? rev && self.rev != null)
          then builtins.substring 0 7 self.rev
          else "unknown";
        buildDate = "1970-01-01T00:00:00Z";
      };
      apps.default = self.apps.${system}.nixai;
      apps.nixai = {
        type = "app";
        program = "${self.packages.${system}.nixai}/bin/nixai";
        meta = {
          description = "Run nixai from the command line";
        };
      };
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
      # Temporarily disabled due to linter issues with yaml.v3 imports
      # checks.lint =
      #   pkgs.runCommand "golangci-lint" {
      #     buildInputs = [pkgs.golangci-lint pkgs.go];
      #   } ''
      #     export HOME=$TMPDIR
      #     export XDG_CACHE_HOME=$TMPDIR/.cache
      #     mkdir -p $XDG_CACHE_HOME
      #     cd ${./.}
      #     ${pkgs.golangci-lint}/bin/golangci-lint run ./... --timeout=10m
      #     touch $out
      #   '';
    })
    // {
      # System-independent modules
      nixosModules.default = import ./modules/nixos.nix;
      homeManagerModules.default = import ./modules/home-manager.nix;

      # Flake-level overlays - provide nixai package for each system
      overlays.default = final: prev: {
        nixai = self.packages.${prev.system}.nixai or (throw "nixai package not available for system ${prev.system}");
      };
    };
}
