package learning

import (
	"context"
	"testing"
)

func TestNewLearningFunction(t *testing.T) {
	lf := NewLearningFunction()

	if lf == nil {
		t.Fatal("NewLearningFunction returned nil")
	}

	if lf.Name() != "learning" {
		t.Errorf("Expected function name 'learning', got '%s'", lf.Name())
	}

	if lf.Description() == "" {
		t.Error("Function description should not be empty")
	}

	// Check that schema has examples
	schema := lf.Schema()
	if len(schema.Examples) == 0 {
		t.Error("Function schema should have examples")
	}
}

func TestLearningFunction_ValidateParameters(t *testing.T) {
	lf := NewLearningFunction()

	tests := []struct {
		name        string
		params      map[string]interface{}
		expectError bool
	}{
		{
			name: "Valid minimal parameters",
			params: map[string]interface{}{
				"topic": "flakes",
			},
			expectError: false,
		},
		{
			name: "Valid parameters with all fields",
			params: map[string]interface{}{
				"topic":       "configuration",
				"level":       "beginner",
				"format":      "tutorial",
				"interactive": true,
				"language":    "english",
			},
			expectError: false,
		},
		{
			name: "Missing required topic parameter",
			params: map[string]interface{}{
				"level": "beginner",
			},
			expectError: true,
		},
		{
			name: "Empty topic parameter",
			params: map[string]interface{}{
				"topic": "",
			},
			expectError: true,
		},
		{
			name: "Topic with only whitespace",
			params: map[string]interface{}{
				"topic": "   ",
			},
			expectError: true,
		},
		{
			name: "Invalid level parameter",
			params: map[string]interface{}{
				"topic": "flakes",
				"level": "expert",
			},
			expectError: true,
		},
		{
			name: "Valid level parameter - beginner",
			params: map[string]interface{}{
				"topic": "flakes",
				"level": "beginner",
			},
			expectError: false,
		},
		{
			name: "Valid level parameter - intermediate",
			params: map[string]interface{}{
				"topic": "flakes",
				"level": "intermediate",
			},
			expectError: false,
		},
		{
			name: "Valid level parameter - advanced",
			params: map[string]interface{}{
				"topic": "flakes",
				"level": "advanced",
			},
			expectError: false,
		},
		{
			name: "Invalid format parameter",
			params: map[string]interface{}{
				"topic":  "flakes",
				"format": "video",
			},
			expectError: true,
		},
		{
			name: "Valid format parameter - tutorial",
			params: map[string]interface{}{
				"topic":  "flakes",
				"format": "tutorial",
			},
			expectError: false,
		},
		{
			name: "Valid format parameter - guide",
			params: map[string]interface{}{
				"topic":  "flakes",
				"format": "guide",
			},
			expectError: false,
		},
		{
			name: "Valid format parameter - reference",
			params: map[string]interface{}{
				"topic":  "flakes",
				"format": "reference",
			},
			expectError: false,
		},
		{
			name: "Valid format parameter - example",
			params: map[string]interface{}{
				"topic":  "flakes",
				"format": "example",
			},
			expectError: false,
		},
		{
			name: "Case insensitive level validation",
			params: map[string]interface{}{
				"topic": "flakes",
				"level": "BEGINNER",
			},
			expectError: false,
		},
		{
			name: "Case insensitive format validation",
			params: map[string]interface{}{
				"topic":  "flakes",
				"format": "TUTORIAL",
			},
			expectError: false,
		},
		{
			name: "Interactive boolean parameter",
			params: map[string]interface{}{
				"topic":       "flakes",
				"interactive": false,
			},
			expectError: false,
		},
		{
			name: "Language parameter",
			params: map[string]interface{}{
				"topic":    "flakes",
				"language": "spanish",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := lf.ValidateParameters(tt.params)
			if tt.expectError && err == nil {
				t.Errorf("Expected validation error for %s, but got none", tt.name)
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no validation error for %s, but got: %v", tt.name, err)
			}
		})
	}
}

