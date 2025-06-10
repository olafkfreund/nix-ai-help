package utils

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// IsFile checks if the given path is a file.
func IsFile(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// IsDirectory checks if the given path is a directory.
func IsDirectory(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

// DirExists checks if the given path exists and is a directory.
func DirExists(path string) bool {
	return IsDirectory(path)
}

// SplitLines splits a string into a slice of lines.
func SplitLines(input string) []string {
	return strings.Split(strings.TrimSpace(input), "\n")
}

// Contains checks if a slice of strings contains a specific string.
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// ValidatePath checks if the provided path is valid and returns an error if not.
func ValidatePath(path string) error {
	if path == "" {
		return errors.New("path cannot be empty")
	}
	if !IsFile(path) && !IsDirectory(path) {
		return errors.New("path does not exist")
	}
	return nil
}

// ExpandHome expands the '~/' prefix in a path to the user's home directory.
func ExpandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		usr, _ := user.Current()
		return filepath.Join(usr.HomeDir, path[2:])
	}
	return path
}

// GetConfigDir returns the config directory for nixai, respecting XDG_CONFIG_HOME or defaulting to $HOME/.config/nixai
func GetConfigDir() (string, error) {
	xdg := os.Getenv("XDG_CONFIG_HOME")
	if xdg != "" {
		return filepath.Join(xdg, "nixai"), nil
	}
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return filepath.Join(usr.HomeDir, ".config", "nixai"), nil
}

// GenerateID generates a unique ID for community resources
func GenerateID() string {
	return fmt.Sprintf("nixai_%d_%s", time.Now().Unix(), randomString(8))
}

// ParseTags parses a comma-separated tag string into a slice
func ParseTags(tagStr string) []string {
	if tagStr == "" {
		return []string{}
	}

	tags := strings.Split(tagStr, ",")
	var result []string
	for _, tag := range tags {
		trimmed := strings.TrimSpace(tag)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// ParseFloat safely parses a string to float64, returning 0 on error
func ParseFloat(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return f
}

// ParseInt safely parses a string to int, returning 0 on error
func ParseInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return i
}

// FindJSONStart finds the starting position of JSON content in a string
func FindJSONStart(s string) int {
	return strings.Index(s, "{")
}

// FindJSONEnd finds the ending position of JSON content starting from a given position
func FindJSONEnd(s string, start int) int {
	if start < 0 || start >= len(s) {
		return -1
	}

	depth := 0
	for i := start; i < len(s); i++ {
		switch s[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return i + 1
			}
		}
	}
	return -1
}

// randomString generates a random string of specified length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// LookPath checks if a command is available in PATH.
func LookPath(cmd string) (string, error) {
	return exec.LookPath(cmd)
}

// RunCommand runs a command with arguments and streams output to stdout/stderr.
func RunCommand(cmd string, args ...string) error {
	c := exec.Command(cmd, args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

// FlakeHasDeployConfig checks if flake.nix contains a deploy-rs config.
func FlakeHasDeployConfig(flakePath string) bool {
	f, err := os.Open(flakePath)
	if err != nil {
		return false
	}
	defer func() { _ = f.Close() }()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), "deploy") {
			return true
		}
	}
	return false
}

// PromptYesNo prompts the user for a yes/no answer and returns true for yes
func PromptYesNo(prompt string) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s (y/N): ", prompt)
		response, err := reader.ReadString('\n')
		if err != nil {
			return false
		}

		response = strings.TrimSpace(strings.ToLower(response))
		switch response {
		case "y", "yes":
			return true
		case "n", "no", "":
			return false
		default:
			fmt.Println("Please answer with 'y' or 'n' (or press Enter for 'no')")
		}
	}
}

