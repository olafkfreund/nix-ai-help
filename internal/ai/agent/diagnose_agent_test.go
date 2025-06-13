package agent

import (
	"context"
	"testing"

	"nix-ai-help/internal/ai/roles"
	"nix-ai-help/internal/nixos"

	"github.com/stretchr/testify/require"
)

func TestNewDiagnoseAgent(t *testing.T) {
	agent := NewDiagnoseAgent()
	require.NotNil(t, agent)
	require.Equal(t, string(roles.RoleDiagnose), agent.role)
	require.NotNil(t, agent.logger)
}

func TestDiagnoseAgent_Query(t *testing.T) {
	agent := NewDiagnoseAgent()
	ctx := context.Background()

	t.Run("valid role", func(t *testing.T) {
		result, err := agent.Query(ctx, "test input", string(roles.RoleDiagnose), nil)
		require.NoError(t, err)
		require.Contains(t, result, "specialized NixOS diagnostic assistant")
		require.Contains(t, result, "test input")
	})

	t.Run("invalid role", func(t *testing.T) {
		_, err := agent.Query(ctx, "test input", "invalid_role", nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "unsupported role")
	})

	t.Run("missing prompt template", func(t *testing.T) {
		// This should not happen with current code, but test the error path
		_, err := agent.Query(ctx, "test input", string(roles.RoleExplainer), nil)
		require.NoError(t, err) // Should work since RoleExplainer exists
	})
}

func TestDiagnoseAgent_QueryWithContext(t *testing.T) {
	agent := NewDiagnoseAgent()
	ctx := context.Background()

	t.Run("with diagnostic context", func(t *testing.T) {
		diagCtx := &DiagnosticContext{
			LogData:         "ERROR: service failed to start",
			ConfigSnippet:   "services.nginx.enable = true;",
			ErrorMessage:    "bind: address already in use",
			UserDescription: "nginx won't start after reboot",
			SystemInfo: &SystemInfo{
				NixOSVersion:  "23.11",
				NixVersion:    "2.18.1",
				Channel:       "nixos-23.11",
				Architecture:  "x86_64-linux",
				IsFlakeSystem: false,
			},
		}

		result, err := agent.Query(ctx, "Why is nginx failing?", string(roles.RoleDiagnose), diagCtx)
		require.NoError(t, err)

		// Check that context is properly included
		require.Contains(t, result, "DIAGNOSTIC CONTEXT")
		require.Contains(t, result, "System Information")
		require.Contains(t, result, "NixOS Version: 23.11")
		require.Contains(t, result, "Log Output")
		require.Contains(t, result, "service failed to start")
		require.Contains(t, result, "Configuration Snippet")
		require.Contains(t, result, "services.nginx.enable = true")
		require.Contains(t, result, "Error Message")
		require.Contains(t, result, "address already in use")
		require.Contains(t, result, "User Description")
		require.Contains(t, result, "nginx won't start")
		require.Contains(t, result, "Why is nginx failing?")
		require.Contains(t, result, "INSTRUCTIONS")
	})

	t.Run("with existing diagnostics", func(t *testing.T) {
		existingDiags := []nixos.Diagnostic{
			{
				Issue:     "Service failed to start",
				ErrorType: "service",
				Severity:  "high",
				Details:   "nginx.service failed with exit code 1",
			},
			{
				Issue:     "Port already in use",
				ErrorType: "network",
				Severity:  "medium",
				Details:   "Another process is using port 80",
			},
		}

		diagCtx := &DiagnosticContext{
			ExistingDiagnostics: existingDiags,
		}

		result, err := agent.Query(ctx, "Help me fix this", string(roles.RoleDiagnose), diagCtx)
		require.NoError(t, err)

		require.Contains(t, result, "Automated Analysis Results")
		require.Contains(t, result, "Service failed to start")
		require.Contains(t, result, "Port already in use")
		require.Contains(t, result, "Severity: high")
		require.Contains(t, result, "Type: service")
	})

	t.Run("minimal context", func(t *testing.T) {
		diagCtx := &DiagnosticContext{
			ErrorMessage: "build failed",
		}

		result, err := agent.Query(ctx, "What went wrong?", string(roles.RoleDiagnose), diagCtx)
		require.NoError(t, err)

		require.Contains(t, result, "Error Message")
		require.Contains(t, result, "build failed")
		require.Contains(t, result, "What went wrong?")
	})
}

