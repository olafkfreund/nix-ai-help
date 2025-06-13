package agent

import (
	"context"
	"strings"
	"testing"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/ai/roles"
)

// MockProvider implements the ai.Provider interface for testing.
type MockMachinesProvider struct {
	response string
}

func (m *MockMachinesProvider) Query(prompt string) (string, error) {
	return m.response, nil
}

func (m *MockMachinesProvider) QueryWithContext(ctx context.Context, prompt string) (string, error) {
	return m.response, nil
}

func (m *MockMachinesProvider) GetPartialResponse() string {
	return ""
}

func (m *MockMachinesProvider) StreamResponse(ctx context.Context, prompt string) (<-chan ai.StreamResponse, error) {
	ch := make(chan ai.StreamResponse, 1)
	ch <- ai.StreamResponse{Content: "mock stream response", Done: true}
	close(ch)
	return ch, nil
}

func (m *MockMachinesProvider) GenerateResponse(ctx context.Context, prompt string) (string, error) {
	return m.response, nil
}

func TestNewMachinesAgent(t *testing.T) {
	provider := &MockMachinesProvider{response: "test response"}
	agent := NewMachinesAgent(provider)

	if agent == nil {
		t.Fatal("NewMachinesAgent returned nil")
	}

	if agent.role != roles.RoleMachines {
		t.Errorf("Expected role %s, got %s", roles.RoleMachines, agent.role)
	}

	if agent.provider != provider {
		t.Error("Provider not set correctly")
	}

	if agent.context == nil {
		t.Error("Context not initialized")
	}
}

func TestMachinesAgent_SetContext(t *testing.T) {
	provider := &MockMachinesProvider{response: "test response"}
	agent := NewMachinesAgent(provider)

	ctx := &MachineContext{
		MachineName:  "test-machine",
		HostName:     "test.example.com",
		Architecture: "x86_64-linux",
		MachineRole:  "webserver",
	}

	agent.SetContext(ctx)

	if agent.context != ctx {
		t.Error("Context not set correctly")
	}

	retrievedCtx := agent.GetContext()
	if retrievedCtx != ctx {
		t.Error("GetContext did not return the set context")
	}
}

func TestMachinesAgent_Query(t *testing.T) {
	provider := &MockMachinesProvider{response: "machine management response"}
	agent := NewMachinesAgent(provider)

	response, err := agent.Query(context.Background(), "How do I deploy to multiple machines?")
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if response != "machine management response" {
		t.Errorf("Expected 'machine management response', got '%s'", response)
	}
}

func TestMachinesAgent_GenerateResponse(t *testing.T) {
	provider := &MockMachinesProvider{response: "generated machine response"}
	agent := NewMachinesAgent(provider)

	response, err := agent.GenerateResponse(context.Background(), "Deploy configuration to all machines")
	if err != nil {
		t.Fatalf("GenerateResponse failed: %v", err)
	}

	if response != "generated machine response" {
		t.Errorf("Expected 'generated machine response', got '%s'", response)
	}
}

func TestMachinesAgent_PlanDeployment(t *testing.T) {
	tests := []struct {
		name         string
		machines     []string
		deployMethod string
		wantContains []string
	}{
		{
			name:         "flakes deployment",
			machines:     []string{"web1", "web2", "db1"},
			deployMethod: "flakes",
			wantContains: []string{"Deployment Planning", "Machine Management Best Practices"},
		},
		{
			name:         "deploy-rs deployment",
			machines:     []string{"server1", "server2"},
			deployMethod: "deploy-rs",
			wantContains: []string{"Deployment Planning", "rollback strategies"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &MockMachinesProvider{response: "deployment planning response"}
			agent := NewMachinesAgent(provider)

			response, err := agent.PlanDeployment(context.Background(), tt.machines, tt.deployMethod)
			if err != nil {
				t.Fatalf("PlanDeployment failed: %v", err)
			}

			for _, want := range tt.wantContains {
				if !strings.Contains(response, want) {
					t.Errorf("Response should contain '%s', got: %s", want, response)
				}
			}

			// Verify context was set correctly
			if agent.context.DeployMethod != tt.deployMethod {
				t.Errorf("Expected deploy method %s, got %s", tt.deployMethod, agent.context.DeployMethod)
			}
		})
	}
}

func TestMachinesAgent_DiagnoseDeploymentIssues(t *testing.T) {
	tests := []struct {
		name        string
		machineName string
		issues      []string
		wantError   bool
	}{
		{
			name:        "ssh connection issues",
			machineName: "web-server-1",
			issues:      []string{"SSH connection timeout", "Permission denied"},
			wantError:   false,
		},
		{
			name:        "build failures",
			machineName: "build-server",
			issues:      []string{"Build failure", "Missing dependency"},
			wantError:   false,
		},
		{
			name:        "network issues",
			machineName: "remote-server",
			issues:      []string{"Network unreachable", "DNS resolution failed"},
			wantError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &MockMachinesProvider{response: "deployment diagnosis response"}
			agent := NewMachinesAgent(provider)

			response, err := agent.DiagnoseDeploymentIssues(context.Background(), tt.machineName, tt.issues)
			if (err != nil) != tt.wantError {
				t.Errorf("DiagnoseDeploymentIssues() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError {
				if !strings.Contains(response, "Deployment Issue Diagnosis") {
					t.Error("Response should contain diagnosis header")
				}

				// Verify context was set correctly
				if agent.context.MachineName != tt.machineName {
					t.Errorf("Expected machine name %s, got %s", tt.machineName, agent.context.MachineName)
				}

				if len(agent.context.Issues) != len(tt.issues) {
					t.Errorf("Expected %d issues, got %d", len(tt.issues), len(agent.context.Issues))
				}
			}
		})
	}
}

