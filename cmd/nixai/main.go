package main

import (
	"log"
	"nix-ai-help/internal/cli"
	"os"
)

func main() {
	// Ensure all logs go to stderr to avoid polluting HTTP responses
	log.SetOutput(os.Stderr)
	// Start the main application logic (calls CLI root command)
	cli.Execute()
}
