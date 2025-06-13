package devenv

import (
	"os"
	"path/filepath"
	"testing"

	"nix-ai-help/internal/ai"
	"nix-ai-help/pkg/logger"
)

// MockAIProvider for testing
type MockAIProvider struct {
	responses map[string]string
}

// Ensure MockAIProvider implements ai.AIProvider
var _ ai.AIProvider = (*MockAIProvider)(nil)

func NewMockAIProvider() *MockAIProvider {
	return &MockAIProvider{
		responses: make(map[string]string),
	}
}

func (m *MockAIProvider) SetResponse(prompt string, response string) {
	m.responses[prompt] = response
}

func (m *MockAIProvider) Query(prompt string) (string, error) {
	if response, exists := m.responses[prompt]; exists {
		return response, nil
	}
	return "python", nil // Default response for template suggestions
}

func (m *MockAIProvider) GenerateResponse(prompt string) (string, error) {
	return m.Query(prompt)
}

func TestService_NewService(t *testing.T) {
	mockAI := NewMockAIProvider()
	log := logger.NewLoggerWithLevel("debug")

	service, err := NewService(mockAI, log)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if service == nil {
		t.Fatal("Expected service to be created")
	}

	// Test that built-in templates are registered
	templates := service.ListTemplates()
	expectedTemplates := []string{"python", "rust", "nodejs", "golang"}

	for _, expected := range expectedTemplates {
		if _, exists := templates[expected]; !exists {
			t.Errorf("Expected template %s to be registered", expected)
		}
	}
}

func TestService_CreateProject_ValidInput(t *testing.T) {
	mockAI := NewMockAIProvider()
	log := logger.NewLoggerWithLevel("debug")
	service, err := NewService(mockAI, log)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	// Create temporary directory for testing
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "test-project")

	err = service.CreateProject("python", "test-project", projectDir,
		map[string]string{"python_version": "311"}, []string{"postgres"})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Check that devenv.nix was created
	devenvPath := filepath.Join(projectDir, "devenv.nix")
	if !fileExists(devenvPath) {
		t.Errorf("Expected devenv.nix to be created at %s", devenvPath)
	}

	// Check that additional files were created
	mainPath := filepath.Join(projectDir, "main.py")
	if !fileExists(mainPath) {
		t.Errorf("Expected main.py to be created at %s", mainPath)
	}

	reqPath := filepath.Join(projectDir, "requirements.txt")
	if !fileExists(reqPath) {
		t.Errorf("Expected requirements.txt to be created at %s", reqPath)
	}
}

func TestService_CreateProject_InvalidProjectName(t *testing.T) {
	mockAI := NewMockAIProvider()
	log := logger.NewLoggerWithLevel("debug")
	service, err := NewService(mockAI, log)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	tmpDir := t.TempDir()

	// Test invalid project names
	invalidNames := []string{
		"",          // empty
		"123test",   // starts with number
		"test@proj", // special characters
		"test proj", // spaces
	}

	for _, name := range invalidNames {
		err := service.CreateProject("python", name, tmpDir, map[string]string{}, []string{})
		if err == nil {
			t.Errorf("Expected error for invalid project name: %s", name)
		}
	}
}

func TestService_CreateProject_InvalidTemplate(t *testing.T) {
	mockAI := NewMockAIProvider()
	log := logger.NewLoggerWithLevel("debug")
	service, err := NewService(mockAI, log)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	tmpDir := t.TempDir()

	err = service.CreateProject("nonexistent", "test-project", tmpDir, map[string]string{}, []string{})
	if err == nil {
		t.Error("Expected error for nonexistent template")
	}
}

func TestService_CreateProject_InvalidServices(t *testing.T) {
	mockAI := NewMockAIProvider()
	log := logger.NewLoggerWithLevel("debug")
	service, err := NewService(mockAI, log)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	tmpDir := t.TempDir()

	// Test unsupported service for rust template
	err = service.CreateProject("rust", "test-project", tmpDir,
		map[string]string{}, []string{"mongodb"}) // rust template doesn't support mongodb
	if err == nil {
		t.Error("Expected error for unsupported service")
	}
}

func TestService_CreateProject_ExistingDevenvFile(t *testing.T) {
	mockAI := NewMockAIProvider()
	log := logger.NewLoggerWithLevel("debug")
	service, err := NewService(mockAI, log)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "existing-project")

	// Create directory and existing devenv.nix
	err = os.MkdirAll(projectDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	devenvPath := filepath.Join(projectDir, "devenv.nix")
	err = os.WriteFile(devenvPath, []byte("# existing file"), 0644)
	if err != nil {
		t.Fatalf("Failed to create existing devenv.nix: %v", err)
	}

	err = service.CreateProject("python", "existing-project", projectDir, map[string]string{}, []string{})
	if err == nil {
		t.Error("Expected error when devenv.nix already exists")
	}
}

func TestService_SuggestTemplate(t *testing.T) {
	mockAI := NewMockAIProvider()
	log := logger.NewLoggerWithLevel("debug")
	service, err := NewService(mockAI, log)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	// Set up mock response
	mockAI.SetResponse("", "python") // Mock will return "python" for any prompt

	suggestion, err := service.SuggestTemplate("web application with database")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if suggestion != "python" {
		t.Errorf("Expected suggestion 'python', got: %s", suggestion)
	}
}

