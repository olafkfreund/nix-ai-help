{
  lib,
  buildGoModule,
  installShellFiles,
  # Optional parameters for version/commit overrides
  version ? "0.1.0",
  src ? ./.,
  rev ? null,
  gitCommit ? null,
  buildDate ? "1970-01-01T00:00:00Z",
}:
buildGoModule rec {
  pname = "nixai";
  inherit version src;

  vendorHash = null;
  doCheck = false;

  subPackages = ["cmd/nixai"];

  nativeBuildInputs = [installShellFiles];

  ldflags = let
    versionString =
      if (rev != null)
      then rev
      else version;
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
