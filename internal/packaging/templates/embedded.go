package templates

import (
	_ "embed"
	"strings"
	"text/template"
)

// Embedded template files - these will be included in the binary when built with Nix flakes
//
//go:embed templates/javascript-npm.nix
var javascriptNpmTemplate string

//go:embed templates/typescript-npm.nix
var typescriptNpmTemplate string

//go:embed templates/python-pip.nix
var pythonPipTemplate string

//go:embed templates/python-poetry.nix
var pythonPoetryTemplate string

//go:embed templates/rust-cargo.nix
var rustCargoTemplate string

//go:embed templates/go-modules.nix
var goModulesTemplate string

//go:embed templates/c-cmake.nix
var cCmakeTemplate string

//go:embed templates/cpp-cmake.nix
var cppCmakeTemplate string

//go:embed templates/default.nix
var defaultTemplate string

// Template function map for use in templates
var templateFuncs = template.FuncMap{
	"lower":     strings.ToLower,
	"upper":     strings.ToUpper,
	"title":     strings.Title,
	"join":      strings.Join,
	"replace":   strings.ReplaceAll,
	"contains":  strings.Contains,
	"hasPrefix": strings.HasPrefix,
	"hasSuffix": strings.HasSuffix,
}

// getEmbeddedTemplates returns a map of embedded template content
func getEmbeddedTemplates() map[string]string {
	return map[string]string{
		"javascript-npm": javascriptNpmTemplate,
		"typescript-npm": typescriptNpmTemplate,
		"python-pip":     pythonPipTemplate,
		"python-poetry":  pythonPoetryTemplate,
		"rust-cargo":     rustCargoTemplate,
		"go-modules":     goModulesTemplate,
		"c-cmake":        cCmakeTemplate,
		"cpp-cmake":      cppCmakeTemplate,
		"default":        defaultTemplate,
	}
}
