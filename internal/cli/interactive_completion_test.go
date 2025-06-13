package cli

import (
	"testing"
)

func TestCommandCompleter(t *testing.T) {
	completer := NewCommandCompleter()

	// Test basic command completion
	testCases := []struct {
		name          string
		input         string
		expectedLen   int
		shouldContain []string
	}{
		{
			name:          "Complete 'co' should include community, config, configure, completion",
			input:         "co",
			expectedLen:   4,
			shouldContain: []string{"community", "config", "configure", "completion"},
		},
		{
			name:          "Complete 'help' should include help",
			input:         "help",
			expectedLen:   1,
			shouldContain: []string{"help"},
		},
		{
			name:          "Complete 'community ' should include subcommands",
			input:         "community ",
			expectedLen:   4,
			shouldContain: []string{"forums", "docs", "matrix", "github"},
		},
		{
			name:          "Complete 'community d' should include docs",
			input:         "community d",
			expectedLen:   1,
			shouldContain: []string{"docs"},
		},
		{
			name:          "Complete 'learn ' should include learn subcommands",
			input:         "learn ",
			expectedLen:   6,
			shouldContain: []string{"basics", "flakes", "packages", "services", "advanced", "troubleshooting"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			line := []rune(tc.input)
			completions, length := completer.Do(line, len(line))

			if len(completions) != tc.expectedLen {
				t.Errorf("Expected %d completions, got %d", tc.expectedLen, len(completions))
			}

			// Convert completions to strings for easier testing
			completionStrs := make([]string, len(completions))
			for i, comp := range completions {
				completionStrs[i] = string(comp)
			}

			for _, expected := range tc.shouldContain {
				found := false
				for _, completion := range completionStrs {
					if completion == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected completion '%s' not found in %v", expected, completionStrs)
				}
			}

			// Basic sanity check on length
			if tc.name == "Complete 'community d' should include docs" && length != 1 {
				t.Errorf("Expected length 1 for partial completion, got %d", length)
			}
		})
	}
}

func TestCommandCompleterWithEmptyInput(t *testing.T) {
	completer := NewCommandCompleter()

	// Test empty input
	line := []rune("")
	completions, length := completer.Do(line, 0)

	// Should return all available commands
	if len(completions) == 0 {
		t.Error("Expected some completions for empty input")
	}

	// Should include basic commands
	completionStrs := make([]string, len(completions))
	for i, comp := range completions {
		completionStrs[i] = string(comp)
	}

	expectedCommands := []string{"help", "community", "config", "ask"}
	for _, expected := range expectedCommands {
		found := false
		for _, completion := range completionStrs {
			if completion == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected basic command '%s' not found in completions", expected)
		}
	}

	if length != 0 {
		t.Errorf("Expected length 0 for empty input, got %d", length)
	}
}

func TestCommandCompleterSubcommands(t *testing.T) {
	completer := NewCommandCompleter()

	// Test that subcommands are properly defined
	if completer.subcommands == nil {
		t.Fatal("Subcommands map should not be nil")
	}

	// Test specific subcommand mappings
	testCases := map[string][]string{
		"community": {"forums", "docs", "matrix", "github"},
		"diagnose":  {"system", "config", "services", "network", "hardware", "performance"},
		"learn":     {"basics", "flakes", "packages", "services", "advanced", "troubleshooting"},
	}

	for cmd, expectedSubs := range testCases {
		actualSubs, exists := completer.subcommands[cmd]
		if !exists {
			t.Errorf("Expected subcommands for '%s' not found", cmd)
			continue
		}

		if len(actualSubs) != len(expectedSubs) {
			t.Errorf("Expected %d subcommands for '%s', got %d", len(expectedSubs), cmd, len(actualSubs))
		}

		for _, expected := range expectedSubs {
			found := false
			for _, actual := range actualSubs {
				if actual == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected subcommand '%s' for command '%s' not found", expected, cmd)
			}
		}
	}
}

func TestInteractiveProcessCommand(t *testing.T) {
	// Test that the processInteractiveCommand function can be called without panicking
	// This is a basic integration test
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("processInteractiveCommand panicked: %v", r)
		}
	}()

	// Test with help command (should not panic)
	processInteractiveCommand("help", false)

	// Test with unknown command (should not panic)
	processInteractiveCommand("nonexistent", false)
}
