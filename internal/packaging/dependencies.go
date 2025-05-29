package packaging

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// analyzeNpmDependencies analyzes package.json for dependencies
func (ra *RepositoryAnalyzer) analyzeNpmDependencies(repoPath string) ([]Dependency, error) {
	var dependencies []Dependency

	packageJSONPath := filepath.Join(repoPath, "package.json")
	content, err := os.ReadFile(packageJSONPath)
	if err != nil {
		return dependencies, nil // No package.json found
	}

	var packageJSON struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}

	if err := json.Unmarshal(content, &packageJSON); err != nil {
		return dependencies, err
	}

	// Add runtime dependencies
	for name, version := range packageJSON.Dependencies {
		dependencies = append(dependencies, Dependency{
			Name:    name,
			Type:    "runtime",
			Version: version,
			System:  false,
		})
	}

	// Add development dependencies
	for name, version := range packageJSON.DevDependencies {
		dependencies = append(dependencies, Dependency{
			Name:    name,
			Type:    "dev",
			Version: version,
			System:  false,
		})
	}

	return dependencies, nil
}

// analyzeCargoDependencies analyzes Cargo.toml for dependencies
func (ra *RepositoryAnalyzer) analyzeCargoDependencies(repoPath string) ([]Dependency, error) {
	var dependencies []Dependency

	cargoTomlPath := filepath.Join(repoPath, "Cargo.toml")
	content, err := os.ReadFile(cargoTomlPath)
	if err != nil {
		return dependencies, nil
	}

	lines := strings.Split(string(content), "\n")
	inDependencies := false
	inDevDependencies := false
	inBuildDependencies := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Check for section headers
		if line == "[dependencies]" {
			inDependencies = true
			inDevDependencies = false
			inBuildDependencies = false
			continue
		} else if line == "[dev-dependencies]" {
			inDependencies = false
			inDevDependencies = true
			inBuildDependencies = false
			continue
		} else if line == "[build-dependencies]" {
			inDependencies = false
			inDevDependencies = false
			inBuildDependencies = true
			continue
		} else if strings.HasPrefix(line, "[") {
			inDependencies = false
			inDevDependencies = false
			inBuildDependencies = false
			continue
		}

		// Parse dependency lines
		if (inDependencies || inDevDependencies || inBuildDependencies) && strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				name := strings.TrimSpace(parts[0])
				version := strings.Trim(strings.TrimSpace(parts[1]), "\"")

				depType := "runtime"
				if inDevDependencies {
					depType = "dev"
				} else if inBuildDependencies {
					depType = "build"
				}

				dependencies = append(dependencies, Dependency{
					Name:    name,
					Type:    depType,
					Version: version,
					System:  false,
				})
			}
		}
	}

	return dependencies, nil
}

// analyzeGoDependencies analyzes go.mod for dependencies
func (ra *RepositoryAnalyzer) analyzeGoDependencies(repoPath string) ([]Dependency, error) {
	var dependencies []Dependency

	goModPath := filepath.Join(repoPath, "go.mod")
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return dependencies, nil
	}

	lines := strings.Split(string(content), "\n")
	inRequire := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "require (") {
			inRequire = true
			continue
		} else if line == ")" && inRequire {
			inRequire = false
			continue
		} else if strings.HasPrefix(line, "require ") {
			// Single-line require
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				dependencies = append(dependencies, Dependency{
					Name:    parts[1],
					Type:    "runtime",
					Version: parts[2],
					System:  false,
				})
			}
			continue
		}

		if inRequire {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				dependencies = append(dependencies, Dependency{
					Name:    parts[0],
					Type:    "runtime",
					Version: parts[1],
					System:  false,
				})
			}
		}
	}

	return dependencies, nil
}

// analyzePythonDependencies analyzes Python dependencies from various files
func (ra *RepositoryAnalyzer) analyzePythonDependencies(repoPath string) ([]Dependency, error) {
	var dependencies []Dependency

	// Check requirements.txt
	reqPath := filepath.Join(repoPath, "requirements.txt")
	if deps, err := ra.parseRequirementsTxt(reqPath); err == nil {
		dependencies = append(dependencies, deps...)
	}

	// Check setup.py
	setupPath := filepath.Join(repoPath, "setup.py")
	if deps, err := ra.parseSetupPy(setupPath); err == nil {
		dependencies = append(dependencies, deps...)
	}

	// Check pyproject.toml
	pyprojectPath := filepath.Join(repoPath, "pyproject.toml")
	if deps, err := ra.parsePyprojectToml(pyprojectPath); err == nil {
		dependencies = append(dependencies, deps...)
	}

	return dependencies, nil
}

// parseRequirementsTxt parses requirements.txt file
func (ra *RepositoryAnalyzer) parseRequirementsTxt(filePath string) ([]Dependency, error) {
	var dependencies []Dependency

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse requirement line (e.g., "package>=1.0.0")
		re := regexp.MustCompile(`^([a-zA-Z0-9_-]+)([>=<~!]+.*)?$`)
		matches := re.FindStringSubmatch(line)
		if len(matches) >= 2 {
			name := matches[1]
			version := ""
			if len(matches) > 2 {
				version = matches[2]
			}

			dependencies = append(dependencies, Dependency{
				Name:    name,
				Type:    "runtime",
				Version: version,
				System:  false,
			})
		}
	}

	return dependencies, scanner.Err()
}

