package function

import (
	"context"
	"fmt"
	"sync"
	"time"

	"nix-ai-help/internal/ai/functionbase"
	"nix-ai-help/pkg/logger"
)

// FunctionManager manages registration and execution of AI functions
type FunctionManager struct {
	functions map[string]functionbase.FunctionInterface
	mutex     sync.RWMutex
	logger    *logger.Logger
}

// NewFunctionManager creates a new function manager
func NewFunctionManager() *FunctionManager {
	return &FunctionManager{
		functions: make(map[string]functionbase.FunctionInterface),
		logger:    logger.NewLogger(),
	}
}

// Register adds a function to the manager
func (fm *FunctionManager) Register(fn functionbase.FunctionInterface) error {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	name := fn.Name()
	if name == "" {
		return fmt.Errorf("function name cannot be empty")
	}

	if _, exists := fm.functions[name]; exists {
		return fmt.Errorf("function '%s' already registered", name)
	}

	fm.functions[name] = fn
	fm.logger.Info(fmt.Sprintf("Registered function: %s", name))
	return nil
}

// Unregister removes a function from the manager
func (fm *FunctionManager) Unregister(name string) error {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	if _, exists := fm.functions[name]; !exists {
		return fmt.Errorf("function '%s' not found", name)
	}

	delete(fm.functions, name)
	fm.logger.Info(fmt.Sprintf("Unregistered function: %s", name))
	return nil
}

// Get retrieves a function by name
func (fm *FunctionManager) Get(name string) (functionbase.FunctionInterface, bool) {
	fm.mutex.RLock()
	defer fm.mutex.RUnlock()

	fn, exists := fm.functions[name]
	return fn, exists
}

// List returns all registered function names
func (fm *FunctionManager) List() []string {
	fm.mutex.RLock()
	defer fm.mutex.RUnlock()

	names := make([]string, 0, len(fm.functions))
	for name := range fm.functions {
		names = append(names, name)
	}
	return names
}

// GetSchemas returns schemas for all registered functions
func (fm *FunctionManager) GetSchemas() map[string]FunctionSchema {
	fm.mutex.RLock()
	defer fm.mutex.RUnlock()

	schemas := make(map[string]FunctionSchema)
	for name, fn := range fm.functions {
		schemas[name] = fn.Schema()
	}
	return schemas
}

// GetSchema returns the schema for a specific function
func (fm *FunctionManager) GetSchema(name string) (FunctionSchema, error) {
	fm.mutex.RLock()
	defer fm.mutex.RUnlock()

	fn, exists := fm.functions[name]
	if !exists {
		return FunctionSchema{}, fmt.Errorf("function '%s' not found", name)
	}

	return fn.Schema(), nil
}

// Execute runs a function with the given parameters
func (fm *FunctionManager) Execute(ctx context.Context, call FunctionCall, options *FunctionOptions) (*FunctionResult, error) {
	startTime := time.Now()

	// Get the function
	fn, exists := fm.Get(call.Name)
	if !exists {
		return functionbase.CreateErrorResult(
			fmt.Errorf("function '%s' not found", call.Name),
			"functionbase.FunctionInterface not found",
		), nil
	}

	// Set default options if not provided
	if options == nil {
		options = &FunctionOptions{
			Timeout: 30 * time.Second,
			Async:   false,
		}
	}

	// Validate parameters
	if err := fn.ValidateParameters(call.Parameters); err != nil {
		return functionbase.CreateErrorResult(err, "Parameter validation failed"), nil
	}

	// Create execution context with timeout
	execCtx := ctx
	if options.Timeout > 0 {
		var cancel context.CancelFunc
		execCtx, cancel = context.WithTimeout(ctx, options.Timeout)
		defer cancel()
	}

	fm.logger.Debug(fmt.Sprintf("Executing function: %s with parameters: %v", call.Name, call.Parameters))

	// Execute the function
	result, err := fn.Execute(execCtx, call.Parameters, options)
	if err != nil {
		fm.logger.Error(fmt.Sprintf("functionbase.FunctionInterface execution failed: %v", err))
		return functionbase.CreateErrorResult(err, "functionbase.FunctionInterface execution failed"), err
	}

	// Set execution duration
	if result != nil {
		result.Duration = time.Since(startTime)
	}

	fm.logger.Debug(fmt.Sprintf("functionbase.FunctionInterface %s executed successfully in %v", call.Name, result.Duration))
	return result, nil
}

