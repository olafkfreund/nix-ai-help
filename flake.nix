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

    devShells.${system} = {
      # Default development shell for local development
      default = pkgs.mkShell {
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

      # Docker development shell for isolated container environment
      # This shell is used inside Docker containers with cloned nixai repository
      docker = pkgs.mkShell {
        name = "nixai-docker-devshell";
        buildInputs = with pkgs; [
          go
          just
          neovim
          git
          curl
          python3
          nodejs
          alejandra # Nix formatter
          nixos-install-tools
          jq # JSON processing for config handling
          htop # System monitoring
          tree # Directory listing
        ];
        shellHook = ''
          echo "🐳 [nixai] Docker development environment ready!"
          echo "📁 Working in isolated container with cloned repository"
          echo "🔧 Available tools: go $(go version | cut -d' ' -f3), just $(just --version)"
          
          # Set up Ollama host for Docker environment
          if [ -z "$OLLAMA_HOST" ]; then
            export OLLAMA_HOST="http://host.docker.internal:11434"
            echo "🤖 Ollama host set to: $OLLAMA_HOST"
          fi
          
          # Set up working directory if available
          if [ -d "/workspace" ]; then
            cd /workspace
            echo "📂 Changed to workspace directory: $(pwd)"
          fi
          
          # Display available justfile commands
          echo ""
          echo "🚀 Available Docker commands:"
          echo "  just build-docker     - Build nixai in container"
          echo "  just run-docker       - Run built nixai"
          echo "  just install-docker   - Install nixai globally"
          echo "  just help            - Show all available commands"
          echo ""
        '';
      };
    };
  };
}
