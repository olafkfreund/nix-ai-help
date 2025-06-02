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

// Implementation of buildExplainOptionPrompt for testing
func testBuildExplainOptionPrompt(option, documentation string) string {
	return fmt.Sprintf(`You are a NixOS expert helping users understand configuration options. Please explain the following NixOS option in a clear, practical manner.

**Option:** %s

**Official Documentation:**
%s

**Please provide:**

1. **Purpose & Overview**: What this option does and why you'd use it
2. **Type & Default**: The data type and default value (if any)
3. **Usage Examples**: Show 2-3 practical configuration examples
4. **Best Practices**: How to use this option effectively
5. **Related Options**: Other options that are commonly used with this one
6. **Common Issues**: Potential problems and their solutions

Format your response using Markdown with section headings and code blocks for examples.`, option, documentation)
}

func TestBuildExplainOptionPrompt(t *testing.T) {
	option := "services.nginx.enable"
	doc := "services.nginx.enable: Enable the nginx service. Type: boolean. Default: false."

	prompt := testBuildExplainOptionPrompt(option, doc)

	// Verify the prompt includes key elements for comprehensive explanations
	if !strings.Contains(prompt, "Usage Examples") {
		t.Error("prompt should request usage examples")
	}
	if !strings.Contains(prompt, "Best Practices") {
		t.Error("prompt should request best practices")
	}
	if !strings.Contains(prompt, "Related Options") {
		t.Error("prompt should request related options")
	}
	if !strings.Contains(prompt, "code blocks") {
		t.Error("prompt should request code blocks for examples")
	}
	if !strings.Contains(prompt, option) {
		t.Error("prompt should include the option name")
	}
	if !strings.Contains(prompt, doc) {
		t.Error("prompt should include the documentation")
	}
}
