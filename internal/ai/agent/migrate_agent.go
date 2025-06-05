package agent

import (
	"context"
	"fmt"
	"strings"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/ai/roles"
)

// MigrateAgent assists with NixOS migrations and system upgrades.
type MigrateAgent struct {
	BaseAgent
}

// MigrationContext provides context for migration operations.
type MigrationContext struct {
	SourceSystem   string   // Current NixOS version/system details
	TargetSystem   string   // Target NixOS version/system
	MigrationType  string   // "version", "machine", "config", "flake"
	CurrentConfig  string   // Current configuration files
	Issues         []string // Known migration issues or concerns
	Hardware       string   // Hardware configuration details
	Services       []string // Running services to migrate
	Packages       []string // Custom packages that need migration
	HomeManager    bool     // Whether Home Manager is used
	FlakeUsage     bool     // Whether flakes are used
	BackupStrategy string   // Backup approach being used
	RollbackPlan   string   // Rollback strategy if migration fails
}

// NewMigrateAgent creates a new MigrateAgent with the specified provider.
func NewMigrateAgent(provider ai.Provider) *MigrateAgent {
	agent := &MigrateAgent{
		BaseAgent: BaseAgent{
			provider: provider,
			role:     roles.RoleMigrate,
		},
	}
	return agent
}

// Query provides migration guidance using the provider's Query method.
func (a *MigrateAgent) Query(ctx context.Context, question string) (string, error) {
	if err := a.validateRole(); err != nil {
		return "", err
	}

	// Build migration-specific prompt with context
	prompt := a.buildMigrationPrompt(question, a.getMigrationContextFromData())

	response, err := a.provider.Query(ctx, prompt)
	if err != nil {
		return "", err
	}

	return a.formatMigrationResponse(response), nil
}

// GenerateResponse provides detailed migration plans using the provider's GenerateResponse method.
func (a *MigrateAgent) GenerateResponse(ctx context.Context, request string) (string, error) {
	if err := a.validateRole(); err != nil {
		return "", err
	}

	// Build comprehensive migration prompt
	prompt := a.buildMigrationPrompt(request, a.getMigrationContextFromData())

	response, err := a.provider.GenerateResponse(ctx, prompt)
	if err != nil {
		return "", err
	}

	return a.formatMigrationResponse(response), nil
}

// SetMigrationContext sets migration-specific context.
func (a *MigrateAgent) SetMigrationContext(context *MigrationContext) {
	a.contextData = context
}

// GetMigrationContext returns the current migration context.
func (a *MigrateAgent) GetMigrationContext() *MigrationContext {
	if ctx, ok := a.contextData.(*MigrationContext); ok {
		return ctx
	}
	return &MigrationContext{}
}

// buildMigrationPrompt constructs a migration-specific prompt.
func (a *MigrateAgent) buildMigrationPrompt(question string, context *MigrationContext) string {
	var prompt strings.Builder

	// Get role-specific prompt template
	if template, exists := roles.RolePromptTemplate[a.role]; exists {
		prompt.WriteString(template)
		prompt.WriteString("\n\n")
	}

	// Add migration context
	prompt.WriteString("Migration Context:\n")

	if context.SourceSystem != "" {
		prompt.WriteString(fmt.Sprintf("- Source System: %s\n", context.SourceSystem))
	}
	if context.TargetSystem != "" {
		prompt.WriteString(fmt.Sprintf("- Target System: %s\n", context.TargetSystem))
	}
	if context.MigrationType != "" {
		prompt.WriteString(fmt.Sprintf("- Migration Type: %s\n", context.MigrationType))
	}
	if context.CurrentConfig != "" {
		prompt.WriteString(fmt.Sprintf("- Current Configuration:\n%s\n", context.CurrentConfig))
	}
	if len(context.Issues) > 0 {
		prompt.WriteString(fmt.Sprintf("- Known Issues: %s\n", strings.Join(context.Issues, ", ")))
	}
	if context.Hardware != "" {
		prompt.WriteString(fmt.Sprintf("- Hardware: %s\n", context.Hardware))
	}
	if len(context.Services) > 0 {
		prompt.WriteString(fmt.Sprintf("- Services: %s\n", strings.Join(context.Services, ", ")))
	}
	if len(context.Packages) > 0 {
		prompt.WriteString(fmt.Sprintf("- Custom Packages: %s\n", strings.Join(context.Packages, ", ")))
	}
	if context.HomeManager {
		prompt.WriteString("- Home Manager: Yes\n")
	}
	if context.FlakeUsage {
		prompt.WriteString("- Using Flakes: Yes\n")
	}
	if context.BackupStrategy != "" {
		prompt.WriteString(fmt.Sprintf("- Backup Strategy: %s\n", context.BackupStrategy))
	}
	if context.RollbackPlan != "" {
		prompt.WriteString(fmt.Sprintf("- Rollback Plan: %s\n", context.RollbackPlan))
	}

	prompt.WriteString("\nMigration Question:\n")
	prompt.WriteString(question)

	return prompt.String()
}

