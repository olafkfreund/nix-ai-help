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
    packages.${system} = {
      nixai = pkgs.buildGoModule {
        pname = "nixai";
        version = "0.1.0";
        src = ./.;
        vendorHash = "sha256-vv+BcoRMniUB+Cbl6zM1slYiaDJp+CX5LQcbFWNC1+Y="; # Using vendored dependencies
        doCheck = false; # Disable tests in Nix build due to network/sandbox restrictions

        # Specify the main package to build
        subPackages = ["cmd/nixai"];

        # Inject version information at build time
        ldflags = let
          version =
            if (self ? rev)
            then self.rev
            else "dirty";
          gitCommit =
            if (self ? rev)
            then builtins.substring 0 7 self.rev
            else "unknown";
          buildDate = "1970-01-01T00:00:00Z"; # Nix builds are reproducible
        in [
          "-X nix-ai-help/pkg/version.Version=${version}"
          "-X nix-ai-help/pkg/version.GitCommit=${gitCommit}"
          "-X nix-ai-help/pkg/version.BuildDate=${buildDate}"
        ];

        meta = {
          description = "A tool for diagnosing and configuring NixOS using AI.";
          license = pkgs.lib.licenses.mit;
          maintainers = [];
        };
      };
      default = self.packages.${system}.nixai;
    };
    apps.${system} = {
      nixai = {
        type = "app";
        program = "${self.packages.${system}.nixai}/bin/nixai";
        meta = {
          description = "Run nixai from the command line";
        };
      };
      default = self.apps.${system}.nixai;
    };

    # NixOS module
    nixosModules.default = import ./modules/nixos.nix;

    # Home Manager module - pass the nixai package
    homeManagerModules.default = import ./modules/home-manager.nix {
      nixaiPackage = self.packages.${system}.nixai;
    };

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
          echo "🐳 [nixai] Docker isolated environment ready!"
          echo "📁 Working with cloned repository (no host mounting)"
          echo "🔧 Available tools: go $(go version | cut -d' ' -f3), just $(just --version)"

          # Set up Ollama host for Docker environment
          if [ -z "$OLLAMA_HOST" ]; then
            export OLLAMA_HOST="http://host.docker.internal:11434"
            echo "🤖 Ollama host set to: $OLLAMA_HOST"
          fi

          # Change to cloned nixai directory
          if [ -d "/home/nixuser/nixai" ]; then
            cd /home/nixuser/nixai
            echo "📂 Changed to cloned nixai directory: $(pwd)"
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
