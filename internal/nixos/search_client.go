package nixos

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"nix-ai-help/pkg/logger"
)

// SearchNixOSClient provides integration with search.nixos.org for package and option verification
type SearchNixOSClient struct {
	baseURL    string
	httpClient *http.Client
	logger     *logger.Logger
}

// Package represents a NixOS package from search.nixos.org
type Package struct {
	Name        string `json:"package_attr_name"`
	Version     string `json:"package_pversion"`
	Description string `json:"package_description"`
	Homepage    string `json:"package_homepage"`
	Available   bool   `json:"available"`
}

// Option represents a NixOS configuration option from search.nixos.org
type Option struct {
	Name        string   `json:"option_name"`
	Description string   `json:"option_description"`
	Type        string   `json:"option_type"`
	Default     string   `json:"option_default"`
	Example     string   `json:"option_example"`
	Defined     []string `json:"option_defined"`
}

// SearchResult represents search results from search.nixos.org
type SearchResult struct {
	Packages []Package `json:"packages"`
	Options  []Option  `json:"options"`
}

// VerificationResult represents the result of verifying answer content against official sources
type VerificationResult struct {
	PackagesVerified          []Package `json:"packages_verified"`
	OptionsVerified           []Option  `json:"options_verified"`
	PackageVerificationFailed bool      `json:"package_verification_failed"`
	OptionVerificationFailed  bool      `json:"option_verification_failed"`
	UnknownPackages           []string  `json:"unknown_packages"`
	UnknownOptions            []string  `json:"unknown_options"`
	Confidence                float64   `json:"confidence"`
}

// NewSearchNixOSClient creates a new client for search.nixos.org API
func NewSearchNixOSClient() *SearchNixOSClient {
	return &SearchNixOSClient{
		baseURL: "https://search.nixos.org/",
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger.NewLogger(),
	}
}

// SearchPackages searches for packages on search.nixos.org
func (c *SearchNixOSClient) SearchPackages(query string) ([]Package, error) {
	endpoint := "packages"
	params := url.Values{
		"query": {query},
		"size":  {"50"},
		"sort":  {"relevance"},
		"type":  {"package"},
	}

	resp, err := c.makeRequest(endpoint, params)
	if err != nil {
		return nil, fmt.Errorf("failed to search packages: %w", err)
	}
	defer resp.Body.Close()

	var searchResult SearchResult
	if err := json.NewDecoder(resp.Body).Decode(&searchResult); err != nil {
		return nil, fmt.Errorf("failed to decode package search response: %w", err)
	}

	return searchResult.Packages, nil
}

// SearchOptions searches for configuration options on search.nixos.org
func (c *SearchNixOSClient) SearchOptions(query string) ([]Option, error) {
	endpoint := "options"
	params := url.Values{
		"query": {query},
		"size":  {"50"},
		"sort":  {"relevance"},
		"type":  {"option"},
	}

	resp, err := c.makeRequest(endpoint, params)
	if err != nil {
		return nil, fmt.Errorf("failed to search options: %w", err)
	}
	defer resp.Body.Close()

	var searchResult SearchResult
	if err := json.NewDecoder(resp.Body).Decode(&searchResult); err != nil {
		return nil, fmt.Errorf("failed to decode option search response: %w", err)
	}

	return searchResult.Options, nil
}

// VerifyPackageExists checks if a specific package exists in the NixOS package repository
func (c *SearchNixOSClient) VerifyPackageExists(packageName string) (bool, *Package, error) {
	packages, err := c.SearchPackages(packageName)
	if err != nil {
		return false, nil, err
	}

	// Look for exact match first
	for _, pkg := range packages {
		if pkg.Name == packageName {
			pkg.Available = true
			return true, &pkg, nil
		}
	}

	// Look for close matches
	for _, pkg := range packages {
		if strings.Contains(pkg.Name, packageName) || strings.Contains(packageName, pkg.Name) {
			pkg.Available = true
			return true, &pkg, nil
		}
	}

	return false, nil, nil
}

// VerifyOptionExists checks if a specific NixOS configuration option exists
func (c *SearchNixOSClient) VerifyOptionExists(optionName string) (bool, *Option, error) {
	options, err := c.SearchOptions(optionName)
	if err != nil {
		return false, nil, err
	}

	// Look for exact match first
	for _, opt := range options {
		if opt.Name == optionName {
			return true, &opt, nil
		}
	}

	// Look for close matches
	for _, opt := range options {
		if strings.Contains(opt.Name, optionName) || strings.Contains(optionName, opt.Name) {
			return true, &opt, nil
		}
	}

	return false, nil, nil
}

