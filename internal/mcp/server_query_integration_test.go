package mcp

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleQuery_OptionAPI(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/options" && r.Method == "POST" {
			w.Header().Set("Content-Type", "application/json")
			// Simulate Elasticsearch response format
			_, _ = w.Write([]byte(`{"took":1,"timed_out":false,"_shards":{"total":1,"successful":1,"skipped":0,"failed":0},"hits":{"total":{"value":1,"relation":"eq"},"max_score":11.525363,"hits":[{"_index":"nixos-43-25.05","_type":"_doc","_id":"test","_score":11.525363,"_source":{"type":"option","option_source":"nixos/modules/services/web-servers/nginx/default.nix","option_name":"services.nginx.enable","option_description":"<rendered-html><p>Whether to enable the Nginx web server.</p>\n</rendered-html>","option_type":"boolean","option_default":"false","option_example":"true","option_flake":null}}]}}`))
			return
		}
		w.WriteHeader(404)
	}))
	defer ts.Close()

	srcs := []string{ts.URL + "/options"}
	srv := NewServer("", srcs)

	req := httptest.NewRequest("GET", "/query?q=services.nginx.enable", nil)
	rw := httptest.NewRecorder()

	srv.handleQuery(rw, req)

	resp := rw.Body.String()
	if want := "Option: services.nginx.enable"; !contains(resp, want) {
		t.Errorf("expected response to contain %q, got: %s", want, resp)
	}
	if want := "Whether to enable Nginx Web Server."; !contains(resp, want) {
		t.Errorf("expected response to contain %q, got: %s", want, resp)
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
