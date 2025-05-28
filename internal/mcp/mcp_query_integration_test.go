package mcp

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMCPQuery_WikiSources(t *testing.T) {
	docs := []string{
		"https://wiki.nixos.org/wiki/NixOS_Wiki",
		"https://nix.dev/manual/nix",
		"https://nixos.org/manual/nixpkgs/stable/",
		"https://nix.dev/manual/nix/2.28/language/",
		"https://nix-community.github.io/home-manager/",
	}
	s := NewServer(":0", docs)

	// Use a common NixOS keyword that should appear in at least one doc
	req := httptest.NewRequest("GET", "/query?q=configuration", nil)
	w := httptest.NewRecorder()

	s.handleQuery(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}
	body := w.Body.String()
	found := false
	for _, src := range docs {
		if strings.Contains(body, src) {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected at least one wiki/manual source in response, got: %s", body)
	}
}
