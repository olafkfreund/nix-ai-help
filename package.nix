{
  lib,
  buildGoModule,
  installShellFiles,
  fetchFromGitHub,
  # Optional parameters for version/commit overrides
  version ? "1.0.2",
  src ? null,
  srcOverride ? null,
  rev ? null,
  gitCommit ? null,
  buildDate ? "1970-01-01T00:00:00Z",
}: let
  # Determine the source to use
  sourceToUse =
    if srcOverride != null
    then srcOverride
    else if src != null
    then src
    else
      fetchFromGitHub {
        owner = "olafkfreund";
        repo = "nix-ai-help";
        rev =
          if rev != null
          then rev
          else "main"; # Use main branch instead of tag for latest code
        sha256 = lib.fakeSha256; # This will need to be updated with actual hash
      };
in
  buildGoModule rec {
    pname = "nixai";
    inherit version;

    src = sourceToUse;

    vendorHash = "sha256-pGyNwzTkHuOzEDOjmkzx0sfb1jHsqb/1FcojsCGR6CY=";
    doCheck = false;

    # Force Go to use modules instead of vendor
    buildFlagsArray = ["-mod=readonly"];

    subPackages = ["cmd/nixai"];

    # Handle source directory structure for standalone installations
    preBuild = ''
      echo "=== Build Environment Debug ==="
      echo "Current directory: $(pwd)"
      echo "Source directory contents:"
      ls -la

      # Check if go.mod exists in current directory
      if [ -f go.mod ]; then
        echo "✓ Found go.mod in current directory"
        echo "Module: $(head -1 go.mod)"
      else
        echo "go.mod not found in current directory, searching..."
        GOMOD_PATH=$(find . -name "go.mod" -type f | head -1)
        if [ -n "$GOMOD_PATH" ]; then
          GOMOD_DIR=$(dirname "$GOMOD_PATH")
          echo "Found go.mod in: $GOMOD_DIR"
          echo "Changing to source directory: $GOMOD_DIR"
          cd "$GOMOD_DIR"
          echo "Now in: $(pwd)"
          echo "Module: $(head -1 go.mod)"
        else
          echo "Error: Cannot find go.mod file in source tree"
          echo "Full directory structure:"
          find . -type f -name "*.go" -o -name "go.*" | head -20
          exit 1
        fi
      fi

      # Force remove vendor directory if it exists anywhere
      echo "Removing any vendor directories..."
      find . -name "vendor" -type d -exec rm -rf {} + 2>/dev/null || true

      echo "=== Build Environment Ready ==="
    '';

    nativeBuildInputs = [installShellFiles];

    ldflags = let
      versionString = version; # Always use the version parameter
      commitString =
        if (gitCommit != null)
        then gitCommit
        else if (rev != null)
        then builtins.substring 0 7 rev
        else "unknown";
    in [
      "-X nix-ai-help/pkg/version.Version=${versionString}"
      "-X nix-ai-help/pkg/version.GitCommit=${commitString}"
      "-X nix-ai-help/pkg/version.BuildDate=${buildDate}"
    ];

    postInstall = ''
      # Generate shell completions if the binary supports it
      installShellCompletion --cmd nixai \
        --bash <($out/bin/nixai completion bash 2>/dev/null || echo "") \
        --fish <($out/bin/nixai completion fish 2>/dev/null || echo "") \
        --zsh <($out/bin/nixai completion zsh 2>/dev/null || echo "") || true
    '';

    meta = {
      description = "A modular, console-based Linux application for solving NixOS configuration problems and assisting with NixOS setup and troubleshooting";
      longDescription = ''
        nixai is a command-line tool that provides AI-powered assistance for NixOS configuration,
        troubleshooting, and package management. It supports multiple AI providers (Ollama, OpenAI,
        Gemini), can analyze logs and configurations, query NixOS documentation, and provides
        modular commands for community, learning, development environments, and more.
      '';
      homepage = "https://github.com/olafkfreund/nix-ai-help";
      license = lib.licenses.mit;
      maintainers = []; # Add your nixpkgs maintainer handle here when submitting to nixpkgs
      platforms = lib.platforms.linux ++ lib.platforms.darwin;
      mainProgram = "nixai";
    };
  }
