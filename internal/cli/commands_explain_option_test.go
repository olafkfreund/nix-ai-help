package cli

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

type mockAIProvider struct{ response string }

func (m *mockAIProvider) Query(prompt string) (string, error) { return m.response, nil }

type mockMCPClient struct{ doc string }

func (m *mockMCPClient) QueryDocumentation(query string) (string, error) { return m.doc, nil }

func TestExplainOptionCmd_Mock(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	option := "services.nginx.enable"
	doc := "services.nginx.enable: Enable the nginx service. Type: boolean. Default: false."
	aiResp := "**services.nginx.enable** enables or disables the nginx web server. Type: boolean. Default: false. Best practice: enable only if you need nginx running."

	cmd := &cobra.Command{
		Use:   "explain-option <option>",
		Short: "Explain a NixOS option using AI and documentation",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 1 {
				t.Fatal("missing option arg")
			}
			mcp := &mockMCPClient{doc: doc}
			ai := &mockAIProvider{response: aiResp}
			fetched, err := mcp.QueryDocumentation(args[0])
			if err != nil {
				t.Fatal(err)
			}
			if fetched == "" {
				t.Fatal("no doc")
			}
			resp, err := ai.Query(fetched)
			if err != nil {
				t.Fatal(err)
			}
			fmt.Println(resp)
		},
	}

	cmd.SetArgs([]string{option})
	if err := cmd.Execute(); err != nil {
		w.Close()
		os.Stdout = oldStdout
		t.Fatalf("command failed: %v", err)
	}
	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()
	if !strings.Contains(output, "nginx web server") {
		t.Errorf("expected AI explanation in output, got: %s", output)
	}
}
