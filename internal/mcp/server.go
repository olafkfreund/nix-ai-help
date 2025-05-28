package mcp

import (
	"fmt"
	"log"
	"net/http"
	"nix-ai-help/internal/config"
)

// Server represents the MCP server that handles requests for NixOS documentation.
type Server struct {
	addr                 string
	documentationSources []string
}

// NewServer creates a new MCP server instance with documentation sources.
func NewServer(addr string, documentationSources []string) *Server {
	return &Server{addr: addr, documentationSources: documentationSources}
}

// NewServerFromConfig creates a new MCP server from a YAML config file.
func NewServerFromConfig(configPath string) (*Server, error) {
	cfg, err := config.LoadYAMLConfig(configPath)
	if err != nil {
		return nil, err
	}
	addr := fmt.Sprintf("%s:%d", cfg.MCPServer.Host, cfg.MCPServer.Port)
	return &Server{addr: addr, documentationSources: cfg.MCPServer.DocumentationSources}, nil
}

// Start initializes and starts the MCP server.
func (s *Server) Start() error {
	http.HandleFunc("/query", s.handleQuery)
	log.Printf("Starting MCP server on %s\n", s.addr)
	return http.ListenAndServe(s.addr, nil)
}

// handleQuery processes incoming requests for NixOS documentation.
func (s *Server) handleQuery(w http.ResponseWriter, r *http.Request) {
	// Example: print documentation sources for now
	fmt.Fprintf(w, "MCP Server: Documentation sources: %v\n", s.documentationSources)
	// TODO: Implement documentation querying logic here.
}
