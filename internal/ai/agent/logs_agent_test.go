package agent

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogsAgent_Query(t *testing.T) {
	tests := []struct {
		name           string
		question       string
		context        *LogsContext
		expectedPrompt string
		mockResponse   string
		expectedError  bool
	}{
		{
			name:     "Basic log analysis query",
			question: "How do I analyze boot logs for errors?",
			context: &LogsContext{
				LogSources:   []string{"systemd", "kernel", "nixos-rebuild"},
				LogLevel:     "error",
				TimeRange:    "last-boot",
				BootMessages: "Boot completed successfully",
				ServiceNames: []string{"nixos-rebuild"},
			},
			expectedPrompt: "How do I analyze boot logs for errors?",
			mockResponse:   "To analyze boot logs for errors, use journalctl -b -p err",
			expectedError:  false,
		},
		{
			name:     "System monitoring query with context",
			question: "What monitoring tools should I use?",
			context: &LogsContext{
				LogSources:   []string{"systemd", "dmesg"},
				ServiceNames: []string{"nginx", "postgresql"},
				LogLevel:     "info",
				TimeRange:    "last-hour",
			},
			expectedPrompt: "What monitoring tools should I use?",
			mockResponse:   "For your setup with nginx and postgresql, consider using prometheus with grafana",
			expectedError:  false,
		},
		{
			name:     "Security audit query",
			question: "How to audit system security from logs?",
			context: &LogsContext{
				AuthLogs:       "/var/log/auth.log entries",
				SecurityEvents: []string{"login", "sudo", "ssh"},
				ErrorMessages:  []string{"Failed login attempt"},
				LogSources:     []string{"auth", "audit"},
			},
			expectedPrompt: "How to audit system security from logs?",
			mockResponse:   "Review audit logs for security events and failed login attempts",
			expectedError:  false,
		},
		{
			name:     "Performance analysis query",
			question: "How to analyze system performance from logs?",
			context: &LogsContext{
				PerformanceIssues: []string{"high CPU usage", "memory pressure"},
				SystemErrors:      []string{"out of memory"},
				KernelMessages:    "kernel performance warnings",
				LogPatterns:       []string{"cpu", "memory", "disk"},
			},
			expectedPrompt: "How to analyze system performance from logs?",
			mockResponse:   "Use journalctl with performance filtering and system metrics",
			expectedError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock provider
			mockProvider := &MockProvider{response: tt.mockResponse}

			// Create agent
			agent := NewLogsAgent(mockProvider)

			// Set context if provided
			if tt.context != nil {
				agent.SetContext(tt.context)
			}

			// Execute query
			result, err := agent.Query(context.Background(), tt.question)

			// Verify results
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// LogsAgent enhances responses with guidance, so expect the enhanced result
				expectedResult := tt.mockResponse + "\n\n---\n**Log Analysis Tips:**\n- Use `journalctl` for systemd service logs\n- Filter logs by time with `--since` and `--until`\n- Use `grep` and `awk` for pattern matching\n- Monitor logs in real-time with `journalctl -f`\n- Check log rotation and retention policies\n- Consider log aggregation tools for complex analysis\n"
				assert.Equal(t, expectedResult, result)
			}
		})
	}
}

func TestLogsAgent_SetContext(t *testing.T) {
	mockProvider := &MockProvider{response: "test response"}
	agent := NewLogsAgent(mockProvider)

	context := &LogsContext{
		LogSources:   []string{"systemd", "kernel"},
		LogLevel:     "info",
		TimeRange:    "last-hour",
		ServiceNames: []string{"sshd", "nginx"},
	}

	agent.SetContext(context)

	// Verify context was set
	assert.Equal(t, context, agent.contextData)
}

func TestLogsAgent_InvalidContext(t *testing.T) {
	mockProvider := &MockProvider{response: "response"}
	agent := NewLogsAgent(mockProvider)

	// Test with invalid context type
	agent.SetContext("invalid context")

	// Should still work but ignore invalid context
	result, err := agent.Query(context.Background(), "test question")

	assert.NoError(t, err)
	// LogsAgent enhances responses with guidance
	expectedResult := "response\n\n---\n**Log Analysis Tips:**\n- Use `journalctl` for systemd service logs\n- Filter logs by time with `--since` and `--until`\n- Use `grep` and `awk` for pattern matching\n- Monitor logs in real-time with `journalctl -f`\n- Check log rotation and retention policies\n- Consider log aggregation tools for complex analysis\n"
	assert.Equal(t, expectedResult, result)
}
