package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"nix-ai-help/internal/config"
	"nix-ai-help/pkg/logger"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/sourcegraph/jsonrpc2"
)

// ElasticSearch configuration for NixOS options
const (
	ElasticSearchUsername    = "aWVSALXpZv"
	ElasticSearchPassword    = "X8gPHnzL52wFEekuxsfQ9cSh"
	ElasticSearchURLTemplate = `https://nixos-search-7-1733963800.us-east-1.bonsaisearch.net:443/%s/_search`
	ElasticSearchIndexPrefix = "latest-*-"
)

// NixOS option structure from ElasticSearch
type NixOSOption struct {
	Type        string `json:"type"`
	Source      string `json:"option_source"`
	Name        string `json:"option_name"`
	Description string `json:"option_description"`
	OptionType  string `json:"option_type"`
	Default     string `json:"option_default"`
	Example     string `json:"option_example"`
	Flake       string `json:"option_flake"`
}

// ElasticSearch response structure
type ESResponse struct {
	Hits struct {
		Hits []struct {
			Source NixOSOption `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

// MCPServer represents the MCP protocol server
type MCPServer struct {
	logger      logger.Logger
	listener    net.Listener
	mu          sync.Mutex
	lspProvider *NixLSPProvider
}

// MCPRequest represents an MCP protocol request
type MCPRequest struct {
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

// MCPResponse represents an MCP protocol response
type MCPResponse struct {
	Result interface{} `json:"result,omitempty"`
	Error  *MCPError   `json:"error,omitempty"`
}

// MCPError represents an error in MCP protocol
type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Tool represents an MCP tool
type Tool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Add a package-level variable to track uptime for metrics and health endpoints
var startTime time.Time

// Handle processes MCP protocol requests
func (m *MCPServer) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	m.logger.Debug(fmt.Sprintf("Handle called | method=%s id=%v", req.Method, req.ID))
	m.mu.Lock()
	defer m.mu.Unlock()

	switch req.Method {
	case "initialize":
		result := map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{
					"listChanged": false,
				},
			},
			"serverInfo": map[string]interface{}{
				"name":    "nixai-mcp-server",
				"version": "1.0.0",
			},
		}
		_ = conn.Reply(ctx, req.ID, result)

	case "tools/list":
		tools := []Tool{
			{
				Name:        "query_nixos_docs",
				Description: "Query NixOS documentation from multiple sources",
			},
			{
				Name:        "explain_nixos_option",
				Description: "Explain NixOS configuration options",
			},
			{
				Name:        "explain_home_manager_option",
				Description: "Explain Home Manager configuration options",
			},
			{
				Name:        "search_nixos_packages",
				Description: "Search for NixOS packages",
			},
			{
				Name:        "complete_nixos_option",
				Description: "Autocomplete NixOS option names for a given prefix",
			},
			{
				Name:        "nix_lsp_completion",
				Description: "Provide LSP-like completion suggestions for Nix files",
			},
			{
				Name:        "nix_lsp_diagnostics",
				Description: "Provide real-time diagnostics and error checking for Nix files",
			},
			{
				Name:        "nix_lsp_hover",
				Description: "Provide hover information and documentation for Nix symbols",
			},
			{
				Name:        "nix_lsp_definition",
				Description: "Provide go-to-definition functionality for Nix symbols",
			},
		}
		_ = conn.Reply(ctx, req.ID, map[string]interface{}{"tools": tools})

	case "tools/call":
		var params struct {
			Name      string                 `json:"name"`
			Arguments map[string]interface{} `json:"arguments"`
		}

		if err := json.Unmarshal(*req.Params, &params); err != nil {
			_ = conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{
				Code:    jsonrpc2.CodeInvalidParams,
				Message: "Invalid parameters",
			})
			return
		}

		switch params.Name {
		case "query_nixos_docs":
			if query, ok := params.Arguments["query"].(string); ok {
				// Extract sources if provided
				var sources []string
				if sourcesArg, ok := params.Arguments["sources"].([]interface{}); ok {
					for _, src := range sourcesArg {
						if srcStr, ok := src.(string); ok {
							sources = append(sources, srcStr)
						}
					}
				}

				result := m.handleDocQuery(query, sources...)
				_ = conn.Reply(ctx, req.ID, map[string]interface{}{
					"content": []map[string]interface{}{
						{
							"type": "text",
							"text": result,
						},
					},
				})
			} else {
				_ = conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{
					Code:    jsonrpc2.CodeInvalidParams,
					Message: "Missing query parameter",
				})
			}

		case "explain_nixos_option":
			if option, ok := params.Arguments["option"].(string); ok {
				result := m.handleOptionExplain(option)
				_ = conn.Reply(ctx, req.ID, map[string]interface{}{
					"content": []map[string]interface{}{
						{
							"type": "text",
							"text": result,
						},
					},
				})
			} else {
				_ = conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{
					Code:    jsonrpc2.CodeInvalidParams,
					Message: "Missing option parameter",
				})
			}

		case "explain_home_manager_option":
			if option, ok := params.Arguments["option"].(string); ok {
				result := m.handleHomeManagerOptionExplain(option)
				_ = conn.Reply(ctx, req.ID, map[string]interface{}{
					"content": []map[string]interface{}{
						{
							"type": "text",
							"text": result,
						},
					},
				})
			} else {
				_ = conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{
					Code:    jsonrpc2.CodeInvalidParams,
					Message: "Missing option parameter",
				})
			}

		case "search_nixos_packages":
			if query, ok := params.Arguments["query"].(string); ok {
				result := m.handlePackageSearch(query)
				_ = conn.Reply(ctx, req.ID, map[string]interface{}{
					"content": []map[string]interface{}{
						{
							"type": "text",
							"text": result,
						},
					},
				})
			} else {
				_ = conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{
					Code:    jsonrpc2.CodeInvalidParams,
					Message: "Missing query parameter",
				})
			}

		case "complete_nixos_option":
			if prefix, ok := params.Arguments["prefix"].(string); ok {
				results := m.handleOptionCompletion(prefix)
				_ = conn.Reply(ctx, req.ID, map[string]interface{}{
					"options": results,
				})
			} else {
				_ = conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{
					Code:    jsonrpc2.CodeInvalidParams,
					Message: "Missing prefix parameter",
				})
			}

		case "nix_lsp_completion":
			if m.lspProvider == nil {
				_ = conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{
					Code:    jsonrpc2.CodeInternalError,
					Message: "LSP provider not initialized",
				})
				return
			}

			fileContent, ok1 := params.Arguments["fileContent"].(string)
			line, ok2 := params.Arguments["line"].(float64)
			character, ok3 := params.Arguments["character"].(float64)

			if !ok1 || !ok2 || !ok3 {
				_ = conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{
					Code:    jsonrpc2.CodeInvalidParams,
					Message: "Missing required parameters: fileContent, line, character",
				})
				return
			}

			position := LSPPosition{Line: int(line), Character: int(character)}
			completions, err := m.lspProvider.ProvideCompletion(fileContent, position)
			if err != nil {
				_ = conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{
					Code:    jsonrpc2.CodeInternalError,
					Message: "Failed to provide completions: " + err.Error(),
				})
				return
			}

			_ = conn.Reply(ctx, req.ID, map[string]interface{}{
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": m.lspProvider.FormatCompletions(completions),
					},
				},
				"completions": completions,
			})

		case "nix_lsp_diagnostics":
			if m.lspProvider == nil {
				_ = conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{
					Code:    jsonrpc2.CodeInternalError,
					Message: "LSP provider not initialized",
				})
				return
			}

			fileContent, ok1 := params.Arguments["fileContent"].(string)
			filePath, ok2 := params.Arguments["filePath"].(string)

			if !ok1 {
				_ = conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{
					Code:    jsonrpc2.CodeInvalidParams,
					Message: "Missing required parameter: fileContent",
				})
				return
			}

			if !ok2 {
				filePath = "untitled.nix" // Default filename
			}

			diagnostics, err := m.lspProvider.ProvideDiagnostics(filePath, fileContent)
			if err != nil {
				_ = conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{
					Code:    jsonrpc2.CodeInternalError,
					Message: "Failed to provide diagnostics: " + err.Error(),
				})
				return
			}

			_ = conn.Reply(ctx, req.ID, map[string]interface{}{
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": m.lspProvider.FormatDiagnostics(diagnostics),
					},
				},
				"diagnostics": diagnostics,
			})

		case "nix_lsp_hover":
			if m.lspProvider == nil {
				_ = conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{
					Code:    jsonrpc2.CodeInternalError,
					Message: "LSP provider not initialized",
				})
				return
			}

			fileContent, ok1 := params.Arguments["fileContent"].(string)
			line, ok2 := params.Arguments["line"].(float64)
			character, ok3 := params.Arguments["character"].(float64)

			if !ok1 || !ok2 || !ok3 {
				_ = conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{
					Code:    jsonrpc2.CodeInvalidParams,
					Message: "Missing required parameters: fileContent, line, character",
				})
				return
			}

			position := LSPPosition{Line: int(line), Character: int(character)}
			hover, err := m.lspProvider.ProvideHover(fileContent, position)
			if err != nil {
				_ = conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{
					Code:    jsonrpc2.CodeInternalError,
					Message: "Failed to provide hover information: " + err.Error(),
				})
				return
			}

			if hover == nil {
				_ = conn.Reply(ctx, req.ID, map[string]interface{}{
					"content": []map[string]interface{}{
						{
							"type": "text",
							"text": "No hover information available",
						},
					},
				})
				return
			}

			_ = conn.Reply(ctx, req.ID, map[string]interface{}{
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": strings.Join(hover.Contents, "\n"),
					},
				},
				"hover": hover,
			})

		case "nix_lsp_definition":
			if m.lspProvider == nil {
				_ = conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{
					Code:    jsonrpc2.CodeInternalError,
					Message: "LSP provider not initialized",
				})
				return
			}

			fileContent, ok1 := params.Arguments["fileContent"].(string)
			line, ok2 := params.Arguments["line"].(float64)
			character, ok3 := params.Arguments["character"].(float64)

			if !ok1 || !ok2 || !ok3 {
				_ = conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{
					Code:    jsonrpc2.CodeInvalidParams,
					Message: "Missing required parameters: fileContent, line, character",
				})
				return
			}

			position := LSPPosition{Line: int(line), Character: int(character)}
			locations, err := m.lspProvider.ProvideDefinition(fileContent, position)
			if err != nil {
				_ = conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{
					Code:    jsonrpc2.CodeInternalError,
					Message: "Failed to provide definition: " + err.Error(),
				})
				return
			}

			if len(locations) == 0 {
				_ = conn.Reply(ctx, req.ID, map[string]interface{}{
					"content": []map[string]interface{}{
						{
							"type": "text",
							"text": "No definition found",
						},
					},
				})
				return
			}

			var result strings.Builder
			result.WriteString("Found definition(s):\n\n")
			for i, loc := range locations {
				result.WriteString(fmt.Sprintf("%d. %s\n", i+1, loc.URI))
			}

			_ = conn.Reply(ctx, req.ID, map[string]interface{}{
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": result.String(),
					},
				},
				"locations": locations,
			})

		default:
			_ = conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{
				Code:    jsonrpc2.CodeMethodNotFound,
				Message: "Unknown tool: " + params.Name,
			})
		}

	default:
		_ = conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{
			Code:    jsonrpc2.CodeMethodNotFound,
			Message: "Method not found: " + req.Method,
		})
	}
}

// Start starts the MCP server on Unix socket
func (m *MCPServer) Start(socketPath string) error {
	// Remove existing socket file if it exists
	_ = os.Remove(socketPath)

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		m.logger.Error(fmt.Sprintf("Failed to listen on Unix socket | socketPath=%s error=%v", socketPath, err))
		return fmt.Errorf("failed to listen on Unix socket %s: %v", socketPath, err)
	}

	// Store listener for cleanup
	m.mu.Lock()
	m.listener = listener
	m.mu.Unlock()

	m.logger.Info(fmt.Sprintf("MCP server listening on Unix socket | socketPath=%s", socketPath))

	// Accept connections in a blocking loop
	for {
		conn, err := listener.Accept()
		if err != nil {
			m.logger.Error(fmt.Sprintf("Failed to accept connection | error=%v", err))
			continue
		}

		go func(conn net.Conn) {
			defer func() { _ = conn.Close() }()
			m.logger.Debug(fmt.Sprintf("New MCP client connected | remoteAddr=%v", conn.RemoteAddr()))

			// Handle connection with JSON-RPC2
			stream := jsonrpc2.NewPlainObjectStream(conn)
			m.logger.Debug("Created buffered stream")

			jsonConn := jsonrpc2.NewConn(context.Background(), stream, m)
			m.logger.Debug("Created JSON-RPC2 connection")
			defer func() { _ = jsonConn.Close() }()

			// Keep connection alive
			m.logger.Debug("Waiting for disconnect notification")
			<-jsonConn.DisconnectNotify()
			m.logger.Debug("MCP client disconnected")
		}(conn)
	}
}

// Stop stops the MCP server
func (m *MCPServer) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.listener != nil {
		_ = m.listener.Close()
		m.listener = nil
	}
}

// handleDocQuery processes documentation queries
func (m *MCPServer) handleDocQuery(query string, sources ...string) string {
	// Add debug header to identify this method is being called
	var debugOutput strings.Builder
	debugOutput.WriteString("==== USING MCP SERVER HANDLE_DOC_QUERY ====\n")
	debugOutput.WriteString(fmt.Sprintf("Query: %s\n", query))
	debugOutput.WriteString(fmt.Sprintf("Sources: %v\n", sources))
	debugOutput.WriteString("===================================\n\n")

	// Create request to process internally
	var requestSources []string
	if len(sources) > 0 {
		requestSources = sources
		if m != nil {
			m.logger.Debug(fmt.Sprintf("handleDocQuery: using provided sources: %v", requestSources))
		}
	} else {
		// Get default sources from the server (by looking at the server field)
		server := findServerInstance()
		if server != nil {
			requestSources = server.documentationSources
			if m != nil {
				m.logger.Debug(fmt.Sprintf("handleDocQuery: using server default sources: %v", requestSources))
			}
		} else {
			// Fallback to well-known sources
			requestSources = []string{
				"nixos-options-es://",
				"https://home-manager-options.extranix.com/options.json",
				"https://wiki.nixos.org/wiki/NixOS_Wiki",
				"https://nix.dev/manual/nix",
			}
			if m != nil {
				m.logger.Debug(fmt.Sprintf("handleDocQuery: using fallback sources: %v", requestSources))
			}
		}
	}

	// Use a buffer to capture output that would normally go to the ResponseWriter
	var buf bytes.Buffer

	// Process each source manually
	for _, src := range requestSources {
		var body string
		var err error

		if m != nil {
			m.logger.Debug(fmt.Sprintf("handleDocQuery: processing source: %s", src))
		}

		if strings.HasPrefix(src, "nixos-options-es://") {
			body, err = fetchNixOSOptionsAPI(src, query)
			if err == nil && !strings.Contains(body, "No documentation found") {
				if m != nil {
					m.logger.Debug(fmt.Sprintf("handleDocQuery: found result in NixOS options API: %s", src))
				}
				return debugOutput.String() + body // Return first good result with debug header
			}
		} else if strings.HasSuffix(src, "/options") {
			body, err = fetchNixOSOptionsAPI(src, query)
			if err == nil && !strings.Contains(body, "No documentation found") {
				if m != nil {
					m.logger.Debug(fmt.Sprintf("handleDocQuery: found result in NixOS options endpoint: %s", src))
				}
				return debugOutput.String() + body // Return first good result with debug header
			}
		} else if strings.HasSuffix(src, "/options.json") {
			body, err = fetchHomeManagerOptionsAPI(src, query)
			if err == nil && !strings.Contains(body, "No documentation found") {
				if m != nil {
					m.logger.Debug(fmt.Sprintf("handleDocQuery: found result in Home Manager options: %s", src))
				}
				return debugOutput.String() + body // Return first good result with debug header
			}
		} else if strings.Contains(src, "nix.dev") {
			body, err = fetchMySTContent(src, query)
			if err == nil && len(body) > 0 {
				if m != nil {
					m.logger.Debug(fmt.Sprintf("handleDocQuery: found result in nix.dev: %s", src))
				}
				return debugOutput.String() + body // Return first good result with debug header
			}
		} else {
			body, err = fetchDocSource(src, query)
			if err == nil && len(body) > 0 {
				if m != nil {
					m.logger.Debug(fmt.Sprintf("handleDocQuery: found partial result in: %s", src))
				}
				buf.WriteString(fmt.Sprintf("%s: %s\n", src, body))
			}
		}

		if err != nil && m != nil {
			m.logger.Debug(fmt.Sprintf("handleDocQuery: error processing source %s: %v", src, err))
		}
	}

	if buf.Len() > 0 {
		if m != nil {
			m.logger.Debug("handleDocQuery: returning combined results")
		}
		return debugOutput.String() + buf.String() // Return combined results with debug header
	}

	if m != nil {
		m.logger.Debug("handleDocQuery: no relevant documentation found")
	}
	return debugOutput.String() + "No relevant documentation found." // Return no results found with debug header
}

// Package-level variable to hold the server instance
var globalServerInstance *Server

// findServerInstance helps locate the server instance for accessing config
func findServerInstance() *Server {
	// Return the global server instance that we track
	if globalServerInstance == nil {
		// Log this issue in a way that doesn't cause further errors if logger is unavailable
		fmt.Fprintf(os.Stderr, "Warning: globalServerInstance is nil in findServerInstance()\n")
	}
	return globalServerInstance
}

// handleOptionExplain processes NixOS option explanations
func (m *MCPServer) handleOptionExplain(option string) string {
	// Directly call fetchNixOSOptionsAPI instead of making a recursive HTTP call
	result, err := fetchNixOSOptionsAPI("nixos-options-es://", option)
	if err != nil {
		return fmt.Sprintf("Error explaining option %s: %v", option, err)
	}
	return result
}

// handleHomeManagerOptionExplain processes Home Manager option explanations
func (m *MCPServer) handleHomeManagerOptionExplain(option string) string {
	// Directly call fetchHomeManagerOptionsAPI instead of making a recursive HTTP call
	result, err := fetchHomeManagerOptionsAPI("https://home-manager-options.extranix.com/options.json", option)
	if err != nil {
		return fmt.Sprintf("Error explaining Home Manager option %s: %v", option, err)
	}
	return result
}

// handlePackageSearch processes package search queries
func (m *MCPServer) handlePackageSearch(query string) string {
	return fmt.Sprintf("Package search for '%s' is not yet implemented in MCP protocol. Use the CLI interface: nixai search pkg %s", query, query)
}

// handleOptionCompletion processes option name completions for a given prefix
func (m *MCPServer) handleOptionCompletion(prefix string) []string {
	// For demo: use a static list, but in real use, query ElasticSearch or in-memory index
	allOptions := []string{
		"services.nginx.enable", "networking.firewall.enable", "programs.zsh.enable", "users.users", "environment.systemPackages", "fonts.fonts", "hardware.opengl.enable", "services.openssh.enable",
		// ... more options ...
	}
	var results []string
	for _, opt := range allOptions {
		if strings.HasPrefix(opt, prefix) {
			results = append(results, opt)
		}
	}
	return results
}

// Server represents the combined HTTP and MCP server
type Server struct {
	addr                 string
	socketPath           string
	documentationSources []string
	logger               *logger.Logger
	debugLogging         bool
	mcpServer            *MCPServer
	configPath           string
	watcher              *fsnotify.Watcher
}

// Add a simple in-memory cache for query results
var (
	cache      = make(map[string]string)
	cacheMutex sync.RWMutex
)

// NewServer creates a new MCP server instance with documentation sources.
func NewServer(addr string, documentationSources []string) *Server {
	log := logger.NewLoggerWithLevel("info")

	// Create and initialize LSP provider
	lspProvider := NewNixLSPProvider(*log)
	if err := lspProvider.LoadNixOSOptions(); err != nil {
		log.Error(fmt.Sprintf("Failed to load NixOS options for LSP: %v", err))
	}

	server := &Server{
		addr:                 addr,
		socketPath:           "/tmp/nixai-mcp.sock", // Default socket path
		documentationSources: documentationSources,
		logger:               log,
		debugLogging:         false,
		mcpServer:            &MCPServer{logger: *log, lspProvider: lspProvider},
	}

	// Set the global server instance for cross-referencing
	globalServerInstance = server

	return server
}

// NewServerWithDebug creates a new MCP server instance with debug logging enabled.
// This is primarily intended for testing purposes.
func NewServerWithDebug(addr string, documentationSources []string) *Server {
	log := logger.NewLoggerWithLevel("debug")

	// Create and initialize LSP provider
	lspProvider := NewNixLSPProvider(*log)
	if err := lspProvider.LoadNixOSOptions(); err != nil {
		log.Error(fmt.Sprintf("Failed to load NixOS options for LSP: %v", err))
	}

	server := &Server{
		addr:                 addr,
		socketPath:           "/tmp/nixai-mcp.sock", // Default socket path
		documentationSources: documentationSources,
		logger:               log,
		debugLogging:         true,
		mcpServer:            &MCPServer{logger: *log, lspProvider: lspProvider},
	}
	
	// Set the global server instance for cross-referencing
	globalServerInstance = server
	
	return server
}

// NewServerFromConfig creates a new MCP server from a YAML config file.
func NewServerFromConfig(configPath string) (*Server, error) {
	// If configPath is empty, use default user config path
	if configPath == "" {
		configPath = os.ExpandEnv("$HOME/.config/nixai/config.yaml")
	}

	// If config file does not exist, create it from embedded default config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Use embedded config instead of reading from file system
		_, err := config.EnsureConfigFileFromEmbedded()
		if err != nil {
			return nil, fmt.Errorf("failed to create user config from embedded default: %w", err)
		}
	}

	// Use LoadUserConfig instead of LoadYAMLConfig since user configs don't have the "default:" wrapper
	userCfg, err := config.LoadUserConfig()
	if err != nil {
		return nil, err
	}
	addr := fmt.Sprintf("%s:%d", userCfg.MCPServer.Host, userCfg.MCPServer.Port)
	socketPath := "/tmp/nixai-mcp.sock" // Default
	if userCfg.MCPServer.SocketPath != "" {
		socketPath = userCfg.MCPServer.SocketPath
	}

	log := logger.NewLoggerWithLevel(userCfg.LogLevel)

	// Create and initialize LSP provider
	lspProvider := NewNixLSPProvider(*log)
	if err := lspProvider.LoadNixOSOptions(); err != nil {
		log.Error(fmt.Sprintf("Failed to load NixOS options for LSP: %v", err))
	}

	srv := &Server{
		addr:                 addr,
		socketPath:           socketPath,
		documentationSources: userCfg.MCPServer.DocumentationSources,
		logger:               log,
		debugLogging:         strings.ToLower(userCfg.LogLevel) == "debug",
		mcpServer:            &MCPServer{logger: *log, lspProvider: lspProvider},
		configPath:           configPath,
	}
	
	// Set the global server instance for cross-referencing
	globalServerInstance = srv

	// Set up config watcher for hot-reload
	watcher, err := fsnotify.NewWatcher()
	if err == nil {
		srv.watcher = watcher
		go srv.watchConfig()
		if err := watcher.Add(configPath); err != nil {
			srv.logger.Error(fmt.Sprintf("Failed to watch config file: %v", err))
		}
	} else {
		srv.logger.Error(fmt.Sprintf("Failed to initialize config watcher: %v", err))
	}

	return srv, nil
}

// watchConfig watches the config file for changes and reloads it
func (s *Server) watchConfig() {
	for {
		select {
		case event, ok := <-s.watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				s.logger.Info("Config file changed, reloading...")
				userCfg, err := config.LoadUserConfig()
				if err == nil {
					s.documentationSources = userCfg.MCPServer.DocumentationSources
					s.logger.Info("Reloaded documentation sources from config.")
					// Optionally reload log level, etc.
				} else {
					s.logger.Error(fmt.Sprintf("Failed to reload config: %v", err))
				}
			}
		case err, ok := <-s.watcher.Errors:
			if !ok {
				return
			}
			s.logger.Error(fmt.Sprintf("Config watcher error: %v", err))
		}
	}
}

// SetSocketPath sets a custom socket path for the MCP server
func (s *Server) SetSocketPath(path string) {
	s.socketPath = path
}

// Start initializes and starts the MCP server with graceful shutdown support.
func (s *Server) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/query", s.handleQuery)

	// Improved /healthz endpoint
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": "ok", "uptime": time.Now().Format(time.RFC3339)})
	})

	// /metrics endpoint (simple Prometheus format)
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		// Example metrics (replace with real metrics as needed)
		_, _ = fmt.Fprintln(w, "# HELP nixai_mcp_requests_total Total number of /query requests")
		_, _ = fmt.Fprintln(w, "# TYPE nixai_mcp_requests_total counter")
		_, _ = fmt.Fprintln(w, "nixai_mcp_requests_total 0")
		_, _ = fmt.Fprintln(w, "# HELP nixai_mcp_uptime_seconds Uptime in seconds")
		uptime := int(time.Since(startTime).Seconds())
		_, _ = fmt.Fprintf(w, "nixai_mcp_uptime_seconds %d\n", uptime)
	})

	shutdownCh := make(chan struct{})
	mux.HandleFunc("/shutdown", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Shutting down MCP server...\n"))
		s.logger.Info("Shutdown endpoint called, shutting down MCP server")
		go func() {
			shutdownCh <- struct{}{}
		}()
	})

	server := &http.Server{
		Addr:    s.addr,
		Handler: mux,
	}

	s.logger.Info(fmt.Sprintf("Starting MCP server | addr=%s", s.addr))

	// Track start time for metrics
	startTime = time.Now()

	// Run HTTP server in goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.ListenAndServe()
	}()

	// Run MCP server in goroutine - but don't capture its result
	// since the MCP server runs indefinitely and should not exit
	go func() {
		// Use the server's socketPath field, which might have been customized
		socketPath := s.socketPath
		if socketPath == "" {
			socketPath = "/tmp/nixai-mcp.sock" // Default fallback
		}

		// Check environment variable for override
		if envPath := os.Getenv("NIXAI_SOCKET_PATH"); envPath != "" {
			socketPath = envPath
		}

		// Start the MCP server (this blocks and shouldn't return unless there's an error)
		if err := s.mcpServer.Start(socketPath); err != nil {
			s.logger.Error(fmt.Sprintf("MCP server encountered an error | error=%v", err))
			// Don't exit the main server if the MCP server exits - just log the error
		}
	}()

	// Wait for shutdown signal or HTTP server error
	select {
	case <-shutdownCh:
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		s.logger.Info("Shutting down MCP server")
		s.mcpServer.Stop()
		return server.Shutdown(ctx)
	case err := <-errCh:
		if strings.Contains(err.Error(), "address already in use") {
			s.logger.Error(fmt.Sprintf("The MCP server could not start because the address is already in use. | error=%v", err))
		}
		s.mcpServer.Stop() // Make sure to stop the MCP server if HTTP server fails
		return err
	}
}

// Levenshtein distance for fuzzy matching
func levenshtein(a, b string) int {
	la, lb := len(a), len(b)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}
	dp := make([][]int, la+1)
	for i := range dp {
		dp[i] = make([]int, lb+1)
	}
	for i := 0; i <= la; i++ {
		dp[i][0] = i
	}
	for j := 0; j <= lb; j++ {
		dp[0][j] = j
	}
	for i := 1; i <= la; i++ {
		for j := 1; j <= lb; j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			dp[i][j] = min(
				dp[i-1][j]+1,
				dp[i][j-1]+1,
				dp[i-1][j-1]+cost,
			)
		}
	}
	return dp[la][lb]
}

func min(a, b, c int) int {
	if a < b && a < c {
		return a
	}
	if b < c {
		return b
	}
	return c
}

// handleQuery processes incoming requests for NixOS documentation.
func (s *Server) handleQuery(w http.ResponseWriter, r *http.Request) {
	var query string
	var sources []string

	// Handle both GET requests with 'q' parameter and POST requests with JSON body
	switch r.Method {
	case "GET":
		query = r.URL.Query().Get("q")
		if query == "" {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = fmt.Fprintln(w, "Missing 'q' query parameter.")
			return
		}
		// Use default sources for GET requests
		sources = s.documentationSources
	case "POST":
		var requestBody struct {
			Query   string   `json:"query"`
			Sources []string `json:"sources,omitempty"`
		}

		// Read the raw request body for debugging
		bodyBytes, _ := io.ReadAll(r.Body)
		s.logger.Info(fmt.Sprintf("Raw request body: %s", string(bodyBytes)))

		// Replace the body for further processing
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = fmt.Fprintln(w, "Invalid JSON body.")
			return
		}
		query = requestBody.Query
		if query == "" {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = fmt.Fprintln(w, "Missing 'query' field in JSON body.")
			return
		}

		// Use sources from request if provided, otherwise use default sources
		if len(requestBody.Sources) > 0 {
			sources = requestBody.Sources
			s.logger.Info(fmt.Sprintf("Using sources from POST request: %v", sources))
		} else {
			sources = s.documentationSources
			s.logger.Info(fmt.Sprintf("Using default sources: %v", s.documentationSources))
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		_, _ = fmt.Fprintln(w, "Method not allowed. Use GET or POST.")
		return
	}

	if s.debugLogging {
		s.logger.Debug(fmt.Sprintf("handleQuery: received query | query=%s sources=%v", query, sources))
	}

	// Helper to write JSON response
	writeJSON := func(result string) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"result": result})
	}

	// Create a cache key that includes both query and sources
	cacheKey := fmt.Sprintf("%s|%s", query, strings.Join(sources, ","))

	// Check cache first
	cacheMutex.RLock()
	if cached, ok := cache[cacheKey]; ok {
		cacheMutex.RUnlock()
		writeJSON(cached)
		return
	}
	cacheMutex.RUnlock()
	// Use the mcpServer's handleDocQuery method for consistency
	s.logger.Debug(fmt.Sprintf("handleQuery: calling handleDocQuery with query=%s and sources=%v", query, sources))
	
	// Debug check if globalServerInstance is set correctly
	if s.debugLogging {
		if globalServerInstance == nil {
			s.logger.Debug("handleQuery: WARNING - globalServerInstance is nil")
		} else {
			s.logger.Debug(fmt.Sprintf("handleQuery: globalServerInstance has %d documentation sources", 
				len(globalServerInstance.documentationSources)))
		}
	}
	
	result := s.mcpServer.handleDocQuery(query, sources...)
	
	// Cache the result
	cacheMutex.Lock()
	cache[cacheKey] = result
	cacheMutex.Unlock()

	// Return the result
	writeJSON(result)
}

func fetchDocSource(urlStr string, queryTerm string) (string, error) {
	if strings.HasSuffix(urlStr, "/options") {
		return fetchNixOSOptionsAPI(urlStr, queryTerm)
	}
	if strings.HasSuffix(urlStr, "/options.json") {
		return fetchHomeManagerOptionsAPI(urlStr, queryTerm)
	}

	// Special handler for MediaWiki sites like wiki.nixos.org
	if strings.Contains(urlStr, "wiki.nixos.org") {
		return fetchMediaWikiContent(urlStr, queryTerm)
	}

	// Special handler for nix.dev documentation that uses MyST
	if strings.Contains(urlStr, "nix.dev") {
		return fetchMySTContent(urlStr, queryTerm)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	// For regular URLs, if we have a query term, try to append it as a search parameter
	if queryTerm != "" {
		parsedURL, err := url.Parse(urlStr)
		if err == nil {
			q := parsedURL.Query()
			q.Set("q", queryTerm)
			parsedURL.RawQuery = q.Encode()
			urlStr = parsedURL.String()
		}
	}

	// #nosec G107 -- urlStr is from trusted config/documentation sources only
	resp, err := client.Get(urlStr)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch %s: %s", urlStr, resp.Status)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// fetchMediaWikiContent uses the MediaWiki API to search for content on wiki.nixos.org
func fetchMediaWikiContent(wikiURL string, query string) (string, error) {
	if query == "" {
		return "", fmt.Errorf("query term required for MediaWiki search")
	}

	// Parse the base URL to extract the domain
	parsedURL, err := url.Parse(wikiURL)
	if err != nil {
		return "", fmt.Errorf("invalid wiki URL: %v", err)
	}

	// Construct the API URL for searching
	apiURL := fmt.Sprintf("%s://%s/w/api.php", parsedURL.Scheme, parsedURL.Host)

	// Create the request URL with query parameters
	reqURL, err := url.Parse(apiURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse API URL: %v", err)
	}

	q := reqURL.Query()
	q.Set("action", "query")
	q.Set("list", "search")
	q.Set("srsearch", query)
	q.Set("format", "json")
	q.Set("srlimit", "5") // Limit to 5 results for conciseness
	reqURL.RawQuery = q.Encode()

	// Make the API request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(reqURL.String())
	if err != nil {
		return "", fmt.Errorf("failed to query MediaWiki API: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("MediaWiki API returned status %d", resp.StatusCode)
	}

	// Parse the JSON response
	var apiResp struct {
		Query struct {
			Search []struct {
				Title   string `json:"title"`
				Snippet string `json:"snippet"`
				PageID  int    `json:"pageid"`
			} `json:"search"`
		} `json:"query"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return "", fmt.Errorf("failed to parse MediaWiki API response: %v", err)
	}

	// Format the search results
	var result strings.Builder
	if len(apiResp.Query.Search) == 0 {
		return "No MediaWiki search results found for: " + query, nil
	}

	result.WriteString("MediaWiki search results for: " + query + "\n\n")

	for _, page := range apiResp.Query.Search {
		// Clean up HTML in snippet
		snippet := stripHTMLTags(page.Snippet)
		snippet = strings.ReplaceAll(snippet, "&quot;", "\"")
		snippet = strings.ReplaceAll(snippet, "&amp;", "&")

		result.WriteString("Title: " + page.Title + "\n")
		result.WriteString("Snippet: " + snippet + "\n")
		result.WriteString("URL: " + fmt.Sprintf("%s://%s/wiki/%s", parsedURL.Scheme, parsedURL.Host, url.PathEscape(page.Title)) + "\n\n")
	}

	return result.String(), nil
}

// Helper to strip HTML tags from ES description fields
func stripHTMLTags(s string) string {
	re := regexp.MustCompile(`<.*?>`)
	return re.ReplaceAllString(s, "")
}

// fetchNixOSOptionsAPI fetches and parses option docs from the NixOS Elasticsearch backend
func fetchNixOSOptionsAPI(_ string, option string) (string, error) {
	if strings.TrimSpace(option) == "" {
		return "", fmt.Errorf("option name required")
	}

	// Create retryable HTTP client
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 3
	retryClient.Logger = nil

	// Build ElasticSearch index URL
	index := ElasticSearchIndexPrefix + "nixos-unstable"
	esURL := fmt.Sprintf(ElasticSearchURLTemplate, index)

	// Build the query body for exact option match
	body := map[string]interface{}{
		"from": 0,
		"size": 3,
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []interface{}{
					map[string]interface{}{"match": map[string]interface{}{"type": "option"}},
					map[string]interface{}{"match": map[string]interface{}{"option_name": option}},
				},
			},
		},
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", esURL, bytes.NewReader(jsonBody))
	if err != nil {
		return "", err
	}

	req.SetBasicAuth(ElasticSearchUsername, ElasticSearchPassword)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := retryClient.StandardClient().Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to query ElasticSearch: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ElasticSearch returned status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Parse response
	var esResp ESResponse
	if err := json.Unmarshal(data, &esResp); err != nil {
		return "", fmt.Errorf("failed to parse ElasticSearch response: %w", err)
	}

	if len(esResp.Hits.Hits) == 0 {
		return "No documentation found for this option in the official NixOS options database.", nil
	}

	// Use the first (best) match
	opt := esResp.Hits.Hits[0].Source
	var result strings.Builder

	result.WriteString(fmt.Sprintf("Option: %s\n", opt.Name))

	if opt.Description != "" {
		cleanDesc := stripHTMLTags(opt.Description)
		result.WriteString(fmt.Sprintf("Description: %s\n", cleanDesc))
	}

	result.WriteString(fmt.Sprintf("Type: %s\n", opt.OptionType))

	if opt.Default != "" {
		result.WriteString(fmt.Sprintf("Default: %s\n", opt.Default))
	}

	if opt.Example != "" && opt.Example != "null" {
		result.WriteString(fmt.Sprintf("Example: %s\n", opt.Example))
	}

	if opt.Source != "" {
		result.WriteString(fmt.Sprintf("Source: %s\n", opt.Source))
	}

	return result.String(), nil
}

