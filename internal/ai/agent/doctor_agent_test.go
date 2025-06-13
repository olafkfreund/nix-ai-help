package agent

import (
	"context"
	"strings"
	"testing"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/ai/roles"
)

// MockDoctorProvider implements the ai.Provider interface for testing
type MockDoctorProvider struct {
	response  string
	err       error
	LastQuery string
}

func (m *MockDoctorProvider) GenerateResponse(ctx context.Context, prompt string) (string, error) {
	m.LastQuery = prompt
	return m.response, m.err
}

func (m *MockDoctorProvider) Query(prompt string) (string, error) {
	return m.response, m.err
}

func (m *MockDoctorProvider) QueryWithContext(ctx context.Context, prompt string) (string, error) {
	return m.response, m.err
}

func (m *MockDoctorProvider) GetPartialResponse() string {
	return ""
}

func (m *MockDoctorProvider) StreamResponse(ctx context.Context, prompt string) (<-chan ai.StreamResponse, error) {
	ch := make(chan ai.StreamResponse, 1)
	ch <- ai.StreamResponse{Content: m.response, Done: true}
	close(ch)
	return ch, nil
}

func TestDoctorAgent_NewDoctorAgent(t *testing.T) {
	mockProvider := &MockDoctorProvider{}
	agent := NewDoctorAgent(mockProvider)

	if agent == nil {
		t.Fatal("NewDoctorAgent returned nil")
	}

	if agent.role != roles.RoleDoctor {
		t.Errorf("Expected role %s, got %s", roles.RoleDoctor, agent.role)
	}

	if agent.provider != mockProvider {
		t.Error("Provider not set correctly")
	}
}

func TestDoctorAgent_Query(t *testing.T) {
	tests := []struct {
		name             string
		question         string
		context          *DoctorContext
		expectedInQuery  string
		providerResponse string
		wantErr          bool
	}{
		{
			name:             "basic health check question",
			question:         "Is my system healthy?",
			expectedInQuery:  "Is my system healthy?",
			providerResponse: "Your system appears to be healthy...",
			wantErr:          false,
		},
		{
			name:     "health check with context",
			question: "What's wrong with my system?",
			context: &DoctorContext{
				SystemHealth:   "degraded",
				ServiceStatus:  "nginx: failed",
				NixStoreHealth: "corrupted entries found",
				SystemErrors:   []string{"out of disk space", "service timeout"},
			},
			expectedInQuery:  "System Health: degraded",
			providerResponse: "Several issues detected...",
			wantErr:          false,
		},
		{
			name:     "performance health check",
			question: "Why is my system slow?",
			context: &DoctorContext{
				PerformanceInfo: "high CPU usage: 90%",
				MemoryInfo:      "8GB/8GB used",
				StorageInfo:     "disk 95% full",
			},
			expectedInQuery:  "Performance Info: high CPU usage: 90%",
			providerResponse: "Performance issues detected...",
			wantErr:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProvider := &MockDoctorProvider{
				response: tt.providerResponse,
			}
			agent := NewDoctorAgent(mockProvider)

			if tt.context != nil {
				agent.SetDoctorContext(tt.context)
			}

			response, err := agent.Query(context.Background(), tt.question)

			if (err != nil) != tt.wantErr {
				t.Errorf("Query() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if response != tt.providerResponse {
					t.Errorf("Query() response = %v, want %v", response, tt.providerResponse)
				}

				if !strings.Contains(mockProvider.LastQuery, tt.expectedInQuery) {
					t.Errorf("Expected query to contain %q, got %q", tt.expectedInQuery, mockProvider.LastQuery)
				}

				if !strings.Contains(mockProvider.LastQuery, tt.question) {
					t.Errorf("Expected query to contain question %q, got %q", tt.question, mockProvider.LastQuery)
				}
			}
		})
	}
}

func TestDoctorAgent_SetDoctorContext(t *testing.T) {
	mockProvider := &MockDoctorProvider{}
	agent := NewDoctorAgent(mockProvider)

	ctx := &DoctorContext{
		SystemHealth:   "healthy",
		ServiceStatus:  "all services running",
		NixStoreHealth: "no issues detected",
	}

	agent.SetDoctorContext(ctx)

	if agent.contextData != ctx {
		t.Error("Doctor context not set correctly")
	}
}

func TestDoctorAgent_formatDoctorContext(t *testing.T) {
	tests := []struct {
		name     string
		context  *DoctorContext
		expected []string // strings that should be in the output
	}{
		{
			name: "comprehensive health context",
			context: &DoctorContext{
				SystemHealth:    "healthy",
				NixStoreHealth:  "verified",
				ChannelStatus:   "up to date",
				ServiceStatus:   "all services running",
				SystemErrors:    []string{"warning: low disk space"},
				WarningMessages: []string{"outdated package detected"},
			},
			expected: []string{
				"System Health: healthy",
				"Nix Store Health: verified",
				"Channel Status: up to date",
				"Service Status: all services running",
				"System Errors: warning: low disk space",
				"Warnings: outdated package detected",
			},
		},
		{
			name: "performance context",
			context: &DoctorContext{
				PerformanceInfo: "CPU: 25%, Memory: 50%",
				StorageInfo:     "50GB free",
				NetworkStatus:   "connected",
			},
			expected: []string{
				"Performance Info: CPU: 25%, Memory: 50%",
				"Storage Info: 50GB free",
				"Network Status: connected",
			},
		},
		{
			name: "minimal context",
			context: &DoctorContext{
				SystemHealth: "unknown",
			},
			expected: []string{
				"System Health: unknown",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProvider := &MockDoctorProvider{}
			agent := NewDoctorAgent(mockProvider)

			result := agent.formatDoctorContext(tt.context)

			for _, expected := range tt.expected {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected output to contain %q, got %q", expected, result)
				}
			}
		})
	}
}
