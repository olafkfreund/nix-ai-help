package nixos

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// Executor provides functionality to execute local commands related to NixOS configuration.
type Executor struct {
	ConfigPath string
}

// NewExecutor creates a new instance of Executor with an optional config path.
func NewExecutor(configPath string) *Executor {
	return &Executor{ConfigPath: configPath}
}

// ExecuteCommand executes a given command and returns its output.
func (e *Executor) ExecuteCommand(command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	if e.ConfigPath != "" {
		cmd.Dir = e.ConfigPath
	}
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// ExecuteNixCommand executes a NixOS specific command and returns the output.
func (e *Executor) ExecuteNixCommand(command string) (string, error) {
	return e.ExecuteCommand("nix", strings.Fields(command)...)
}

// SearchNixPackages searches for Nix packages using `nix search nixpkgs <query> --json` and returns a parsed result.
// Now supports fuzzy matching for multi-word queries.
func (e *Executor) SearchNixPackages(query string) (string, error) {
	args := []string{"search", "nixpkgs", "--json"}
	if strings.TrimSpace(query) != "" {
		queryTerm := strings.TrimSpace(query)
		args = append(args, queryTerm)
	}
	output, err := e.ExecuteCommand("nix", args...)
	if err != nil {
		return output, err
	}
	// Sanitize output: extract JSON object from first '{' to last '}'
	start := strings.Index(output, "{")
	end := strings.LastIndex(output, "}")
	if start == -1 || end == -1 || end <= start {
		return output, fmt.Errorf("could not find JSON object in output")
	}
	jsonStr := output[start : end+1]
	// Parse JSON output
	type NixPkg struct {
		AttrPath    string   `json:"attrPath"`
		Pname       string   `json:"pname"`
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Homepage    string   `json:"homepage"`
		Version     string   `json:"version"`
		Platforms   []string `json:"platforms"`
	}
	var pkgs map[string]NixPkg
	err = json.Unmarshal([]byte(jsonStr), &pkgs)
	if err != nil {
		return output, err
	}

	// Fuzzy match: if no results and query has multiple words, try to match any word in name/description
	if len(pkgs) == 0 && strings.Contains(query, " ") {
		// Try again with regex '^' to get all packages, then filter
		allPkgsOut, err := e.ExecuteCommand("nix", "search", "nixpkgs", "^", "--json")
		if err == nil {
			start := strings.Index(allPkgsOut, "{")
			end := strings.LastIndex(allPkgsOut, "}")
			if start != -1 && end != -1 && end > start {
				allPkgsJson := allPkgsOut[start : end+1]
				var allPkgs map[string]NixPkg
				if json.Unmarshal([]byte(allPkgsJson), &allPkgs) == nil {
					words := strings.Fields(strings.ToLower(query))
					for attr, pkg := range allPkgs {
						name := strings.ToLower(pkg.Pname)
						desc := strings.ToLower(pkg.Description)
						match := false
						for _, w := range words {
							if strings.Contains(name, w) || strings.Contains(desc, w) {
								match = true
								break
							}
						}
						if match {
							pkgs[attr] = pkg
						}
					}
				}
			}
		}
	}

	// ANSI color codes
	blue := "\033[1;34m"
	reset := "\033[0m"
	header := blue + "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"
	header += "📦 Nixpkgs Package Results" + reset + "\n"
	header += blue + "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" + reset + "\n"
	var lines []string
	lines = append(lines, header)
	for attr, pkg := range pkgs {
		desc := pkg.Description
		if desc == "" {
			desc = pkg.Name
		}
		name := pkg.Pname
		if name == "" {
			name = pkg.Name
		}
		// Blue bullet for package
		line := blue + "• " + name + reset + " (" + attr + ") - " + desc
		if pkg.Version != "" {
			line += " [v" + pkg.Version + "]"
		}
		if pkg.Homepage != "" {
			line += "\n    " + blue + pkg.Homepage + reset
		}
		lines = append(lines, line)
	}
	lines = append(lines, blue+"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"+reset)
	return strings.Join(lines, "\n"), nil
}

// SearchNixPackagesForAutocomplete searches for Nix packages using `nix search nixpkgs <query> --json` and returns a list of package names for autocomplete.
// Now supports fuzzy matching for multi-word queries and a max result limit for performance.
func (e *Executor) SearchNixPackagesForAutocomplete(query string, max int) ([]string, error) {
	args := []string{"search", "nixpkgs", "--json"}
	if strings.TrimSpace(query) != "" {
		queryTerm := strings.TrimSpace(query)
		args = append(args, queryTerm)
	}
	output, err := e.ExecuteCommand("nix", args...)
	if err != nil {
		return nil, err
	}
	start := strings.Index(output, "{")
	end := strings.LastIndex(output, "}")
	if start == -1 || end == -1 || end <= start {
		return nil, fmt.Errorf("could not find JSON object in output")
	}
	jsonStr := output[start : end+1]
	type NixPkg struct {
		AttrPath    string `json:"attrPath"`
		Pname       string `json:"pname"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	var pkgs map[string]NixPkg
	err = json.Unmarshal([]byte(jsonStr), &pkgs)
	if err != nil {
		return nil, err
	}
	// Fuzzy match: if no results and query has multiple words, try to match any word in name/description
	if len(pkgs) == 0 && strings.Contains(query, " ") {
		allPkgsOut, err := e.ExecuteCommand("nix", "search", "nixpkgs", "^", "--json")
		if err == nil {
			start := strings.Index(allPkgsOut, "{")
			end := strings.LastIndex(allPkgsOut, "}")
			if start != -1 && end != -1 && end > start {
				allPkgsJson := allPkgsOut[start : end+1]
				var allPkgs map[string]NixPkg
				if json.Unmarshal([]byte(allPkgsJson), &allPkgs) == nil {
					words := strings.Fields(strings.ToLower(query))
					for attr, pkg := range allPkgs {
						name := strings.ToLower(pkg.Pname)
						desc := strings.ToLower(pkg.Description)
						match := false
						for _, w := range words {
							if strings.Contains(name, w) || strings.Contains(desc, w) {
								match = true
								break
							}
						}
						if match {
							pkgs[attr] = pkg
						}
					}
				}
			}
		}
	}
	// Collect up to max package names for autocomplete
	var names []string
	for _, pkg := range pkgs {
		if pkg.Pname != "" {
			names = append(names, pkg.Pname)
		} else if pkg.Name != "" {
			names = append(names, pkg.Name)
		}
		if max > 0 && len(names) >= max {
			break
		}
	}
	return names, nil
}

// SearchNixServices searches for NixOS services using `nix search nixos` and returns the output.
func (e *Executor) SearchNixServices(query string) (string, error) {
	return e.ExecuteCommand("nix", "search", "nixos", query)
}

// InstallNixPackage installs a package using `nix-env -iA` and returns the output.
func (e *Executor) InstallNixPackage(attr string) (string, error) {
	return e.ExecuteCommand("nix-env", "-iA", attr)
}

// ShowNixOSOptions runs 'nixos-option <option>' and returns the output.
func (e *Executor) ShowNixOSOptions(option string) (string, error) {
	return e.ExecuteCommand("nixos-option", option)
}

// ListServiceOptions lists all available options for a given NixOS service using 'nixos-option --find services.<name>'.
func (e *Executor) ListServiceOptions(service string) (string, error) {
	if strings.TrimSpace(service) == "" {
		return "", fmt.Errorf("service name is required")
	}
	return e.ExecuteCommand("nixos-option", "--find", fmt.Sprintf("services.%s", service))
}
