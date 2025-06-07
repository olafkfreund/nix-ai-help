package function

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"nix-ai-help/pkg/logger"
	"nix-ai-help/pkg/utils"
)

// CLIIntegration provides CLI integration for function calling
type CLIIntegration struct {
	registry *FunctionManager
	logger   *logger.Logger
}

// NewCLIIntegration creates a new CLI integration
func NewCLIIntegration() *CLIIntegration {
	return &CLIIntegration{
		registry: GetGlobalRegistry(),
		logger:   logger.NewLogger(),
	}
}

// ExecuteFromCLI executes a function call from CLI parameters
func (cli *CLIIntegration) ExecuteFromCLI(functionName string, paramsJSON string, showProgress bool) error {
	cli.logger.Info(fmt.Sprintf("Executing function: %s", functionName))

	// Parse parameters
	var params map[string]interface{}
	if paramsJSON != "" {
		if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
			return fmt.Errorf("failed to parse parameters JSON: %v", err)
		}
	} else {
		params = make(map[string]interface{})
	}

	// Create function call
	call := CreateCallWithContext(context.Background(), functionName, params)

	// Set up options
	options := &FunctionOptions{
		Timeout: 5 * time.Minute,
	}

	// Set up progress reporting if requested
	var progressChan chan Progress
	if showProgress {
		progressChan = make(chan Progress, 10)
		options.ProgressCallback = func(progress Progress) {
			progressChan <- progress
		}

		// Start progress reporting goroutine
		go cli.reportProgress(progressChan)
	}

	// Execute the function
	result, err := cli.registry.Execute(context.Background(), call, options)

	// Close progress channel
	if progressChan != nil {
		close(progressChan)
	}

	if err != nil {
		return fmt.Errorf("function execution failed: %v", err)
	}

	// Display result
	cli.displayResult(result)
	return nil
}

// reportProgress displays progress updates
func (cli *CLIIntegration) reportProgress(progressChan <-chan Progress) {
	for progress := range progressChan {
		fmt.Printf("\r%s [%d/%d] %.1f%% - %s",
			progress.Stage,
			progress.Current,
			progress.Total,
			progress.Percentage,
			progress.Message)

		if progress.Percentage >= 100 {
			fmt.Println() // New line when complete
		}
	}
}

// displayResult formats and displays the function result
func (cli *CLIIntegration) displayResult(result *FunctionResult) {
	fmt.Println(utils.FormatDivider())
	fmt.Println(utils.FormatHeader("Function Execution Result"))
	fmt.Println(utils.FormatDivider())

	if result.Success {
		fmt.Println(utils.FormatKeyValue("Status", "✅ Success"))
		if result.Duration > 0 {
			fmt.Println(utils.FormatKeyValue("Duration", result.Duration.String()))
		}

		fmt.Println("\n" + utils.FormatHeader("Result Data"))
		cli.displayResultData(result.Data)
	} else {
		fmt.Println(utils.FormatKeyValue("Status", "❌ Failed"))
		if result.Error != "" {
			fmt.Println(utils.FormatKeyValue("Error", result.Error))
		}
	}

	fmt.Println(utils.FormatDivider())
}

// displayResultData formats and displays the result data based on its type
func (cli *CLIIntegration) displayResultData(data interface{}) {
	switch v := data.(type) {
	case map[string]interface{}:
		cli.displayMapData(v, 0)
	case []interface{}:
		cli.displayArrayData(v, 0)
	default:
		// Try to marshal as JSON for structured display
		if jsonData, err := json.MarshalIndent(data, "", "  "); err == nil {
			fmt.Println(string(jsonData))
		} else {
			fmt.Printf("%+v\n", data)
		}
	}
}

// displayMapData displays map data with proper formatting
func (cli *CLIIntegration) displayMapData(data map[string]interface{}, indent int) {
	indentStr := ""
	for i := 0; i < indent; i++ {
		indentStr += "  "
	}

	for key, value := range data {
		switch v := value.(type) {
		case string:
			if len(v) > 100 {
				// For long strings, show them in a more readable format
				fmt.Printf("%s%s:\n%s%s\n", indentStr, key, indentStr+"  ", v)
			} else {
				fmt.Printf("%s%s: %s\n", indentStr, key, v)
			}
		case []interface{}:
			fmt.Printf("%s%s:\n", indentStr, key)
			cli.displayArrayData(v, indent+1)
		case map[string]interface{}:
			fmt.Printf("%s%s:\n", indentStr, key)
			cli.displayMapData(v, indent+1)
		default:
			fmt.Printf("%s%s: %+v\n", indentStr, key, v)
		}
	}
}

