package agent

import (
	"context"
	"strings"
	"testing"

	"nix-ai-help/internal/ai/roles"
)

func TestNewDevenvAgent(t *testing.T) {
	provider := &MockProvider{response: "test response"}
	agent := NewDevenvAgent(provider)

	if agent == nil {
		t.Fatal("NewDevenvAgent returned nil")
	}

	if agent.provider != provider {
		t.Error("Provider not set correctly")
	}

	if agent.role != roles.RoleDevenv {
		t.Errorf("Expected role %s, got %s", roles.RoleDevenv, agent.role)
	}
}

func TestDevenvAgent_Query(t *testing.T) {
	tests := []struct {
		name         string
		question     string
		context      *DevenvContext
		providerResp string
		wantErr      bool
		checkOutput  func(string) bool
	}{
		{
			name:         "basic devenv query",
			question:     "How do I set up a Rust development environment?",
			providerResp: "To set up a Rust development environment...",
			wantErr:      false,
			checkOutput: func(s string) bool {
				return len(s) > 0 && strings.Contains(s, "Development Environment Guidance")
			},
		},
		{
			name:     "devenv with context",
			question: "What tools do I need for this Node.js project?",
			context: &DevenvContext{
				ProjectType: "web",
				Languages:   []string{"javascript", "typescript"},
				BuildSystem: "npm",
				Flakes:      true,
				Direnv:      true,
			},
			providerResp: "For a Node.js project with TypeScript...",
			wantErr:      false,
			checkOutput: func(s string) bool {
				return strings.Contains(s, "Development Environment Guidance") && strings.Contains(s, "Best Practices")
			},
		},
		{
			name:         "python development environment",
			question:     "How do I create a Python development environment with poetry?",
			providerResp: "Python development environment with poetry requires...",
			wantErr:      false,
			checkOutput: func(s string) bool {
				return strings.Contains(s, "Development Environment Guidance")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &MockProvider{response: tt.providerResp}
			agent := NewDevenvAgent(provider)

			if tt.context != nil {
				agent.SetDevenvContext(tt.context)
			}

			result, err := agent.Query(context.Background(), tt.question)

			if (err != nil) != tt.wantErr {
				t.Errorf("Query() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !tt.checkOutput(result) {
				t.Errorf("Query() result doesn't meet expectations: %s", result)
			}
		})
	}
}

func TestDevenvAgent_GenerateResponse(t *testing.T) {
	tests := []struct {
		name         string
		request      string
		context      *DevenvContext
		providerResp string
		wantErr      bool
		checkOutput  func(string) bool
	}{
		{
			name:         "shell.nix generation",
			request:      "Generate a shell.nix for Go development",
			providerResp: "Here's a shell.nix for Go development...",
			wantErr:      false,
			checkOutput: func(s string) bool {
				return strings.Contains(s, "Development Environment Guidance") && strings.Contains(s, "declarative")
			},
		},
		{
			name:    "comprehensive development environment",
			request: "Create a complete development environment setup",
			context: &DevenvContext{
				ProjectType:      "web",
				Languages:        []string{"javascript", "rust"},
				Tools:            []string{"nodejs", "cargo", "webpack"},
				Services:         []string{"postgresql", "redis"},
				Frameworks:       []string{"react", "actix-web"},
				BuildSystem:      "npm",
				TestingFramework: "jest",
				LintingTools:     []string{"eslint", "rustfmt"},
				Flakes:           true,
				Direnv:           true,
			},
			providerResp: "Complete development environment setup includes...",
			wantErr:      false,
			checkOutput: func(s string) bool {
				return strings.Contains(s, "Development Environment Guidance") && strings.Contains(s, "Best Practices")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &MockProvider{response: tt.providerResp}
			agent := NewDevenvAgent(provider)

			if tt.context != nil {
				agent.SetDevenvContext(tt.context)
			}

			result, err := agent.GenerateResponse(context.Background(), tt.request)

			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !tt.checkOutput(result) {
				t.Errorf("GenerateResponse() result doesn't meet expectations: %s", result)
			}
		})
	}
}

func TestDevenvAgent_SetDevenvContext(t *testing.T) {
	agent := NewDevenvAgent(&MockProvider{})
	context := &DevenvContext{
		ProjectType: "web",
		Languages:   []string{"javascript", "typescript"},
		Tools:       []string{"nodejs", "webpack"},
		BuildSystem: "npm",
		Flakes:      true,
		Direnv:      true,
	}

	agent.SetDevenvContext(context)
	retrieved := agent.GetDevenvContext()

	if retrieved.ProjectType != context.ProjectType {
		t.Errorf("Expected ProjectType %s, got %s", context.ProjectType, retrieved.ProjectType)
	}

	if len(retrieved.Languages) != len(context.Languages) {
		t.Errorf("Expected %d languages, got %d", len(context.Languages), len(retrieved.Languages))
	}

	if len(retrieved.Tools) != len(context.Tools) {
		t.Errorf("Expected %d tools, got %d", len(context.Tools), len(retrieved.Tools))
	}

	if retrieved.BuildSystem != context.BuildSystem {
		t.Errorf("Expected BuildSystem %s, got %s", context.BuildSystem, retrieved.BuildSystem)
	}

	if retrieved.Flakes != context.Flakes {
		t.Errorf("Expected Flakes %v, got %v", context.Flakes, retrieved.Flakes)
	}

	if retrieved.Direnv != context.Direnv {
		t.Errorf("Expected Direnv %v, got %v", context.Direnv, retrieved.Direnv)
	}
}

func TestDevenvAgent_GetDevenvContext(t *testing.T) {
	agent := NewDevenvAgent(&MockProvider{})

	// Test with no context set
	context := agent.GetDevenvContext()
	if context == nil {
		t.Error("GetDevenvContext() returned nil")
	}

	// Test with context set
	devenvCtx := &DevenvContext{
		ProjectType: "web",
		Languages:   []string{"go"},
		Flakes:      true,
	}
	agent.SetDevenvContext(devenvCtx)

	retrieved := agent.GetDevenvContext()
	if retrieved.ProjectType != devenvCtx.ProjectType {
		t.Errorf("Expected ProjectType %s, got %s", devenvCtx.ProjectType, retrieved.ProjectType)
	}
}

func TestDevenvAgent_AnalyzeProject(t *testing.T) {
	tests := []struct {
		name         string
		projectPath  string
		projectType  string
		providerResp string
		wantErr      bool
		checkOutput  func(string) bool
	}{
		{
			name:         "rust project analysis",
			projectPath:  "/home/user/my-rust-project",
			projectType:  "rust",
			providerResp: "Rust project analysis shows...",
			wantErr:      false,
			checkOutput: func(s string) bool {
				return strings.Contains(s, "Development Environment Guidance")
			},
		},
		{
			name:         "node.js project analysis",
			projectPath:  "/home/user/my-node-app",
			projectType:  "nodejs",
			providerResp: "Node.js project requires...",
			wantErr:      false,
			checkOutput: func(s string) bool {
				return strings.Contains(s, "Development Environment Guidance") && strings.Contains(s, "declarative")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &MockProvider{response: tt.providerResp}
			agent := NewDevenvAgent(provider)

			result, err := agent.AnalyzeProject(context.Background(), tt.projectPath, tt.projectType)

			if (err != nil) != tt.wantErr {
				t.Errorf("AnalyzeProject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !tt.checkOutput(result) {
				t.Errorf("AnalyzeProject() result doesn't meet expectations: %s", result)
			}

			// Verify context was set correctly
			ctx := agent.GetDevenvContext()
			if ctx.ProjectType != tt.projectType {
				t.Errorf("Expected ProjectType %s, got %s", tt.projectType, ctx.ProjectType)
			}
			if ctx.ProjectRoot != tt.projectPath {
				t.Errorf("Expected ProjectRoot %s, got %s", tt.projectPath, ctx.ProjectRoot)
			}
		})
	}
}

func TestDevenvAgent_GenerateShellNix(t *testing.T) {
	tests := []struct {
		name         string
		context      *DevenvContext
		providerResp string
		wantErr      bool
		checkOutput  func(string) bool
	}{
		{
			name: "rust shell.nix generation",
			context: &DevenvContext{
				ProjectType: "rust",
				Languages:   []string{"rust"},
				Tools:       []string{"cargo", "rustc"},
				BuildSystem: "cargo",
			},
			providerResp: "{ pkgs ? import <nixpkgs> {} }: pkgs.mkShell...",
			wantErr:      false,
			checkOutput: func(s string) bool {
				return strings.Contains(s, "Development Environment Guidance") && strings.Contains(s, "direnv")
			},
		},
		{
			name: "multi-language shell.nix",
			context: &DevenvContext{
				ProjectType: "fullstack",
				Languages:   []string{"javascript", "rust", "python"},
				Tools:       []string{"nodejs", "cargo", "python3"},
				Services:    []string{"postgresql"},
			},
			providerResp: "Multi-language shell.nix with Node.js, Rust, Python...",
			wantErr:      false,
			checkOutput: func(s string) bool {
				return strings.Contains(s, "Development Environment Guidance")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &MockProvider{response: tt.providerResp}
			agent := NewDevenvAgent(provider)

			result, err := agent.GenerateShellNix(context.Background(), tt.context)

			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateShellNix() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !tt.checkOutput(result) {
				t.Errorf("GenerateShellNix() result doesn't meet expectations: %s", result)
			}
		})
	}
}

func TestDevenvAgent_GenerateFlakeNix(t *testing.T) {
	tests := []struct {
		name         string
		context      *DevenvContext
		providerResp string
		wantErr      bool
		checkOutput  func(string) bool
	}{
		{
			name: "go flake.nix generation",
			context: &DevenvContext{
				ProjectType: "go",
				Languages:   []string{"go"},
				Tools:       []string{"go", "golangci-lint"},
				BuildSystem: "go",
			},
			providerResp: "Go flake.nix with development shell and build outputs...",
			wantErr:      false,
			checkOutput: func(s string) bool {
				return strings.Contains(s, "Development Environment Guidance") && strings.Contains(s, "cachix")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &MockProvider{response: tt.providerResp}
			agent := NewDevenvAgent(provider)

			result, err := agent.GenerateFlakeNix(context.Background(), tt.context)

			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateFlakeNix() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !tt.checkOutput(result) {
				t.Errorf("GenerateFlakeNix() result doesn't meet expectations: %s", result)
			}

			// Verify Flakes was set to true
			ctx := agent.GetDevenvContext()
			if !ctx.Flakes {
				t.Error("Expected Flakes to be set to true")
			}
		})
	}
}

func TestDevenvAgent_SetupDirenv(t *testing.T) {
	tests := []struct {
		name         string
		context      *DevenvContext
		providerResp string
		wantErr      bool
		checkOutput  func(string) bool
	}{
		{
			name: "direnv setup for flake project",
			context: &DevenvContext{
				ProjectType: "web",
				Flakes:      true,
			},
			providerResp: "Direnv setup with flakes requires...",
			wantErr:      false,
			checkOutput: func(s string) bool {
				return strings.Contains(s, "Development Environment Guidance")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &MockProvider{response: tt.providerResp}
			agent := NewDevenvAgent(provider)

			result, err := agent.SetupDirenv(context.Background(), tt.context)

			if (err != nil) != tt.wantErr {
				t.Errorf("SetupDirenv() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !tt.checkOutput(result) {
				t.Errorf("SetupDirenv() result doesn't meet expectations: %s", result)
			}

			// Verify Direnv was set to true
			ctx := agent.GetDevenvContext()
			if !ctx.Direnv {
				t.Error("Expected Direnv to be set to true")
			}
		})
	}
}

func TestDevenvAgent_OptimizeBuildPerformance(t *testing.T) {
	tests := []struct {
		name         string
		context      *DevenvContext
		providerResp string
		wantErr      bool
		checkOutput  func(string) bool
	}{
		{
			name: "optimize rust build performance",
			context: &DevenvContext{
				ProjectType: "rust",
				BuildSystem: "cargo",
				Languages:   []string{"rust"},
			},
			providerResp: "Rust build optimization includes sccache, link-time optimization...",
			wantErr:      false,
			checkOutput: func(s string) bool {
				return strings.Contains(s, "Development Environment Guidance")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &MockProvider{response: tt.providerResp}
			agent := NewDevenvAgent(provider)

			result, err := agent.OptimizeBuildPerformance(context.Background(), tt.context)

			if (err != nil) != tt.wantErr {
				t.Errorf("OptimizeBuildPerformance() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !tt.checkOutput(result) {
				t.Errorf("OptimizeBuildPerformance() result doesn't meet expectations: %s", result)
			}
		})
	}
}

func TestDevenvAgent_TroubleshootEnvironment(t *testing.T) {
	tests := []struct {
		name         string
		issues       []string
		context      *DevenvContext
		providerResp string
		wantErr      bool
		checkOutput  func(string) bool
	}{
		{
			name:   "package not found error",
			issues: []string{"package 'nodejs' not found", "shell hook fails"},
			context: &DevenvContext{
				ProjectType: "web",
				Languages:   []string{"javascript"},
			},
			providerResp: "Package not found errors usually indicate...",
			wantErr:      false,
			checkOutput: func(s string) bool {
				return strings.Contains(s, "Development Environment Guidance")
			},
		},
		{
			name:   "build failures",
			issues: []string{"cargo build fails", "linker errors"},
			context: &DevenvContext{
				ProjectType: "rust",
				BuildSystem: "cargo",
			},
			providerResp: "Cargo build failures with linker errors...",
			wantErr:      false,
			checkOutput: func(s string) bool {
				return strings.Contains(s, "Development Environment Guidance") && strings.Contains(s, "Pin package")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &MockProvider{response: tt.providerResp}
			agent := NewDevenvAgent(provider)

			result, err := agent.TroubleshootEnvironment(context.Background(), tt.issues, tt.context)

			if (err != nil) != tt.wantErr {
				t.Errorf("TroubleshootEnvironment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !tt.checkOutput(result) {
				t.Errorf("TroubleshootEnvironment() result doesn't meet expectations: %s", result)
			}
		})
	}
}

func TestDevenvAgent_buildDevenvPrompt(t *testing.T) {
	agent := NewDevenvAgent(&MockProvider{})
	context := &DevenvContext{
		ProjectType:      "web",
		Languages:        []string{"javascript", "typescript"},
		Tools:            []string{"nodejs", "webpack"},
		Services:         []string{"postgresql"},
		Frameworks:       []string{"react"},
		BuildSystem:      "npm",
		TestingFramework: "jest",
		LintingTools:     []string{"eslint", "prettier"},
		Flakes:           true,
		Direnv:           true,
	}

	prompt := agent.buildDevenvPrompt("How do I set up this environment?", context)

	// Check that context information is included
	if !strings.Contains(prompt, "web") {
		t.Error("Project type not included in prompt")
	}
	if !strings.Contains(prompt, "javascript") {
		t.Error("Languages not included in prompt")
	}
	if !strings.Contains(prompt, "nodejs") {
		t.Error("Tools not included in prompt")
	}
	if !strings.Contains(prompt, "postgresql") {
		t.Error("Services not included in prompt")
	}
	if !strings.Contains(prompt, "react") {
		t.Error("Frameworks not included in prompt")
	}
	if !strings.Contains(prompt, "npm") {
		t.Error("Build system not included in prompt")
	}
	if !strings.Contains(prompt, "jest") {
		t.Error("Testing framework not included in prompt")
	}
	if !strings.Contains(prompt, "eslint") {
		t.Error("Linting tools not included in prompt")
	}
	if !strings.Contains(prompt, "Using Flakes: Yes") {
		t.Error("Flakes status not included in prompt")
	}
	if !strings.Contains(prompt, "Using direnv: Yes") {
		t.Error("Direnv status not included in prompt")
	}
	if !strings.Contains(prompt, "How do I set up this environment?") {
		t.Error("Original question not included in prompt")
	}
}

func TestDevenvAgent_formatDevenvResponse(t *testing.T) {
	agent := NewDevenvAgent(&MockProvider{})
	response := "Here are the development environment setup steps..."

	formatted := agent.formatDevenvResponse(response)

	if !strings.Contains(formatted, "🛠️ Development Environment Guidance") {
		t.Error("Development environment guidance header not found")
	}
	if !strings.Contains(formatted, "Here are the development environment setup steps...") {
		t.Error("Original response not included")
	}
	if !strings.Contains(formatted, "📋 Development Environment Best Practices") {
		t.Error("Best practices section not found")
	}
	if !strings.Contains(formatted, "declarative configuration") {
		t.Error("Declarative configuration reminder not found")
	}
	if !strings.Contains(formatted, "direnv") {
		t.Error("Direnv reminder not found")
	}
	if !strings.Contains(formatted, "cachix") {
		t.Error("Cachix recommendation not found")
	}
}
