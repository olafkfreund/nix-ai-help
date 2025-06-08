package mcp

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"nix-ai-help/pkg/logger"
	"strings"
	"testing"
)

func TestHandleQuery_DebugLogging(t *testing.T) {
	var buf bytes.Buffer
	customLogger := logger.NewLoggerWithLevelAndWriter("debug", &buf)

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
	srv := &Server{
		addr:                 "",
		socketPath:           "/tmp/nixai-mcp.sock",
		documentationSources: srcs,
		logger:               customLogger,
		debugLogging:         true,
		mcpServer:            &MCPServer{logger: *customLogger},
	}

	req := httptest.NewRequest("GET", "/query?q=services.nginx.enable", nil)
	rw := httptest.NewRecorder()

	srv.handleQuery(rw, req)

	logOutput := buf.String()

	if !strings.Contains(logOutput, "handleDocQuery: processing source") {
		t.Errorf("expected debug log for source query, got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "handleDocQuery: found result in NixOS options endpoint") {
		t.Errorf("expected debug log for structured doc, got: %s", logOutput)
	}
}