// ExecuteWithProgress runs a function with progress reporting
func (fm *FunctionManager) ExecuteWithProgress(ctx context.Context, call FunctionCall, options *FunctionOptions, progressChan chan<- Progress) (*FunctionResult, error) {
	if options == nil {
		options = &FunctionOptions{}
	}

	// Set up progress callback
	if progressChan != nil {
		options.ProgressCallback = func(progress Progress) {
			select {
			case progressChan <- progress:
			case <-ctx.Done():
				return
			}
		}
	}

	return fm.Execute(ctx, call, options)
}

// ValidateCall validates a function call without executing it
func (fm *FunctionManager) ValidateCall(call FunctionCall) error {
	fn, exists := fm.Get(call.Name)
	if !exists {
		return fmt.Errorf("function '%s' not found", call.Name)
	}

	return fn.ValidateParameters(call.Parameters)
}

// HasFunction checks if a function is registered
func (fm *FunctionManager) HasFunction(name string) bool {
	_, exists := fm.Get(name)
	return exists
}

// Count returns the number of registered functions
func (fm *FunctionManager) Count() int {
	fm.mutex.RLock()
	defer fm.mutex.RUnlock()
	return len(fm.functions)
}

// Clear removes all registered functions
func (fm *FunctionManager) Clear() {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	count := len(fm.functions)
	fm.functions = make(map[string]functionbase.FunctionInterface)
	fm.logger.Info(fmt.Sprintf("Cleared %d functions from registry", count))
}

// GetFunctionInfo returns detailed information about a function
func (fm *FunctionManager) GetFunctionInfo(name string) (map[string]interface{}, error) {
	fn, exists := fm.Get(name)
	if !exists {
		return nil, fmt.Errorf("function '%s' not found", name)
	}

	schema := fn.Schema()
	info := map[string]interface{}{
		"name":        fn.Name(),
		"description": fn.Description(),
		"parameters":  schema.Parameters,
		"examples":    schema.Examples,
	}

	return info, nil
}

// RegisterMultiple registers multiple functions at once
func (fm *FunctionManager) RegisterMultiple(functions []functionbase.FunctionInterface) []error {
	var errors []error

	for _, fn := range functions {
		if err := fm.Register(fn); err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

// CreateCall is a helper to create a FunctionCall
func CreateCall(name string, params map[string]interface{}) FunctionCall {
	return FunctionCall{
		Name:       name,
		Parameters: params,
		Timestamp:  time.Now(),
	}
}

// CreateCallWithContext is a helper to create a FunctionCall with context
func CreateCallWithContext(ctx context.Context, name string, params map[string]interface{}) FunctionCall {
	return FunctionCall{
		Name:       name,
		Parameters: params,
		Context:    ctx,
		Timestamp:  time.Now(),
	}
}

// Default global function manager
var defaultManager *FunctionManager
var once sync.Once

// GetDefaultManager returns the global function manager instance
func GetDefaultManager() *FunctionManager {
	once.Do(func() {
		defaultManager = NewFunctionManager()
	})
	return defaultManager
}

// Convenience functions for the global manager

// RegisterGlobal registers a function in the global manager
func RegisterGlobal(fn functionbase.FunctionInterface) error {
	return GetDefaultManager().Register(fn)
}

// ExecuteGlobal executes a function using the global manager
func ExecuteGlobal(ctx context.Context, call FunctionCall, options *FunctionOptions) (*FunctionResult, error) {
	return GetDefaultManager().Execute(ctx, call, options)
}

// GetGlobal retrieves a function from the global manager
func GetGlobal(name string) (functionbase.FunctionInterface, bool) {
	return GetDefaultManager().Get(name)
}

// ListGlobal returns all functions from the global manager
func ListGlobal() []string {
	return GetDefaultManager().List()
}
