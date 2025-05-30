package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"nix-ai-help/internal/config"
	"nix-ai-help/pkg/logger"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

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
	logger   logger.Logger
	listener net.Listener
	mu       sync.Mutex
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

// Handle processes MCP protocol requests
func (m *MCPServer) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	log.Printf("[DEBUG] Handle called with method: %s, ID: %v", req.Method, req.ID)
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
		conn.Reply(ctx, req.ID, result)

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
		}
		conn.Reply(ctx, req.ID, map[string]interface{}{"tools": tools})

	case "tools/call":
		var params struct {
			Name      string                 `json:"name"`
			Arguments map[string]interface{} `json:"arguments"`
		}

		if err := json.Unmarshal(*req.Params, &params); err != nil {
			conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{
				Code:    jsonrpc2.CodeInvalidParams,
				Message: "Invalid parameters",
			})
			return
		}

		switch params.Name {
		case "query_nixos_docs":
			if query, ok := params.Arguments["query"].(string); ok {
				result := m.handleDocQuery(query)
				conn.Reply(ctx, req.ID, map[string]interface{}{
					"content": []map[string]interface{}{
						{
							"type": "text",
							"text": result,
						},
					},
				})
			} else {
				conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{
					Code:    jsonrpc2.CodeInvalidParams,
					Message: "Missing query parameter",
				})
			}

		case "explain_nixos_option":
			if option, ok := params.Arguments["option"].(string); ok {
				result := m.handleOptionExplain(option)
				conn.Reply(ctx, req.ID, map[string]interface{}{
					"content": []map[string]interface{}{
						{
							"type": "text",
							"text": result,
						},
					},
				})
			} else {
				conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{
					Code:    jsonrpc2.CodeInvalidParams,
					Message: "Missing option parameter",
				})
			}

		case "explain_home_manager_option":
			if option, ok := params.Arguments["option"].(string); ok {
				result := m.handleHomeManagerOptionExplain(option)
				conn.Reply(ctx, req.ID, map[string]interface{}{
					"content": []map[string]interface{}{
						{
							"type": "text",
							"text": result,
						},
					},
				})
			} else {
				conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{
					Code:    jsonrpc2.CodeInvalidParams,
					Message: "Missing option parameter",
				})
			}

		case "search_nixos_packages":
			if query, ok := params.Arguments["query"].(string); ok {
				result := m.handlePackageSearch(query)
				conn.Reply(ctx, req.ID, map[string]interface{}{
					"content": []map[string]interface{}{
						{
							"type": "text",
							"text": result,
						},
					},
				})
			} else {
				conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{
					Code:    jsonrpc2.CodeInvalidParams,
					Message: "Missing query parameter",
				})
			}

		default:
			conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{
				Code:    jsonrpc2.CodeMethodNotFound,
				Message: "Unknown tool: " + params.Name,
			})
		}

	default:
		conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{
			Code:    jsonrpc2.CodeMethodNotFound,
			Message: "Method not found: " + req.Method,
		})
	}
}

// Start starts the MCP server on Unix socket
func (m *MCPServer) Start(socketPath string) error {
	// Remove existing socket file if it exists
	os.Remove(socketPath)

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return fmt.Errorf("failed to listen on Unix socket %s: %v", socketPath, err)
	}

	// Store listener for cleanup
	m.mu.Lock()
	m.listener = listener
	m.mu.Unlock()

	m.logger.Info(fmt.Sprintf("MCP server listening on Unix socket: %s", socketPath))

	// Accept connections in a blocking loop
	for {
		conn, err := listener.Accept()
		if err != nil {
			m.logger.Error(fmt.Sprintf("Failed to accept connection: %v", err))
			continue
		}

		go func(conn net.Conn) {
			defer conn.Close()
			log.Printf("[DEBUG] New MCP client connected from %v", conn.RemoteAddr())

			// Handle connection with JSON-RPC2
			stream := jsonrpc2.NewPlainObjectStream(conn)
			log.Printf("[DEBUG] Created buffered stream")

			jsonConn := jsonrpc2.NewConn(context.Background(), stream, m)
			log.Printf("[DEBUG] Created JSON-RPC2 connection")
			defer jsonConn.Close()

			// Keep connection alive
			log.Printf("[DEBUG] Waiting for disconnect notification...")
			<-jsonConn.DisconnectNotify()
			log.Printf("[DEBUG] MCP client disconnected")
		}(conn)
	}
}