// displayArrayData displays array data with proper formatting
func (cli *CLIIntegration) displayArrayData(data []interface{}, indent int) {
	indentStr := ""
	for i := 0; i < indent; i++ {
		indentStr += "  "
	}

	for i, item := range data {
		switch v := item.(type) {
		case string:
			fmt.Printf("%s%d. %s\n", indentStr, i+1, v)
		case map[string]interface{}:
			fmt.Printf("%s%d.\n", indentStr, i+1)
			cli.displayMapData(v, indent+1)
		default:
			fmt.Printf("%s%d. %+v\n", indentStr, i+1, v)
		}
	}
}

// ListFunctions displays all available functions
func (cli *CLIIntegration) ListFunctions() {
	fmt.Println(utils.FormatHeader("Available AI Functions"))
	fmt.Println(utils.FormatDivider())

	functions := ListAvailableFunctions()
	if len(functions) == 0 {
		fmt.Println("No functions registered")
		return
	}

	for name, description := range functions {
		fmt.Println(utils.FormatKeyValue(name, description))
	}

	fmt.Println(utils.FormatDivider())
}

// ShowFunctionSchema displays the schema for a specific function
func (cli *CLIIntegration) ShowFunctionSchema(functionName string) error {
	schema, err := GetFunctionSchema(functionName)
	if err != nil {
		return err
	}

	fmt.Println(utils.FormatHeader(fmt.Sprintf("Function Schema: %s", functionName)))
	fmt.Println(utils.FormatDivider())

	fmt.Println(utils.FormatKeyValue("Name", schema.Name))
	fmt.Println(utils.FormatKeyValue("Description", schema.Description))

	if len(schema.Parameters) > 0 {
		fmt.Println("\n" + utils.FormatHeader("Parameters"))
		for _, param := range schema.Parameters {
			required := ""
			if param.Required {
				required = " (required)"
			}

			paramInfo := fmt.Sprintf("%s%s - %s", param.Type, required, param.Description)
			if len(param.Enum) > 0 {
				paramInfo += fmt.Sprintf(" [options: %v]", param.Enum)
			}

			fmt.Println(utils.FormatKeyValue(param.Name, paramInfo))
		}
	}

	if len(schema.Examples) > 0 {
		fmt.Println("\n" + utils.FormatHeader("Examples"))
		for i, example := range schema.Examples {
			fmt.Printf("%d. %s\n", i+1, example.Description)
			if paramJSON, err := json.MarshalIndent(example.Parameters, "   ", "  "); err == nil {
				fmt.Printf("   Parameters: %s\n", string(paramJSON))
			}
			fmt.Printf("   Expected: %s\n\n", example.Expected)
		}
	}

	fmt.Println(utils.FormatDivider())
	return nil
}

// ValidateCall validates a function call from CLI
func (cli *CLIIntegration) ValidateCall(functionName string, paramsJSON string) error {
	var params map[string]interface{}
	if paramsJSON != "" {
		if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
			return fmt.Errorf("failed to parse parameters JSON: %v", err)
		}
	} else {
		params = make(map[string]interface{})
	}

	call := CreateCall(functionName, params)
	if err := ValidateFunction(call); err != nil {
		return fmt.Errorf("validation failed: %v", err)
	}

	fmt.Println(utils.FormatKeyValue("Validation", "✅ Passed"))
	return nil
}

// InteractiveMode starts an interactive mode for function calling
func (cli *CLIIntegration) InteractiveMode() {
	fmt.Println(utils.FormatHeader("AI Function Calling - Interactive Mode"))
	fmt.Println("Type 'help' for available commands, 'exit' to quit")
	fmt.Println(utils.FormatDivider())

	for {
		fmt.Print("nixai-func> ")

		var input string
		fmt.Scanln(&input)

		switch input {
		case "exit", "quit":
			fmt.Println("Goodbye!")
			return
		case "help":
			cli.showInteractiveHelp()
		case "list":
			cli.ListFunctions()
		default:
			fmt.Printf("Unknown command: %s. Type 'help' for available commands.\n", input)
		}
	}
}

// showInteractiveHelp displays help for interactive mode
func (cli *CLIIntegration) showInteractiveHelp() {
	fmt.Println(utils.FormatHeader("Interactive Mode Commands"))
	fmt.Println(utils.FormatKeyValue("help", "Show this help message"))
	fmt.Println(utils.FormatKeyValue("list", "List all available functions"))
	fmt.Println(utils.FormatKeyValue("exit/quit", "Exit interactive mode"))
	fmt.Println(utils.FormatDivider())
}

// CreateSampleCall creates a sample function call for testing
func CreateSampleCall() FunctionCall {
	params := map[string]interface{}{
		"user_description":   "My nginx service won't start after rebuilding NixOS",
		"log_data":           "systemctl status nginx shows: Job for nginx.service failed because the control process exited with error code",
		"analysis_type":      "service",
		"include_steps":      true,
		"include_prevention": true,
		"system_info": map[string]interface{}{
			"nixos_version":   "23.11",
			"nix_version":     "2.18.1",
			"is_flake_system": true,
		},
	}

	return CreateCallWithContext(context.Background(), "diagnose", params)
}
