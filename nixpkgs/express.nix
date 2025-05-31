{
  lib,
  fetchFromGitHub,
  buildNpmPackage,
}:
buildNpmPackage rec {
  pname = "express";
  version = "4.18.1";

  src = fetchFromGitHub {
    owner = "expressjs";
    repo = "express";
    rev = "v${version}";
    sha256 = "0000000000000000000000000000000000000000000000000000"; # TODO: update with actual hash
  };

  npmFlags = ["--production"];

  meta = with lib; {
    description = "Fast, unopinionated, JavaScript web framework";
    homepage = "https://expressjs.com/";
    license = licenses.mit;
    maintainers = [];
    platforms = platforms.unix;
  };

  doCheck = true;
}
# Note: Replace the sha256 with the correct value after running 'nix-build' once to get the fixed-output hash error.
# The builder.sh script is not needed when using buildNpmPackage.

