package packages

import (
	"context"
	"testing"

	"nix-ai-help/internal/ai/functionbase"
)

func TestPackagesFunction_Implementation(t *testing.T) {
	function := NewPackagesFunction()

	// Test that function implements required interface
	var _ functionbase.FunctionInterface = function

	// Test function name
	if function.Name() != "packages" {
		t.Errorf("Expected function name 'packages', got '%s'", function.Name())
	}

	// Test function description
	desc := function.Description()
	if desc == "" {
		t.Error("Function description should not be empty")
	}

	// Test parameters
	params := function.Schema().Parameters
	if len(params) == 0 {
		t.Error("Function should have parameters")
	}

	// Verify required parameters exist
	paramNames := make(map[string]bool)
	for _, param := range params {
		paramNames[param.Name] = true
	}

	requiredParams := []string{"operation"}
	for _, reqParam := range requiredParams {
		if !paramNames[reqParam] {
			t.Errorf("Required parameter '%s' not found", reqParam)
		}
	}
}

func TestPackagesFunction_Execute_BasicOperations(t *testing.T) {
	function := NewPackagesFunction()
	ctx := context.Background()

	testCases := []struct {
		name   string
		params map[string]interface{}
		valid  bool
	}{
		{
			name: "Search operation",
			params: map[string]interface{}{
				"operation":    "search",
				"search_query": "text editor",
			},
			valid: true,
		},
		{
			name: "Install operation",
			params: map[string]interface{}{
				"operation":    "install",
				"package_name": "vim",
			},
			valid: true,
		},
		{
			name: "Info operation",
			params: map[string]interface{}{
				"operation":    "info",
				"package_name": "git",
			},
			valid: true,
		},
		{
			name: "List operation",
			params: map[string]interface{}{
				"operation": "list",
			},
			valid: true,
		},
		{
			name: "Invalid operation",
			params: map[string]interface{}{
				"operation": "invalid_op",
			},
			valid: false,
		},
		{
			name: "Missing operation",
			params: map[string]interface{}{
				"package_name": "vim",
			},
			valid: false,
		},
		{
			name: "Search without query or package name",
			params: map[string]interface{}{
				"operation": "search",
			},
			valid: false,
		},
		{
			name: "Install without package name",
			params: map[string]interface{}{
				"operation": "install",
			},
			valid: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			options := &functionbase.FunctionOptions{}
			result, err := function.Execute(ctx, tc.params, options)

			if tc.valid {
				if err != nil {
					t.Errorf("Expected no error for valid case, got: %v", err)
				}
				if result == nil {
					t.Error("Expected result for valid case, got nil")
				}
			} else {
				if err == nil {
					t.Error("Expected error for invalid case, got none")
				}
			}
		})
	}
}

func TestPackagesFunction_RequestValidation(t *testing.T) {
	function := NewPackagesFunction()

	testCases := []struct {
		name    string
		request *PackagesRequest
		valid   bool
	}{
		{
			name: "Valid search request",
			request: &PackagesRequest{
				Operation:   "search",
				SearchQuery: "text editor",
			},
			valid: true,
		},
		{
			name: "Valid install request",
			request: &PackagesRequest{
				Operation:   "install",
				PackageName: "vim",
			},
			valid: true,
		},
		{
			name: "Valid info request",
			request: &PackagesRequest{
				Operation:   "info",
				PackageName: "git",
				ShowDetails: true,
			},
			valid: true,
		},
		{
			name: "Valid versions request",
			request: &PackagesRequest{
				Operation:    "versions",
				PackageName:  "nodejs",
				ShowVersions: true,
			},
			valid: true,
		},
		{
			name: "Invalid operation",
			request: &PackagesRequest{
				Operation: "invalid",
			},
			valid: false,
		},
		{
			name: "Install without package name",
			request: &PackagesRequest{
				Operation: "install",
			},
			valid: false,
		},
		{
			name: "Search without query or package name",
			request: &PackagesRequest{
				Operation: "search",
			},
			valid: false,
		},
		{
			name: "Compare without package name",
			request: &PackagesRequest{
				Operation: "compare",
			},
			valid: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := function.validateRequest(tc.request)
			if tc.valid && err != nil {
				t.Errorf("Expected valid request, got error: %v", err)
			}
			if !tc.valid && err == nil {
				t.Error("Expected invalid request, got no error")
			}
		})
	}
}

func TestPackagesFunction_ParameterParsing(t *testing.T) {
	function := NewPackagesFunction()

	testCases := []struct {
		name     string
		params   map[string]interface{}
		expected *PackagesRequest
		hasError bool
	}{
		{
			name: "Complete search request",
			params: map[string]interface{}{
				"operation":      "search",
				"search_query":   "development tools",
				"channel":        "unstable",
				"include_unfree": true,
				"limit":          20,
				"sort_by":        "popularity",
			},
			expected: &PackagesRequest{
				Operation:     "search",
				SearchQuery:   "development tools",
				Channel:       "unstable",
				IncludeUnfree: true,
				Limit:         20,
				SortBy:        "popularity",
			},
			hasError: false,
		},
		{
			name: "Install request with config type",
			params: map[string]interface{}{
				"operation":      "install",
				"package_name":   "firefox",
				"install_method": "declarative",
				"config_type":    "system",
			},
			expected: &PackagesRequest{
				Operation:     "install",
				PackageName:   "firefox",
				InstallMethod: "declarative",
				ConfigType:    "system",
				Limit:         10,          // default
				SortBy:        "relevance", // default
			},
			hasError: false,
		},
		{
			name: "Missing operation",
			params: map[string]interface{}{
				"package_name": "vim",
			},
			expected: nil,
			hasError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := function.parseRequest(tc.params)

			if tc.hasError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result.Operation != tc.expected.Operation {
				t.Errorf("Expected operation %s, got %s", tc.expected.Operation, result.Operation)
			}

			if result.PackageName != tc.expected.PackageName {
				t.Errorf("Expected package_name %s, got %s", tc.expected.PackageName, result.PackageName)
			}

			if result.SearchQuery != tc.expected.SearchQuery {
				t.Errorf("Expected search_query %s, got %s", tc.expected.SearchQuery, result.SearchQuery)
			}

			if result.Limit != tc.expected.Limit {
				t.Errorf("Expected limit %d, got %d", tc.expected.Limit, result.Limit)
			}

			if result.SortBy != tc.expected.SortBy {
				t.Errorf("Expected sort_by %s, got %s", tc.expected.SortBy, result.SortBy)
			}
		})
	}
}

func TestPackagesFunction_Operations(t *testing.T) {
	function := NewPackagesFunction()

	// Test valid operations
	validOps := []string{
		"search", "install", "info", "list", "update", "remove",
		"versions", "dependencies", "alternatives", "compare", "help",
	}

	for _, op := range validOps {
		t.Run("Valid_"+op, func(t *testing.T) {
			request := &PackagesRequest{Operation: op}

			// Add required fields for specific operations
			switch op {
			case "install", "info", "versions", "dependencies", "remove", "compare":
				request.PackageName = "test-package"
			case "search":
				request.SearchQuery = "test query"
			}

			err := function.validateRequest(request)
			if err != nil {
				t.Errorf("Operation %s should be valid, got error: %v", op, err)
			}
		})
	}

	// Test invalid operations
	invalidOps := []string{"invalid", "bad", "unknown"}
	for _, op := range invalidOps {
		t.Run("Invalid_"+op, func(t *testing.T) {
			request := &PackagesRequest{Operation: op}
			err := function.validateRequest(request)
			if err == nil {
				t.Errorf("Operation %s should be invalid", op)
			}
		})
	}
}
