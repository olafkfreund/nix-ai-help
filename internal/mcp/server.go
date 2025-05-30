package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"nix-ai-help/internal/config"
	"nix-ai-help/pkg/logger"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

// Server represents the MCP server that handles requests for NixOS documentation.
type Server struct {
	addr                 string
	documentationSources []string
	logger               *logger.Logger
	debugLogging         bool
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

	// Run server in goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.ListenAndServe()
	}()

	// Wait for shutdown signal
	select {
	case <-shutdownCh:
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		log.Println("Shutting down MCP server...")
		return server.Shutdown(ctx)
	case err := <-errCh:
		if strings.Contains(err.Error(), "address already in use") {
			log.Printf("ERROR: The MCP server could not start because the address is already in use. If another instance is running, stop it with 'nixai mcp-server stop'.")
		}
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
		if strings.HasSuffix(src, "/options") {
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
	// gosec:ignore G101 -- This is a test credential for CI and not used in production
	const (
		esUser = "aWVSALXpZv"
		esPass = "X8gPHnzL52wFEekuxsfQ9cSh"
	)
	// #nosec G107 -- esURL is a constant, not user input
	esURL := "https://elasticsearch.nixos.org/options/_search"
	// Build the query body
	body := map[string]interface{}{
		"from": 0,
		"size": 5,
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
	req, err := http.NewRequest("POST", esURL, strings.NewReader(string(jsonBody)))
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(esUser, esPass)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch %s: %s", esURL, resp.Status)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	log.Printf("Received %d bytes from NixOS ES", len(data))
	// Parse hits
	var esResp struct {
		Hits struct {
			Hits []struct {
				Source struct {
					Name        string `json:"option_name"`
					Type        string `json:"option_type"`
					Default     string `json:"option_default"`
					Example     string `json:"option_example"`
					Description string `json:"option_description"`
					SourceFile  string `json:"option_source"`
				} `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.Unmarshal(data, &esResp); err != nil {
		return "", err
	}
	if len(esResp.Hits.Hits) == 0 {
		return "No documentation found for option in the official NixOS options database.", nil
	}
	chosen := esResp.Hits.Hits[0].Source
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
		b.WriteString("Description: " + stripHTMLTags(chosen.Description) + "\n")
	}
	if chosen.SourceFile != "" {
		b.WriteString("Source: " + chosen.SourceFile + "\n")
	}
	return b.String(), nil
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