// Stop stops the MCP server
func (m *MCPServer) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.listener != nil {
		m.listener.Close()
		m.listener = nil
	}
}

// handleDocQuery processes documentation queries
func (m *MCPServer) handleDocQuery(query string) string {
	// Use existing QueryDocumentation logic with the configured server
	client := NewMCPClient("http://localhost:8081") // Use correct port
	result, err := client.QueryDocumentation(query)
	if err != nil {
		return fmt.Sprintf("Error querying documentation: %v", err)
	}
	return result
}

// handleOptionExplain processes NixOS option explanations
func (m *MCPServer) handleOptionExplain(option string) string {
	// Use the same logic as the HTTP server
	client := NewMCPClient("http://localhost:8081")
	result, err := client.QueryDocumentation(option)
	if err != nil {
		return fmt.Sprintf("Error explaining option %s: %v", option, err)
	}
	return result
}

// handleHomeManagerOptionExplain processes Home Manager option explanations
func (m *MCPServer) handleHomeManagerOptionExplain(option string) string {
	// Use the same logic as the HTTP server for Home Manager options
	client := NewMCPClient("http://localhost:8081")
	result, err := client.QueryDocumentation(option)
	if err != nil {
		return fmt.Sprintf("Error explaining Home Manager option %s: %v", option, err)
	}
	return result
}

// handlePackageSearch processes package search queries
func (m *MCPServer) handlePackageSearch(query string) string {
	return fmt.Sprintf("Package search for '%s' is not yet implemented in MCP protocol. Use the CLI interface: nixai search pkg %s", query, query)
}

// Server represents the combined HTTP and MCP server
type Server struct {
	addr                 string
	documentationSources []string
	logger               *logger.Logger
	debugLogging         bool
	mcpServer            *MCPServer
}

// Add a simple in-memory cache for query results
var (
	cache      = make(map[string]string)
	cacheMutex sync.RWMutex
)

// NewServer creates a new MCP server instance with documentation sources.
func NewServer(addr string, documentationSources []string) *Server {
	return &Server{
		addr:                 addr,
		documentationSources: documentationSources,
		logger:               logger.NewLoggerWithLevel("info"), // Default to info level
		debugLogging:         false,
		mcpServer:            &MCPServer{logger: *logger.NewLoggerWithLevel("info")},
	}
}

// NewServerWithDebug creates a new MCP server instance with debug logging enabled.
// This is primarily intended for testing purposes.
func NewServerWithDebug(addr string, documentationSources []string) *Server {
	return &Server{
		addr:                 addr,
		documentationSources: documentationSources,
		logger:               logger.NewLoggerWithLevel("debug"), // Enable debug level
		debugLogging:         true,
		mcpServer:            &MCPServer{logger: *logger.NewLoggerWithLevel("debug")},
	}
}

// NewServerFromConfig creates a new MCP server from a YAML config file.
func NewServerFromConfig(configPath string) (*Server, error) {
	cfg, err := config.LoadYAMLConfig(configPath)
	if err != nil {
		return nil, err
	}
	addr := fmt.Sprintf("%s:%d", cfg.MCPServer.Host, cfg.MCPServer.Port)
	return &Server{
		addr:                 addr,
		documentationSources: cfg.MCPServer.DocumentationSources,
		logger:               logger.NewLoggerWithLevel(cfg.LogLevel),
		debugLogging:         strings.ToLower(cfg.LogLevel) == "debug",
		mcpServer:            &MCPServer{logger: *logger.NewLoggerWithLevel(cfg.LogLevel)},
	}, nil
}