func TestLearningFunction_Execute(t *testing.T) {
	lf := NewLearningFunction()
	ctx := context.Background()

	tests := []struct {
		name           string
		params         map[string]interface{}
		expectSuccess  bool
		expectDataType string
	}{
		{
			name: "Execute with minimal parameters",
			params: map[string]interface{}{
				"topic": "flakes",
			},
			expectSuccess:  true,
			expectDataType: "*learning.LearningResponse",
		},
		{
			name: "Execute with all parameters",
			params: map[string]interface{}{
				"topic":       "configuration",
				"level":       "intermediate",
				"format":      "guide",
				"interactive": true,
				"language":    "english",
			},
			expectSuccess:  true,
			expectDataType: "*learning.LearningResponse",
		},
		{
			name: "Execute with empty topic",
			params: map[string]interface{}{
				"topic": "",
			},
			expectSuccess: false,
		},
		{
			name: "Execute with flakes topic",
			params: map[string]interface{}{
				"topic": "flakes",
				"level": "beginner",
			},
			expectSuccess:  true,
			expectDataType: "*learning.LearningResponse",
		},
		{
			name: "Execute with configuration topic",
			params: map[string]interface{}{
				"topic": "configuration",
				"level": "advanced",
			},
			expectSuccess:  true,
			expectDataType: "*learning.LearningResponse",
		},
		{
			name: "Execute with packages topic",
			params: map[string]interface{}{
				"topic": "packages",
				"level": "intermediate",
			},
			expectSuccess:  true,
			expectDataType: "*learning.LearningResponse",
		},
		{
			name: "Execute with home-manager topic",
			params: map[string]interface{}{
				"topic": "home-manager",
				"level": "beginner",
			},
			expectSuccess:  true,
			expectDataType: "*learning.LearningResponse",
		},
		{
			name: "Execute with derivations topic",
			params: map[string]interface{}{
				"topic": "derivations",
				"level": "advanced",
			},
			expectSuccess:  true,
			expectDataType: "*learning.LearningResponse",
		},
		{
			name: "Execute with nix language topic",
			params: map[string]interface{}{
				"topic": "nix language",
				"level": "intermediate",
			},
			expectSuccess:  true,
			expectDataType: "*learning.LearningResponse",
		},
		{
			name: "Execute with unknown topic",
			params: map[string]interface{}{
				"topic": "unknown-topic",
				"level": "beginner",
			},
			expectSuccess:  true,
			expectDataType: "*learning.LearningResponse",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := lf.Execute(ctx, tt.params, nil)

			if err != nil {
				t.Fatalf("Execute returned error: %v", err)
			}

			if result == nil {
				t.Fatal("Execute returned nil result")
			}

			if tt.expectSuccess && !result.Success {
				t.Errorf("Expected successful execution for %s, but got failure", tt.name)
			}

			if !tt.expectSuccess && result.Success {
				t.Errorf("Expected failed execution for %s, but got success", tt.name)
			}

			if tt.expectSuccess && result.Data != nil {
				// Verify the response structure
				response, ok := result.Data.(*LearningResponse)
				if !ok {
					t.Errorf("Expected response to be *LearningResponse, got %T", result.Data)
				} else {
					// Verify response fields
					if response.Topic == "" {
						t.Error("Response topic should not be empty")
					}
					if response.Level == "" {
						t.Error("Response level should not be empty")
					}
					if len(response.Resources) == 0 {
						t.Error("Response should have at least one resource")
					}
					// Verify each resource has required fields
					for i, resource := range response.Resources {
						if resource.Title == "" {
							t.Errorf("Resource %d should have a title", i)
						}
						if resource.Type == "" {
							t.Errorf("Resource %d should have a type", i)
						}
						if resource.Description == "" {
							t.Errorf("Resource %d should have a description", i)
						}
						if resource.Difficulty == "" {
							t.Errorf("Resource %d should have a difficulty", i)
						}
					}
				}
			}
		})
	}
}

