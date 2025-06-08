{ lib, rustPlatform, fetchFromGitHub }:

rustPlatform.buildRustPackage rec {
  pname = "{{.ProjectName}}";
  version = "{{.Version}}";

  src = fetchFromGitHub {
    owner = "{{.Owner}}";
    repo = "{{.ProjectName}}";
    rev = "v${version}";
    sha256 = lib.fakeHash;
  };

  cargoHash = lib.fakeHash;

{{- if .BuildInputs}}
  buildInputs = [
{{- range .BuildInputs}}
    {{.}}
{{- end}}
  ];
{{- end}}

{{- if .NativeBuildInputs}}
  nativeBuildInputs = [
{{- range .NativeBuildInputs}}
    {{.}}
{{- end}}
  ];
{{- end}}

{{- if .CheckPhase}}
  doCheck = true;
  checkPhase = ''
{{.CheckPhase}}
  '';
{{- else}}
  doCheck = true;
  checkPhase = ''
    cargo test
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
