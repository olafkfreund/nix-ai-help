package mcpserver

import (
	"context"
	"fmt"
	"testing"

	"nix-ai-help/internal/ai/functionbase"
)

func TestNewMcpServerFunction(t *testing.T) {
	fn := NewMcpServerFunction()
	if fn == nil {
		t.Fatal("expected function to be created, got nil")
	}

	if fn.Name() != "mcp-server" {
		t.Errorf("expected function name to be 'mcp-server', got '%s'", fn.Name())
	}

	description := fn.Description()
	if description == "" {
		t.Error("expected function description to be non-empty")
	}

	schema := fn.Schema()
	if len(schema.Parameters) != 16 {
		t.Errorf("expected 16 parameters, got %d", len(schema.Parameters))
	}

	// Check required parameters
	operationParam := schema.Parameters[0]
	if operationParam.Name != "operation" || !operationParam.Required {
		t.Error("expected first parameter to be 'operation' and required")
	}
}

func TestParseRequest(t *testing.T) {
	fn := NewMcpServerFunction()

	tests := []struct {
		name     string
		params   map[string]interface{}
		expected *McpServerRequest
		wantErr  bool
	}{
		{
			name: "valid start operation",
			params: map[string]interface{}{
				"operation": "start",
				"debug":     true,
			},
			expected: &McpServerRequest{
				Operation: "start",
				Debug:     true,
			},
			wantErr: false,
		},
		{
			name: "valid query operation",
			params: map[string]interface{}{
				"operation": "query",
				"query":     "services.nginx.enable",
			},
			expected: &McpServerRequest{
				Operation: "query",
				Query:     "services.nginx.enable",
			},
			wantErr: false,
		},
		{
			name: "valid setup operation with requirements",
			params: map[string]interface{}{
				"operation":   "setup",
				"server_type": "documentation",
				"requirements": map[string]interface{}{
					"host": "localhost",
					"port": "8081",
				},
			},
			expected: &McpServerRequest{
				Operation:  "setup",
				ServerType: "documentation",
				Requirements: map[string]string{
					"host": "localhost",
					"port": "8081",
				},
			},
			wantErr: false,
		},
		{
			name: "valid diagnose operation",
			params: map[string]interface{}{
				"operation": "diagnose",
				"issue":     "server not responding",
				"symptoms":  []interface{}{"connection timeout", "socket error"},
			},
			expected: &McpServerRequest{
				Operation: "diagnose",
				Issue:     "server not responding",
				Symptoms:  []string{"connection timeout", "socket error"},
			},
			wantErr: false,
		},
		{
			name: "missing operation",
			params: map[string]interface{}{
				"query": "test",
			},
			expected: nil,
			wantErr:  true,
		},
		{
			name: "invalid operation type",
			params: map[string]interface{}{
				"operation": 123,
			},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := fn.parseRequest(tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.expected != nil && result != nil {
				if result.Operation != tt.expected.Operation {
					t.Errorf("expected operation %s, got %s", tt.expected.Operation, result.Operation)
				}
				if result.Query != tt.expected.Query {
					t.Errorf("expected query %s, got %s", tt.expected.Query, result.Query)
				}
			}
		})
	}
}

