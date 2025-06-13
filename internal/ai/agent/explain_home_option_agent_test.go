package agent

import (
	"context"
	"strings"
	"testing"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/ai/roles"
)

// MockProviderForHomeOptions implements the Provider interface for testing
type MockProviderForHomeOptions struct {
	response  string
	err       error
	LastQuery string
}

func (m *MockProviderForHomeOptions) Query(prompt string) (string, error) {
	return m.response, m.err
}

func (m *MockProviderForHomeOptions) QueryWithContext(ctx context.Context, prompt string) (string, error) {
	return m.response, m.err
}

func (m *MockProviderForHomeOptions) GetPartialResponse() string {
	return ""
}

func (m *MockProviderForHomeOptions) StreamResponse(ctx context.Context, prompt string) (<-chan ai.StreamResponse, error) {
	ch := make(chan ai.StreamResponse, 1)
	ch <- ai.StreamResponse{Content: m.response, Done: true}
	close(ch)
	return ch, nil
}

func (m *MockProviderForHomeOptions) GenerateResponse(ctx context.Context, prompt string) (string, error) {
	return m.response, m.err
}

func TestExplainHomeOptionAgent_NewExplainHomeOptionAgent(t *testing.T) {
	mockProvider := &MockProviderForHomeOptions{}
	agent := NewExplainHomeOptionAgent(mockProvider, nil)

	if agent == nil {
		t.Fatal("NewExplainHomeOptionAgent returned nil")
	}
}

func TestExplainHomeOptionAgent_SetRole(t *testing.T) {
	mockProvider := &MockProviderForHomeOptions{}
	agent := NewExplainHomeOptionAgent(mockProvider, nil)

	err := agent.SetRole(roles.RoleExplainHomeOption)

	if err != nil {
		t.Errorf("SetRole failed: %v", err)
	}
}

func TestExplainHomeOptionAgent_QueryWithContext(t *testing.T) {
	mockProvider := &MockProviderForHomeOptions{
		response: "Git is a version control system...",
	}
	agent := NewExplainHomeOptionAgent(mockProvider, nil)

	// Set role
	err := agent.SetRole(roles.RoleExplainHomeOption)
	if err != nil {
		t.Fatalf("SetRole failed: %v", err)
	}

	question := "What does programs.git.enable do?"
	homeCtx := &HomeOptionContext{
		OptionPath:        "programs.git.enable",
		Category:          "User Programs",
		ProgramName:       "git",
		ConfigFiles:       []string{".gitconfig", ".gitignore_global"},
		DotfileLocation:   "$HOME/.config/git/",
		RelatedOpts:       []string{"programs.git.userName", "programs.git.userEmail"},
		SystemIntegration: "Complements system-wide program configuration with user-specific settings",
	}

	response, err := agent.QueryWithContext(context.Background(), question, homeCtx)

	if err != nil {
		t.Errorf("QueryWithContext failed: %v", err)
	}
	if response == "" {
		t.Error("Expected non-empty response")
	}
	if response != mockProvider.response {
		t.Errorf("QueryWithContext() = %v, want %v", response, mockProvider.response)
	}

	// Verify prompt contains expected elements
	if !strings.Contains(mockProvider.LastQuery, "programs.git.enable") {
		t.Error("Expected query to contain option path")
	}
	if !strings.Contains(mockProvider.LastQuery, "User Programs") {
		t.Error("Expected query to contain category")
	}
	if !strings.Contains(mockProvider.LastQuery, question) {
		t.Error("Expected query to contain original question")
	}
}

func TestExplainHomeOptionAgent_DetermineDotfileLocation(t *testing.T) {
	agent := NewExplainHomeOptionAgent(nil, nil)

	tests := []struct {
		name             string
		optionPath       string
		expectedLocation string
	}{
		{
			name:             "git program option",
			optionPath:       "programs.git.enable",
			expectedLocation: "$HOME/.config/git/",
		},
		{
			name:             "zsh shell option",
			optionPath:       "programs.zsh.enable",
			expectedLocation: "$HOME/.zshrc and $HOME/.config/zsh/",
		},
		{
			name:             "unknown option",
			optionPath:       "unknown.option",
			expectedLocation: "varies by application",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := agent.determineDotfileLocation(tt.optionPath)
			if result != tt.expectedLocation {
				t.Errorf("determineDotfileLocation() = %v, want %v", result, tt.expectedLocation)
			}
		})
	}
}

func TestExplainHomeOptionAgent_GetConfigFiles(t *testing.T) {
	agent := NewExplainHomeOptionAgent(nil, nil)

	tests := []struct {
		name          string
		optionPath    string
		expectedFiles []string
	}{
		{
			name:          "git configuration files",
			optionPath:    "programs.git.enable",
			expectedFiles: []string{".gitconfig", ".gitignore_global"},
		},
		{
			name:          "zsh configuration files",
			optionPath:    "programs.zsh.enable",
			expectedFiles: []string{".zshrc", ".zshenv", ".zprofile"},
		},
		{
			name:          "unknown option",
			optionPath:    "unknown.option",
			expectedFiles: []string{"unknown configuration files"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := agent.getConfigFiles(tt.optionPath)
			if len(result) != len(tt.expectedFiles) {
				t.Errorf("getConfigFiles() returned %d files, want %d", len(result), len(tt.expectedFiles))
			}
			for i, expected := range tt.expectedFiles {
				if i >= len(result) || result[i] != expected {
					t.Errorf("getConfigFiles() = %v, want %v", result, tt.expectedFiles)
					break
				}
			}
		})
	}
}

func TestExplainHomeOptionAgent_FindRelatedHomeOptions(t *testing.T) {
	agent := NewExplainHomeOptionAgent(nil, nil)

	tests := []struct {
		name          string
		optionPath    string
		expectedCount int
		shouldContain []string
	}{
		{
			name:          "git related options",
			optionPath:    "programs.git.enable",
			expectedCount: 3,
			shouldContain: []string{"programs.git.userName", "programs.git.userEmail", "programs.git.aliases"},
		},
		{
			name:          "firefox related options",
			optionPath:    "programs.firefox.enable",
			expectedCount: 2,
			shouldContain: []string{"programs.firefox.profiles", "programs.firefox.extensions"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			related := agent.findRelatedHomeOptions(tt.optionPath)

			if len(related) != tt.expectedCount {
				t.Errorf("findRelatedHomeOptions() returned %d options, want %d", len(related), tt.expectedCount)
			}

			for _, shouldContain := range tt.shouldContain {
				found := false
				for _, rel := range related {
					if rel == shouldContain {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected related options to contain %q, but it didn't", shouldContain)
				}
			}

			// Verify original option is not included
			for _, rel := range related {
				if rel == tt.optionPath {
					t.Errorf("Expected related options not to contain original option %q", tt.optionPath)
				}
			}
		})
	}
}
