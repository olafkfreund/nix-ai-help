package config

import (
	"encoding/json"
	"os"
	"os/user"
	"path/filepath"
	"time"

	yaml "gopkg.in/yaml.v3"
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
    # Basic AI models configuration (subset for embedded config)
    ai_models:
        providers:
            ollama:
                name: "Ollama"
                description: "Local AI model provider"
                type: "local"
                base_url: "http://localhost:11434"
                available: true
                supports_streaming: true
                supports_tools: true
                requires_api_key: false
                models:
                    llama3:
                        name: "Llama 3"
                        description: "Meta's Llama 3 model"
                        type: "chat"
                        context_window: 8192
                        max_tokens: 4096
                        recommended_for: ["nixos", "general"]
                        default: true
        selection_preferences:
            default_provider: "ollama"
            default_models:
                ollama: "llama3"
            task_models:
                nixos_config:
                    primary: ["ollama:llama3"]
                    fallback: []
        discovery:
            auto_discover: true
            cache_duration: 3600
            check_timeout: 10
            max_retries: 2
    log_level: info
    mcp_server:
        host: localhost
        port: 8081
        socket_path: /tmp/nixai-mcp.sock
        auto_start: false
        documentation_sources:
            - https://wiki.nixos.org/wiki/NixOS_Wiki
            - https://nix.dev/manual/nix
            - https://nix.dev/ 
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

// NixOSContext represents the detected NixOS configuration context
type NixOSContext struct {
	// System Detection
	UsesFlakes      bool   `yaml:"uses_flakes" json:"uses_flakes"`
	UsesChannels    bool   `yaml:"uses_channels" json:"uses_channels"`
	NixOSConfigPath string `yaml:"nixos_config_path" json:"nixos_config_path"`
	SystemType      string `yaml:"system_type" json:"system_type"` // "nixos", "nix-darwin", "home-manager-only", "unknown"

	// Home Manager
	HasHomeManager        bool   `yaml:"has_home_manager" json:"has_home_manager"`
	HomeManagerType       string `yaml:"home_manager_type" json:"home_manager_type"` // "standalone", "module", "none"
	HomeManagerConfigPath string `yaml:"home_manager_config_path" json:"home_manager_config_path"`

	// Version Information
	NixOSVersion string `yaml:"nixos_version" json:"nixos_version"`
	NixVersion   string `yaml:"nix_version" json:"nix_version"`

	// Configuration Analysis
	ConfigurationFiles []string `yaml:"configuration_files" json:"configuration_files"`
	EnabledServices    []string `yaml:"enabled_services" json:"enabled_services"`
	InstalledPackages  []string `yaml:"installed_packages" json:"installed_packages"`

	// File Paths
	FlakeFile         string `yaml:"flake_file" json:"flake_file"`
	ConfigurationNix  string `yaml:"configuration_nix" json:"configuration_nix"`
	HardwareConfigNix string `yaml:"hardware_config_nix" json:"hardware_config_nix"`

	// Cache Information
	LastDetected    time.Time `yaml:"last_detected" json:"last_detected"`
	CacheValid      bool      `yaml:"cache_valid" json:"cache_valid"`
	DetectionErrors []string  `yaml:"detection_errors,omitempty" json:"detection_errors,omitempty"`
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

// AI Models Configuration Structures

// AIModelConfig represents a single AI model configuration
type AIModelConfig struct {
	Name             string   `yaml:"name" json:"name"`
	Description      string   `yaml:"description" json:"description"`
	Size             string   `yaml:"size,omitempty" json:"size,omitempty"`
	Type             string   `yaml:"type" json:"type"` // chat, code, completion
	ContextWindow    int      `yaml:"context_window" json:"context_window"`
	MaxTokens        int      `yaml:"max_tokens" json:"max_tokens"`
	RecommendedFor   []string `yaml:"recommended_for" json:"recommended_for"`
	RequiresDownload bool     `yaml:"requires_download,omitempty" json:"requires_download,omitempty"`
	CostTier         string   `yaml:"cost_tier,omitempty" json:"cost_tier,omitempty"` // budget, standard, premium
	Default          bool     `yaml:"default,omitempty" json:"default,omitempty"`
}

// AIProviderConfig represents a single AI provider configuration
type AIProviderConfig struct {
	Name              string                   `yaml:"name" json:"name"`
	Description       string                   `yaml:"description" json:"description"`
	Type              string                   `yaml:"type" json:"type"` // local, cloud, custom
	BaseURL           string                   `yaml:"base_url" json:"base_url"`
	Available         bool                     `yaml:"available" json:"available"`
	SupportsStreaming bool                     `yaml:"supports_streaming" json:"supports_streaming"`
	SupportsTools     bool                     `yaml:"supports_tools" json:"supports_tools"`
	RequiresAPIKey    bool                     `yaml:"requires_api_key" json:"requires_api_key"`
	EnvVar            string                   `yaml:"env_var,omitempty" json:"env_var,omitempty"`
	Models            map[string]AIModelConfig `yaml:"models" json:"models"`
}

// TaskModelPreferences represents model preferences for specific tasks
type TaskModelPreferences struct {
	Primary  []string `yaml:"primary" json:"primary"`
	Fallback []string `yaml:"fallback" json:"fallback"`
}

// AISelectionPreferences represents model selection preferences
type AISelectionPreferences struct {
	DefaultProvider string                          `yaml:"default_provider" json:"default_provider"`
	DefaultModels   map[string]string               `yaml:"default_models" json:"default_models"`
	TaskModels      map[string]TaskModelPreferences `yaml:"task_models" json:"task_models"`
}

// AIDiscoveryConfig represents model discovery configuration
type AIDiscoveryConfig struct {
	AutoDiscover  bool `yaml:"auto_discover" json:"auto_discover"`
	CacheDuration int  `yaml:"cache_duration" json:"cache_duration"`
	CheckTimeout  int  `yaml:"check_timeout" json:"check_timeout"`
	MaxRetries    int  `yaml:"max_retries" json:"max_retries"`
}

// AIModelsConfig represents the complete AI models configuration
type AIModelsConfig struct {
	Providers            map[string]AIProviderConfig `yaml:"providers" json:"providers"`
	SelectionPreferences AISelectionPreferences      `yaml:"selection_preferences" json:"selection_preferences"`
	Discovery            AIDiscoveryConfig           `yaml:"discovery" json:"discovery"`
}

type YAMLConfig struct {
	AIProvider  string            `yaml:"ai_provider" json:"ai_provider"`
	LogLevel    string            `yaml:"log_level" json:"log_level"`
	AIModels    AIModelsConfig    `yaml:"ai_models" json:"ai_models"`
	MCPServer   MCPServerConfig   `yaml:"mcp_server" json:"mcp_server"`
	Nixos       NixosConfig       `yaml:"nixos" json:"nixos"`
	Diagnostics DiagnosticsConfig `yaml:"diagnostics" json:"diagnostics"`
	Commands    CommandsConfig    `yaml:"commands" json:"commands"`
	Devenv      DevenvConfig      `yaml:"devenv" json:"devenv"`
	CustomAI    CustomAIConfig    `yaml:"custom_ai" json:"custom_ai"`
	Discourse   DiscourseConfig   `yaml:"discourse" json:"discourse"`
}

type UserConfig struct {
	AIProvider   string            `yaml:"ai_provider" json:"ai_provider"`
	AIModel      string            `yaml:"ai_model" json:"ai_model"`
	NixosFolder  string            `yaml:"nixos_folder" json:"nixos_folder"`
	LogLevel     string            `yaml:"log_level" json:"log_level"`
	AIModels     AIModelsConfig    `yaml:"ai_models" json:"ai_models"`
	MCPServer    MCPServerConfig   `yaml:"mcp_server" json:"mcp_server"`
	Nixos        NixosConfig       `yaml:"nixos" json:"nixos"`
	Diagnostics  DiagnosticsConfig `yaml:"diagnostics" json:"diagnostics"`
	Commands     CommandsConfig    `yaml:"commands" json:"commands"`
	Devenv       DevenvConfig      `yaml:"devenv" json:"devenv"`
	CustomAI     CustomAIConfig    `yaml:"custom_ai" json:"custom_ai"`
	Discourse    DiscourseConfig   `yaml:"discourse" json:"discourse"`
	NixOSContext NixOSContext      `yaml:"nixos_context" json:"nixos_context"`
}

func DefaultUserConfig() *UserConfig {
	return &UserConfig{
		AIProvider:  "ollama",
		AIModel:     "llama3",
		NixosFolder: "~/nixos-config",
		LogLevel:    "info",
		AIModels: AIModelsConfig{
			Providers: map[string]AIProviderConfig{
				"ollama": {
					Name:              "Ollama",
					Description:       "Local AI model provider for privacy-focused inference",
					Type:              "local",
					BaseURL:           "http://localhost:11434",
					Available:         true,
					SupportsStreaming: true,
					SupportsTools:     true,
					RequiresAPIKey:    false,
					Models: map[string]AIModelConfig{
						"llama3": {
							Name:             "Llama 3",
							Description:      "Meta's Llama 3 model for general-purpose tasks",
							Size:             "8B",
							Type:             "chat",
							ContextWindow:    8192,
							MaxTokens:        4096,
							RecommendedFor:   []string{"nixos", "general", "coding"},
							RequiresDownload: true,
							Default:          true,
						},
					},
				},
				"gemini": {
					Name:              "Google Gemini",
					Description:       "Google's advanced AI models via API",
					Type:              "cloud",
					BaseURL:           "https://generativelanguage.googleapis.com",
					Available:         true,
					SupportsStreaming: true,
					SupportsTools:     true,
					RequiresAPIKey:    true,
					EnvVar:            "GEMINI_API_KEY",
					Models: map[string]AIModelConfig{
						"gemini-1.5-flash": {
							Name:           "Gemini 1.5 Flash",
							Description:    "Fast and efficient Gemini model",
							Type:           "chat",
							ContextWindow:  1048576,
							MaxTokens:      8192,
							RecommendedFor: []string{"fast", "general", "nixos"},
							CostTier:       "standard",
							Default:        true,
						},
					},
				},
				"openai": {
					Name:              "OpenAI",
					Description:       "OpenAI's GPT models via API",
					Type:              "cloud",
					BaseURL:           "https://api.openai.com",
					Available:         true,
					SupportsStreaming: true,
					SupportsTools:     true,
					RequiresAPIKey:    true,
					EnvVar:            "OPENAI_API_KEY",
					Models: map[string]AIModelConfig{
						"gpt-3.5-turbo": {
							Name:           "GPT-3.5 Turbo",
							Description:    "Fast and cost-effective model",
							Type:           "chat",
							ContextWindow:  16385,
							MaxTokens:      4096,
							RecommendedFor: []string{"general", "fast", "budget"},
							CostTier:       "standard",
							Default:        true,
						},
					},
				},
			},
			SelectionPreferences: AISelectionPreferences{
				DefaultProvider: "ollama",
				DefaultModels: map[string]string{
					"ollama": "llama3",
					"gemini": "gemini-1.5-flash",
					"openai": "gpt-3.5-turbo",
				},
				TaskModels: map[string]TaskModelPreferences{
					"nixos_config": {
						Primary:  []string{"ollama:llama3", "gemini:gemini-1.5-flash"},
						Fallback: []string{"openai:gpt-3.5-turbo"},
					},
					"general_help": {
						Primary:  []string{"ollama:llama3", "gemini:gemini-1.5-flash"},
						Fallback: []string{"openai:gpt-3.5-turbo"},
					},
				},
			},
			Discovery: AIDiscoveryConfig{
				AutoDiscover:  true,
				CacheDuration: 3600,
				CheckTimeout:  10,
				MaxRetries:    2,
			},
		},
		MCPServer: MCPServerConfig{
			Host:       "localhost",
			Port:       8081,
			SocketPath: "/tmp/nixai-mcp.sock",
			AutoStart:  false,
			DocumentationSources: []string{
				"https://wiki.nixos.org/wiki/NixOS_Wiki",
				"https://nix.dev/manual/nix",
				"https://nix.dev/",
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
		NixOSContext: NixOSContext{
			UsesFlakes:            false,
			UsesChannels:          false,
			NixOSConfigPath:       "",
			SystemType:            "unknown",
			HasHomeManager:        false,
			HomeManagerType:       "none",
			HomeManagerConfigPath: "",
			NixOSVersion:          "",
			NixVersion:            "",
			ConfigurationFiles:    []string{},
			EnabledServices:       []string{},
			InstalledPackages:     []string{},
			FlakeFile:             "",
			ConfigurationNix:      "",
			HardwareConfigNix:     "",
			LastDetected:          time.Time{},
			CacheValid:            false,
			DetectionErrors:       []string{},
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
			AIModels:    embeddedCfg.AIModels,
			MCPServer:   embeddedCfg.MCPServer,
			Nixos:       embeddedCfg.Nixos,
			Diagnostics: embeddedCfg.Diagnostics,
			Commands:    embeddedCfg.Commands,
			Devenv:      embeddedCfg.Devenv,
			CustomAI:    embeddedCfg.CustomAI,
			Discourse:   embeddedCfg.Discourse,
			NixOSContext: NixOSContext{
				UsesFlakes:            false,
				UsesChannels:          false,
				NixOSConfigPath:       "",
				SystemType:            "unknown",
				HasHomeManager:        false,
				HomeManagerType:       "none",
				HomeManagerConfigPath: "",
				NixOSVersion:          "",
				NixVersion:            "",
				ConfigurationFiles:    []string{},
				EnabledServices:       []string{},
				InstalledPackages:     []string{},
				FlakeFile:             "",
				ConfigurationNix:      "",
				HardwareConfigNix:     "",
				LastDetected:          time.Time{},
				CacheValid:            false,
				DetectionErrors:       []string{},
			},
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
