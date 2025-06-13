package agent

import (
	"context"
	"errors"
	"testing"
)

func TestNewMcpServerAgent(t *testing.T) {
	provider := &MockProvider{}
	agent := NewMcpServerAgent(provider)

	if agent == nil {
		t.Fatal("Expected agent to be created")
	}

	if agent.provider != provider {
		t.Error("Expected provider to be set")
	}
}

func TestMcpServerAgent_SetContext(t *testing.T) {
	agent := NewMcpServerAgent(&MockProvider{})

	// Test valid context
	ctx := &McpServerContext{
		AvailableServers: []string{"server1", "server2"},
		ServerStatus:     map[string]string{"server1": "running", "server2": "stopped"},
		ErrorLogs:        []string{"connection timeout", "auth failed"},
		ServerConfig:     map[string]interface{}{"port": 8080, "host": "localhost"},
		SecuritySettings: map[string]interface{}{"auth": "enabled", "ssl": true},
	}

	err := agent.SetContext(ctx)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Test invalid context
	err = agent.SetContext("invalid")
	if err == nil {
		t.Error("Expected error for invalid context type")
	}

	// Test nil context
	err = agent.SetContext(nil)
	if err != nil {
		t.Errorf("Expected no error for nil context, got: %v", err)
	}
}

func TestMcpServerAgent_Query(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		provider *MockProvider
		wantErr  bool
	}{
		{
			name:     "successful MCP server query",
			input:    "How do I set up an MCP server?",
			provider: &MockProvider{response: "Mock response"},
			wantErr:  false,
		},
		{
			name:     "MCP configuration query",
			input:    "What are the best practices for MCP server security?",
			provider: &MockProvider{response: "Mock response"},
			wantErr:  false,
		},
		{
			name:     "provider error",
			input:    "Help with MCP troubleshooting",
			provider: &MockProvider{err: errors.New("mock error")},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := NewMcpServerAgent(tt.provider)

			_, err := agent.Query(context.Background(), tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("Query() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMcpServerAgent_GenerateResponse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		provider *MockProvider
		wantErr  bool
	}{
		{
			name:     "generate MCP setup guide",
			input:    "Create a complete MCP server setup guide",
			provider: &MockProvider{response: "Mock response"},
			wantErr:  false,
		},
		{
			name:     "generate troubleshooting guide",
			input:    "Help diagnose MCP server connection issues",
			provider: &MockProvider{response: "Mock response"},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := NewMcpServerAgent(tt.provider)

			_, err := agent.GenerateResponse(context.Background(), tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateResponse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMcpServerAgent_SetupMcpServer(t *testing.T) {
	tests := []struct {
		name         string
		serverType   string
		requirements map[string]interface{}
		provider     *MockProvider
		wantErr      bool
	}{
		{
			name:       "setup documentation server",
			serverType: "documentation",
			requirements: map[string]interface{}{
				"sources": []string{"nixos.org", "nix.dev"},
				"auth":    false,
			},
			provider: &MockProvider{response: "Mock response"},
			wantErr:  false,
		},
		{
			name:       "setup secure server",
			serverType: "secure",
			requirements: map[string]interface{}{
				"auth":    true,
				"ssl":     true,
				"logging": "detailed",
			},
			provider: &MockProvider{response: "Mock response"},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := NewMcpServerAgent(tt.provider)

			_, err := agent.SetupMcpServer(tt.serverType, tt.requirements)

			if (err != nil) != tt.wantErr {
				t.Errorf("SetupMcpServer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMcpServerAgent_DiagnoseMcpIssues(t *testing.T) {
	tests := []struct {
		name     string
		issue    string
		symptoms []string
		provider *MockProvider
		wantErr  bool
	}{
		{
			name:     "connection issues",
			issue:    "server not responding",
			symptoms: []string{"timeout errors", "connection refused"},
			provider: &MockProvider{response: "Mock response"},
			wantErr:  false,
		},
		{
			name:     "authentication problems",
			issue:    "authentication failures",
			symptoms: []string{"401 errors", "invalid credentials"},
			provider: &MockProvider{response: "Mock response"},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := NewMcpServerAgent(tt.provider)

			_, err := agent.DiagnoseMcpIssues(tt.issue, tt.symptoms)

			if (err != nil) != tt.wantErr {
				t.Errorf("DiagnoseMcpIssues() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMcpServerAgent_OptimizeMcpPerformance(t *testing.T) {
	tests := []struct {
		name             string
		performanceGoals []string
		currentMetrics   map[string]interface{}
		provider         *MockProvider
		wantErr          bool
	}{
		{
			name:             "optimize response time",
			performanceGoals: []string{"reduce latency", "increase throughput"},
			currentMetrics: map[string]interface{}{
				"avg_response_time": "500ms",
				"requests_per_sec":  100,
			},
			provider: &MockProvider{response: "Mock response"},
			wantErr:  false,
		},
		{
			name:             "optimize memory usage",
			performanceGoals: []string{"reduce memory usage", "improve caching"},
			currentMetrics: map[string]interface{}{
				"memory_usage":    "2GB",
				"cache_hit_ratio": 0.7,
			},
			provider: &MockProvider{response: "Mock response"},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := NewMcpServerAgent(tt.provider)

			_, err := agent.OptimizeMcpPerformance(tt.performanceGoals, tt.currentMetrics)

			if (err != nil) != tt.wantErr {
				t.Errorf("OptimizeMcpPerformance() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMcpServerAgent_ManageMcpSecurity(t *testing.T) {
	tests := []struct {
		name             string
		securityConcerns []string
		currentConfig    map[string]interface{}
		provider         *MockProvider
		wantErr          bool
	}{
		{
			name:             "enhance authentication",
			securityConcerns: []string{"weak authentication", "no rate limiting"},
			currentConfig: map[string]interface{}{
				"auth_method": "basic",
				"rate_limit":  "none",
			},
			provider: &MockProvider{response: "Mock response"},
			wantErr:  false,
		},
		{
			name:             "improve encryption",
			securityConcerns: []string{"no SSL", "weak ciphers"},
			currentConfig: map[string]interface{}{
				"ssl":          false,
				"cipher_suite": "default",
			},
			provider: &MockProvider{response: "Mock response"},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := NewMcpServerAgent(tt.provider)

			_, err := agent.ManageMcpSecurity(tt.securityConcerns, tt.currentConfig)

			if (err != nil) != tt.wantErr {
				t.Errorf("ManageMcpSecurity() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMcpServerAgent_IntegrateMcpServer(t *testing.T) {
	tests := []struct {
		name                    string
		targetApp               string
		integrationRequirements []string
		provider                *MockProvider
		wantErr                 bool
	}{
		{
			name:      "integrate with VS Code",
			targetApp: "vscode",
			integrationRequirements: []string{
				"extension support",
				"real-time updates",
				"authentication",
			},
			provider: &MockProvider{response: "Mock response"},
			wantErr:  false,
		},
		{
			name:      "integrate with CLI tool",
			targetApp: "nixai",
			integrationRequirements: []string{
				"command-line interface",
				"configuration file",
				"error handling",
			},
			provider: &MockProvider{response: "Mock response"},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := NewMcpServerAgent(tt.provider)

			_, err := agent.IntegrateMcpServer(tt.targetApp, tt.integrationRequirements)

			if (err != nil) != tt.wantErr {
				t.Errorf("IntegrateMcpServer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMcpServerAgent_MonitorMcpServer(t *testing.T) {
	tests := []struct {
		name            string
		monitoringScope []string
		alertingNeeds   []string
		provider        *MockProvider
		wantErr         bool
	}{
		{
			name:            "comprehensive monitoring",
			monitoringScope: []string{"performance", "security", "availability"},
			alertingNeeds:   []string{"email", "slack", "pagerduty"},
			provider:        &MockProvider{response: "Mock response"},
			wantErr:         false,
		},
		{
			name:            "basic monitoring",
			monitoringScope: []string{"uptime", "errors"},
			alertingNeeds:   []string{"email"},
			provider:        &MockProvider{response: "Mock response"},
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := NewMcpServerAgent(tt.provider)

			_, err := agent.MonitorMcpServer(tt.monitoringScope, tt.alertingNeeds)

			if (err != nil) != tt.wantErr {
				t.Errorf("MonitorMcpServer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMcpServerAgent_formatMcpServerContext(t *testing.T) {
	tests := []struct {
		name    string
		context interface{}
		want    string
	}{
		{
			name:    "nil context",
			context: nil,
			want:    "No specific MCP server context provided.",
		},
		{
			name: "comprehensive MCP context",
			context: &McpServerContext{
				AvailableServers: []string{"docs-server", "api-server"},
				ServerStatus:     map[string]string{"docs-server": "running", "api-server": "stopped"},
				ErrorLogs:        []string{"connection timeout", "auth failed"},
				ServerConfig:     map[string]interface{}{"port": 8080},
				SecuritySettings: map[string]interface{}{"auth": true},
			},
			want: "MCP Server Environment:",
		},
		{
			name: "minimal context",
			context: &McpServerContext{
				AvailableServers: []string{"test-server"},
			},
			want: "MCP Server Environment:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := NewMcpServerAgent(&MockProvider{})
			agent.SetContext(tt.context)

			result := agent.formatMcpServerContext()

			if !contains(result, tt.want) {
				t.Errorf("formatMcpServerContext() = %v, want to contain %v", result, tt.want)
			}
		})
	}
}

func TestMcpServerAgent_SetRole(t *testing.T) {
	agent := NewMcpServerAgent(&MockProvider{})

	agent.SetRole("custom-mcp-role")

	if agent.role != "custom-mcp-role" {
		t.Errorf("Expected role to be 'custom-mcp-role', got %s", agent.role)
	}
}

func TestMcpServerAgent_ValidationErrors(t *testing.T) {
	// Test with nil provider
	agent := NewMcpServerAgent(nil)

	_, err := agent.Query(context.Background(), "test")
	if err == nil {
		t.Error("Expected error with nil provider")
	}

	_, err = agent.GenerateResponse(context.Background(), "test")
	if err == nil {
		t.Error("Expected error with nil provider")
	}

	_, err = agent.SetupMcpServer("test", nil)
	if err == nil {
		t.Error("Expected error with nil provider")
	}
}
