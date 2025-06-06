{
  lib,
  buildGoModule,
  rev ? null,
  ...
}:
buildGoModule {
  pname = "nixai";
  version = "0.1.0";
  src = builtins.path {
    name = "nix-ai-help";
    path = ./.;
  };
  vendorHash = null;
  modVendor = true;
  proxyVendor = true;
  doCheck = false;
  subPackages = ["cmd/nixai"];
  ldflags = let
    version =
      if (rev != null)
      then rev
      else "dirty";
    gitCommit =
      if (rev != null)
      then builtins.substring 0 7 rev
      else "unknown";
    buildDate = "1970-01-01T00:00:00Z";
  in [
    "-X nix-ai-help/pkg/version.Version=${version}"
    "-X nix-ai-help/pkg/version.GitCommit=${gitCommit}"
    "-X nix-ai-help/pkg/version.BuildDate=${buildDate}"
  ];
  meta = {
    description = "A tool for diagnosing and configuring NixOS using AI.";
    license = lib.licenses.mit;
    maintainers = ["olafkfreund"];
  };
}
