package config

import (
	"encoding/json"
	"io/ioutil"
	"os"

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
	Retries int `yaml:"retries" json:"retries"`
}

type YAMLConfig struct {
	AIProvider  string            `yaml:"ai_provider" json:"ai_provider"`
	LogLevel    string            `yaml:"log_level" json:"log_level"`
	MCPServer   MCPServerConfig   `yaml:"mcp_server" json:"mcp_server"`
	Nixos       NixosConfig       `yaml:"nixos" json:"nixos"`
	Diagnostics DiagnosticsConfig `yaml:"diagnostics" json:"diagnostics"`
	Commands    CommandsConfig    `yaml:"commands" json:"commands"`
}

func LoadConfig(filePath string) (*Config, error) {
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

	return ioutil.WriteFile(filePath, bytes, 0644)
}

func LoadYAMLConfig(filePath string) (*YAMLConfig, error) {
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
