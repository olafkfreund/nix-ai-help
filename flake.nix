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
      vendorSha256 = null;
      subPackages = ["cmd/nixai"];
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
    devShells.${system}.default = pkgs.mkShell {
      buildInputs = [pkgs.go pkgs.git pkgs.curl pkgs.nix];
      shellHook = ''
        export GOPATH=$(pwd)/go
        export PATH=$GOPATH/bin:$PATH
      '';
    };
  };
}
