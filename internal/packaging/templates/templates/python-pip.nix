{ lib, python3Packages, fetchFromGitHub }:

python3Packages.buildPythonApplication rec {
  pname = "{{.ProjectName}}";
  version = "{{.Version}}";
  format = "setuptools";

  src = fetchFromGitHub {
    owner = "{{.Owner}}";
    repo = "{{.ProjectName}}";
    rev = "v${version}";
    sha256 = lib.fakeHash;
  };

{{- if .Dependencies}}
  propagatedBuildInputs = with python3Packages; [
{{- range $name, $version := .Dependencies}}
    {{$name}}
{{- end}}
  ];
{{- end}}

{{- if .DevDependencies}}
  nativeCheckInputs = with python3Packages; [
{{- range $name, $version := .DevDependencies}}
    {{$name}}
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
    python -m pytest
  '';
{{- end}}

  pythonImportsCheck = [ "{{.ProjectName | replace "-" "_"}}" ];

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
