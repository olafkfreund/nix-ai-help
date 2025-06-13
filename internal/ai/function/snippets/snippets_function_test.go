package snippets

import (
	"context"
	"testing"

	"nix-ai-help/internal/ai/functionbase"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSnippetsFunction(t *testing.T) {
	function := NewSnippetsFunction()

	assert.NotNil(t, function)
	assert.Equal(t, "snippets", function.Name())
	assert.Equal(t, "Manage NixOS configuration snippets for reusable code blocks and common patterns", function.Description())
	assert.NotNil(t, function.logger)

	// Test schema
	schema := function.Schema()
	assert.Equal(t, "snippets", schema.Name)
	assert.NotEmpty(t, schema.Parameters)

	// Verify required parameters
	operationParam := findParameter(schema.Parameters, "operation")
	require.NotNil(t, operationParam)
	assert.True(t, operationParam.Required)
	assert.Contains(t, operationParam.Enum, "list")
	assert.Contains(t, operationParam.Enum, "search")
	assert.Contains(t, operationParam.Enum, "show")
	assert.Contains(t, operationParam.Enum, "add")
}

func TestSnippetsFunction_ValidateParameters(t *testing.T) {
	function := NewSnippetsFunction()

	tests := []struct {
		name        string
		params      map[string]interface{}
		expectError bool
	}{
		{
			name: "valid list operation",
			params: map[string]interface{}{
				"operation": "list",
			},
			expectError: false,
		},
		{
			name: "valid search operation",
			params: map[string]interface{}{
				"operation": "search",
				"query":     "gaming",
			},
			expectError: false,
		},
		{
			name: "valid show operation",
			params: map[string]interface{}{
				"operation": "show",
				"name":      "nvidia-gaming",
			},
			expectError: false,
		},
		{
			name: "valid add operation",
			params: map[string]interface{}{
				"operation":   "add",
				"name":        "test-snippet",
				"content":     "# Test content",
				"description": "Test snippet",
				"category":    "custom",
			},
			expectError: false,
		},
		{
			name:        "missing operation parameter",
			params:      map[string]interface{}{},
			expectError: true,
		},
		{
			name: "invalid operation",
			params: map[string]interface{}{
				"operation": "invalid-operation",
			},
			expectError: true,
		},
		{
			name: "non-string operation",
			params: map[string]interface{}{
				"operation": 123,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := function.ValidateParameters(tt.params)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSnippetsFunction_Execute_List(t *testing.T) {
	function := NewSnippetsFunction()
	ctx := context.Background()

	params := map[string]interface{}{
		"operation": "list",
	}

	result, err := function.Execute(ctx, params, nil)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)

	data, ok := result.Data.(map[string]interface{})
	require.True(t, ok)

	assert.Equal(t, "list", data["operation"])
	assert.Contains(t, data, "total_count")
	assert.Contains(t, data, "snippets")
	assert.Contains(t, data, "categories")
	assert.Contains(t, data, "popular_tags")
	assert.Contains(t, data, "statistics")

	// Verify snippets structure
	snippets, ok := data["snippets"].([]map[string]interface{})
	require.True(t, ok)
	assert.Greater(t, len(snippets), 0)

	for _, snippet := range snippets {
		assert.Contains(t, snippet, "name")
		assert.Contains(t, snippet, "description")
		assert.Contains(t, snippet, "category")
		assert.Contains(t, snippet, "tags")
		assert.Contains(t, snippet, "language")
	}
}

func TestSnippetsFunction_Execute_ListWithFilter(t *testing.T) {
	function := NewSnippetsFunction()
	ctx := context.Background()

	params := map[string]interface{}{
		"operation": "list",
		"filter": map[string]interface{}{
			"category": "gaming",
		},
	}

	result, err := function.Execute(ctx, params, nil)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)

	data, ok := result.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "list", data["operation"])
}

func TestSnippetsFunction_Execute_Search(t *testing.T) {
	function := NewSnippetsFunction()
	ctx := context.Background()

	tests := []struct {
		name     string
		query    string
		expected int
	}{
		{
			name:     "search gaming",
			query:    "gaming",
			expected: 1,
		},
		{
			name:     "search nvidia",
			query:    "nvidia",
			expected: 1,
		},
		{
			name:     "search security",
			query:    "security",
			expected: 1,
		},
		{
			name:     "search ssh",
			query:    "ssh",
			expected: 1,
		},
		{
			name:     "search nonexistent",
			query:    "nonexistent",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := map[string]interface{}{
				"operation": "search",
				"query":     tt.query,
			}

			result, err := function.Execute(ctx, params, nil)

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.True(t, result.Success)

			data, ok := result.Data.(map[string]interface{})
			require.True(t, ok)

			assert.Equal(t, "search", data["operation"])
			assert.Equal(t, tt.query, data["query"])
			assert.Equal(t, tt.expected, data["results_count"])
			assert.Contains(t, data, "results")
			assert.Contains(t, data, "suggestions")
		})
	}
}

func TestSnippetsFunction_Execute_SearchMissingQuery(t *testing.T) {
	function := NewSnippetsFunction()
	ctx := context.Background()

	params := map[string]interface{}{
		"operation": "search",
	}

	result, err := function.Execute(ctx, params, nil)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "query parameter is required")
}

