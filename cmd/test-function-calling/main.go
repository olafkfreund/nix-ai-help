package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"nix-ai-help/internal/ai/function"
	"nix-ai-help/pkg/logger"
)

func main() {
	var (
		functionName  = flag.String("function", "diagnose", "Function name to execute")
		paramsJSON    = flag.String("params", "", "Function parameters as JSON")
		showProgress  = flag.Bool("progress", true, "Show progress during execution")
		listFunctions = flag.Bool("list", false, "List all available functions")
		showSchema    = flag.String("schema", "", "Show schema for specific function")
		validate      = flag.Bool("validate", false, "Only validate parameters, don't execute")
		interactive   = flag.Bool("interactive", false, "Start interactive mode")
		sample        = flag.Bool("sample", false, "Run a sample diagnose call")
	)
	flag.Parse()

	logger := logger.NewLogger()
	cli := function.NewCLIIntegration()

	// Handle different modes
	if *listFunctions {
		cli.ListFunctions()
		return
	}

	if *showSchema != "" {
		if err := cli.ShowFunctionSchema(*showSchema); err != nil {
			logger.Error(fmt.Sprintf("Error showing schema: %v", err))
			os.Exit(1)
		}
		return
	}

	if *interactive {
		cli.InteractiveMode()
		return
	}

	if *sample {
		runSampleTest(cli, logger)
		return
	}

	// Handle validation or execution
	if *validate {
		if err := cli.ValidateCall(*functionName, *paramsJSON); err != nil {
			logger.Error(fmt.Sprintf("Validation failed: %v", err))
			os.Exit(1)
		}
		fmt.Println("✅ Parameters are valid")
		return
	}

	// Execute function
	if err := cli.ExecuteFromCLI(*functionName, *paramsJSON, *showProgress); err != nil {
		logger.Error(fmt.Sprintf("Execution failed: %v", err))
		os.Exit(1)
	}
}

func runSampleTest(cli *function.CLIIntegration, logger *logger.Logger) {
	fmt.Println("🧪 Running sample diagnose function test...")
	fmt.Println(strings.Repeat("=", 60))

	// Create sample call
	call := function.CreateSampleCall()

	// Set up options with progress reporting
	options := &function.FunctionOptions{
		Timeout: 2 * time.Minute,
		ProgressCallback: func(progress function.Progress) {
			fmt.Printf("\r[%s] %d/%d (%.1f%%) - %s",
				progress.Stage,
				progress.Current,
				progress.Total,
				progress.Percentage,
				progress.Message)
			if progress.Percentage >= 100 {
				fmt.Println()
			}
		},
	}

	// Execute
	registry := function.GetGlobalRegistry()
	result, err := registry.Execute(context.Background(), call, options)

	if err != nil {
		logger.Error(fmt.Sprintf("Sample test failed: %v", err))
		os.Exit(1)
	}

	// Display results
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("📊 SAMPLE TEST RESULTS")
	fmt.Println(strings.Repeat("=", 60))

	if result.Success {
		fmt.Printf("✅ Status: Success\n")
		fmt.Printf("⏱️  Duration: %v\n", result.Duration)
		fmt.Println("\n📋 Result Data:")

		if resultJSON, err := json.MarshalIndent(result.Data, "", "  "); err == nil {
			fmt.Println(string(resultJSON))
		} else {
			fmt.Printf("%+v\n", result.Data)
		}
	} else {
		fmt.Printf("❌ Status: Failed\n")
		fmt.Printf("🚨 Error: %s\n", result.Error)
	}

	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("✨ Sample test completed!")
}