// parseSetupPy attempts to extract dependencies from setup.py
func (ra *RepositoryAnalyzer) parseSetupPy(filePath string) ([]Dependency, error) {
	var dependencies []Dependency

	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Simple regex-based parsing for install_requires
	re := regexp.MustCompile(`install_requires\s*=\s*\[(.*?)\]`)
	matches := re.FindStringSubmatch(string(content))
	if len(matches) > 1 {
		reqsStr := matches[1]
		// Extract quoted strings
		reqRe := regexp.MustCompile(`['"]([^'"]+)['"]`)
		reqMatches := reqRe.FindAllStringSubmatch(reqsStr, -1)
		for _, match := range reqMatches {
			if len(match) > 1 {
				req := match[1]
				// Parse requirement
				nameRe := regexp.MustCompile(`^([a-zA-Z0-9_-]+)`)
				nameMatch := nameRe.FindStringSubmatch(req)
				if len(nameMatch) > 1 {
					dependencies = append(dependencies, Dependency{
						Name:    nameMatch[1],
						Type:    "runtime",
						Version: "",
						System:  false,
					})
				}
			}
		}
	}

	return dependencies, nil
}

// parsePyprojectToml parses pyproject.toml file
func (ra *RepositoryAnalyzer) parsePyprojectToml(filePath string) ([]Dependency, error) {
	var dependencies []Dependency

	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content), "\n")
	inDependencies := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line == "dependencies = [" {
			inDependencies = true
			continue
		} else if line == "]" && inDependencies {
			inDependencies = false
			continue
		}

		if inDependencies && strings.Contains(line, "\"") {
			// Extract dependency from quoted string
			re := regexp.MustCompile(`"([^"]+)"`)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				req := matches[1]
				nameRe := regexp.MustCompile(`^([a-zA-Z0-9_-]+)`)
				nameMatch := nameRe.FindStringSubmatch(req)
				if len(nameMatch) > 1 {
					dependencies = append(dependencies, Dependency{
						Name:    nameMatch[1],
						Type:    "runtime",
						Version: "",
						System:  false,
					})
				}
			}
		}
	}

	return dependencies, nil
}

// analyzeCMakeDependencies analyzes CMakeLists.txt for dependencies
func (ra *RepositoryAnalyzer) analyzeCMakeDependencies(repoPath string) ([]Dependency, error) {
	var dependencies []Dependency

	cmakeFiles := []string{
		filepath.Join(repoPath, "CMakeLists.txt"),
		filepath.Join(repoPath, "cmake", "CMakeLists.txt"),
	}

	for _, cmakeFile := range cmakeFiles {
		if deps, err := ra.parseCMakeFile(cmakeFile); err == nil {
			dependencies = append(dependencies, deps...)
		}
	}

	return dependencies, nil
}

// parseCMakeFile parses CMakeLists.txt for dependencies
func (ra *RepositoryAnalyzer) parseCMakeFile(filePath string) ([]Dependency, error) {
	var dependencies []Dependency

	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Look for find_package calls
	re := regexp.MustCompile(`find_package\s*\(\s*([a-zA-Z0-9_]+)`)
	matches := re.FindAllStringSubmatch(string(content), -1)
	for _, match := range matches {
		if len(match) > 1 {
			dependencies = append(dependencies, Dependency{
				Name:   strings.ToLower(match[1]),
				Type:   "build",
				System: true,
			})
		}
	}

	// Look for pkg_check_modules calls
	pkgRe := regexp.MustCompile(`pkg_check_modules\s*\([^)]*\s+([a-zA-Z0-9_-]+)`)
	pkgMatches := pkgRe.FindAllStringSubmatch(string(content), -1)
	for _, match := range pkgMatches {
		if len(match) > 1 {
			dependencies = append(dependencies, Dependency{
				Name:   match[1],
				Type:   "build",
				System: true,
			})
		}
	}

	return dependencies, nil
}

// analyzeMesonDependencies analyzes meson.build for dependencies
func (ra *RepositoryAnalyzer) analyzeMesonDependencies(repoPath string) ([]Dependency, error) {
	var dependencies []Dependency

	mesonFile := filepath.Join(repoPath, "meson.build")
	content, err := os.ReadFile(mesonFile)
	if err != nil {
		return nil, err
	}

	// Look for dependency() calls
	re := regexp.MustCompile(`dependency\s*\(\s*['"]([^'"]+)['"]`)
	matches := re.FindAllStringSubmatch(string(content), -1)
	for _, match := range matches {
		if len(match) > 1 {
			dependencies = append(dependencies, Dependency{
				Name:   match[1],
				Type:   "build",
				System: true,
			})
		}
	}

	return dependencies, nil
}