// formatMigrationResponse formats the AI response for migration guidance.
func (a *MigrateAgent) formatMigrationResponse(response string) string {
	// Add migration-specific formatting and guidance
	var formatted strings.Builder

	formatted.WriteString("🔄 Migration Guidance:\n\n")
	formatted.WriteString(response)

	// Add common migration reminders
	formatted.WriteString("\n\n📋 Migration Checklist Reminders:")
	formatted.WriteString("\n• Always backup your current configuration")
	formatted.WriteString("\n• Test the migration on a non-production system first")
	formatted.WriteString("\n• Have a rollback plan ready")
	formatted.WriteString("\n• Update your hardware-configuration.nix if moving machines")
	formatted.WriteString("\n• Verify all services restart correctly after migration")
	formatted.WriteString("\n• Check that all user data and home configurations are preserved")

	return formatted.String()
}

// getMigrationContextFromData extracts migration context from stored data.
func (a *MigrateAgent) getMigrationContextFromData() *MigrationContext {
	if ctx, ok := a.contextData.(*MigrationContext); ok {
		return ctx
	}
	return &MigrationContext{}
}

// AnalyzeMigrationPath analyzes the migration requirements and suggests a path.
func (a *MigrateAgent) AnalyzeMigrationPath(ctx context.Context, sourceSystem, targetSystem string) (string, error) {
	migrationCtx := &MigrationContext{
		SourceSystem:  sourceSystem,
		TargetSystem:  targetSystem,
		MigrationType: "version",
	}

	a.SetMigrationContext(migrationCtx)

	question := fmt.Sprintf("Analyze the migration path from %s to %s. What are the key steps, potential issues, and best practices for this migration?", sourceSystem, targetSystem)

	return a.GenerateResponse(ctx, question)
}

// GenerateMigrationPlan creates a detailed migration plan.
func (a *MigrateAgent) GenerateMigrationPlan(ctx context.Context, migrationCtx *MigrationContext) (string, error) {
	a.SetMigrationContext(migrationCtx)

	var request strings.Builder
	request.WriteString("Generate a comprehensive migration plan including:")
	request.WriteString("\n1. Pre-migration preparation steps")
	request.WriteString("\n2. Backup procedures")
	request.WriteString("\n3. Step-by-step migration process")
	request.WriteString("\n4. Post-migration verification")
	request.WriteString("\n5. Rollback procedures if needed")
	request.WriteString("\n6. Common issues and troubleshooting")

	return a.GenerateResponse(ctx, request.String())
}

// DiagnoseMigrationIssues helps troubleshoot migration problems.
func (a *MigrateAgent) DiagnoseMigrationIssues(ctx context.Context, issues []string, context *MigrationContext) (string, error) {
	context.Issues = issues
	a.SetMigrationContext(context)

	question := fmt.Sprintf("Help diagnose and resolve these migration issues: %s", strings.Join(issues, ", "))

	return a.GenerateResponse(ctx, question)
}
