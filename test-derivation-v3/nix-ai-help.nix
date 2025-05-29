{
  stdenv,
  lib,
  buildGoModule,
  fetchFromGitHub,
}: {
  lib,
  buildGoModule,
  fetchFromGitHub,
}:
buildGoModule rec {
  pname = "nix-ai-help";
  version = "1.0.0";

  src = fetchFromGitHub {
    owner = "olafkfreund";
    repo = "nix-ai-help";
    rev = "v${version}";
    sha256 = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="; # Replace with actual hash
  };

  vendorHash = "sha256-BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB="; # Replace with actual hash

  # For Go projects, dependencies are typically handled by go.mod
  # Only add buildInputs/nativeBuildInputs if you need system libraries
  nativeBuildInputs = [];
  buildInputs = [];

  ldflags = ["-s" "-w"];

  doCheck = true;

  meta = with lib; {
    description = "AI-powered Nix configuration assistant and package management tool";
    homepage = "https://github.com/olafkfreund/nix-ai-help";
    license = licenses.mit;
    maintainers = []; # Add maintainer info
    platforms = platforms.unix;
  };
}
