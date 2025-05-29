package mcp

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleQuery_HomeManagerOptionAPI(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/options.json") {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"options":[{"name":"programs.zsh.enable","description":"Enable the Zsh shell.","type":"boolean","default":"false","example":"true","readOnly":false,"loc":["programs.zsh"]}]}`))
			return
		}
		w.WriteHeader(404)
	}))
	defer ts.Close()

	srcs := []string{ts.URL + "/options.json"}
	srv := NewServerWithDebug("", srcs)

	req := httptest.NewRequest("GET", "/query?q=programs.zsh.enable", nil)
	rw := httptest.NewRecorder()

	srv.handleQuery(rw, req)

	resp := rw.Body.String()
	if want := "Option: programs.zsh.enable"; !strings.Contains(resp, want) {
		t.Errorf("expected response to contain %q, got: %s", want, resp)
	}
	if want := "Enable the Zsh shell."; !strings.Contains(resp, want) {
		t.Errorf("expected response to contain %q, got: %s", want, resp)
	}
}
