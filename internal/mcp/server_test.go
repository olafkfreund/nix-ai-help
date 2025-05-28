package mcp

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleQuery_FuzzySearch(t *testing.T) {
	docs := []string{
		"https://raw.githubusercontent.com/NixOS/nixpkgs/master/README.md", // should contain 'nixpkgs'
	}
	s := NewServer(":0", docs)
	req := httptest.NewRequest("GET", "/query?q=nixpkgs", nil)
	w := httptest.NewRecorder()

	s.handleQuery(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}
	body := w.Body.String()
	if !strings.Contains(strings.ToLower(body), "nixpkgs") {
		t.Errorf("expected result to contain 'nixpkgs', got: %s", body)
	}
}