func TestDiagnoseAgent_GenerateResponse(t *testing.T) {
	agent := NewDiagnoseAgent()
	ctx := context.Background()

	result, err := agent.GenerateResponse(ctx, "test", string(roles.RoleDiagnose), nil)
	require.NoError(t, err)
	require.Contains(t, result, "test")
}

func TestDiagnoseAgent_SetRole(t *testing.T) {
	agent := NewDiagnoseAgent()
	agent.SetRole("new_role")
	require.Equal(t, "new_role", agent.role)
}

func TestDiagnoseAgent_SetContext(t *testing.T) {
	agent := NewDiagnoseAgent()
	context := "test context"
	agent.SetContext(context)
	require.Equal(t, context, agent.contextData)
}

func TestBuildDiagnosticContext(t *testing.T) {
	diagnostics := []nixos.Diagnostic{
		{Issue: "test issue", ErrorType: "test", Severity: "low"},
	}

	ctx := BuildDiagnosticContext(
		"log data",
		"config snippet",
		"error message",
		"user description",
		diagnostics,
	)

	require.Equal(t, "log data", ctx.LogData)
	require.Equal(t, "config snippet", ctx.ConfigSnippet)
	require.Equal(t, "error message", ctx.ErrorMessage)
	require.Equal(t, "user description", ctx.UserDescription)
	require.Len(t, ctx.ExistingDiagnostics, 1)
	require.Equal(t, "test issue", ctx.ExistingDiagnostics[0].Issue)
}

func TestDiagnosticContext_AddSystemInfo(t *testing.T) {
	ctx := &DiagnosticContext{}
	sysInfo := &SystemInfo{
		NixOSVersion:  "23.11",
		IsFlakeSystem: true,
	}

	ctx.AddSystemInfo(sysInfo)
	require.Equal(t, sysInfo, ctx.SystemInfo)
	require.Equal(t, "23.11", ctx.SystemInfo.NixOSVersion)
	require.True(t, ctx.SystemInfo.IsFlakeSystem)
}

func TestDiagnosticContext_AddCommandOutput(t *testing.T) {
	ctx := &DiagnosticContext{}
	output := "command output here"

	ctx.AddCommandOutput(output)
	require.Equal(t, output, ctx.CommandOutput)
}

func TestDiagnoseAgent_buildDiagnosticPrompt(t *testing.T) {
	agent := NewDiagnoseAgent()

	t.Run("with full context", func(t *testing.T) {
		basePrompt := "Base diagnostic prompt"
		input := "User input question"

		diagCtx := &DiagnosticContext{
			LogData:         "log content",
			ConfigSnippet:   "nix config",
			ErrorMessage:    "error occurred",
			UserDescription: "user describes problem",
			CommandOutput:   "command result",
			SystemInfo: &SystemInfo{
				NixOSVersion:  "23.11",
				IsFlakeSystem: true,
			},
			ExistingDiagnostics: []nixos.Diagnostic{
				{Issue: "test issue", ErrorType: "test", Severity: "medium"},
			},
		}

		result := agent.buildDiagnosticPrompt(basePrompt, input, diagCtx)

		// Check all sections are included
		require.Contains(t, result, basePrompt)
		require.Contains(t, result, "DIAGNOSTIC CONTEXT")
		require.Contains(t, result, "System Information")
		require.Contains(t, result, "Error Message")
		require.Contains(t, result, "Log Output")
		require.Contains(t, result, "Configuration Snippet")
		require.Contains(t, result, "Command Output")
		require.Contains(t, result, "User Description")
		require.Contains(t, result, "Automated Analysis Results")
		require.Contains(t, result, "USER INPUT")
		require.Contains(t, result, "INSTRUCTIONS")

		// Check content is included
		require.Contains(t, result, "NixOS Version: 23.11")
		require.Contains(t, result, "Flake System: true")
		require.Contains(t, result, "log content")
		require.Contains(t, result, "nix config")
		require.Contains(t, result, "error occurred")
		require.Contains(t, result, "command result")
		require.Contains(t, result, "user describes problem")
		require.Contains(t, result, "test issue")
		require.Contains(t, result, input)
	})

	t.Run("with minimal context", func(t *testing.T) {
		basePrompt := "Base prompt"
		input := "Question"

		result := agent.buildDiagnosticPrompt(basePrompt, input, nil)

		require.Contains(t, result, basePrompt)
		require.Contains(t, result, "USER INPUT")
		require.Contains(t, result, input)
		require.Contains(t, result, "INSTRUCTIONS")
		require.NotContains(t, result, "DIAGNOSTIC CONTEXT")
	})
}
