package devenv

// Template represents a devenv template that can generate development environment configurations
type Template interface {
	Name() string
	Description() string
	Generate(config TemplateConfig) (*DevenvConfig, error)
	Validate(config TemplateConfig) error
	RequiredInputs() []InputField
	SupportedServices() []string
}

// TemplateConfig holds the configuration options for template generation
type TemplateConfig struct {
	ProjectName string            `yaml:"project_name"`
	Directory   string            `yaml:"directory"`
	Language    string            `yaml:"language"`
	Options     map[string]string `yaml:"options"`
	Services    []string          `yaml:"services"`
	Packages    []string          `yaml:"packages"`
	EnvVars     map[string]string `yaml:"env_vars"`
}

// InputField represents a configuration input field that templates can request
type InputField struct {
	Name        string   `yaml:"name"`
	Type        string   `yaml:"type"` // string, bool, choice, multi-choice
	Description string   `yaml:"description"`
	Required    bool     `yaml:"required"`
	Default     string   `yaml:"default"`
	Choices     []string `yaml:"choices,omitempty"`
}

// DevenvConfig represents the complete devenv.nix configuration
type DevenvConfig struct {
	Languages   map[string]interface{} `yaml:"languages"`
	Packages    []string               `yaml:"packages"`
	Services    map[string]interface{} `yaml:"services"`
	Environment map[string]string      `yaml:"environment"`
	Scripts     map[string]interface{} `yaml:"scripts"`
	PreCommit   map[string]interface{} `yaml:"pre_commit,omitempty"`
	EnterShell  string                 `yaml:"enter_shell,omitempty"`
	ExitShell   string                 `yaml:"exit_shell,omitempty"`
	Dotenv      map[string]interface{} `yaml:"dotenv,omitempty"`
}

// ServiceConfig represents configuration for a specific service
type ServiceConfig struct {
	Enable   bool                   `yaml:"enable"`
	Settings map[string]interface{} `yaml:"settings,omitempty"`
}
