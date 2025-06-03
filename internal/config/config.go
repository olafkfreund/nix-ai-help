package config

import (
	"encoding/json"
	"os"
	"os/user"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// EmbeddedDefaultConfig contains the default configuration YAML that gets compiled into the binary.
// This eliminates the need for external config files when installing via nix build.
const EmbeddedDefaultConfig = `default:
    ai_provider: ollama  # Options: openai, ollama, gemini, custom
    ai_model: llama3
    # Custom AI provider configuration (used if ai_provider: custom)
    custom_ai:
        base_url: http://localhost:8080/api/generate  # HTTP API endpoint URL
        headers:  # Optional custom headers (e.g., for authentication)
            Authorization: "Bearer your-api-key-here"
            # Content-Type: "application/json"  # Set automatically if not provided
    log_level: info
    mcp_server:
        host: localhost
        port: 8081
        socket_path: /tmp/nixai-mcp.sock
        auto_start: false
        documentation_sources:
            - https://wiki.nixos.org/wiki/NixOS_Wiki
            - https://nix.dev/manual/nix
            - https://nixos.org/manual/nixpkgs/stable/
            - https://nix.dev/manual/nix/2.28/language/
            - https://nix-community.github.io/home-manager/
    nixos:
        config_path: /etc/nixos/configuration.nix
        log_path: /var/log/nixos.log
    diagnostics:
        enabled: true
        threshold: 5
        error_patterns:
            - name: example_error
              pattern: '(?i)example error regex'
              error_type: custom
              severity: high
              description: Example error description
    commands:
        timeout: 30
        retries: 3
    devenv:
        default_directory: "."
        auto_init_git: true
        templates:
            python:
                enabled: true
                default_version: "311"
                default_package_manager: "pip"
            rust:
                enabled: true
                default_version: "stable"
            nodejs:
                enabled: true
                default_version: "20"
                default_package_manager: "npm"
            golang:
                enabled: true
                default_version: "1.21"
    discourse:
        base_url: "https://discourse.nixos.org"
        api_key: ""  # Optional: set via DISCOURSE_API_KEY environment variable
        username: ""  # Optional: set via DISCOURSE_USERNAME environment variable
        enabled: true
`

type Config struct {
	AIProvider string `json:"ai_provider"`
	MCPServer  string `json:"mcp_server"`
	LogLevel   string `json:"log_level"`
	// Add other configuration fields as needed
}

type MCPServerConfig struct {
	Host                 string   `yaml:"host" json:"host"`
	Port                 int      `yaml:"port" json:"port"`
	SocketPath           string   `yaml:"socket_path" json:"socket_path"`
	AutoStart            bool     `yaml:"auto_start" json:"auto_start"`
	DocumentationSources []string `yaml:"documentation_sources" json:"documentation_sources"`
}

type NixosConfig struct {
	ConfigPath string `yaml:"config_path" json:"config_path"`
	LogPath    string `yaml:"log_path" json:"log_path"`
}

// ErrorPatternConfig allows user-defined error patterns for diagnostics
// Pattern is a regex string
// Example YAML:
//   - name: my_error
//     pattern: '(?i)my error regex'
//     error_type: custom
//     severity: high
//     description: My custom error

type ErrorPatternConfig struct {
	Name        string `yaml:"name" json:"name"`
	Pattern     string `yaml:"pattern" json:"pattern"`
	ErrorType   string `yaml:"error_type" json:"error_type"`
	Severity    string `yaml:"severity" json:"severity"`
	Description string `yaml:"description" json:"description"`
}

type DiagnosticsConfig struct {
	Enabled       bool                 `yaml:"enabled" json:"enabled"`
	Threshold     int                  `yaml:"threshold" json:"threshold"`
	ErrorPatterns []ErrorPatternConfig `yaml:"error_patterns" json:"error_patterns"`
}

type CommandsConfig struct {
	Timeout int `yaml:"timeout" json:"timeout"`
	Retries int `yaml:"retries"`
}

type DevenvTemplateConfig struct {
	Enabled               bool   `yaml:"enabled" json:"enabled"`
	DefaultVersion        string `yaml:"default_version" json:"default_version"`
	DefaultPackageManager string `yaml:"default_package_manager" json:"default_package_manager"`
}

type DevenvConfig struct {
	DefaultDirectory string                          `yaml:"default_directory" json:"default_directory"`
	AutoInitGit      bool                            `yaml:"auto_init_git" json:"auto_init_git"`
	Templates        map[string]DevenvTemplateConfig `yaml:"templates" json:"templates"`
}

// CustomAIConfig holds config for a custom AI provider
type CustomAIConfig struct {
	BaseURL string            `yaml:"base_url" json:"base_url"`
	Headers map[string]string `yaml:"headers" json:"headers"`
}

// DiscourseConfig holds configuration for Discourse integration
type DiscourseConfig struct {
	BaseURL  string `yaml:"base_url" json:"base_url"`
	APIKey   string `yaml:"api_key" json:"api_key"`
	Username string `yaml:"username" json:"username"`
	Enabled  bool   `yaml:"enabled" json:"enabled"`
}

type YAMLConfig struct {
	AIProvider  string            `yaml:"ai_provider" json:"ai_provider"`
	LogLevel    string            `yaml:"log_level" json:"log_level"`
	MCPServer   MCPServerConfig   `yaml:"mcp_server" json:"mcp_server"`
	Nixos       NixosConfig       `yaml:"nixos" json:"nixos"`
	Diagnostics DiagnosticsConfig `yaml:"diagnostics" json:"diagnostics"`
	Commands    CommandsConfig    `yaml:"commands" json:"commands"`
	Devenv      DevenvConfig      `yaml:"devenv" json:"devenv"`
	CustomAI    CustomAIConfig    `yaml:"custom_ai" json:"custom_ai"`
	Discourse   DiscourseConfig   `yaml:"discourse" json:"discourse"`
}

type UserConfig struct {
	AIProvider  string            `yaml:"ai_provider" json:"ai_provider"`
	AIModel     string            `yaml:"ai_model" json:"ai_model"`
	NixosFolder string            `yaml:"nixos_folder" json:"nixos_folder"`
	LogLevel    string            `yaml:"log_level" json:"log_level"`
	MCPServer   MCPServerConfig   `yaml:"mcp_server" json:"mcp_server"`
	Nixos       NixosConfig       `yaml:"nixos" json:"nixos"`
	Diagnostics DiagnosticsConfig `yaml:"diagnostics" json:"diagnostics"`
	Commands    CommandsConfig    `yaml:"commands" json:"commands"`
	Devenv      DevenvConfig      `yaml:"devenv" json:"devenv"`
	CustomAI    CustomAIConfig    `yaml:"custom_ai" json:"custom_ai"`
	Discourse   DiscourseConfig   `yaml:"discourse" json:"discourse"`
}

func DefaultUserConfig() *UserConfig {
	return &UserConfig{
		AIProvider:  "ollama",
		AIModel:     "llama3",
		NixosFolder: "~/nixos-config",
		LogLevel:    "info",
		MCPServer: MCPServerConfig{
			Host:       "localhost",
			Port:       8081,
			SocketPath: "/tmp/nixai-mcp.sock",
			AutoStart:  false,
			DocumentationSources: []string{
				"https://wiki.nixos.org/wiki/NixOS_Wiki",
				"https://nix.dev/manual/nix",
				"https://nixos.org/manual/nixpkgs/stable/",
				"https://nix.dev/manual/nix/2.28/language/",
				"https://nix-community.github.io/home-manager/",
			},
		},
		Nixos: NixosConfig{
			ConfigPath: "~/nixos-config/configuration.nix",
			LogPath:    "/var/log/nixos/nixos-rebuild.log",
		},
		Diagnostics: DiagnosticsConfig{
			Enabled:   true,
			Threshold: 1,
			ErrorPatterns: []ErrorPatternConfig{
				{
					Name:        "example_error",
					Pattern:     "(?i)example error regex",
					ErrorType:   "custom",
					Severity:    "high",
					Description: "Example error description",
				},
			},
		},
		Commands: CommandsConfig{Timeout: 30, Retries: 2},
		Devenv: DevenvConfig{
			DefaultDirectory: ".",
			AutoInitGit:      true,
			Templates: map[string]DevenvTemplateConfig{
				"python": {
					Enabled:               true,
					DefaultVersion:        "311",
					DefaultPackageManager: "pip",
				},
				"rust": {
					Enabled:        true,
					DefaultVersion: "stable",
				},
				"nodejs": {
					Enabled:               true,
					DefaultVersion:        "20",
					DefaultPackageManager: "npm",
				},
				"golang": {
					Enabled:        true,
					DefaultVersion: "1.21",
				},
			},
		},
		Discourse: DiscourseConfig{
			BaseURL:  "https://discourse.nixos.org",
			APIKey:   "", // Optional, can be set via environment variable
			Username: "", // Optional, can be set via environment variable
			Enabled:  true,
		},
	}
}

func ConfigFilePath() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	configDir := filepath.Join(usr.HomeDir, ".config", "nixai")
	return filepath.Join(configDir, "config.yaml"), nil
}

