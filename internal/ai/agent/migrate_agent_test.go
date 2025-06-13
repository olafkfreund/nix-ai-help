package agent

import (
	"context"
	"strings"
	"testing"

	"nix-ai-help/internal/ai/roles"
)

func TestNewMigrateAgent(t *testing.T) {
	provider := &MockProvider{response: "test response"}
	agent := NewMigrateAgent(provider)

	if agent == nil {
		t.Fatal("NewMigrateAgent returned nil")
	}

	if agent.provider != provider {
		t.Error("Provider not set correctly")
	}

	if agent.role != roles.RoleMigrate {
		t.Errorf("Expected role %s, got %s", roles.RoleMigrate, agent.role)
	}
}

func TestMigrateAgent_Query(t *testing.T) {
	tests := []struct {
		name         string
		question     string
		context      *MigrationContext
		providerResp string
		wantErr      bool
		checkOutput  func(string) bool
	}{
		{
			name:         "basic migration query",
			question:     "How do I migrate from NixOS 23.05 to 23.11?",
			providerResp: "To migrate from NixOS 23.05 to 23.11...",
			wantErr:      false,
			checkOutput: func(s string) bool {
				return len(s) > 0 && strings.Contains(s, "Migration Guidance")
			},
		},
		{
			name:     "migration with context",
			question: "What are the steps for machine migration?",
			context: &MigrationContext{
				SourceSystem:  "NixOS 23.05 x86_64",
				TargetSystem:  "NixOS 23.11 x86_64",
				MigrationType: "machine",
				HomeManager:   true,
			},
			providerResp: "For machine migration with Home Manager...",
			wantErr:      false,
			checkOutput: func(s string) bool {
				return strings.Contains(s, "Migration Guidance") && strings.Contains(s, "Checklist Reminders")
			},
		},
		{
			name:         "flake migration query",
			question:     "How do I convert my configuration to flakes during migration?",
			providerResp: "Converting to flakes during migration requires...",
			wantErr:      false,
			checkOutput: func(s string) bool {
				return strings.Contains(s, "Migration Guidance")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &MockProvider{response: tt.providerResp}
			agent := NewMigrateAgent(provider)

			if tt.context != nil {
				agent.SetMigrationContext(tt.context)
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

func TestMigrateAgent_GenerateResponse(t *testing.T) {
	tests := []struct {
		name         string
		request      string
		context      *MigrationContext
		providerResp string
		wantErr      bool
		checkOutput  func(string) bool
	}{
		{
			name:         "migration plan generation",
			request:      "Generate a migration plan for upgrading to latest NixOS",
			providerResp: "Migration Plan: 1. Backup system...",
			wantErr:      false,
			checkOutput: func(s string) bool {
				return strings.Contains(s, "Migration Guidance") && strings.Contains(s, "backup")
			},
		},
		{
			name:    "comprehensive migration with full context",
			request: "Create detailed migration steps",
			context: &MigrationContext{
				SourceSystem:   "NixOS 22.11",
				TargetSystem:   "NixOS 23.11",
				MigrationType:  "version",
				Services:       []string{"nginx", "postgresql"},
				Packages:       []string{"custom-package"},
				HomeManager:    true,
				FlakeUsage:     true,
				BackupStrategy: "borgbackup",
			},
			providerResp: "Detailed migration steps with services and packages...",
			wantErr:      false,
			checkOutput: func(s string) bool {
				return strings.Contains(s, "Migration Guidance") && strings.Contains(s, "Checklist Reminders")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &MockProvider{response: tt.providerResp}
			agent := NewMigrateAgent(provider)

			if tt.context != nil {
				agent.SetMigrationContext(tt.context)
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

func TestMigrateAgent_SetMigrationContext(t *testing.T) {
	agent := NewMigrateAgent(&MockProvider{})
	context := &MigrationContext{
		SourceSystem:  "NixOS 23.05",
		TargetSystem:  "NixOS 23.11",
		MigrationType: "version",
		HomeManager:   true,
	}

	agent.SetMigrationContext(context)
	retrieved := agent.GetMigrationContext()

	if retrieved.SourceSystem != context.SourceSystem {
		t.Errorf("Expected SourceSystem %s, got %s", context.SourceSystem, retrieved.SourceSystem)
	}

	if retrieved.TargetSystem != context.TargetSystem {
		t.Errorf("Expected TargetSystem %s, got %s", context.TargetSystem, retrieved.TargetSystem)
	}

	if retrieved.MigrationType != context.MigrationType {
		t.Errorf("Expected MigrationType %s, got %s", context.MigrationType, retrieved.MigrationType)
	}

	if retrieved.HomeManager != context.HomeManager {
		t.Errorf("Expected HomeManager %v, got %v", context.HomeManager, retrieved.HomeManager)
	}
}

func TestMigrateAgent_GetMigrationContext(t *testing.T) {
	agent := NewMigrateAgent(&MockProvider{})

	// Test with no context set
	context := agent.GetMigrationContext()
	if context == nil {
		t.Error("GetMigrationContext() returned nil")
	}

	// Test with context set
	migrationCtx := &MigrationContext{
		SourceSystem: "NixOS 22.11",
		TargetSystem: "NixOS 23.11",
	}
	agent.SetMigrationContext(migrationCtx)

	retrieved := agent.GetMigrationContext()
	if retrieved.SourceSystem != migrationCtx.SourceSystem {
		t.Errorf("Expected SourceSystem %s, got %s", migrationCtx.SourceSystem, retrieved.SourceSystem)
	}
}

func TestMigrateAgent_AnalyzeMigrationPath(t *testing.T) {
	tests := []struct {
		name         string
		sourceSystem string
		targetSystem string
		providerResp string
		wantErr      bool
		checkOutput  func(string) bool
	}{
		{
			name:         "version migration analysis",
			sourceSystem: "NixOS 23.05",
			targetSystem: "NixOS 23.11",
			providerResp: "Migration analysis: Key steps include...",
			wantErr:      false,
			checkOutput: func(s string) bool {
				return strings.Contains(s, "Migration Guidance")
			},
		},
		{
			name:         "major version migration",
			sourceSystem: "NixOS 22.11",
			targetSystem: "NixOS 24.05",
			providerResp: "Major version migration requires careful planning...",
			wantErr:      false,
			checkOutput: func(s string) bool {
				return strings.Contains(s, "Migration Guidance") && strings.Contains(s, "backup")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &MockProvider{response: tt.providerResp}
			agent := NewMigrateAgent(provider)

			result, err := agent.AnalyzeMigrationPath(context.Background(), tt.sourceSystem, tt.targetSystem)

			if (err != nil) != tt.wantErr {
				t.Errorf("AnalyzeMigrationPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !tt.checkOutput(result) {
				t.Errorf("AnalyzeMigrationPath() result doesn't meet expectations: %s", result)
			}

			// Verify context was set correctly
			ctx := agent.GetMigrationContext()
			if ctx.SourceSystem != tt.sourceSystem {
				t.Errorf("Expected SourceSystem %s, got %s", tt.sourceSystem, ctx.SourceSystem)
			}
			if ctx.TargetSystem != tt.targetSystem {
				t.Errorf("Expected TargetSystem %s, got %s", tt.targetSystem, ctx.TargetSystem)
			}
		})
	}
}

func TestMigrateAgent_GenerateMigrationPlan(t *testing.T) {
	tests := []struct {
		name         string
		context      *MigrationContext
		providerResp string
		wantErr      bool
		checkOutput  func(string) bool
	}{
		{
			name: "comprehensive migration plan",
			context: &MigrationContext{
				SourceSystem:   "NixOS 23.05",
				TargetSystem:   "NixOS 23.11",
				MigrationType:  "version",
				Services:       []string{"nginx", "postgresql"},
				HomeManager:    true,
				BackupStrategy: "rsync",
			},
			providerResp: "Comprehensive migration plan: 1. Pre-migration...",
			wantErr:      false,
			checkOutput: func(s string) bool {
				return strings.Contains(s, "Migration Guidance") && strings.Contains(s, "backup")
			},
		},
		{
			name: "flake migration plan",
			context: &MigrationContext{
				SourceSystem:  "NixOS 23.05",
				TargetSystem:  "NixOS 23.11",
				MigrationType: "flake",
				FlakeUsage:    true,
			},
			providerResp: "Flake migration plan includes...",
			wantErr:      false,
			checkOutput: func(s string) bool {
				return strings.Contains(s, "Migration Guidance")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &MockProvider{response: tt.providerResp}
			agent := NewMigrateAgent(provider)

			result, err := agent.GenerateMigrationPlan(context.Background(), tt.context)

			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateMigrationPlan() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !tt.checkOutput(result) {
				t.Errorf("GenerateMigrationPlan() result doesn't meet expectations: %s", result)
			}
		})
	}
}

func TestMigrateAgent_DiagnoseMigrationIssues(t *testing.T) {
	tests := []struct {
		name         string
		issues       []string
		context      *MigrationContext
		providerResp string
		wantErr      bool
		checkOutput  func(string) bool
	}{
		{
			name:   "service startup issues",
			issues: []string{"nginx failed to start", "postgresql connection refused"},
			context: &MigrationContext{
				SourceSystem:  "NixOS 23.05",
				TargetSystem:  "NixOS 23.11",
				MigrationType: "version",
			},
			providerResp: "Service startup issues can be resolved by...",
			wantErr:      false,
			checkOutput: func(s string) bool {
				return strings.Contains(s, "Migration Guidance")
			},
		},
		{
			name:   "configuration conflicts",
			issues: []string{"configuration option deprecated", "module conflicts"},
			context: &MigrationContext{
				SourceSystem:  "NixOS 22.11",
				TargetSystem:  "NixOS 23.11",
				MigrationType: "version",
				FlakeUsage:    true,
			},
			providerResp: "Configuration conflicts during migration...",
			wantErr:      false,
			checkOutput: func(s string) bool {
				return strings.Contains(s, "Migration Guidance") && strings.Contains(s, "backup")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &MockProvider{response: tt.providerResp}
			agent := NewMigrateAgent(provider)

			result, err := agent.DiagnoseMigrationIssues(context.Background(), tt.issues, tt.context)

			if (err != nil) != tt.wantErr {
				t.Errorf("DiagnoseMigrationIssues() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !tt.checkOutput(result) {
				t.Errorf("DiagnoseMigrationIssues() result doesn't meet expectations: %s", result)
			}

			// Verify issues were set in context
			ctx := agent.GetMigrationContext()
			if len(ctx.Issues) != len(tt.issues) {
				t.Errorf("Expected %d issues, got %d", len(tt.issues), len(ctx.Issues))
			}
		})
	}
}

func TestMigrateAgent_buildMigrationPrompt(t *testing.T) {
	agent := NewMigrateAgent(&MockProvider{})
	context := &MigrationContext{
		SourceSystem:   "NixOS 23.05",
		TargetSystem:   "NixOS 23.11",
		MigrationType:  "version",
		Services:       []string{"nginx"},
		HomeManager:    true,
		BackupStrategy: "borgbackup",
	}

	prompt := agent.buildMigrationPrompt("How do I migrate?", context)

	// Check that context information is included
	if !strings.Contains(prompt, "NixOS 23.05") {
		t.Error("Source system not included in prompt")
	}
	if !strings.Contains(prompt, "NixOS 23.11") {
		t.Error("Target system not included in prompt")
	}
	if !strings.Contains(prompt, "version") {
		t.Error("Migration type not included in prompt")
	}
	if !strings.Contains(prompt, "nginx") {
		t.Error("Services not included in prompt")
	}
	if !strings.Contains(prompt, "Home Manager: Yes") {
		t.Error("Home Manager status not included in prompt")
	}
	if !strings.Contains(prompt, "borgbackup") {
		t.Error("Backup strategy not included in prompt")
	}
	if !strings.Contains(prompt, "How do I migrate?") {
		t.Error("Original question not included in prompt")
	}
}

func TestMigrateAgent_formatMigrationResponse(t *testing.T) {
	agent := NewMigrateAgent(&MockProvider{})
	response := "Here are the migration steps..."

	formatted := agent.formatMigrationResponse(response)

	if !strings.Contains(formatted, "🔄 Migration Guidance") {
		t.Error("Migration guidance header not found")
	}
	if !strings.Contains(formatted, "Here are the migration steps...") {
		t.Error("Original response not included")
	}
	if !strings.Contains(formatted, "📋 Migration Checklist Reminders") {
		t.Error("Checklist reminders not found")
	}
	if !strings.Contains(formatted, "backup") {
		t.Error("Backup reminder not found")
	}
	if !strings.Contains(formatted, "rollback") {
		t.Error("Rollback reminder not found")
	}
}