// fetchHomeManagerOptionsAPI fetches and parses option docs from home-manager-options.extranix.com or a compatible endpoint
// It's optimized for js-search which powers the search functionality on the home-manager-options site
func fetchHomeManagerOptionsAPI(baseURL, option string) (string, error) {
	if strings.TrimSpace(option) == "" {
		return "", fmt.Errorf("option name required")
	}

	// Prepare query for js-search - handle both exact and partial searches
	// js-search performs both prefix and infix matching by default
	apiURL := baseURL + "?query=" + url.QueryEscape(option)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(apiURL)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch %s: %s", apiURL, resp.Status)
	}

	var result struct {
		Options []struct {
			Name        string   `json:"name"`
			Description string   `json:"description"`
			Type        string   `json:"type"`
			Default     string   `json:"default"`
			Example     string   `json:"example"`
			ReadOnly    bool     `json:"readOnly"`
			Loc         []string `json:"loc"`
		} `json:"options"`
	}

	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&result); err != nil {
		return "", err
	}
	if len(result.Options) == 0 {
		return "No documentation found for option.", nil
	}

	// Define option type for ranking
	type optionType struct {
		Name        string
		Description string
		Type        string
		Default     string
		Example     string
		ReadOnly    bool
		Loc         []string
	}

	// Rank the options by relevance to the query
	type rankedOption struct {
		option optionType
		score  int
	}

	var rankedOptions []rankedOption
	for _, opt := range result.Options {
		// Calculate score (lower is better)
		score := 100 // Base score

		// Exact match is best
		if opt.Name == option {
			score = 0
		} else if strings.Contains(opt.Name, option) {
			// Partial match in name is good
			score = 10
		} else if strings.Contains(strings.ToLower(opt.Name), strings.ToLower(option)) {
			// Case-insensitive match is slightly worse
			score = 20
		} else if opt.Description != "" && strings.Contains(strings.ToLower(opt.Description), strings.ToLower(option)) {
			// Match in description is okay
			score = 30
		}

		rankedOptions = append(rankedOptions, rankedOption{
			option: optionType{
				Name:        opt.Name,
				Description: opt.Description,
				Type:        opt.Type,
				Default:     opt.Default,
				Example:     opt.Example,
				ReadOnly:    opt.ReadOnly,
				Loc:         opt.Loc,
			},
			score: score,
		})
	}

	// Sort by score
	sort.Slice(rankedOptions, func(i, j int) bool {
		return rankedOptions[i].score < rankedOptions[j].score
	})

	// If we have multiple results, format them differently
	var b strings.Builder

	if len(rankedOptions) == 1 || rankedOptions[0].score < 10 {
		// Single result or exact match - detailed format
		chosen := rankedOptions[0].option
		b.WriteString("Option: " + chosen.Name + "\n")
		b.WriteString("Type: " + chosen.Type + "\n")
		if chosen.Default != "" {
			b.WriteString("Default: " + chosen.Default + "\n")
		}
		if chosen.Example != "" {
			b.WriteString("Example: " + chosen.Example + "\n")
		}
		if chosen.Description != "" {
			b.WriteString("Description: " + chosen.Description + "\n")
		}
		if len(chosen.Loc) > 0 {
			b.WriteString("Location: " + strings.Join(chosen.Loc, ", ") + "\n")
		}
		if chosen.ReadOnly {
			b.WriteString("(Read-only option)\n")
		}
	} else {
		// Multiple results - summarized format
		b.WriteString(fmt.Sprintf("Found %d Home Manager options matching '%s':\n\n",
			len(rankedOptions), option))

		// Show top 5 results
		limit := 5
		if len(rankedOptions) < limit {
			limit = len(rankedOptions)
		}

		for i := 0; i < limit; i++ {
			opt := rankedOptions[i].option
			b.WriteString(fmt.Sprintf("%d. %s\n", i+1, opt.Name))
			b.WriteString(fmt.Sprintf("   Type: %s\n", opt.Type))
			if opt.Description != "" {
				desc := opt.Description
				if len(desc) > 80 {
					desc = desc[:77] + "..."
				}
				b.WriteString(fmt.Sprintf("   %s\n", desc))
			}
			b.WriteString("\n")
		}
	}

	return b.String(), nil
}

