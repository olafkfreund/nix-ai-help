{ lib, buildNpmPackage, fetchFromGitHub }:

buildNpmPackage rec {
  pname = "{{.ProjectName}}";
  version = "{{.Version}}";

  src = fetchFromGitHub {
    owner = "{{.Owner}}";
    repo = "{{.ProjectName}}";
    rev = "v${version}";
    sha256 = lib.fakeHash;
  };

  npmDepsHash = lib.fakeHash;

  # The prepack script runs the build step
  npmPackFlags = [ "--ignore-scripts" ];
  
  NODE_OPTIONS = "--openssl-legacy-provider";

{{- if .BuildPhase}}
  buildPhase = ''
{{.BuildPhase}}
  '';
{{- end}}

{{- if .InstallPhase}}
  installPhase = ''
{{.InstallPhase}}
  '';
{{- else}}
  installPhase = ''
    runHook preInstall
    mkdir -p $out/bin
    cp -r dist/* $out/
    runHook postInstall
  '';
{{- end}}

{{- if .CheckPhase}}
  doCheck = true;
  checkPhase = ''
{{.CheckPhase}}
  '';
{{- end}}

  meta = with lib; {
    description = "{{.Description}}";
    homepage = "{{.Homepage}}";
{{- if .License}}
    license = licenses.{{.License | lower}};
{{- end}}
    maintainers = with maintainers; [ ];
    platforms = platforms.all;
  };
}