func TestMachinesAgent_OptimizeMachineConfiguration(t *testing.T) {
	tests := []struct {
		name         string
		machineName  string
		machineRole  string
		requirements []string
	}{
		{
			name:         "web server optimization",
			machineName:  "web-01",
			machineRole:  "webserver",
			requirements: []string{"high performance", "SSL termination", "load balancing"},
		},
		{
			name:         "database server optimization",
			machineName:  "db-primary",
			machineRole:  "database",
			requirements: []string{"high memory", "fast storage", "backup automation"},
		},
		{
			name:         "development machine optimization",
			machineName:  "dev-workstation",
			machineRole:  "development",
			requirements: []string{"multiple languages", "containerization", "debugging tools"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &MockMachinesProvider{response: "configuration optimization response"}
			agent := NewMachinesAgent(provider)

			response, err := agent.OptimizeMachineConfiguration(context.Background(), tt.machineName, tt.machineRole, tt.requirements)
			if err != nil {
				t.Fatalf("OptimizeMachineConfiguration failed: %v", err)
			}

			if !strings.Contains(response, "Machine Configuration Optimization") {
				t.Error("Response should contain optimization header")
			}

			// Verify context was set correctly
			if agent.context.MachineName != tt.machineName {
				t.Errorf("Expected machine name %s, got %s", tt.machineName, agent.context.MachineName)
			}

			if agent.context.MachineRole != tt.machineRole {
				t.Errorf("Expected machine role %s, got %s", tt.machineRole, agent.context.MachineRole)
			}
		})
	}
}

func TestMachinesAgent_MonitorMachineHealth(t *testing.T) {
	tests := []struct {
		name          string
		machines      []string
		healthMetrics []string
	}{
		{
			name:          "basic monitoring",
			machines:      []string{"web1", "web2", "db1"},
			healthMetrics: []string{"CPU usage", "Memory usage", "Disk space"},
		},
		{
			name:          "advanced monitoring",
			machines:      []string{"app-server", "cache-server"},
			healthMetrics: []string{"Response time", "Error rate", "Connection count", "Queue length"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &MockMachinesProvider{response: "health monitoring response"}
			agent := NewMachinesAgent(provider)

			response, err := agent.MonitorMachineHealth(context.Background(), tt.machines, tt.healthMetrics)
			if err != nil {
				t.Fatalf("MonitorMachineHealth failed: %v", err)
			}

			if !strings.Contains(response, "Machine Health Monitoring") {
				t.Error("Response should contain monitoring header")
			}

			if !strings.Contains(response, "Machine Management Best Practices") {
				t.Error("Response should contain best practices")
			}
		})
	}
}

func TestMachinesAgent_ManageFlakeMigration(t *testing.T) {
	tests := []struct {
		name      string
		flakePath string
		machines  []string
	}{
		{
			name:      "small deployment",
			flakePath: "/etc/nixos",
			machines:  []string{"server1", "server2"},
		},
		{
			name:      "large deployment",
			flakePath: "/home/admin/nixos-config",
			machines:  []string{"web1", "web2", "web3", "db1", "db2", "cache1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &MockMachinesProvider{response: "flake migration response"}
			agent := NewMachinesAgent(provider)

			response, err := agent.ManageFlakeMigration(context.Background(), tt.flakePath, tt.machines)
			if err != nil {
				t.Fatalf("ManageFlakeMigration failed: %v", err)
			}

			if !strings.Contains(response, "Flake Migration Planning") {
				t.Error("Response should contain migration planning header")
			}

			// Verify context was set correctly
			if agent.context.FlakePath != tt.flakePath {
				t.Errorf("Expected flake path %s, got %s", tt.flakePath, agent.context.FlakePath)
			}
		})
	}
}

func TestMachinesAgent_SetupDeployRs(t *testing.T) {
	tests := []struct {
		name        string
		hosts       []string
		interactive bool
	}{
		{
			name:        "interactive setup",
			hosts:       []string{"web-server", "db-server"},
			interactive: true,
		},
		{
			name:        "automated setup",
			hosts:       []string{"server1", "server2", "server3"},
			interactive: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &MockMachinesProvider{response: "deploy-rs setup response"}
			agent := NewMachinesAgent(provider)

			response, err := agent.SetupDeployRs(context.Background(), tt.hosts, tt.interactive)
			if err != nil {
				t.Fatalf("SetupDeployRs failed: %v", err)
			}

			if !strings.Contains(response, "Deploy-rs Configuration") {
				t.Error("Response should contain deploy-rs configuration header")
			}

			// Verify context was set correctly
			if agent.context.DeployMethod != "deploy-rs" {
				t.Errorf("Expected deploy method 'deploy-rs', got %s", agent.context.DeployMethod)
			}
		})
	}
}

