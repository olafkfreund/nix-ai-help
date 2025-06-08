{ lib, stdenv, fetchFromGitHub }:

stdenv.mkDerivation rec {
  pname = "{{.ProjectName}}";
  version = "{{.Version}}";

  src = fetchFromGitHub {
    owner = "{{.Owner}}";
    repo = "{{.ProjectName}}";
    rev = "v${version}";
    sha256 = lib.fakeHash;
  };

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

{{- if .ConfigureFlags}}
  configureFlags = [
{{- range .ConfigureFlags}}
    "{{.}}"
{{- end}}
  ];
{{- end}}

{{- if .MakeFlags}}
  makeFlags = [
{{- range .MakeFlags}}
    "{{.}}"
{{- end}}
  ];
{{- end}}

{{- if .BuildPhase}}
  buildPhase = ''
{{.BuildPhase}}
  '';
{{- end}}

{{- if .InstallPhase}}
  installPhase = ''
{{.InstallPhase}}
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
