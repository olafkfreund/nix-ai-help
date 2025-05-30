package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	AIProvider string `json:"ai_provider"`
	MCPServer  string `json:"mcp_server"`
	LogLevel   string `json:"log_level"`
	// Add other configuration fields as needed
}

type MCPServerConfig struct {
	Host                 string   `yaml:"host" json:"host"`
	Port                 int      `yaml:"port" json:"port"`
	DocumentationSources []string `yaml:"documentation_sources" json:"documentation_sources"`
}

type NixosConfig struct {
	ConfigPath string `yaml:"config_path" json:"config_path"`
	LogPath    string `yaml:"log_path" json:"log_path"`
}

type DiagnosticsConfig struct {
	Enabled   bool `yaml:"enabled" json:"enabled"`
	Threshold int  `yaml:"threshold" json:"threshold"`
}

type CommandsConfig struct {
	Timeout int `yaml:"timeout" json:"timeout"`
	Retries int `yaml:"retries"`
}

type YAMLConfig struct {
	AIProvider  string            `yaml:"ai_provider" json:"ai_provider"`
	LogLevel    string            `yaml:"log_level" json:"log_level"`
	MCPServer   MCPServerConfig   `yaml:"mcp_server" json:"mcp_server"`
	Nixos       NixosConfig       `yaml:"nixos" json:"nixos"`
	Diagnostics DiagnosticsConfig `yaml:"diagnostics" json:"diagnostics"`
	Commands    CommandsConfig    `yaml:"commands" json:"commands"`
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
}

func DefaultUserConfig() *UserConfig {
	return &UserConfig{
		AIProvider:  "ollama",
		AIModel:     "llama3",
		NixosFolder: "~/nixos-config",
		LogLevel:    "info",
		MCPServer: MCPServerConfig{
			Host: "localhost",
			Port: 8080,
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
		Diagnostics: DiagnosticsConfig{Enabled: true, Threshold: 1},
		Commands:    CommandsConfig{Timeout: 30, Retries: 2},
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
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
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
	return ioutil.WriteFile(filePath, bytes, 0644)
}

func LoadYAMLConfig(filePath string) (*YAMLConfig, error) {
	// #nosec G304 -- Config file paths are validated and not user-supplied
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
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
