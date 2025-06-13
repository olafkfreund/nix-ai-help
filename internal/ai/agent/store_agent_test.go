package agent

import (
	"context"
	"nix-ai-help/internal/ai"
	"strings"
	"testing"
)

// MockProvider for testing StoreAgent
type MockStoreProvider struct {
	response string
}

func (m *MockStoreProvider) GenerateResponse(ctx context.Context, prompt string) (string, error) {
	return m.response, nil
}

func (m *MockStoreProvider) Query(query string) (string, error) {
	return m.response, nil
}

func (m *MockStoreProvider) QueryWithContext(ctx context.Context, prompt string) (string, error) {
	return m.response, nil
}

func (m *MockStoreProvider) GetPartialResponse() string {
	return ""
}

func (m *MockStoreProvider) StreamResponse(ctx context.Context, prompt string) (<-chan ai.StreamResponse, error) {
	ch := make(chan ai.StreamResponse, 1)
	ch <- ai.StreamResponse{Content: "mock stream response", Done: true}
	close(ch)
	return ch, nil
}

func TestNewStoreAgent(t *testing.T) {
	provider := &MockStoreProvider{response: "test"}
	agent := NewStoreAgent(provider)

	if agent == nil {
		t.Fatal("NewStoreAgent returned nil")
	}

	if agent.provider != provider {
		t.Error("Provider not set correctly")
	}

	if agent.context == nil {
		t.Error("Context not initialized")
	}
}

func TestStoreAgent_SetAndGetContext(t *testing.T) {
	provider := &MockStoreProvider{response: "test"}
	agent := NewStoreAgent(provider)

	ctx := &StoreContext{
		StorePath:     "/nix/store",
		StoreSize:     "50GB",
		FreeSpace:     "10GB",
		TargetPaths:   []string{"/nix/store/abc123-package"},
		OperationType: "analysis",
	}

	agent.SetContext(ctx)

	retrievedCtx := agent.GetContext()
	if retrievedCtx.StorePath != ctx.StorePath {
		t.Errorf("Expected StorePath %s, got %s", ctx.StorePath, retrievedCtx.StorePath)
	}
	if retrievedCtx.StoreSize != ctx.StoreSize {
		t.Errorf("Expected StoreSize %s, got %s", ctx.StoreSize, retrievedCtx.StoreSize)
	}
	if len(retrievedCtx.TargetPaths) != len(ctx.TargetPaths) {
		t.Errorf("Expected %d target paths, got %d", len(ctx.TargetPaths), len(retrievedCtx.TargetPaths))
	}
}

