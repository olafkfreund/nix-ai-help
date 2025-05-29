package main

import (
	"fmt"
	"nix-ai-help/internal/ai"
	"os"
)

func main() {
	provider := ai.NewOllamaProvider("llama3")

	// Simple test prompt
	prompt := `Generate a complete Nix derivation for a Go project named "test-app".
The derivation should include:
- Function signature with inputs
- pname and version attributes
- src attribute
- buildGoModule call
- meta section

Return ONLY the Nix code, no explanation.`

	fmt.Printf("=== PROMPT ===\n%s\n\n", prompt)

	response, err := provider.Query(prompt)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("=== RESPONSE ===\n%s\n", response)
	fmt.Printf("=== RESPONSE LENGTH ===\n%d characters\n", len(response))
}