// GenerateMinimalDeployConfig appends a minimal deploy-rs config to flake.nix.
func GenerateMinimalDeployConfig(flakePath string) error {
	f, err := os.OpenFile(flakePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	config := `\n  # deploy-rs minimal config\n  deploy = {\n    nodes = {\n      example = {\n        hostname = \"your-host\";\n        sshUser = \"nixos\";\n        profiles.system = {\n          path = self.nixosConfigurations.example.config.system.build.toplevel;\n        };\n      };\n    };\n  };\n`
	_, err = f.WriteString(config)
	return err
}

// GenerateDeployRsConfig creates a comprehensive deploy-rs configuration based on existing nixosConfigurations
func GenerateDeployRsConfig(flakeDir string, interactive bool) error {
	// Get the flake.nix path
	flakeFile := filepath.Join(flakeDir, "flake.nix")
	if !IsFile(flakeFile) {
		return fmt.Errorf("flake.nix not found at %s", flakeFile)
	}

	// Get existing hosts from nixosConfigurations
	hosts, err := GetFlakeHosts(flakeDir)
	if err != nil {
		return fmt.Errorf("failed to get hosts from flake.nix: %v", err)
	}

	if len(hosts) == 0 {
		return fmt.Errorf("no hosts found in nixosConfigurations")
	}

	// Read the current flake.nix content
	content, err := os.ReadFile(flakeFile)
	if err != nil {
		return fmt.Errorf("failed to read flake.nix: %v", err)
	}

	flakeContent := string(content)

	// Check if deploy-rs input is already present
	if !strings.Contains(flakeContent, "deploy-rs") {
		// Add deploy-rs input
		if err := addDeployRsInput(flakeFile, flakeContent); err != nil {
			return fmt.Errorf("failed to add deploy-rs input: %v", err)
		}
		// Re-read the updated content
		content, _ = os.ReadFile(flakeFile)
		flakeContent = string(content)
	}

	// Generate deploy configuration
	deployConfig, err := generateDeployNodes(hosts, interactive)
	if err != nil {
		return fmt.Errorf("failed to generate deploy nodes: %v", err)
	}

	// Add deploy configuration to flake
	return addDeployConfig(flakeFile, flakeContent, deployConfig)
}

// addDeployRsInput adds deploy-rs to the flake inputs
func addDeployRsInput(flakeFile, content string) error {
	// Find the inputs section
	inputsStart := strings.Index(content, "inputs = {")
	if inputsStart == -1 {
		return fmt.Errorf("inputs section not found in flake.nix")
	}

	// Find the end of inputs section
	inputsEnd := findClosingBrace(content, inputsStart+10)
	if inputsEnd == -1 {
		return fmt.Errorf("could not find end of inputs section")
	}

	// Insert deploy-rs input before the closing brace
	beforeInputsEnd := content[:inputsEnd-1]
	afterInputsEnd := content[inputsEnd-1:]

	// Check if there are other inputs (to add comma if needed)
	hasOtherInputs := strings.Contains(beforeInputsEnd[inputsStart:], ".url")
	deployInput := ""
	if hasOtherInputs {
		deployInput = "\n    deploy-rs.url = \"github:serokell/deploy-rs\";"
	} else {
		deployInput = "\n    deploy-rs.url = \"github:serokell/deploy-rs\";"
	}

	newContent := beforeInputsEnd + deployInput + "\n  " + afterInputsEnd

	return os.WriteFile(flakeFile, []byte(newContent), 0644)
}

// findClosingBrace finds the matching closing brace for an opening brace
func findClosingBrace(content string, start int) int {
	depth := 1
	for i := start; i < len(content); i++ {
		switch content[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return i + 1
			}
		}
	}
	return -1
}

// generateDeployNodes creates deploy configuration for each host
func generateDeployNodes(hosts []string, interactive bool) (string, error) {
	var nodes []string

	for _, host := range hosts {
		var hostname, sshUser string

		if interactive {
			// Prompt for hostname
			fmt.Printf("Enter hostname/IP for '%s' (press Enter for '%s'): ", host, host)
			var input string
			_, _ = fmt.Scanln(&input)
			if strings.TrimSpace(input) == "" {
				hostname = host
			} else {
				hostname = strings.TrimSpace(input)
			}

			// Prompt for SSH user
			fmt.Printf("Enter SSH user for '%s' (press Enter for 'nixos'): ", host)
			_, _ = fmt.Scanln(&input)
			if strings.TrimSpace(input) == "" {
				sshUser = "nixos"
			} else {
				sshUser = strings.TrimSpace(input)
			}
		} else {
			// Use defaults for non-interactive mode
			hostname = host
			sshUser = "nixos"
		}

		node := fmt.Sprintf(`      %s = {
        hostname = "%s";
        sshUser = "%s";
        profiles.system = {
          path = self.nixosConfigurations.%s.config.system.build.toplevel;
        };
      };`, host, hostname, sshUser, host)

		nodes = append(nodes, node)
	}

	return fmt.Sprintf(`
  # Deploy-rs configuration generated by nixai
  deploy = {
    nodes = {
%s
    };
  };`, strings.Join(nodes, "\n")), nil
}