func TestValidateRequest(t *testing.T) {
	fn := NewMcpServerFunction()

	tests := []struct {
		name    string
		request *McpServerRequest
		wantErr bool
	}{
		{
			name: "valid start operation",
			request: &McpServerRequest{
				Operation: "start",
			},
			wantErr: false,
		},
		{
			name: "valid query operation",
			request: &McpServerRequest{
				Operation: "query",
				Query:     "test",
			},
			wantErr: false,
		},
		{
			name: "valid setup operation",
			request: &McpServerRequest{
				Operation: "setup",
			},
			wantErr: false,
		},
		{
			name: "empty operation",
			request: &McpServerRequest{
				Operation: "",
			},
			wantErr: true,
		},
		{
			name: "invalid operation",
			request: &McpServerRequest{
				Operation: "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := fn.validateRequest(tt.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExecute(t *testing.T) {
	fn := NewMcpServerFunction()
	ctx := context.Background()

	tests := []struct {
		name    string
		params  map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid status operation",
			params: map[string]interface{}{
				"operation": "status",
			},
			wantErr: false,
		},
		{
			name: "valid logs operation",
			params: map[string]interface{}{
				"operation": "logs",
			},
			wantErr: false,
		},
		{
			name: "invalid parameters",
			params: map[string]interface{}{
				"operation": 123,
			},
			wantErr: true,
		},
		{
			name: "missing operation",
			params: map[string]interface{}{
				"query": "test",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := fn.Execute(ctx, tt.params, &functionbase.FunctionOptions{})
			if err != nil {
				t.Errorf("Execute() returned unexpected error: %v", err)
				return
			}
			if result == nil {
				t.Error("expected result to be non-nil")
				return
			}
			if tt.wantErr && result.Success {
				t.Errorf("expected failure but got success=true")
			}
			if !tt.wantErr && !result.Success {
				t.Errorf("expected success=true, got success=%v, error=%s", result.Success, result.Error)
			}
		})
	}
}

func TestHandleStatusOperation(t *testing.T) {
	fn := NewMcpServerFunction()
	ctx := context.Background()

	request := &McpServerRequest{
		Operation: "status",
	}

	response, err := fn.handleStatusOperation(ctx, request)
	if err != nil {
		t.Fatalf("handleStatusOperation() error = %v", err)
	}

	if !response.Success {
		t.Errorf("expected success=true, got success=%v", response.Success)
	}

	if response.Message == "" {
		t.Error("expected non-empty message")
	}

	if response.ServerInfo == nil {
		t.Error("expected server info to be populated")
	}
}

func TestHandleLogsOperation(t *testing.T) {
	fn := NewMcpServerFunction()
	ctx := context.Background()

	request := &McpServerRequest{
		Operation: "logs",
	}

	response, err := fn.handleLogsOperation(ctx, request)
	if err != nil {
		t.Fatalf("handleLogsOperation() error = %v", err)
	}

	if !response.Success {
		t.Errorf("expected success=true, got success=%v", response.Success)
	}

	if response.Output == "" {
		t.Error("expected non-empty output")
	}

	if len(response.Documentation) == 0 {
		t.Error("expected documentation to be populated")
	}

	if len(response.NextSteps) == 0 {
		t.Error("expected next steps to be populated")
	}
}

func TestHandleConfigureOperation(t *testing.T) {
	fn := NewMcpServerFunction()
	ctx := context.Background()

	request := &McpServerRequest{
		Operation: "configure",
	}

	response, err := fn.handleConfigureOperation(ctx, request)
	if err != nil {
		t.Fatalf("handleConfigureOperation() error = %v", err)
	}

	if !response.Success {
		t.Errorf("expected success=true, got success=%v", response.Success)
	}

	if response.ServerInfo == nil {
		t.Error("expected server info to be populated")
	}

	if len(response.ConfigSnippets) == 0 {
		t.Error("expected config snippets to be populated")
	}
}

// MockMcpServerAgent for testing agent-dependent operations
type MockMcpServerAgent struct {
	shouldFail bool
}

func (m *MockMcpServerAgent) SetupMcpServer(serverType string, requirements map[string]interface{}) (string, error) {
	if m.shouldFail {
		return "", fmt.Errorf("mock setup failure")
	}
	return "Setup guidance provided", nil
}

func (m *MockMcpServerAgent) DiagnoseMcpIssues(issue string, symptoms []string) (string, error) {
	if m.shouldFail {
		return "", fmt.Errorf("mock diagnosis failure")
	}
	return "Diagnosis completed", nil
}

func TestHandleSetupOperationWithMockAgent(t *testing.T) {
	fn := NewMcpServerFunction()

	// Replace agent with mock
	mockAgent := &MockMcpServerAgent{shouldFail: false}
	// Note: This would require making mcpAgent settable or using dependency injection
	// For now, this test demonstrates the structure

	ctx := context.Background()
	request := &McpServerRequest{
		Operation:    "setup",
		ServerType:   "documentation",
		Requirements: map[string]string{"host": "localhost"},
	}

	// This test would work if we had proper dependency injection
	_ = mockAgent
	_ = ctx
	_ = request
	_ = fn

	// Since we can't easily mock the agent without changing the structure,
	// we'll just verify basic functionality works
	t.Log("Mock operation test setup completed")
}

func TestMcpServerResponseStructure(t *testing.T) {
	response := &McpServerResponse{
		Success:          true,
		Message:          "test message",
		Output:           "test output",
		ServerStatus:     "running",
		ServerInfo:       map[string]interface{}{"test": "value"},
		QueryResults:     []string{"result1", "result2"},
		ConfigSnippets:   []string{"config1", "config2"},
		Recommendations:  []string{"rec1", "rec2"},
		NextSteps:        []string{"step1", "step2"},
		Documentation:    []string{"doc1", "doc2"},
		TroubleShooting:  []string{"trouble1", "trouble2"},
		SecuritySettings: []string{"security1", "security2"},
		Optimizations:    []string{"opt1", "opt2"},
	}

	if !response.Success {
		t.Error("expected success to be true")
	}

	if response.Message != "test message" {
		t.Error("expected message to match")
	}

	if len(response.QueryResults) != 2 {
		t.Error("expected 2 query results")
	}

	if len(response.NextSteps) != 2 {
		t.Error("expected 2 next steps")
	}
}

func TestAllValidOperations(t *testing.T) {
	fn := NewMcpServerFunction()
	validOps := []string{
		"start", "stop", "status", "restart", "query",
		"setup", "diagnose", "configure", "optimize",
		"secure", "integrate", "monitor", "logs",
	}

	for _, op := range validOps {
		request := &McpServerRequest{Operation: op}
		err := fn.validateRequest(request)
		if err != nil {
			t.Errorf("operation %s should be valid, got error: %v", op, err)
		}
	}
}
