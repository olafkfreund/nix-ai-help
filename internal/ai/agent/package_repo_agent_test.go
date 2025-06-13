package agent

import (
	"context"
	"testing"

	"nix-ai-help/internal/ai/roles"

	"github.com/stretchr/testify/require"
)

func TestPackageRepoAgent_Query(t *testing.T) {
	mockProvider := &MockProvider{response: "package-repo agent response"}
	agent := NewPackageRepoAgent(mockProvider)

	repoCtx := &PackageRepoContext{
		RepositoryURL:   "https://github.com/example/rust-project",
		ProjectLanguage: "rust",
		BuildSystem:     "cargo",
		Dependencies:    []string{"serde", "tokio", "clap"},
	}
	agent.SetContext(repoCtx)

	input := "Generate a Nix derivation for this Rust project"
	resp, err := agent.Query(context.Background(), input)
	require.NoError(t, err)
	require.Contains(t, resp, "package-repo agent")
}

func TestPackageRepoAgent_GenerateResponse(t *testing.T) {
	mockProvider := &MockProvider{response: "package-repo agent response"}
	agent := NewPackageRepoAgent(mockProvider)

	repoCtx := &PackageRepoContext{
		RepositoryURL:   "https://github.com/example/nodejs-app",
		ProjectLanguage: "javascript",
		BuildSystem:     "npm",
		Dependencies:    []string{"express", "lodash", "axios"},
	}
	agent.SetContext(repoCtx)

	input := "Help me package this Node.js application for Nix"
	resp, err := agent.GenerateResponse(context.Background(), input)
	require.NoError(t, err)
	require.Contains(t, resp, "package-repo agent response")
}

func TestPackageRepoAgent_SetRole(t *testing.T) {
	mockProvider := &MockProvider{}
	agent := NewPackageRepoAgent(mockProvider)

	// Test setting a valid role
	err := agent.SetRole(roles.RolePackageRepo)
	require.NoError(t, err)
	require.Equal(t, roles.RolePackageRepo, agent.role)

	// Test setting context
	repoCtx := &PackageRepoContext{ProjectLanguage: "python"}
	agent.SetContext(repoCtx)
	require.Equal(t, repoCtx, agent.contextData)
}

func TestPackageRepoAgent_InvalidRole(t *testing.T) {
	mockProvider := &MockProvider{}
	agent := NewPackageRepoAgent(mockProvider)
	// Manually set an invalid role to test validation
	agent.role = ""
	_, err := agent.Query(context.Background(), "test question")
	require.Error(t, err)
	require.Contains(t, err.Error(), "role not set")
}

func TestPackageRepoContext_Formatting(t *testing.T) {
	repoCtx := &PackageRepoContext{
		RepositoryURL:    "https://github.com/example/python-package",
		RepositoryPath:   "/home/user/projects/python-package",
		ProjectLanguage:  "python",
		BuildSystem:      "setuptools",
		Dependencies:     []string{"requests", "numpy", "pandas", "pytest"},
		PackageManagers:  []string{"requirements.txt", "setup.py", "pyproject.toml"},
		LicenseInfo:      "BSD-3-Clause",
		ProjectMetadata:  map[string]string{"name": "mypackage", "version": "2.1.0", "author": "Example Developer"},
		SourceFiles:      []string{"mypackage/__init__.py", "mypackage/core.py", "mypackage/utils.py"},
		ConfigFiles:      []string{"setup.py", "pyproject.toml", "tox.ini", ".flake8"},
		BuildScripts:     []string{"setup.py", "scripts/build.sh"},
		TestCommands:     []string{"pytest", "python -m unittest", "tox"},
		Documentation:    "Comprehensive Python package for data processing with extensive test suite",
		ExistingNix:      "{ lib, buildPythonPackage, fetchFromGitHub }: buildPythonPackage rec { ... }",
		PackageVersion:   "2.1.0",
		ArchitectureReqs: []string{"x86_64-linux", "aarch64-linux", "x86_64-darwin"},
	}

	// Test that context can be created and has expected fields
	require.NotEmpty(t, repoCtx.RepositoryURL)
	require.NotEmpty(t, repoCtx.RepositoryPath)
	require.Equal(t, "python", repoCtx.ProjectLanguage)
	require.Equal(t, "setuptools", repoCtx.BuildSystem)
	require.Len(t, repoCtx.Dependencies, 4)
	require.Len(t, repoCtx.PackageManagers, 3)
	require.Equal(t, "BSD-3-Clause", repoCtx.LicenseInfo)
	require.Len(t, repoCtx.ProjectMetadata, 3)
	require.Len(t, repoCtx.SourceFiles, 3)
	require.Len(t, repoCtx.ConfigFiles, 4)
	require.Len(t, repoCtx.BuildScripts, 2)
	require.Len(t, repoCtx.TestCommands, 3)
	require.NotEmpty(t, repoCtx.Documentation)
	require.NotEmpty(t, repoCtx.ExistingNix)
	require.Equal(t, "2.1.0", repoCtx.PackageVersion)
	require.Len(t, repoCtx.ArchitectureReqs, 3)
}