func TestLearningFunction_ParseRequest(t *testing.T) {
	lf := NewLearningFunction()

	tests := []struct {
		name                string
		params              map[string]interface{}
		expectedTopic       string
		expectedLevel       string
		expectedFormat      string
		expectedLang        string
		expectedInteractive bool
	}{
		{
			name: "Minimal parameters",
			params: map[string]interface{}{
				"topic": "flakes",
			},
			expectedTopic:       "flakes",
			expectedLevel:       "beginner",
			expectedFormat:      "tutorial",
			expectedLang:        "english",
			expectedInteractive: false,
		},
		{
			name: "All parameters provided",
			params: map[string]interface{}{
				"topic":       "configuration",
				"level":       "Advanced",
				"format":      "Guide",
				"interactive": true,
				"language":    "Spanish",
			},
			expectedTopic:       "configuration",
			expectedLevel:       "advanced",
			expectedFormat:      "guide",
			expectedLang:        "Spanish",
			expectedInteractive: true,
		},
		{
			name: "Parameters with whitespace",
			params: map[string]interface{}{
				"topic":    "  flakes  ",
				"level":    "  INTERMEDIATE  ",
				"format":   "  EXAMPLE  ",
				"language": "  French  ",
			},
			expectedTopic:       "flakes",
			expectedLevel:       "intermediate",
			expectedFormat:      "example",
			expectedLang:        "French",
			expectedInteractive: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request, err := lf.parseRequest(tt.params)
			if err != nil {
				t.Fatalf("parseRequest returned error: %v", err)
			}

			if request.Topic != tt.expectedTopic {
				t.Errorf("Expected topic '%s', got '%s'", tt.expectedTopic, request.Topic)
			}
			if request.Level != tt.expectedLevel {
				t.Errorf("Expected level '%s', got '%s'", tt.expectedLevel, request.Level)
			}
			if request.Format != tt.expectedFormat {
				t.Errorf("Expected format '%s', got '%s'", tt.expectedFormat, request.Format)
			}
			if request.Language != tt.expectedLang {
				t.Errorf("Expected language '%s', got '%s'", tt.expectedLang, request.Language)
			}
			if request.Interactive != tt.expectedInteractive {
				t.Errorf("Expected interactive '%v', got '%v'", tt.expectedInteractive, request.Interactive)
			}
		})
	}
}

func TestLearningFunction_GenerateLearningResources(t *testing.T) {
	lf := NewLearningFunction()

	tests := []struct {
		name           string
		request        *LearningRequest
		expectedMinRes int
		expectContent  bool
	}{
		{
			name: "Flakes resources",
			request: &LearningRequest{
				Topic:  "flakes",
				Level:  "beginner",
				Format: "tutorial",
			},
			expectedMinRes: 2, // Should have flake-specific + official docs
			expectContent:  true,
		},
		{
			name: "Configuration resources",
			request: &LearningRequest{
				Topic:  "configuration",
				Level:  "intermediate",
				Format: "guide",
			},
			expectedMinRes: 2,
			expectContent:  false,
		},
		{
			name: "Package resources",
			request: &LearningRequest{
				Topic:  "packages",
				Level:  "advanced",
				Format: "reference",
			},
			expectedMinRes: 3, // Package-specific + official docs
			expectContent:  false,
		},
		{
			name: "Home Manager resources",
			request: &LearningRequest{
				Topic:  "home-manager",
				Level:  "beginner",
				Format: "tutorial",
			},
			expectedMinRes: 3, // Home Manager-specific + official docs
			expectContent:  false,
		},
		{
			name: "Unknown topic resources",
			request: &LearningRequest{
				Topic:  "unknown-topic",
				Level:  "beginner",
				Format: "tutorial",
			},
			expectedMinRes: 3, // General resources + official docs
			expectContent:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resources := lf.generateLearningResources(tt.request)

			if len(resources) < tt.expectedMinRes {
				t.Errorf("Expected at least %d resources, got %d", tt.expectedMinRes, len(resources))
			}

			// Check if any resource has content when expected
			hasContent := false
			for _, resource := range resources {
				if resource.Content != "" {
					hasContent = true
					break
				}
			}

			if tt.expectContent && !hasContent {
				t.Error("Expected at least one resource to have content")
			}

			// Verify all resources have required fields
			for i, resource := range resources {
				if resource.Title == "" {
					t.Errorf("Resource %d missing title", i)
				}
				if resource.Type == "" {
					t.Errorf("Resource %d missing type", i)
				}
				if resource.Description == "" {
					t.Errorf("Resource %d missing description", i)
				}
				if resource.Difficulty == "" {
					t.Errorf("Resource %d missing difficulty", i)
				}
			}
		})
	}
}