// VerifyAnswer verifies an AI-generated answer against official NixOS sources
func (c *SearchNixOSClient) VerifyAnswer(ctx context.Context, answer string) (*VerificationResult, error) {
	result := &VerificationResult{
		PackagesVerified:          []Package{},
		OptionsVerified:           []Option{},
		PackageVerificationFailed: false,
		OptionVerificationFailed:  false,
		UnknownPackages:           []string{},
		UnknownOptions:            []string{},
		Confidence:                1.0,
	}

	// Extract package references from the answer
	packageNames := c.extractPackageNames(answer)
	for _, packageName := range packageNames {
		exists, pkg, err := c.VerifyPackageExists(packageName)
		if err != nil {
			c.logger.Printf("Failed to verify package %s: %v", packageName, err)
			continue
		}

		if exists && pkg != nil {
			result.PackagesVerified = append(result.PackagesVerified, *pkg)
		} else {
			result.UnknownPackages = append(result.UnknownPackages, packageName)
			result.PackageVerificationFailed = true
		}
	}

	// Extract option references from the answer
	optionNames := c.extractOptionNames(answer)
	for _, optionName := range optionNames {
		exists, opt, err := c.VerifyOptionExists(optionName)
		if err != nil {
			c.logger.Printf("Failed to verify option %s: %v", optionName, err)
			continue
		}

		if exists && opt != nil {
			result.OptionsVerified = append(result.OptionsVerified, *opt)
		} else {
			result.UnknownOptions = append(result.UnknownOptions, optionName)
			result.OptionVerificationFailed = true
		}
	}

	// Calculate confidence based on verification results
	result.Confidence = c.calculateVerificationConfidence(result)

	return result, nil
}

// extractPackageNames extracts potential package names from answer text
func (c *SearchNixOSClient) extractPackageNames(answer string) []string {
	var packages []string

	// Common patterns for package references in NixOS answers
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`with pkgs; \[([^\]]+)\]`),                         // with pkgs; [package1 package2]
		regexp.MustCompile(`pkgs\.([a-zA-Z0-9\-_]+)`),                         // pkgs.package-name
		regexp.MustCompile(`environment\.systemPackages.*?([a-zA-Z0-9\-_]+)`), // in systemPackages context
		regexp.MustCompile(`nix-shell -p ([a-zA-Z0-9\-_\s]+)`),                // nix-shell -p package
	}

	for _, pattern := range patterns {
		matches := pattern.FindAllStringSubmatch(answer, -1)
		for _, match := range matches {
			if len(match) > 1 {
				// Split multiple packages and clean them
				pkgList := strings.Fields(match[1])
				for _, pkg := range pkgList {
					cleaned := strings.Trim(pkg, "[]{}()\"',")
					if len(cleaned) > 2 && isValidPackageName(cleaned) {
						packages = append(packages, cleaned)
					}
				}
			}
		}
	}

	// Remove duplicates
	return removeDuplicates(packages)
}

// extractOptionNames extracts potential NixOS configuration option names from answer text
func (c *SearchNixOSClient) extractOptionNames(answer string) []string {
	var options []string

	// Pattern for NixOS configuration options
	optionPattern := regexp.MustCompile(`([a-zA-Z][a-zA-Z0-9]*(?:\.[a-zA-Z][a-zA-Z0-9]*)+)\s*=`)

	matches := optionPattern.FindAllStringSubmatch(answer, -1)
	for _, match := range matches {
		if len(match) > 1 {
			option := match[1]
			if isValidOptionName(option) {
				options = append(options, option)
			}
		}
	}

	// Remove duplicates
	return removeDuplicates(options)
}

// calculateVerificationConfidence calculates confidence based on verification results
func (c *SearchNixOSClient) calculateVerificationConfidence(result *VerificationResult) float64 {
	totalReferences := len(result.PackagesVerified) + len(result.OptionsVerified) +
		len(result.UnknownPackages) + len(result.UnknownOptions)

	if totalReferences == 0 {
		return 1.0 // No references to verify, assume good
	}

	verifiedReferences := len(result.PackagesVerified) + len(result.OptionsVerified)
	return float64(verifiedReferences) / float64(totalReferences)
}

// makeRequest makes an HTTP request to the search.nixos.org API
func (c *SearchNixOSClient) makeRequest(endpoint string, params url.Values) (*http.Response, error) {
	fullURL := c.baseURL + endpoint + "?" + params.Encode()

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "nixai/1.0")
	req.Header.Set("Accept", "application/json")

	return c.httpClient.Do(req)
}

// Helper functions

func isValidPackageName(name string) bool {
	// Basic validation for package names
	return regexp.MustCompile(`^[a-zA-Z0-9\-_]+$`).MatchString(name) &&
		len(name) > 1 &&
		!strings.HasPrefix(name, "-") &&
		!strings.HasSuffix(name, "-")
}

func isValidOptionName(name string) bool {
	// Basic validation for option names (must have at least one dot)
	return strings.Contains(name, ".") &&
		regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9]*(?:\.[a-zA-Z][a-zA-Z0-9]*)+$`).MatchString(name)
}

func removeDuplicates(slice []string) []string {
	keys := make(map[string]bool)
	result := []string{}

	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}

	return result
}
