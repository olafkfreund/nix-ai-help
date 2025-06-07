# Function System Documentation

This document provides comprehensive documentation for the nixai function calling system, which powers AI-driven NixOS assistance through structured, validated tool execution.

## Table of Contents

- [System Overview](#system-overview)
- [Architecture](#architecture)
- [Function Catalog](#function-catalog)
- [Function Manager API](#function-manager-api)
- [Parameter Schema System](#parameter-schema-system)
- [Execution Options](#execution-options)
- [Error Handling](#error-handling)
- [Integration with CLI](#integration-with-cli)
- [Best Practices](#best-practices)
- [Development Guide](#development-guide)

## System Overview

The nixai function system provides a robust framework for AI-powered tool execution with:

- **29 specialized functions** covering all aspects of NixOS management
- **JSON schema validation** for all function parameters
- **Async execution support** with progress reporting
- **Centralized registry** for function discovery and management
- **Comprehensive error handling** with detailed diagnostics
- **CLI integration** for both direct and AI-mediated execution

### Key Features

- **Type Safety**: All functions use strict JSON schema validation
- **Modularity**: Each function is self-contained with its own agent and logic
- **Consistency**: Standardized interface through BaseFunction
- **Performance**: Async execution with timeout and progress tracking
- **Extensibility**: Easy to add new functions following established patterns

## Architecture

The function system is built on several key components:

### Core Components

1. **FunctionInterface** - Base interface all functions must implement
2. **BaseFunction** - Common functionality and schema validation
3. **FunctionManager** - Registration, discovery, and execution management
4. **Global Registry** - Centralized function access point
5. **Parameter System** - Type-safe parameter definition and validation

### Class Hierarchy

```
FunctionInterface (interface)
└── BaseFunction (base implementation)
    └── Individual Functions (ask, build, community, etc.)
```

### Data Flow

```
CLI Command → Function Call → Validation → Execution → Result
                    ↓              ↓           ↓
              Parameter       Schema      Progress
              Parsing      Validation    Reporting
```

## Function Catalog

The nixai function system includes 29 specialized functions:

### Core System Functions

| Function | Description | Key Operations |
|----------|-------------|----------------|
| **ask** | Direct AI question answering | query, context, provider-selection |
| **help** | System help and documentation | commands, examples, usage |
| **doctor** | System health diagnostics | check, repair, optimize |
| **diagnose** | Issue analysis and troubleshooting | logs, errors, recommendations |

### Configuration Management

| Function | Description | Key Operations |
|----------|-------------|----------------|
| **config** | NixOS configuration generation and management | generate, validate, migrate, optimize |
| **configure** | Interactive configuration modification | get, set, add, remove, validate |
| **migrate** | Configuration migration and upgrades | backup, convert, preview, rollback |
| **flakes** | Flake management and operations | init, update, show, lock |

### Package Management

| Function | Description | Key Operations |
|----------|-------------|----------------|
| **packages** | Package search and installation guidance | search, install, configure, dependencies |
| **package-repo** | Repository analysis and derivation generation | analyze, generate, validate, test |
| **store** | Nix store management and optimization | query, usage, optimize, gc |
| **gc** | Garbage collection and cleanup | collect, analyze, roots, schedule |

### Development Environment

| Function | Description | Key Operations |
|----------|-------------|----------------|
| **devenv** | Development environment setup | create, configure, languages, tools |
| **templates** | Project template management | list, create, init, customize |
| **build** | Build system operations | build, test, deploy, debug |
| **completion** | Shell completion generation | bash, zsh, fish, powershell |

### Hardware and System

| Function | Description | Key Operations |
|----------|-------------|----------------|
| **hardware** | Hardware detection and configuration | detect, configure, drivers, optimization |
| **machines** | Machine management and deployment | create, configure, security, performance |
| **logs** | System log analysis | parse, filter, analyze, diagnose |
| **snippets** | Configuration snippet management | search, create, share, apply |

### Learning and Community

| Function | Description | Key Operations |
|----------|-------------|----------------|
| **learning** | Educational content and tutorials | tutorials, examples, concepts, practice |
| **community** | Community resource access | forum, docs, issues, tutorials |
| **search** | Universal search across NixOS resources | packages, options, docs, community |

### Documentation and Options

| Function | Description | Key Operations |
|----------|-------------|----------------|
| **explain-option** | NixOS option explanation | describe, examples, usage, related |
| **explain-home-option** | Home Manager option explanation | describe, examples, configuration, integration |

### Advanced Features

| Function | Description | Key Operations |
|----------|-------------|----------------|
| **interactive** | Interactive shell mode | start, execute, history, settings |
| **neovim** | Neovim configuration management | setup, plugins, lsp, themes |
| **mcp-server** | MCP server management | setup, configure, monitor, troubleshoot |

## Function Manager API

The FunctionManager provides the central interface for function operations:

### Core Methods

```go
// Registration
func (fm *FunctionManager) Register(fn FunctionInterface) error
func (fm *FunctionManager) Unregister(name string) error

// Discovery
func (fm *FunctionManager) Get(name string) (FunctionInterface, bool)
func (fm *FunctionManager) List() []string
func (fm *FunctionManager) GetSchema(name string) (FunctionSchema, error)

// Execution
func (fm *FunctionManager) Execute(ctx context.Context, call FunctionCall, options *FunctionOptions) (*FunctionResult, error)
func (fm *FunctionManager) ExecuteWithProgress(ctx context.Context, call FunctionCall, options *FunctionOptions, progressChan chan<- Progress) (*FunctionResult, error)

// Validation
func (fm *FunctionManager) ValidateCall(call FunctionCall) error
func (fm *FunctionManager) HasFunction(name string) bool
```

### Usage Examples

```go
// Get the global function manager
fm := function.GetGlobalRegistry()

// List all available functions
functions := fm.List()

// Get a specific function
fn, exists := fm.Get("packages")

// Create and execute a function call
call := function.CreateCall("packages", map[string]interface{}{
    "query": "git",
    "operation": "search",
})

result, err := fm.Execute(context.Background(), call, nil)
```

## Parameter Schema System

Functions use a robust parameter schema system for validation:

### Parameter Types

- **string** - Text values with optional enum, pattern, length constraints
- **integer** - Numeric values with optional min/max bounds
- **boolean** - True/false values
- **array** - Lists of values with optional item type constraints
- **object** - Complex nested structures

### Schema Definition

```go
// Example parameter definition
parameters := []functionbase.FunctionParameter{
    functionbase.StringParam("query", "Search query", true),
    functionbase.StringParamWithEnum("operation", "Operation type", true, 
        []string{"search", "install", "remove"}),
    functionbase.IntParam("limit", "Maximum results", false, 10),
    functionbase.BoolParam("detailed", "Show detailed information", false),
    functionbase.ArrayParam("tags", "Filter tags", false),
    functionbase.ObjectParam("options", "Additional options", false),
}
```

### Validation Features

- **Required parameter checking** - Ensures all required parameters are provided
- **Type validation** - Verifies parameter types match schema
- **Enum validation** - Restricts string values to predefined options
- **Range validation** - Enforces numeric min/max bounds
- **Pattern validation** - Validates strings against regex patterns
- **Length validation** - Enforces string length constraints

## Execution Options

Functions support various execution options:

### FunctionOptions Structure

```go
type FunctionOptions struct {
    Timeout         time.Duration     // Execution timeout
    Async           bool              // Asynchronous execution
    ProgressCallback ProgressCallback // Progress reporting
    Metadata        map[string]interface{} // Additional context
}
```

### Execution Modes

1. **Synchronous** - Default mode, blocks until completion
2. **Asynchronous** - Non-blocking execution with progress callbacks
3. **With Timeout** - Automatic cancellation after specified duration
4. **With Progress** - Real-time progress reporting for long operations

### Usage Examples

```go
// Synchronous execution with timeout
options := &FunctionOptions{
    Timeout: 30 * time.Second,
}

// Asynchronous execution with progress
progressChan := make(chan Progress, 10)
options := &FunctionOptions{
    Async: true,
    ProgressCallback: func(progress Progress) {
        fmt.Printf("Progress: %d%%\n", progress.Percentage)
    },
}

result, err := fm.ExecuteWithProgress(ctx, call, options, progressChan)
```

## Error Handling

The function system provides comprehensive error handling:

### Error Types

1. **Validation Errors** - Parameter schema violations
2. **Execution Errors** - Runtime failures during function execution
3. **Timeout Errors** - Operations exceeding specified timeouts
4. **Not Found Errors** - Requests for non-existent functions

### Error Response Format

```go
type FunctionResult struct {
    Success   bool                   `json:"success"`
    Data      map[string]interface{} `json:"data,omitempty"`
    Error     string                 `json:"error,omitempty"`
    Metadata  map[string]interface{} `json:"metadata,omitempty"`
    Duration  time.Duration          `json:"duration"`
    Timestamp time.Time              `json:"timestamp"`
}
```

### Error Handling Best Practices

- Always check function existence before execution
- Validate parameters before calling Execute
- Handle timeouts gracefully in long-running operations
- Use structured error responses for client applications
- Log errors with sufficient context for debugging

## Integration with CLI

Functions integrate seamlessly with the nixai CLI system:

### Direct Function Execution

```bash
# Direct function calls
nixai packages --query "git" --operation "search"
nixai config --operation "generate" --target "desktop"
nixai doctor --check "system"
```

### AI-Mediated Execution

```bash
# AI determines appropriate functions
nixai "I need help setting up a development environment for Rust"
nixai "How do I configure my graphics drivers?"
nixai "Find packages related to web development"
```

### Function Discovery

```bash
# List all available functions
nixai help functions

# Get detailed function information
nixai help function packages
nixai help function config
```

## Best Practices

### For Users

1. **Use descriptive queries** - Provide specific details about your needs
2. **Leverage function discovery** - Use `nixai help functions` to explore capabilities
3. **Validate complex operations** - Use dry-run modes when available
4. **Check function documentation** - Review parameter schemas before direct calls

### For Developers

1. **Follow naming conventions** - Use clear, descriptive function names
2. **Implement comprehensive validation** - Use all relevant parameter constraints
3. **Provide helpful examples** - Include usage examples in function schemas
4. **Handle errors gracefully** - Return structured error responses
5. **Support progress reporting** - Implement progress callbacks for long operations
6. **Write comprehensive tests** - Cover all parameter combinations and edge cases

## Development Guide

### Creating New Functions

1. **Define the function structure**:
```go
type MyFunction struct {
    *functionbase.BaseFunction
    agent  *agent.MyAgent
    logger *logger.Logger
}
```

2. **Implement the required interface**:
```go
func (f *MyFunction) Execute(ctx context.Context, params map[string]interface{}, options *FunctionOptions) (*FunctionResult, error) {
    // Implementation
}
```

3. **Define parameter schema**:
```go
parameters := []functionbase.FunctionParameter{
    functionbase.StringParam("operation", "Operation to perform", true),
    // Additional parameters...
}
```

4. **Register the function**:
```go
func NewMyFunction() *MyFunction {
    baseFunc := functionbase.NewBaseFunction(
        "my-function",
        "Description of what this function does",
        parameters,
    )
    
    return &MyFunction{
        BaseFunction: baseFunc,
        agent:        agent.NewMyAgent(),
        logger:       logger.NewLogger(),
    }
}
```

5. **Add to registry** in `internal/ai/function/registry.go`:
```go
{"my-function", myfunction.NewMyFunction()},
```

### Testing Functions

1. **Unit tests** - Test individual function logic
2. **Schema validation tests** - Verify parameter validation
3. **Integration tests** - Test function manager integration
4. **CLI integration tests** - Verify command-line interface

### Function Examples

Each function should include comprehensive examples:

```go
Examples: []functionbase.FunctionExample{
    {
        Description: "Search for Git packages",
        Parameters: map[string]interface{}{
            "query": "git",
            "operation": "search",
            "limit": 10,
        },
        Expected: "Returns a list of Git-related packages",
    },
}
```

## Function Reference

For detailed information about specific functions, their parameters, and usage examples, see:

- Individual function implementations in `internal/ai/function/*/`
- CLI help: `nixai help function <function-name>`
- Function schemas: Available through the FunctionManager API
- Test files: `*_test.go` files for real-world usage examples

## Troubleshooting

### Common Issues

1. **Function not found** - Ensure function is registered in registry.go
2. **Parameter validation fails** - Check parameter types and constraints
3. **Timeout errors** - Increase timeout or optimize function performance
4. **Agent initialization fails** - Verify AI provider configuration

### Debug Mode

Enable debug logging to trace function execution:

```bash
export LOG_LEVEL=debug
nixai packages --query "git"
```

### Validation Testing

Test function calls before execution:

```go
err := fm.ValidateCall(call)
if err != nil {
    log.Printf("Validation failed: %v", err)
}
```

---

The nixai function system provides a powerful, type-safe framework for AI-driven NixOS operations. By following the patterns and best practices outlined in this documentation, users and developers can effectively leverage and extend the system's capabilities.