// fetchMySTContent handles documentation pages using MyST format like nix.dev
func fetchMySTContent(docURL string, query string) (string, error) {
	if query == "" {
		return "", fmt.Errorf("query term required for MyST documentation search")
	}

	// First try to find a specific page that might be related to the query
	// by using URL path components derived from the query terms
	parsedURL, err := url.Parse(docURL)
	if err != nil {
		return "", fmt.Errorf("invalid documentation URL: %v", err)
	}

	// Clean up query to create possible URL paths
	normalizedQuery := strings.ToLower(strings.TrimSpace(query))
	normalizedQuery = strings.ReplaceAll(normalizedQuery, " ", "-")

	// Try several possible paths based on query terms
	possiblePaths := []string{
		normalizedQuery,
		"manual/" + normalizedQuery,
		"tutorials/" + normalizedQuery,
		"concepts/" + normalizedQuery,
		"language/" + normalizedQuery,
		"reference/" + normalizedQuery,
	}

	// Results to accumulate relevant content
	var results []struct {
		Title   string
		URL     string
		Content string
	}

	client := &http.Client{Timeout: 10 * time.Second}

	// Check each possible path
	for _, path := range possiblePaths {
		// Create URL for this potential path
		pageURL := fmt.Sprintf("%s://%s/%s",
			parsedURL.Scheme,
			parsedURL.Host,
			strings.TrimPrefix(path, "/"))

		// Attempt to fetch this specific page
		resp, err := client.Get(pageURL)
		if err != nil || resp.StatusCode != http.StatusOK {
			if resp != nil {
				resp.Body.Close()
			}
			continue // Try next path
		}

		// Successfully found a page, now extract content
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			continue
		}

		// Extract title and main content from the HTML
		title := extractHtmlTitle(string(body))
		content := extractMainContent(string(body))

		// If we found content, add it to results
		if content != "" {
			results = append(results, struct {
				Title   string
				URL     string
				Content string
			}{
				Title:   title,
				URL:     pageURL,
				Content: content,
			})
		}
	}

	// If we didn't find any specific pages, try a site-wide search
	if len(results) == 0 {
		// Some documentation sites have a search.json file
		searchURL := fmt.Sprintf("%s://%s/search.json", parsedURL.Scheme, parsedURL.Host)
		resp, err := client.Get(searchURL)

		// If search.json is available, use it
		if err == nil && resp.StatusCode == http.StatusOK {
			var searchIndex struct {
				Documents []struct {
					Location string `json:"location"`
					Title    string `json:"title"`
					Text     string `json:"text"`
				} `json:"documents"`
			}

			if err := json.NewDecoder(resp.Body).Decode(&searchIndex); err == nil {
				// Search through the documents for our query
				for _, doc := range searchIndex.Documents {
					if strings.Contains(strings.ToLower(doc.Text), strings.ToLower(query)) ||
						strings.Contains(strings.ToLower(doc.Title), strings.ToLower(query)) {

						results = append(results, struct {
							Title   string
							URL     string
							Content string
						}{
							Title:   doc.Title,
							URL:     fmt.Sprintf("%s://%s%s", parsedURL.Scheme, parsedURL.Host, doc.Location),
							Content: extractRelevantSnippet(doc.Text, query),
						})

						// Limit results to top 5
						if len(results) >= 5 {
							break
						}
					}
				}
			}
			resp.Body.Close()
		}
	}

	// Format the results
	var result strings.Builder
	if len(results) == 0 {
		result.WriteString(fmt.Sprintf("No relevant documentation found for '%s' on %s\n", query, parsedURL.Host))
		result.WriteString(fmt.Sprintf("Try browsing the documentation directly at %s\n", docURL))
		return result.String(), nil
	}

	result.WriteString(fmt.Sprintf("Documentation results for '%s':\n\n", query))

	for i, entry := range results {
		result.WriteString(fmt.Sprintf("%d. %s\n", i+1, entry.Title))
		result.WriteString(fmt.Sprintf("   URL: %s\n", entry.URL))
		result.WriteString(fmt.Sprintf("   %s\n\n", entry.Content))
	}

	return result.String(), nil
}

