package main

import (
	"fmt"
	"nix-ai-help/internal/mcp"
)

func main() {
	client := mcp.NewMCPClient("http://localhost:8081")
	doc, err := client.QueryDocumentation("services.nginx.enable")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Doc received: '%s'\n", doc)
		fmt.Printf("Doc length: %d\n", len(doc))
	}
}