func TestService_SuggestTemplate_NoAI(t *testing.T) {
	log := logger.NewLoggerWithLevel("debug")
	service, err := NewService(nil, log)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	_, err = service.SuggestTemplate("web application")
	if err == nil {
		t.Error("Expected error when AI provider is nil")
	}
}

func TestValidationFunctions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
		testFunc func(string) bool
	}{
		{"Valid project name", "my-project", true, isValidProjectName},
		{"Invalid project name - empty", "", false, isValidProjectName},
		{"Invalid project name - starts with number", "123project", false, isValidProjectName},
		{"Invalid project name - special chars", "my@project", false, isValidProjectName},
		{"Valid service name", "postgres", true, isValidServiceName},
		{"Invalid service name", "invalidservice", false, isValidServiceName},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.testFunc(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for input: %s", tt.expected, result, tt.input)
			}
		})
	}
}

func TestTemplate_PythonTemplate(t *testing.T) {
	template := &PythonTemplate{}

	// Test basic template properties
	if template.Name() != "python" {
		t.Errorf("Expected name 'python', got: %s", template.Name())
	}

	if template.Description() == "" {
		t.Error("Expected non-empty description")
	}

	// Test supported services
	services := template.SupportedServices()
	expectedServices := []string{"postgres", "redis", "mysql", "mongodb"}
	if len(services) != len(expectedServices) {
		t.Errorf("Expected %d services, got %d", len(expectedServices), len(services))
	}

	// Test config generation
	config := TemplateConfig{
		ProjectName: "test-project",
		Directory:   "/tmp/test",
		Language:    "python",
		Options:     map[string]string{"python_version": "311"},
		Services:    []string{"postgres"},
	}

	devenvConfig, err := template.Generate(config)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if devenvConfig == nil {
		t.Fatal("Expected devenv config to be generated")
		return
	}

	// Check Python language configuration
	if devenvConfig.Languages == nil {
		t.Error("Expected languages configuration to be present")
		return
	}

	if python, exists := devenvConfig.Languages["python"]; !exists {
		t.Error("Expected Python language configuration")
	} else {
		pythonConfig := python.(map[string]interface{})
		if pythonConfig["enable"] != true {
			t.Error("Expected Python to be enabled")
		}
		if pythonConfig["version"] != "311" {
			t.Errorf("Expected Python version 311, got: %v", pythonConfig["version"])
		}
	}

	// Check services configuration
	if devenvConfig.Services != nil {
		if postgres, exists := devenvConfig.Services["postgres"]; !exists {
			t.Error("Expected postgres service configuration")
		} else {
			postgresConfig := postgres.(map[string]interface{})
			if postgresConfig["enable"] != true {
				t.Error("Expected postgres to be enabled")
			}
		}
	}
}

func TestTemplate_RustTemplate(t *testing.T) {
	template := &RustTemplate{}

	config := TemplateConfig{
		ProjectName: "rust-project",
		Directory:   "/tmp/rust-test",
		Language:    "rust",
		Options:     map[string]string{"with_wasm": "true"},
		Services:    []string{},
	}

	devenvConfig, err := template.Generate(config)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Check Rust language configuration
	if rust, exists := devenvConfig.Languages["rust"]; !exists {
		t.Error("Expected Rust language configuration")
	} else {
		rustConfig := rust.(map[string]interface{})
		if rustConfig["enable"] != true {
			t.Error("Expected Rust to be enabled")
		}
	}

	// Check WASM package
	found := false
	for _, pkg := range devenvConfig.Packages {
		if pkg == "wasm-pack" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected wasm-pack package to be included")
	}
}

func TestTemplate_NodejsTemplate(t *testing.T) {
	template := &NodejsTemplate{}

	config := TemplateConfig{
		ProjectName: "node-project",
		Directory:   "/tmp/node-test",
		Language:    "nodejs",
		Options:     map[string]string{"package_manager": "yarn"},
		Services:    []string{},
	}

	devenvConfig, err := template.Generate(config)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Check Node.js language configuration
	if js, exists := devenvConfig.Languages["javascript"]; !exists {
		t.Error("Expected JavaScript language configuration")
	} else {
		jsConfig := js.(map[string]interface{})
		if jsConfig["enable"] != true {
			t.Error("Expected JavaScript to be enabled")
		}
	}

	// Check yarn package
	found := false
	for _, pkg := range devenvConfig.Packages {
		if pkg == "yarn" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected yarn package to be included")
	}
}

func TestTemplate_GolangTemplate(t *testing.T) {
	template := &GolangTemplate{}

	config := TemplateConfig{
		ProjectName: "go-project",
		Directory:   "/tmp/go-test",
		Language:    "golang",
		Options:     map[string]string{"with_grpc": "true"},
		Services:    []string{},
	}

	devenvConfig, err := template.Generate(config)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Check Go language configuration
	if golang, exists := devenvConfig.Languages["go"]; !exists {
		t.Error("Expected Go language configuration")
	} else {
		goConfig := golang.(map[string]interface{})
		if goConfig["enable"] != true {
			t.Error("Expected Go to be enabled")
		}
	}

	// Check gRPC packages
	grpcPackages := []string{"protobuf", "protoc-gen-go", "protoc-gen-go-grpc"}
	for _, expected := range grpcPackages {
		found := false
		for _, pkg := range devenvConfig.Packages {
			if pkg == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected %s package to be included", expected)
		}
	}
}

// Helper function to check if file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
