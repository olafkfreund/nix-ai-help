{
  description = "NixAI: A console-based application for diagnosing and configuring NixOS using AI models.",
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  },
  outputs = {
    packages.x86_64-linux.nixai = let
      src = ./.;
    in
      pkgs.mkShell {
        buildInputs = [
          pkgs.go
          pkgs.git
          pkgs.curl
          pkgs.nix
        ];
        shellHook = ''
          export GOPATH=$(pwd)/go
          export PATH=$GOPATH/bin:$PATH
        '';
      };
  },
  meta = {
    description = "A tool for diagnosing and configuring NixOS using AI.",
    license = "MIT",
    maintainers = with pkgs.lib.maintainers; [ yourGitHubUsername ],
  }
}