// Start initializes and starts the MCP server with graceful shutdown support.
func (s *Server) Start() error {
	// Redirect log output to stderr to avoid polluting HTTP responses
	log.SetOutput(os.Stderr)
	mux := http.NewServeMux()
	mux.HandleFunc("/query", s.handleQuery)

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	shutdownCh := make(chan struct{})
	mux.HandleFunc("/shutdown", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Shutting down MCP server...\n"))
		go func() {
			shutdownCh <- struct{}{}
		}()
	})

	server := &http.Server{
		Addr:    s.addr,
		Handler: mux,
	}

	log.Printf("Starting MCP server on %s\n", s.addr)

	// Run HTTP server in goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.ListenAndServe()
	}()

	// Run MCP server in goroutine - but don't capture its result
	// since the MCP server runs indefinitely and should not exit
	go func() {
		// Load config to get socket path, fallback to default
		cfg, err := config.LoadYAMLConfig("configs/default.yaml")
		socketPath := "/tmp/nixai-mcp.sock" // Default
		if err == nil && cfg.MCPServer.SocketPath != "" {
			socketPath = cfg.MCPServer.SocketPath
		}

		// Start the MCP server (this blocks and shouldn't return unless there's an error)
		if err := s.mcpServer.Start(socketPath); err != nil {
			log.Printf("ERROR: MCP server encountered an error: %v", err)
			// Don't exit the main server if the MCP server exits - just log the error
		}
	}()

	// Wait for shutdown signal or HTTP server error
	select {
	case <-shutdownCh:
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		log.Println("Shutting down MCP server...")
		s.mcpServer.Stop()
		return server.Shutdown(ctx)
	case err := <-errCh:
		if strings.Contains(err.Error(), "address already in use") {
			log.Printf("ERROR: The MCP server could not start because the address is already in use. If another instance is running, stop it with 'nixai mcp-server stop'.")
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

	// Handle both GET requests with 'q' parameter and POST requests with JSON body
	if r.Method == "GET" {
		query = r.URL.Query().Get("q")
		if query == "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "Missing 'q' query parameter.")
			return
		}
	} else if r.Method == "POST" {
		var requestBody struct {
			Query string `json:"query"`
		}
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "Invalid JSON body.")
			return
		}
		query = requestBody.Query
		if query == "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "Missing 'query' field in JSON body.")
			return
		}
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintln(w, "Method not allowed. Use GET or POST.")
		return
	}

	if s.debugLogging {
		log.Printf("[DEBUG] handleQuery: received query: %s", query)
	}

	// Helper to write JSON response
	writeJSON := func(result string) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"result": result})
	}

	// Check cache first
	cacheMutex.RLock()
	if cached, ok := cache[query]; ok {
		cacheMutex.RUnlock()
		writeJSON(cached)
		return
	}
	cacheMutex.RUnlock()

	type result struct {
		line   string
		source string
		score  int // lower is better
	}
	var results []result
	var structuredNoDoc bool
	for _, src := range s.documentationSources {
		if s.debugLogging {
			log.Printf("[DEBUG] Querying documentation source: %s for option: %s", src, query)
		}
		var body string
		var err error
		isStructured := false
		if strings.HasPrefix(src, "nixos-options-es://") {
			body, err = fetchNixOSOptionsAPI(src, query)
			isStructured = true
		} else if strings.HasSuffix(src, "/options") {
			body, err = fetchNixOSOptionsAPI(src, query)
			isStructured = true
		} else if strings.HasSuffix(src, "/options.json") {
			body, err = fetchHomeManagerOptionsAPI(src, query)
			isStructured = true
		} else {
			body, err = fetchDocSource(src)
		}
		if err != nil {
			if s.debugLogging {
				log.Printf("[DEBUG] Error querying source %s: %v", src, err)
			}
			continue
		}
		if isStructured {
			clean := strings.TrimSpace(body)
			if clean != "" && !strings.HasPrefix(clean, "No documentation found") {
				if s.debugLogging {
					log.Printf("[DEBUG] Structured doc found from %s: %s", src, clean)
				}
				// Return immediately if a structured doc is found
				cacheMutex.Lock()
				cache[query] = clean
				cacheMutex.Unlock()
				writeJSON(clean)
				return
			} else if strings.HasPrefix(clean, "No documentation found") {
				if s.debugLogging {
					log.Printf("[DEBUG] No documentation found for %s in %s", query, src)
				}
				structuredNoDoc = true
			}
			continue
		}
		for _, line := range strings.Split(body, "\n") {
			clean := strings.TrimSpace(line)
			if clean == "" {
				continue
			}
			// Fuzzy match: score by Levenshtein distance to query, prefer substring matches
			lowerLine := strings.ToLower(clean)
			lowerQuery := strings.ToLower(query)
			score := levenshtein(lowerLine, lowerQuery)
			if strings.Contains(lowerLine, lowerQuery) {
				score -= 5 // prefer direct substring matches
			}
			if score < 20 { // only keep reasonably close matches
				results = append(results, result{line: clean, source: src, score: score})
			}
		}
	}

	if len(results) == 0 {
		if structuredNoDoc {
			writeJSON("No documentation found for this option in the official NixOS/Home Manager option databases.")
			return
		}
		writeJSON("No relevant documentation found.")
		return
	}

	// If any structured doc (score==0, isStructured) exists, return only those as the response
	var structuredResults []result
	for _, r := range results {
		if r.score == 0 && strings.HasPrefix(r.source, "https://search.nixos.org/options") {
			structuredResults = append(structuredResults, r)
		}
	}
	if len(structuredResults) > 0 {
		// Always return the first structured doc block as the response
		response := structuredResults[0].line
		cacheMutex.Lock()
		cache[query] = response
		cacheMutex.Unlock()
		writeJSON(response)
		return
	}

	// Show top 10 ranked results
	maxResults := 10
	if len(results) < maxResults {
		maxResults = len(results)
	}
	var out []string
	for i := 0; i < maxResults; i++ {
		out = append(out, fmt.Sprintf("%s: %s", results[i].source, results[i].line))
	}
	response := strings.Join(out, "\n---\n")

	// Cache result
	cacheMutex.Lock()
	cache[query] = response
	cacheMutex.Unlock()

	writeJSON(response)
}

func fetchDocSource(urlStr string) (string, error) {
	if strings.HasSuffix(urlStr, "/options") {
		return fetchNixOSOptionsAPI(urlStr, "")
	}
	if strings.HasSuffix(urlStr, "/options.json") {
		return fetchHomeManagerOptionsAPI(urlStr, "")
	}
	client := &http.Client{Timeout: 10 * time.Second}
	// #nosec G107 -- urlStr is from trusted config/documentation sources only
	resp, err := client.Get(urlStr)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch %s: %s", urlStr, resp.Status)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(data), nil
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
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ElasticSearch returned status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Debug logging for the response size
	log.Printf("[DEBUG] Received %d bytes from NixOS ES", len(data))

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
func fetchHomeManagerOptionsAPI(baseURL, option string) (string, error) {
	if strings.TrimSpace(option) == "" {
		return "", fmt.Errorf("option name required")
	}
	apiURL := baseURL + "?query=" + url.QueryEscape(option)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(apiURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
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
	// Prefer exact match, else first result
	chosen := result.Options[0]
	for _, opt := range result.Options {
		if opt.Name == option {
			chosen = opt
			break
		}
	}
	var b strings.Builder
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
	return b.String(), nil
}