func TestSnippetsFunction_Execute_Show(t *testing.T) {
	function := NewSnippetsFunction()
	ctx := context.Background()

	tests := []struct {
		name        string
		snippetName string
		expectError bool
	}{
		{
			name:        "show nvidia-gaming snippet",
			snippetName: "nvidia-gaming",
			expectError: false,
		},
		{
			name:        "show ssh-hardening snippet",
			snippetName: "ssh-hardening",
			expectError: false,
		},
		{
			name:        "show nonexistent snippet",
			snippetName: "nonexistent",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := map[string]interface{}{
				"operation": "show",
				"name":      tt.snippetName,
			}

			result, err := function.Execute(ctx, params, nil)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.True(t, result.Success)

				data, ok := result.Data.(map[string]interface{})
				require.True(t, ok)

				assert.Equal(t, tt.snippetName, data["name"])
				assert.Contains(t, data, "description")
				assert.Contains(t, data, "content")
				assert.Contains(t, data, "category")
				assert.Contains(t, data, "tags")
			}
		})
	}
}

func TestSnippetsFunction_Execute_ShowMissingName(t *testing.T) {
	function := NewSnippetsFunction()
	ctx := context.Background()

	params := map[string]interface{}{
		"operation": "show",
	}

	result, err := function.Execute(ctx, params, nil)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "name parameter is required")
}

func TestSnippetsFunction_Execute_Add(t *testing.T) {
	function := NewSnippetsFunction()
	ctx := context.Background()

	params := map[string]interface{}{
		"operation":   "add",
		"name":        "test-snippet",
		"content":     "# Test configuration\nservices.test.enable = true;",
		"description": "Test snippet for unit testing",
		"category":    "development",
		"tags":        []interface{}{"test", "development", "sample"},
	}

	result, err := function.Execute(ctx, params, nil)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)

	data, ok := result.Data.(map[string]interface{})
	require.True(t, ok)

	assert.Equal(t, "add", data["operation"])
	assert.Equal(t, "test-snippet", data["name"])
	assert.Equal(t, "Test snippet for unit testing", data["description"])
	assert.Equal(t, "development", data["category"])
	assert.Contains(t, data, "content")
	assert.Contains(t, data, "tags")
	assert.Contains(t, data, "saved_to")
}