func TestStoreAgent_AnalyzeStoreHealth(t *testing.T) {
	tests := []struct {
		name        string
		storePath   string
		response    string
		wantError   bool
		wantContain []string
	}{
		{
			name:      "successful health analysis",
			storePath: "/nix/store",
			response:  "Store health analysis: All systems operational. Store integrity verified.",
			wantError: false,
			wantContain: []string{
				"Store Health Analysis",
				"Store health analysis",
				"Store Operation Safety Reminders",
			},
		},
		{
			name:      "custom store path analysis",
			storePath: "/custom/nix/store",
			response:  "Custom store analysis completed successfully.",
			wantError: false,
			wantContain: []string{
				"Store Health Analysis",
				"Custom store analysis",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &MockStoreProvider{response: tt.response}
			agent := NewStoreAgent(provider)

			result, err := agent.AnalyzeStoreHealth(context.Background(), tt.storePath)

			if (err != nil) != tt.wantError {
				t.Errorf("AnalyzeStoreHealth() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError {
				for _, want := range tt.wantContain {
					if !strings.Contains(result, want) {
						t.Errorf("AnalyzeStoreHealth() result missing %q", want)
					}
				}

				// Verify context was set
				ctx := agent.GetContext()
				if ctx.StorePath != tt.storePath {
					t.Errorf("Expected context StorePath %s, got %s", tt.storePath, ctx.StorePath)
				}
				if ctx.OperationType != "health_analysis" {
					t.Errorf("Expected OperationType 'health_analysis', got %s", ctx.OperationType)
				}
			}
		})
	}
}

func TestStoreAgent_QueryStorePaths(t *testing.T) {
	tests := []struct {
		name        string
		paths       []string
		queryType   string
		response    string
		wantError   bool
		wantContain []string
	}{
		{
			name:      "dependency query",
			paths:     []string{"/nix/store/abc123-package", "/nix/store/def456-library"},
			queryType: "dependencies",
			response:  "Dependency analysis: Found 5 runtime dependencies and 12 build dependencies.",
			wantError: false,
			wantContain: []string{
				"Store Path Query Results",
				"Dependency analysis",
				"Store Operation Safety Reminders",
			},
		},
		{
			name:      "closure analysis",
			paths:     []string{"/nix/store/xyz789-app"},
			queryType: "closure",
			response:  "Closure analysis: Total size 2.5GB with 45 store paths.",
			wantError: false,
			wantContain: []string{
				"Store Path Query Results",
				"Closure analysis",
			},
		},
		{
			name:      "reverse dependencies",
			paths:     []string{"/nix/store/lib123-shared"},
			queryType: "reverse-deps",
			response:  "Reverse dependency analysis: 8 packages depend on this library.",
			wantError: false,
			wantContain: []string{
				"Store Path Query Results",
				"Reverse dependency analysis",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &MockStoreProvider{response: tt.response}
			agent := NewStoreAgent(provider)

			result, err := agent.QueryStorePaths(context.Background(), tt.paths, tt.queryType)

			if (err != nil) != tt.wantError {
				t.Errorf("QueryStorePaths() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError {
				for _, want := range tt.wantContain {
					if !strings.Contains(result, want) {
						t.Errorf("QueryStorePaths() result missing %q", want)
					}
				}

				// Verify context was set
				ctx := agent.GetContext()
				if len(ctx.TargetPaths) != len(tt.paths) {
					t.Errorf("Expected %d target paths, got %d", len(tt.paths), len(ctx.TargetPaths))
				}
				if ctx.QueryType != tt.queryType {
					t.Errorf("Expected QueryType %s, got %s", tt.queryType, ctx.QueryType)
				}
				if ctx.OperationType != "path_query" {
					t.Errorf("Expected OperationType 'path_query', got %s", ctx.OperationType)
				}
			}
		})
	}
}

func TestStoreAgent_OptimizeGarbageCollection(t *testing.T) {
	tests := []struct {
		name        string
		generations []string
		dryRun      bool
		response    string
		wantError   bool
		wantContain []string
	}{
		{
			name:        "safe garbage collection",
			generations: []string{"1", "2", "3"},
			dryRun:      true,
			response:    "Garbage collection analysis: Can safely remove 3 old generations, saving 2.1GB.",
			wantError:   false,
			wantContain: []string{
				"Garbage Collection Strategy",
				"Garbage collection analysis",
				"Store Operation Safety Reminders",
			},
		},
		{
			name:        "production garbage collection",
			generations: []string{"5", "10", "15"},
			dryRun:      false,
			response:    "Production GC strategy: Recommended approach with staged deletion.",
			wantError:   false,
			wantContain: []string{
				"Garbage Collection Strategy",
				"Production GC strategy",
			},
		},
		{
			name:        "aggressive cleanup",
			generations: []string{},
			dryRun:      true,
			response:    "Aggressive cleanup analysis: All unreferenced paths identified.",
			wantError:   false,
			wantContain: []string{
				"Garbage Collection Strategy",
				"Aggressive cleanup analysis",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &MockStoreProvider{response: tt.response}
			agent := NewStoreAgent(provider)

			result, err := agent.OptimizeGarbageCollection(context.Background(), tt.generations, tt.dryRun)

			if (err != nil) != tt.wantError {
				t.Errorf("OptimizeGarbageCollection() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError {
				for _, want := range tt.wantContain {
					if !strings.Contains(result, want) {
						t.Errorf("OptimizeGarbageCollection() result missing %q", want)
					}
				}

				// Verify context was set
				ctx := agent.GetContext()
				if len(ctx.Generations) != len(tt.generations) {
					t.Errorf("Expected %d generations, got %d", len(tt.generations), len(ctx.Generations))
				}
				if ctx.DryRun != tt.dryRun {
					t.Errorf("Expected DryRun %v, got %v", tt.dryRun, ctx.DryRun)
				}
				if ctx.OperationType != "garbage_collection" {
					t.Errorf("Expected OperationType 'garbage_collection', got %s", ctx.OperationType)
				}
			}
		})
	}
}

func TestStoreAgent_RepairStoreIntegrity(t *testing.T) {
	tests := []struct {
		name           string
		corruptedPaths []string
		response       string
		wantError      bool
		wantContain    []string
	}{
		{
			name:           "single corruption repair",
			corruptedPaths: []string{"/nix/store/abc123-corrupted"},
			response:       "Integrity repair: Successfully repaired 1 corrupted path using backup source.",
			wantError:      false,
			wantContain: []string{
				"Store Integrity Repair",
				"Integrity repair",
				"Store Operation Safety Reminders",
			},
		},
		{
			name:           "multiple corruption repair",
			corruptedPaths: []string{"/nix/store/abc123-corrupted", "/nix/store/def456-damaged"},
			response:       "Multiple corruption repair: Analyzed 2 paths, repair strategy developed.",
			wantError:      false,
			wantContain: []string{
				"Store Integrity Repair",
				"Multiple corruption repair",
			},
		},
		{
			name:           "prevention guidance",
			corruptedPaths: []string{},
			response:       "Integrity analysis: Store is healthy, providing prevention guidance.",
			wantError:      false,
			wantContain: []string{
				"Store Integrity Repair",
				"Integrity analysis",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &MockStoreProvider{response: tt.response}
			agent := NewStoreAgent(provider)

			result, err := agent.RepairStoreIntegrity(context.Background(), tt.corruptedPaths)

			if (err != nil) != tt.wantError {
				t.Errorf("RepairStoreIntegrity() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError {
				for _, want := range tt.wantContain {
					if !strings.Contains(result, want) {
						t.Errorf("RepairStoreIntegrity() result missing %q", want)
					}
				}

				// Verify context was set
				ctx := agent.GetContext()
				if len(ctx.CorruptedPaths) != len(tt.corruptedPaths) {
					t.Errorf("Expected %d corrupted paths, got %d", len(tt.corruptedPaths), len(ctx.CorruptedPaths))
				}
				if ctx.OperationType != "integrity_repair" {
					t.Errorf("Expected OperationType 'integrity_repair', got %s", ctx.OperationType)
				}
			}
		})
	}
}

func TestStoreAgent_OptimizeStorePerformance(t *testing.T) {
	tests := []struct {
		name              string
		performanceIssues []string
		response          string
		wantError         bool
		wantContain       []string
	}{
		{
			name:              "cache optimization",
			performanceIssues: []string{"slow cache access", "network timeouts"},
			response:          "Performance optimization: Cache configuration improved, network settings tuned.",
			wantError:         false,
			wantContain: []string{
				"Store Performance Optimization",
				"Performance optimization",
				"Store Operation Safety Reminders",
			},
		},
		{
			name:              "general optimization",
			performanceIssues: []string{"slow builds", "high disk usage"},
			response:          "General optimization: Build parallelism increased, disk usage patterns analyzed.",
			wantError:         false,
			wantContain: []string{
				"Store Performance Optimization",
				"General optimization",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &MockStoreProvider{response: tt.response}
			agent := NewStoreAgent(provider)

			result, err := agent.OptimizeStorePerformance(context.Background(), tt.performanceIssues)

			if (err != nil) != tt.wantError {
				t.Errorf("OptimizeStorePerformance() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError {
				for _, want := range tt.wantContain {
					if !strings.Contains(result, want) {
						t.Errorf("OptimizeStorePerformance() result missing %q", want)
					}
				}

				// Verify context was set
				ctx := agent.GetContext()
				if len(ctx.Issues) != len(tt.performanceIssues) {
					t.Errorf("Expected %d performance issues, got %d", len(tt.performanceIssues), len(ctx.Issues))
				}
				if ctx.OperationType != "performance_optimization" {
					t.Errorf("Expected OperationType 'performance_optimization', got %s", ctx.OperationType)
				}
			}
		})
	}
}

func TestStoreAgent_ManageStoreCopying(t *testing.T) {
	tests := []struct {
		name        string
		sourceStore string
		targetStore string
		paths       []string
		response    string
		wantError   bool
		wantContain []string
	}{
		{
			name:        "local to remote copy",
			sourceStore: "/nix/store",
			targetStore: "ssh://remote/nix/store",
			paths:       []string{"/nix/store/abc123-package"},
			response:    "Store copying: Optimized network transfer strategy for 1 package to remote store.",
			wantError:   false,
			wantContain: []string{
				"Store Copying Management",
				"Store copying",
				"Store Operation Safety Reminders",
			},
		},
		{
			name:        "bulk migration",
			sourceStore: "/old/nix/store",
			targetStore: "/new/nix/store",
			paths:       []string{"/nix/store/pkg1", "/nix/store/pkg2", "/nix/store/pkg3"},
			response:    "Bulk migration: Strategy for migrating 3 packages with dependency preservation.",
			wantError:   false,
			wantContain: []string{
				"Store Copying Management",
				"Bulk migration",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &MockStoreProvider{response: tt.response}
			agent := NewStoreAgent(provider)

			result, err := agent.ManageStoreCopying(context.Background(), tt.sourceStore, tt.targetStore, tt.paths)

			if (err != nil) != tt.wantError {
				t.Errorf("ManageStoreCopying() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError {
				for _, want := range tt.wantContain {
					if !strings.Contains(result, want) {
						t.Errorf("ManageStoreCopying() result missing %q", want)
					}
				}

				// Verify context was set
				ctx := agent.GetContext()
				if len(ctx.TargetPaths) != len(tt.paths) {
					t.Errorf("Expected %d target paths, got %d", len(tt.paths), len(ctx.TargetPaths))
				}
				if ctx.OperationType != "store_copying" {
					t.Errorf("Expected OperationType 'store_copying', got %s", ctx.OperationType)
				}
			}
		})
	}
}

func TestStoreAgent_BuildPrompt(t *testing.T) {
	provider := &MockStoreProvider{response: "test"}
	agent := NewStoreAgent(provider)

	// Set some context
	agent.SetContext(&StoreContext{
		StorePath:     "/nix/store",
		StoreSize:     "50GB",
		OperationType: "test_operation",
		TargetPaths:   []string{"/nix/store/test"},
		Issues:        []string{"test issue"},
	})

	prompt := agent.buildPrompt("Test task", map[string]interface{}{
		"test_param": "test_value",
		"operation":  "test_op",
	})

	// Check that prompt contains expected sections
	expectedSections := []string{
		"Test task",
		"Store Context",
		"Store Path: /nix/store",
		"Store Size: 50GB",
		"Operation Type: test_operation",
		"Target Paths: [/nix/store/test]",
		"Known Issues: [test issue]",
		"Operation Details",
		"Test Param: test_value",
		"Operation: test_op",
		"Requirements",
		"Provide specific nix-store commands",
		"Include safety warnings",
		"Consider system stability",
	}

	for _, section := range expectedSections {
		if !strings.Contains(prompt, section) {
			t.Errorf("buildPrompt() missing expected section: %q", section)
		}
	}
}

func TestStoreAgent_FormatStoreResponse(t *testing.T) {
	provider := &MockStoreProvider{response: "test"}
	agent := NewStoreAgent(provider)

	response := "Test response content"
	operation := "Test Operation"

	formatted := agent.formatStoreResponse(response, operation)

	expectedSections := []string{
		"# Test Operation",
		"Test response content",
		"Store Operation Safety Reminders",
		"Always backup important data",
		"Test operations with --dry-run",
		"Verify store integrity",
		"Monitor disk space",
		"Keep track of active roots",
	}

	for _, section := range expectedSections {
		if !strings.Contains(formatted, section) {
			t.Errorf("formatStoreResponse() missing expected section: %q", section)
		}
	}
}
