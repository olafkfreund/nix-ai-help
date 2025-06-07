package function

import (
	"fmt"
	"sync"

	"nix-ai-help/internal/ai/function/ask"
	"nix-ai-help/internal/ai/function/build"
	"nix-ai-help/internal/ai/function/community"
	"nix-ai-help/internal/ai/function/completion"
	configfunction "nix-ai-help/internal/ai/function/config"
	"nix-ai-help/internal/ai/function/configure"
	"nix-ai-help/internal/ai/function/devenv"
	"nix-ai-help/internal/ai/function/diagnose"
	"nix-ai-help/internal/ai/function/doctor"
	explainHomeoption "nix-ai-help/internal/ai/function/explain-home-option"
	explainoption "nix-ai-help/internal/ai/function/explain-option"
	"nix-ai-help/internal/ai/function/flakes"
	"nix-ai-help/internal/ai/function/gc"
	"nix-ai-help/internal/ai/function/hardware"
	"nix-ai-help/internal/ai/function/help"
	"nix-ai-help/internal/ai/function/interactive"
	"nix-ai-help/internal/ai/function/learning"
	"nix-ai-help/internal/ai/function/logs"
	"nix-ai-help/internal/ai/function/machines"
	mcpserver "nix-ai-help/internal/ai/function/mcp-server"
	"nix-ai-help/internal/ai/function/migrate"
	"nix-ai-help/internal/ai/function/neovim"
	packagerepo "nix-ai-help/internal/ai/function/package-repo"
	"nix-ai-help/internal/ai/function/packages"
	"nix-ai-help/internal/ai/function/search"
	"nix-ai-help/internal/ai/function/snippets"
	"nix-ai-help/internal/ai/function/store"
	"nix-ai-help/internal/ai/function/templates"
	"nix-ai-help/internal/ai/functionbase"
	"nix-ai-help/pkg/logger"
)

var (
	globalRegistry *FunctionManager
	registryOnce   sync.Once
)

// GetGlobalRegistry returns the global function registry with all functions registered
func GetGlobalRegistry() *FunctionManager {
	registryOnce.Do(func() {
		globalRegistry = NewFunctionManager()
		registerAllFunctions()
	})
	return globalRegistry
}

// registerAllFunctions registers all available AI functions
func registerAllFunctions() {
	log := logger.NewLogger()

	// Register all implemented functions
	functions := []struct {
		name string
		fn   functionbase.FunctionInterface
	}{
		{"ask", ask.NewAskFunction()},
		{"build", build.NewBuildFunction()},
		{"community", community.NewCommunityFunction()},
		{"completion", completion.NewCompletionFunction()},
		{"config", configfunction.NewConfigFunction()},
		{"configure", configure.NewConfigureFunction()},
		{"devenv", devenv.NewDevenvFunction()},
		{"diagnose", diagnose.NewDiagnoseFunction()},
		{"doctor", doctor.NewDoctorFunction()},
		{"explain-home-option", explainHomeoption.NewExplainHomeOptionFunction()},
		{"explain-option", explainoption.NewExplainOptionFunction()},
		{"flakes", flakes.NewFlakesFunction()},
		{"gc", gc.NewGcFunction()},
		{"hardware", hardware.NewHardwareFunction()},
		{"help", help.NewHelpFunction()},
		{"interactive", interactive.NewInteractiveFunction()},
		{"learning", learning.NewLearningFunction()},
		{"logs", logs.NewLogsFunction()},
		{"machines", machines.NewMachinesFunction()},
		{"mcp-server", mcpserver.NewMcpServerFunction()},
		{"migrate", migrate.NewMigrateFunction()},
		{"neovim", neovim.NewNeovimFunction()},
		{"packages", packages.NewPackagesFunction()},
		{"package-repo", packagerepo.NewPackageRepoFunction()},
		{"search", search.NewSearchFunction()},
		{"snippets", snippets.NewSnippetsFunction()},
		{"store", store.NewStoreFunction()},
		{"templates", templates.NewTemplatesFunction()},
	}

	successCount := 0
	for _, f := range functions {
		if err := globalRegistry.Register(f.fn); err != nil {
			log.Error(fmt.Sprintf("Failed to register function %s: %v", f.name, err))
		} else {
			log.Info(fmt.Sprintf("Registered function successfully: %s", f.name))
			successCount++
		}
	}

	log.Info(fmt.Sprintf("Function registry initialized: %d/%d functions registered", successCount, len(functions)))
}

// ListAvailableFunctions returns a map of function names to their descriptions
func ListAvailableFunctions() map[string]string {
	registry := GetGlobalRegistry()
	functions := make(map[string]string)

	for _, name := range registry.List() {
		if fn, exists := registry.Get(name); exists {
			functions[name] = fn.Description()
		}
	}

	return functions
}

// GetFunctionSchema returns the schema for a specific function
func GetFunctionSchema(name string) (FunctionSchema, error) {
	registry := GetGlobalRegistry()
	return registry.GetSchema(name)
}

// GetAllFunctionSchemas returns schemas for all registered functions
func GetAllFunctionSchemas() map[string]FunctionSchema {
	registry := GetGlobalRegistry()
	return registry.GetSchemas()
}

// ExecuteFunction is a convenience function to execute a function by name
func ExecuteFunction(call FunctionCall, options *FunctionOptions) (*FunctionResult, error) {
	registry := GetGlobalRegistry()
	return registry.Execute(call.Context, call, options)
}

// ValidateFunction validates a function call without executing it
func ValidateFunction(call FunctionCall) error {
	registry := GetGlobalRegistry()
	return registry.ValidateCall(call)
}

// FunctionExists checks if a function is registered
func FunctionExists(name string) bool {
	registry := GetGlobalRegistry()
	return registry.HasFunction(name)
}