// addDeployConfig adds the deploy configuration to the flake outputs
func addDeployConfig(flakeFile, content, deployConfig string) error {
	// Find the outputs function and its opening brace
	outputsPattern := "outputs = {"
	outputsStart := strings.Index(content, outputsPattern)
	if outputsStart == -1 {
		return fmt.Errorf("outputs section not found in flake.nix")
	}

	// Look for the final closing brace of the outputs section
	// We need to find the very last closing brace before the end of the file
	lastBracePos := strings.LastIndex(content, "};")
	if lastBracePos == -1 {
		return fmt.Errorf("could not find closing of outputs section")
	}

	// Insert deploy config before the last closing brace
	beforeLastBrace := content[:lastBracePos]
	afterLastBrace := content[lastBracePos:]

	newContent := beforeLastBrace + deployConfig + "\n" + afterLastBrace

	return os.WriteFile(flakeFile, []byte(newContent), 0644)
}

// GetFlakeHosts returns a list of hostnames from nixosConfigurations in flake.nix.
// If flakePath is empty, it defaults to ~/.config/nixos/
func GetFlakeHosts(flakePath string, debug ...bool) ([]string, error) {
	// Check environment variable for debug mode
	envDebug := os.Getenv("NIXAI_DEBUG") == "1"
	isDebug := envDebug || (len(debug) > 0 && debug[0])

	// Ensure nix command is available
	nixPath, err := exec.LookPath("nix")
	if err != nil {
		if isDebug {
			fmt.Fprintf(os.Stderr, "[nixai debug] nix command not found: %v\n", err)
		}
		return nil, fmt.Errorf("nix command not found: %v", err)
	}

	if isDebug {
		fmt.Fprintf(os.Stderr, "[nixai debug] Using nix binary: %s\n", nixPath)
	}

	// Determine the flake directory
	var flakeDir string
	if flakePath != "" {
		flakeDir = flakePath
	} else {
		// Default to ~/.config/nixos/ for NixOS configurations
		usr, err := user.Current()
		if err != nil {
			return nil, fmt.Errorf("failed to get current user: %v", err)
		}
		flakeDir = filepath.Join(usr.HomeDir, ".config", "nixos")
	}

	// Expand home directory if needed
	flakeDir = ExpandHome(flakeDir)

	if isDebug {
		fmt.Fprintf(os.Stderr, "[nixai debug] Using flake directory: %s\n", flakeDir)
	}

	// Check if flake.nix exists in the directory
	flakeFile := filepath.Join(flakeDir, "flake.nix")
	if !IsFile(flakeFile) {
		return nil, fmt.Errorf("flake.nix not found at %s", flakeFile)
	}

	cmdArgs := []string{"eval", "--json", ".#nixosConfigurations", "--apply", "builtins.attrNames"}
	cmd := exec.Command("nix", cmdArgs...)

	// Set working directory to the flake directory
	cmd.Dir = flakeDir
	if isDebug {
		fmt.Fprintf(os.Stderr, "[nixai debug] Setting working directory: %s\n", flakeDir)
	}

	// Set up environment
	cmd.Env = os.Environ()

	if isDebug {
		fmt.Fprintf(os.Stderr, "[nixai debug] Full command: %s %s\n", nixPath, strings.Join(cmdArgs, " "))
		fmt.Fprintf(os.Stderr, "[nixai debug] Environment PATH: %s\n", os.Getenv("PATH"))
	}

	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run the command
	err = cmd.Run()
	if err != nil {
		errMsg := fmt.Sprintf("nix eval failed: %v", err)
		if isDebug {
			fmt.Fprintf(os.Stderr, "[nixai debug] %s\n", errMsg)
			if stderr.Len() > 0 {
				fmt.Fprintf(os.Stderr, "[nixai debug] nix stderr: %s\n", stderr.String())
			}
			if stdout.Len() > 0 {
				fmt.Fprintf(os.Stderr, "[nixai debug] nix stdout: %s\n", stdout.String())
			}
		}
		return nil, fmt.Errorf("%s: %s", errMsg, stderr.String())
	}

	output := stdout.String()
	if isDebug {
		fmt.Fprintf(os.Stderr, "[nixai debug] Raw output: %s\n", output)
	}

	var hosts []string
	err = json.Unmarshal([]byte(output), &hosts)
	if err != nil {
		errMsg := fmt.Sprintf("JSON unmarshal failed: %v", err)
		if isDebug {
			fmt.Fprintf(os.Stderr, "[nixai debug] %s\nOutput: %s\n", errMsg, output)
		}
		return nil, fmt.Errorf("%s: %s", errMsg, output)
	}

	if isDebug {
		fmt.Fprintf(os.Stderr, "[nixai debug] Successfully parsed hosts: %v\n", hosts)
	}

	return hosts, nil
}