// Helper functions for HTML/content extraction
func extractHtmlTitle(html string) string {
	titleRegex := regexp.MustCompile(`<title[^>]*>(.*?)</title>`)
	matches := titleRegex.FindStringSubmatch(html)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return "Documentation Page" // Default title if not found
}

func extractMainContent(html string) string {
	// Try to find main content section
	// This is a simplified approach - for production use, consider a proper HTML parser
	mainContentRegex := regexp.MustCompile(`<main[^>]*>([\s\S]*?)</main>`)
	matches := mainContentRegex.FindStringSubmatch(html)
	if len(matches) > 1 {
		// Clean up HTML tags and normalize whitespace
		content := stripHTMLTags(matches[1])
		content = strings.Join(strings.Fields(content), " ")
		if len(content) > 500 {
			return content[:500] + "..." // Return just the first 500 chars
		}
		return content
	}

	// If no main tag, try article or div with content/main class
	articleRegex := regexp.MustCompile(`<article[^>]*>([\s\S]*?)</article>`)
	matches = articleRegex.FindStringSubmatch(html)
	if len(matches) > 1 {
		content := stripHTMLTags(matches[1])
		content = strings.Join(strings.Fields(content), " ")
		if len(content) > 500 {
			return content[:500] + "..."
		}
		return content
	}

	// Try content divs as last resort
	contentDivRegex := regexp.MustCompile(`<div[^>]*class="[^"]*content[^"]*"[^>]*>([\s\S]*?)</div>`)
	matches = contentDivRegex.FindStringSubmatch(html)
	if len(matches) > 1 {
		content := stripHTMLTags(matches[1])
		content = strings.Join(strings.Fields(content), " ")
		if len(content) > 500 {
			return content[:500] + "..."
		}
		return content
	}

	return "Content extraction failed. Please visit the page directly."
}

func extractRelevantSnippet(text, query string) string {
	lowerText := strings.ToLower(text)
	lowerQuery := strings.ToLower(query)

	// Find position of query in text
	pos := strings.Index(lowerText, lowerQuery)
	if pos < 0 {
		// If exact query not found, try each word
		queryWords := strings.Fields(lowerQuery)
		for _, word := range queryWords {
			if pos = strings.Index(lowerText, word); pos >= 0 {
				break
			}
		}
	}

	// If still not found, just return beginning of text
	if pos < 0 {
		if len(text) > 300 {
			return text[:300] + "..."
		}
		return text
	}

	// Extract snippet around the match
	start := pos - 150
	if start < 0 {
		start = 0
	}

	end := pos + len(query) + 150
	if end > len(text) {
		end = len(text)
	}

	// Add ellipsis if we're not at the beginning/end
	prefix := ""
	if start > 0 {
		prefix = "..."
	}

	suffix := ""
	if end < len(text) {
		suffix = "..."
	}

	return prefix + text[start:end] + suffix
}
