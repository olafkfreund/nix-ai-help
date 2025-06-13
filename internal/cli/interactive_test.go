package cli

import (
	"strings"
	"testing"
)

// TestParseCommandArgs tests the argument parsing functionality
func TestParseCommandArgs(t *testing.T) {
	testCases := []struct {
		input    string
		expected []string
	}{
		{"simple command", []string{"simple", "command"}},
		{"command with 'quoted args'", []string{"command", "with", "quoted args"}},
		{"command with \"double quotes\"", []string{"command", "with", "double quotes"}},
		{"mixed 'single' and \"double\" quotes", []string{"mixed", "single", "and", "double", "quotes"}},
		{"", []string{}},
		{"   spaces   ", []string{"spaces"}},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := parseCommandArgs(tc.input)
			if len(result) != len(tc.expected) {
				t.Errorf("Expected %d args, got %d", len(tc.expected), len(result))
			}
			for i, expected := range tc.expected {
				if i >= len(result) || result[i] != expected {
					t.Errorf("Expected arg %d to be %s, got %s", i, expected, result[i])
				}
			}
		})
	}
}

// TestCommandParsing tests various command parsing scenarios
func TestCommandParsing(t *testing.T) {
	testCases := []struct {
		input    string
		expected []string
	}{
		{"help", []string{"help"}},
		{"explain-option services.openssh.enable", []string{"explain-option", "services.openssh.enable"}},
		{"search nginx", []string{"search", "nginx"}},
		{"ask 'how do I configure nginx?'", []string{"ask", "how do I configure nginx?"}},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := parseCommandArgs(tc.input)
			if len(result) != len(tc.expected) {
				t.Errorf("Expected %d parts, got %d", len(tc.expected), len(result))
			}
			for i, expected := range tc.expected {
				if i >= len(result) || result[i] != expected {
					t.Errorf("Expected part %d to be '%s', got '%s'", i, expected, result[i])
				}
			}
		})
	}
}

// TestAskCommand tests the ask functionality
func TestAskCommand(t *testing.T) {
	// Test basic ask functionality
	question := "How do I install packages in NixOS?"

	// Mock the AI response (this would require mocking the AI provider)
	// For now, just test that the function exists and can be called
	answer, err := handleAsk(question)

	// We expect this to potentially fail due to missing AI provider configuration
	// but we want to ensure the function exists and handles errors gracefully
	if err == nil && answer == "" {
		t.Errorf("Expected either an answer or an error, got neither")
	}
}

// TestDirectCommandExecution tests the direct command execution functionality
func TestDirectCommandExecution(t *testing.T) {
	testCases := []struct {
		command      string
		args         []string
		shouldHandle bool
	}{
		{"config", []string{"show"}, true},
		{"community", []string{}, true},
		{"configure", []string{"wizard"}, true},
		{"nonexistent", []string{}, false},
	}

	for _, tc := range testCases {
		t.Run(tc.command, func(t *testing.T) {
			var buf strings.Builder
			handled, _ := RunDirectCommand(tc.command, tc.args, &buf)

			if handled != tc.shouldHandle {
				t.Errorf("Expected command '%s' to be handled: %v, got: %v", tc.command, tc.shouldHandle, handled)
			}
		})
	}
}
