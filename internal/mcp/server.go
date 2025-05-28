package mcp

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"nix-ai-help/internal/config"
	"sort"
	"strings"
	"sync"
	"time"
)

// Server represents the MCP server that handles requests for NixOS documentation.
type Server struct {
	addr                 string
	documentationSources []string
}

// Add a simple in-memory cache for query results
var (
	cache      = make(map[string]string)
	cacheMutex sync.RWMutex
)

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
	query := r.URL.Query().Get("q")
	if query == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "Missing 'q' query parameter.")
		return
	}

	// Check cache first
	cacheMutex.RLock()
	if cached, ok := cache[query]; ok {
		cacheMutex.RUnlock()
		fmt.Fprint(w, cached)
		return
	}
	cacheMutex.RUnlock()

	type result struct {
		line   string
		source string
		score  int // lower is better
	}
	var results []result
	for _, src := range s.documentationSources {
		body, err := fetchDocSource(src)
		if err != nil {
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
		fmt.Fprintln(w, "No relevant documentation found.")
		return
	}

	// Sort by score (best match first)
	sort.Slice(results, func(i, j int) bool {
		return results[i].score < results[j].score
	})

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

	fmt.Fprint(w, response)
}

func fetchDocSource(url string) (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch %s: %s", url, resp.Status)
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