func TestLearningFunction_GenerateNextSteps(t *testing.T) {
	lf := NewLearningFunction()

	tests := []struct {
		name           string
		topic          string
		level          string
		expectedMinLen int
	}{
		{
			name:           "Beginner level",
			topic:          "flakes",
			level:          "beginner",
			expectedMinLen: 4,
		},
		{
			name:           "Intermediate level",
			topic:          "configuration",
			level:          "intermediate",
			expectedMinLen: 4,
		},
		{
			name:           "Advanced level",
			topic:          "packages",
			level:          "advanced",
			expectedMinLen: 4,
		},
		{
			name:           "Flakes topic specific",
			topic:          "flakes",
			level:          "intermediate",
			expectedMinLen: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextSteps := lf.generateNextSteps(tt.topic, tt.level)

			if len(nextSteps) < tt.expectedMinLen {
				t.Errorf("Expected at least %d next steps, got %d", tt.expectedMinLen, len(nextSteps))
			}

			// Verify all steps are non-empty
			for i, step := range nextSteps {
				if step == "" {
					t.Errorf("Next step %d is empty", i)
				}
			}
		})
	}
}

func TestLearningFunction_GeneratePrerequisites(t *testing.T) {
	lf := NewLearningFunction()

	tests := []struct {
		name           string
		topic          string
		level          string
		expectedMinLen int
	}{
		{
			name:           "Beginner level",
			topic:          "flakes",
			level:          "beginner",
			expectedMinLen: 2,
		},
		{
			name:           "Intermediate level",
			topic:          "configuration",
			level:          "intermediate",
			expectedMinLen: 3,
		},
		{
			name:           "Advanced level",
			topic:          "packages",
			level:          "advanced",
			expectedMinLen: 3,
		},
		{
			name:           "Derivations topic specific",
			topic:          "derivations",
			level:          "advanced",
			expectedMinLen: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prerequisites := lf.generatePrerequisites(tt.topic, tt.level)

			if len(prerequisites) < tt.expectedMinLen {
				t.Errorf("Expected at least %d prerequisites, got %d", tt.expectedMinLen, len(prerequisites))
			}

			// Verify all prerequisites are non-empty
			for i, prereq := range prerequisites {
				if prereq == "" {
					t.Errorf("Prerequisite %d is empty", i)
				}
			}
		})
	}
}

func TestLearningFunction_EstimateLearningTime(t *testing.T) {
	lf := NewLearningFunction()

	tests := []struct {
		name     string
		topic    string
		level    string
		expected string
	}{
		{
			name:     "Flakes beginner",
			topic:    "flakes",
			level:    "beginner",
			expected: "2-3 hours",
		},
		{
			name:     "Configuration advanced",
			topic:    "configuration",
			level:    "advanced",
			expected: "12-16 hours",
		},
		{
			name:     "Unknown topic intermediate",
			topic:    "unknown",
			level:    "intermediate",
			expected: "4-8 hours",
		},
		{
			name:     "Invalid level defaults",
			topic:    "flakes",
			level:    "invalid",
			expected: "2-4 hours",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			time := lf.estimateLearningTime(tt.topic, tt.level)
			if time != tt.expected {
				t.Errorf("Expected time '%s', got '%s'", tt.expected, time)
			}
		})
	}
}

func TestLearningFunction_Contains(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		item     string
		expected bool
	}{
		{
			name:     "Item exists",
			slice:    []string{"beginner", "intermediate", "advanced"},
			item:     "intermediate",
			expected: true,
		},
		{
			name:     "Item does not exist",
			slice:    []string{"beginner", "intermediate", "advanced"},
			item:     "expert",
			expected: false,
		},
		{
			name:     "Empty slice",
			slice:    []string{},
			item:     "beginner",
			expected: false,
		},
		{
			name:     "Empty item",
			slice:    []string{"beginner", "intermediate"},
			item:     "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(tt.slice, tt.item)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestLearningFunction_TopicSpecificResources(t *testing.T) {
	lf := NewLearningFunction()

	// Test specific topic resource generation methods
	tests := []struct {
		name         string
		topic        string
		level        string
		format       string
		minResources int
	}{
		{
			name:         "Flake resources beginner",
			topic:        "flake",
			level:        "beginner",
			format:       "tutorial",
			minResources: 1,
		},
		{
			name:         "Configuration resources advanced",
			topic:        "config",
			level:        "advanced",
			format:       "guide",
			minResources: 1,
		},
		{
			name:         "Package resources intermediate",
			topic:        "package",
			level:        "intermediate",
			format:       "reference",
			minResources: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &LearningRequest{
				Topic:  tt.topic,
				Level:  tt.level,
				Format: tt.format,
			}

			resources := lf.generateLearningResources(request)

			if len(resources) < tt.minResources {
				t.Errorf("Expected at least %d resources for %s, got %d", tt.minResources, tt.topic, len(resources))
			}
		})
	}
}