// --- Snippets/Template helpers ---
// Minimal snippet/template types and utils for listing

type Snippet struct {
	Name        string
	Description string
	Path        string
}

type Template struct {
	Name        string
	Description string
	Path        string
}

// GetSnippetsDir returns the default snippets directory (e.g. ~/.config/nixai/snippets)
func GetSnippetsDir() string {
	dir, err := GetConfigDir()
	if err != nil {
		return "./snippets"
	}
	return filepath.Join(dir, "snippets")
}

// ListSnippets returns a list of snippets in the snippets directory
func ListSnippets(snippetDir string) ([]Snippet, error) {
	files, err := os.ReadDir(snippetDir)
	if err != nil {
		return nil, err
	}
	var snippets []Snippet
	for _, f := range files {
		if f.IsDir() || !strings.HasSuffix(f.Name(), ".nix") {
			continue
		}
		path := filepath.Join(snippetDir, f.Name())
		desc := ""
		file, err := os.Open(path)
		if err == nil {
			scanner := bufio.NewScanner(file)
			if scanner.Scan() {
				line := scanner.Text()
				if strings.HasPrefix(line, "#") {
					desc = strings.TrimPrefix(line, "#")
					desc = strings.TrimSpace(desc)
				}
			}
			file.Close()
		}
		snippets = append(snippets, Snippet{
			Name:        strings.TrimSuffix(f.Name(), ".nix"),
			Description: desc,
			Path:        path,
		})
	}
	return snippets, nil
}

// ListTemplates returns a list of templates in the templates directory
func ListTemplates() ([]Template, error) {
	dir, err := GetConfigDir()
	if err != nil {
		return nil, err
	}
	templateDir := filepath.Join(dir, "templates")
	files, err := os.ReadDir(templateDir)
	if err != nil {
		return nil, err
	}
	var templates []Template
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		path := filepath.Join(templateDir, f.Name())
		desc := ""
		file, err := os.Open(path)
		if err == nil {
			scanner := bufio.NewScanner(file)
			if scanner.Scan() {
				line := scanner.Text()
				if strings.HasPrefix(line, "#") {
					desc = strings.TrimPrefix(line, "#")
					desc = strings.TrimSpace(desc)
				}
			}
			file.Close()
		}
		templates = append(templates, Template{
			Name:        f.Name(),
			Description: desc,
			Path:        path,
		})
	}
	return templates, nil
}

// FormatDuration formats a time.Duration into a human-readable string
func FormatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%.0fms", float64(d.Nanoseconds())/1e6)
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.1fm", d.Minutes())
	}
	return fmt.Sprintf("%.1fh", d.Hours())
}