func TestSnippetsFunction_Execute_AddMissingParameters(t *testing.T) {
	function := NewSnippetsFunction()
	ctx := context.Background()

	tests := []struct {
		name        string
		params      map[string]interface{}
		expectedErr string
	}{
		{
			name: "missing name",
			params: map[string]interface{}{
				"operation": "add",
				"content":   "test content",
			},
			expectedErr: "name parameter is required",
		},
		{
			name: "missing content",
			params: map[string]interface{}{
				"operation": "add",
				"name":      "test-snippet",
			},
			expectedErr: "content parameter is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := function.Execute(ctx, tt.params, nil)

			assert.Error(t, err)
			assert.Nil(t, result)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestSnippetsFunction_Execute_Remove(t *testing.T) {
	function := NewSnippetsFunction()
	ctx := context.Background()

	params := map[string]interface{}{
		"operation": "remove",
		"name":      "test-snippet",
	}

	result, err := function.Execute(ctx, params, nil)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)

	data, ok := result.Data.(map[string]interface{})
	require.True(t, ok)

	assert.Equal(t, "remove", data["operation"])
	assert.Equal(t, "test-snippet", data["name"])
	assert.Equal(t, "success", data["status"])
	assert.Contains(t, data, "removed_from")
	assert.Contains(t, data, "backup_created")
}

func TestSnippetsFunction_Execute_Apply(t *testing.T) {
	function := NewSnippetsFunction()
	ctx := context.Background()

	params := map[string]interface{}{
		"operation":   "apply",
		"name":        "nvidia-gaming",
		"output_path": "/etc/nixos/configuration.nix",
		"merge":       true,
	}

	result, err := function.Execute(ctx, params, nil)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)

	data, ok := result.Data.(map[string]interface{})
	require.True(t, ok)

	assert.Equal(t, "apply", data["operation"])
	assert.Equal(t, "nvidia-gaming", data["snippet_name"])
	assert.Equal(t, "/etc/nixos/configuration.nix", data["output_path"])
	assert.Equal(t, true, data["merge_mode"])
	assert.Contains(t, data, "changes")
	assert.Contains(t, data, "next_steps")
}

func TestSnippetsFunction_Execute_Edit(t *testing.T) {
	function := NewSnippetsFunction()
	ctx := context.Background()

	params := map[string]interface{}{
		"operation":   "edit",
		"name":        "test-snippet",
		"content":     "# Updated content",
		"description": "Updated description",
	}

	result, err := function.Execute(ctx, params, nil)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)
}

func TestSnippetsFunction_Execute_Export(t *testing.T) {
	function := NewSnippetsFunction()
	ctx := context.Background()

	params := map[string]interface{}{
		"operation":   "export",
		"output_path": "/tmp/snippets-export.yaml",
	}

	result, err := function.Execute(ctx, params, nil)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)
}

func TestSnippetsFunction_Execute_Import(t *testing.T) {
	function := NewSnippetsFunction()
	ctx := context.Background()

	params := map[string]interface{}{
		"operation":   "import",
		"source_path": "/tmp/snippets-import.yaml",
	}

	result, err := function.Execute(ctx, params, nil)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)
}

func TestSnippetsFunction_Execute_Organize(t *testing.T) {
	function := NewSnippetsFunction()
	ctx := context.Background()

	params := map[string]interface{}{
		"operation": "organize",
	}

	result, err := function.Execute(ctx, params, nil)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)
}

func TestSnippetsFunction_Execute_InvalidOperation(t *testing.T) {
	function := NewSnippetsFunction()
	ctx := context.Background()

	params := map[string]interface{}{
		"operation": "invalid-operation",
	}

	result, err := function.Execute(ctx, params, nil)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "unsupported snippet operation")
}

func TestSnippetsFunction_Execute_MissingOperation(t *testing.T) {
	function := NewSnippetsFunction()
	ctx := context.Background()

	params := map[string]interface{}{}

	result, err := function.Execute(ctx, params, nil)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "operation parameter is required")
}

func TestSnippetsFunction_matchesFilter(t *testing.T) {
	function := NewSnippetsFunction()

	snippet := map[string]interface{}{
		"name":     "test-snippet",
		"category": "gaming",
		"tags":     []string{"nvidia", "gaming"},
		"language": "nix",
	}

	tests := []struct {
		name     string
		filter   map[string]interface{}
		expected bool
	}{
		{
			name:     "matching category",
			filter:   map[string]interface{}{"category": "gaming"},
			expected: true,
		},
		{
			name:     "non-matching category",
			filter:   map[string]interface{}{"category": "security"},
			expected: false,
		},
		{
			name:     "matching language",
			filter:   map[string]interface{}{"language": "nix"},
			expected: true,
		},
		{
			name:     "empty filter",
			filter:   map[string]interface{}{},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := function.matchesFilter(snippet, tt.filter)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper function to find a parameter by name
func findParameter(params []functionbase.FunctionParameter, name string) *functionbase.FunctionParameter {
	for i := range params {
		if params[i].Name == name {
			return &params[i]
		}
	}
	return nil
}
