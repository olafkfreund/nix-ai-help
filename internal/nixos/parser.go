package nixos

import (
	"errors"
	"strings"
)

// ParseLog takes a log output as a string and returns a structured representation of the log.
func ParseLog(log string) (map[string]interface{}, error) {
	if log == "" {
		return nil, errors.New("log input is empty")
	}

	parsedLog := make(map[string]interface{})
	lines := strings.Split(log, "\n")

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		// Simple parsing logic: split by the first space
		parts := strings.SplitN(line, " ", 2)
		if len(parts) < 2 {
			parsedLog[parts[0]] = nil
		} else {
			parsedLog[parts[0]] = parts[1]
		}
	}

	return parsedLog, nil
}

// ParseNixConfig takes a NixOS configuration file content as a string and returns a structured representation.
func ParseNixConfig(config string) (map[string]interface{}, error) {
	if config == "" {
		return nil, errors.New("configuration input is empty")
	}

	parsedConfig := make(map[string]interface{})
	lines := strings.Split(config, "\n")

	for _, line := range lines {
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "#") {
			continue // Skip empty lines and comments
		}
		// Simple parsing logic: split by the first '='
		parts := strings.SplitN(line, "=", 2)
		if len(parts) < 2 {
			parsedConfig[parts[0]] = nil
		} else {
			parsedConfig[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}

	return parsedConfig, nil
}
