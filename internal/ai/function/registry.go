package function

import (
	"fmt"
	"sync"

	"nix-ai-help/internal/ai/function/ask"
	"nix-ai-help/internal/ai/function/community"
	"nix-ai-help/internal/ai/function/diagnose"
	explainHomeoption "nix-ai-help/internal/ai/function/explain-home-option"
	explainoption "nix-ai-help/internal/ai/function/explain-option"
	"nix-ai-help/internal/ai/function/flakes"
	"nix-ai-help/internal/ai/function/learning"
	mcpserver "nix-ai-help/internal/ai/function/mcp-server"
	packagerepo "nix-ai-help/internal/ai/function/package-repo"
	"nix-ai-help/internal/ai/function/packages"
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
	logger := logger.NewLogger()

	// Register all implemented functions
	functions := []struct {
		name string
		fn   functionbase.FunctionInterface
	}{
		{"ask", ask.NewAskFunction()},
		{"community", community.NewCommunityFunction()},
		{"diagnose", diagnose.NewDiagnoseFunction()},
		{"explain-home-option", explainHomeoption.NewExplainHomeOptionFunction()},
		{"explain-option", explainoption.NewExplainOptionFunction()},
		{"flakes", flakes.NewFlakesFunction()},
		{"learning", learning.NewLearningFunction()},
		{"mcp-server", mcpserver.NewMcpServerFunction()},
		{"packages", packages.NewPackagesFunction()},
		{"package-repo", packagerepo.NewPackageRepoFunction()},
	}

	successCount := 0
	for _, f := range functions {
		if err := globalRegistry.Register(f.fn); err != nil {
			logger.Error(fmt.Sprintf("Failed to register %s function: %v", f.name, err))
		} else {
			logger.Info(fmt.Sprintf("Registered %s function successfully", f.name))
			successCount++
		}
	}

	logger.Info(fmt.Sprintf("Function registry initialized with %d/%d functions", successCount, len(functions)))
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