func TestMachinesAgent_buildContextualPrompt(t *testing.T) {
	provider := &MockMachinesProvider{response: "test response"}
	agent := NewMachinesAgent(provider)

	// Set some context
	ctx := &MachineContext{
		MachineName:  "test-machine",
		MachineRole:  "webserver",
		DeployMethod: "flakes",
		HealthStatus: "healthy",
	}
	agent.SetContext(ctx)

	prompt := agent.buildContextualPrompt("Test role prompt", "Test user input")

	expectedParts := []string{
		"Test role prompt",
		"Machine Context:",
		"test-machine",
		"webserver",
		"flakes",
		"healthy",
		"User Request:",
		"Test user input",
	}

	for _, part := range expectedParts {
		if !strings.Contains(prompt, part) {
			t.Errorf("Prompt should contain '%s', got: %s", part, prompt)
		}
	}
}

func TestMachinesAgent_formatMachineContext(t *testing.T) {
	provider := &MockMachinesProvider{response: "test response"}
	agent := NewMachinesAgent(provider)

	ctx := &MachineContext{
		MachineName:  "web-server-01",
		HostName:     "web01.example.com",
		Architecture: "x86_64-linux",
		MachineRole:  "webserver",
		DeployMethod: "deploy-rs",
		DeployStatus: "success",
		HealthStatus: "healthy",
		Issues:       []string{"high CPU usage", "low disk space"},
	}

	formatted := agent.formatMachineContext(ctx)

	expectedParts := []string{
		"Machine: web-server-01",
		"Hostname: web01.example.com",
		"Architecture: x86_64-linux",
		"Role: webserver",
		"Deploy Method: deploy-rs",
		"Deploy Status: success",
		"Health Status: healthy",
		"Issues: [high CPU usage low disk space]",
	}

	for _, part := range expectedParts {
		if !strings.Contains(formatted, part) {
			t.Errorf("Formatted context should contain '%s', got: %s", part, formatted)
		}
	}
}

func TestMachinesAgent_buildPrompt(t *testing.T) {
	provider := &MockMachinesProvider{response: "test response"}
	agent := NewMachinesAgent(provider)

	// Set context
	ctx := &MachineContext{
		MachineName:  "test-machine",
		MachineRole:  "database",
		DeployMethod: "flakes",
		HealthStatus: "degraded",
		Issues:       []string{"connection timeout"},
	}
	agent.SetContext(ctx)

	details := map[string]interface{}{
		"target_machines": []string{"db1", "db2"},
		"operation_type":  "health check",
	}

	prompt := agent.buildPrompt("Test machine operation", details)

	expectedParts := []string{
		"**Task**: Test machine operation",
		"**Machine Context**:",
		"Machine Name: test-machine",
		"Machine Role: database",
		"Deployment Method: flakes",
		"Health Status: degraded",
		"Known Issues: [connection timeout]",
		"**Operation Details**:",
		"Target Machines: [db1 db2]",
		"Operation Type: health check",
		"**Requirements**:",
		"Provide specific commands and configuration examples",
		"Include safety measures and rollback strategies",
	}

	for _, part := range expectedParts {
		if !strings.Contains(prompt, part) {
			t.Errorf("Prompt should contain '%s', got: %s", part, prompt)
		}
	}
}

func TestMachinesAgent_formatMachineResponse(t *testing.T) {
	provider := &MockMachinesProvider{response: "test response"}
	agent := NewMachinesAgent(provider)

	response := agent.formatMachineResponse("Test machine response content", "Test Operation")

	expectedParts := []string{
		"# Test Operation",
		"Test machine response content",
		"🔧 Machine Management Best Practices",
		"Test deployments in staging environment first",
		"Maintain rollback strategies for all deployments",
		"Monitor machine health and resource usage",
		"Keep machine configurations in version control",
	}

	for _, part := range expectedParts {
		if !strings.Contains(response, part) {
			t.Errorf("Response should contain '%s', got: %s", part, response)
		}
	}
}

func TestMachinesAgent_RoleValidation(t *testing.T) {
	provider := &MockMachinesProvider{response: "test response"}
	agent := NewMachinesAgent(provider)

	// Test that the agent starts with the correct role
	if agent.role != roles.RoleMachines {
		t.Errorf("Expected role %s, got %s", roles.RoleMachines, agent.role)
	}

	// Test role validation
	err := agent.validateRole()
	if err != nil {
		t.Errorf("Role validation failed: %v", err)
	}

	// Test with invalid role
	agent.role = "invalid-role"
	err = agent.validateRole()
	if err == nil {
		t.Error("Expected role validation to fail for invalid role")
	}
}
