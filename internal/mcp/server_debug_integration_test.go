package mcp

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleQuery_DebugLogging(t *testing.T) {
	var buf bytes.Buffer
	origLog := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(origLog)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/options" && r.Method == "POST" {
			w.Header().Set("Content-Type", "application/json")
			// Simulate Elasticsearch response format
			w.Write([]byte(`{"took":1,"timed_out":false,"_shards":{"total":1,"successful":1,"skipped":0,"failed":0},"hits":{"total":{"value":1,"relation":"eq"},"max_score":11.525363,"hits":[{"_index":"nixos-43-25.05","_type":"_doc","_id":"test","_score":11.525363,"_source":{"type":"option","option_source":"nixos/modules/services/web-servers/nginx/default.nix","option_name":"services.nginx.enable","option_description":"<rendered-html><p>Whether to enable the Nginx web server.</p>\n</rendered-html>","option_type":"boolean","option_default":"false","option_example":"true","option_flake":null}}]}}`))
			return
		}
		w.WriteHeader(404)
	}))
	defer ts.Close()

	srcs := []string{ts.URL + "/options"}
	srv := NewServerWithDebug("", srcs)

	req := httptest.NewRequest("GET", "/query?q=services.nginx.enable", nil)
	rw := httptest.NewRecorder()

	srv.handleQuery(rw, req)

	logOutput := buf.String()

	if !strings.Contains(logOutput, "[DEBUG] Querying documentation source") {
		t.Errorf("expected debug log for source query, got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "[DEBUG] Received") && !strings.Contains(logOutput, "bytes from NixOS ES") {
		t.Errorf("expected debug log for ES response, got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "[DEBUG] Structured doc found") {
		t.Errorf("expected debug log for structured doc, got: %s", logOutput)
	}
}