func EnsureConfigFile() (string, error) {
	path, err := ConfigFilePath()
	if err != nil {
		return "", err
	}
	// #nosec G304 -- Config file paths are validated and not user-supplied
	if _, err := os.Stat(path); os.IsNotExist(err) {
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0700); err != nil {
			return "", err
		}
		cfg := DefaultUserConfig()
		data, err := yaml.Marshal(cfg)
		if err != nil {
			return "", err
		}
		// #nosec G306 -- Config files are not sensitive, 0644 is intentional for user config
		if err := os.WriteFile(path, data, 0600); err != nil {
			return "", err
		}
	}
	return path, nil
}

func LoadUserConfig() (*UserConfig, error) {
	path, err := EnsureConfigFile()
	if err != nil {
		return nil, err
	}
	// #nosec G304 -- Config file paths are validated and not user-supplied
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg UserConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func SaveUserConfig(cfg *UserConfig) error {
	path, err := ConfigFilePath()
	if err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	// #nosec G306 -- Config files are not sensitive, 0644 is intentional for user config
	return os.WriteFile(path, data, 0600)
}

func LoadConfig(filePath string) (*Config, error) {
	// #nosec G304 -- Config file paths are validated and not user-supplied
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(bytes, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func SaveConfig(filePath string, config *Config) error {
	bytes, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	// #nosec G306 -- Config files are not sensitive, 0644 is intentional for user config
	return os.WriteFile(filePath, bytes, 0644)
}

func LoadYAMLConfig(filePath string) (*YAMLConfig, error) {
	// #nosec G304 -- Config file paths are validated and not user-supplied
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var config struct {
		Default YAMLConfig `yaml:"default"`
	}
	if err := yaml.Unmarshal(bytes, &config); err != nil {
		return nil, err
	}

	return &config.Default, nil
}

// LoadEmbeddedYAMLConfig loads the embedded YAML configuration
func LoadEmbeddedYAMLConfig() (*YAMLConfig, error) {
	var config struct {
		Default YAMLConfig `yaml:"default"`
	}
	if err := yaml.Unmarshal([]byte(EmbeddedDefaultConfig), &config); err != nil {
		return nil, err
	}

	return &config.Default, nil
}

// EnsureConfigFileFromEmbedded creates user config from embedded default if it doesn't exist
func EnsureConfigFileFromEmbedded() (string, error) {
	path, err := ConfigFilePath()
	if err != nil {
		return "", err
	}

	// If config file doesn't exist, create it from embedded default
	if _, err := os.Stat(path); os.IsNotExist(err) {
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0700); err != nil {
			return "", err
		}

		// Parse embedded config and extract the content under 'default:' key
		embeddedCfg, err := LoadEmbeddedYAMLConfig()
		if err != nil {
			return "", err
		}

		// Convert to UserConfig structure and write as YAML
		userCfg := &UserConfig{
			AIProvider:  embeddedCfg.AIProvider,
			AIModel:     "llama3",         // Default model
			NixosFolder: "~/nixos-config", // Default folder
			LogLevel:    embeddedCfg.LogLevel,
			MCPServer:   embeddedCfg.MCPServer,
			Nixos:       embeddedCfg.Nixos,
			Diagnostics: embeddedCfg.Diagnostics,
			Commands:    embeddedCfg.Commands,
			Devenv:      embeddedCfg.Devenv,
			CustomAI:    embeddedCfg.CustomAI,
			Discourse:   embeddedCfg.Discourse,
		}

		// Marshal to YAML and write to user config file
		data, err := yaml.Marshal(userCfg)
		if err != nil {
			return "", err
		}

		if err := os.WriteFile(path, data, 0600); err != nil {
			return "", err
		}
	}
	return path, nil
}
