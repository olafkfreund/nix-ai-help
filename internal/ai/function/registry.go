package function

import (
	"fmt"
	"sync"

	"nix-ai-help/internal/ai/function/diagnose"
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

	// Register diagnose function
	diagnoseFunc := diagnose.NewDiagnoseFunction()
	if err := globalRegistry.Register(diagnoseFunc); err != nil {
		logger.Error(fmt.Sprintf("Failed to register diagnose function: %v", err))
	} else {
		logger.Info("Registered diagnose function successfully")
	}

	// TODO: Add more functions here as they are implemented
	// buildFunc := build.NewBuildFunction()
	// globalRegistry.Register(buildFunc)

	// communityFunc := community.NewCommunityFunction()
	// globalRegistry.Register(communityFunc)

	logger.Info(fmt.Sprintf("Function registry initialized with %d functions", globalRegistry.Count()))
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
