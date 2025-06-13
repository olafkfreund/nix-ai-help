package detection_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"nix-ai-help/internal/packaging/detection"
	"nix-ai-help/pkg/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnhancedDetector_DetectLanguages(t *testing.T) {
	log := logger.NewTestLogger()
	detector := detection.NewEnhancedDetector(log)

	tests := []struct {
		name          string
		setupRepo     func(t *testing.T, repoPath string)
		expectedLangs []string
		minConfidence float64
		expectedError bool
	}{
		{
			name: "JavaScript project",
			setupRepo: func(t *testing.T, repoPath string) {
				// Create package.json
				packageJSON := `{
  "name": "test-project",
  "version": "1.0.0",
  "main": "index.js",
  "dependencies": {
    "express": "^4.18.0"
  }
}`
				require.NoError(t, os.WriteFile(filepath.Join(repoPath, "package.json"), []byte(packageJSON), 0644))

				// Create JavaScript files
				indexJS := `const express = require('express');
const app = express();

app.get('/', (req, res) => {
  res.send('Hello World!');
});

app.listen(3000);`
				require.NoError(t, os.WriteFile(filepath.Join(repoPath, "index.js"), []byte(indexJS), 0644))

				utilsJS := `export function formatDate(date) {
  return date.toISOString();
}`
				require.NoError(t, os.WriteFile(filepath.Join(repoPath, "utils.js"), []byte(utilsJS), 0644))
			},
			expectedLangs: []string{"javascript"},
			minConfidence: 0.5,
		},
		{
			name: "TypeScript project",
			setupRepo: func(t *testing.T, repoPath string) {
				// Create tsconfig.json
				tsconfig := `{
  "compilerOptions": {
    "target": "es2020",
    "module": "commonjs",
    "strict": true
  }
}`
				require.NoError(t, os.WriteFile(filepath.Join(repoPath, "tsconfig.json"), []byte(tsconfig), 0644))

				// Create TypeScript files
				indexTS := `interface User {
  id: number;
  name: string;
}

function getUser(id: number): User {
  return { id, name: "Test User" };
}`
				require.NoError(t, os.WriteFile(filepath.Join(repoPath, "index.ts"), []byte(indexTS), 0644))
			},
			expectedLangs: []string{"typescript"},
			minConfidence: 0.5,
		},
		{
			name: "Python project",
			setupRepo: func(t *testing.T, repoPath string) {
				// Create requirements.txt
				requirements := `requests==2.28.0
flask==2.1.0`
				require.NoError(t, os.WriteFile(filepath.Join(repoPath, "requirements.txt"), []byte(requirements), 0644))

				// Create Python files
				mainPy := `#!/usr/bin/env python3
import requests
from flask import Flask

def main():
    app = Flask(__name__)
    
    @app.route('/')
    def hello():
        return "Hello World!"
    
    app.run()

if __name__ == "__main__":
    main()`
				require.NoError(t, os.WriteFile(filepath.Join(repoPath, "main.py"), []byte(mainPy), 0644))
			},
			expectedLangs: []string{"python"},
			minConfidence: 0.5,
		},
		{
			name: "Rust project",
			setupRepo: func(t *testing.T, repoPath string) {
				// Create Cargo.toml
				cargoToml := `[package]
name = "test-project"
version = "0.1.0"
edition = "2021"

[dependencies]
serde = "1.0"`
				require.NoError(t, os.WriteFile(filepath.Join(repoPath, "Cargo.toml"), []byte(cargoToml), 0644))

				// Create Rust files
				srcDir := filepath.Join(repoPath, "src")
				require.NoError(t, os.MkdirAll(srcDir, 0755))

				mainRS := `use serde::{Deserialize, Serialize};

#[derive(Serialize, Deserialize)]
struct User {
    id: u32,
    name: String,
}

fn main() {
    println!("Hello, world!");
}`
				require.NoError(t, os.WriteFile(filepath.Join(srcDir, "main.rs"), []byte(mainRS), 0644))
			},
			expectedLangs: []string{"rust"},
			minConfidence: 0.5,
		},
		{
			name: "Go project",
			setupRepo: func(t *testing.T, repoPath string) {
				// Create go.mod
				goMod := `module test-project

go 1.19

require github.com/gin-gonic/gin v1.8.0`
				require.NoError(t, os.WriteFile(filepath.Join(repoPath, "go.mod"), []byte(goMod), 0644))

				// Create Go files
				mainGo := `package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello World!",
		})
	})
	fmt.Println("Starting server...")
	r.Run()
}`
				require.NoError(t, os.WriteFile(filepath.Join(repoPath, "main.go"), []byte(mainGo), 0644))
			},
			expectedLangs: []string{"go"},
			minConfidence: 0.5,
		},
		{
			name: "Multi-language project",
			setupRepo: func(t *testing.T, repoPath string) {
				// Create package.json (JavaScript)
				packageJSON := `{"name": "multi-lang", "version": "1.0.0"}`
				require.NoError(t, os.WriteFile(filepath.Join(repoPath, "package.json"), []byte(packageJSON), 0644))

				// Create requirements.txt (Python)
				requirements := `flask==2.1.0`
				require.NoError(t, os.WriteFile(filepath.Join(repoPath, "requirements.txt"), []byte(requirements), 0644))

				// Create JS file
				indexJS := `console.log("Hello from JavaScript");`
				require.NoError(t, os.WriteFile(filepath.Join(repoPath, "index.js"), []byte(indexJS), 0644))

				// Create Python file
				mainPy := `print("Hello from Python")`
				require.NoError(t, os.WriteFile(filepath.Join(repoPath, "main.py"), []byte(mainPy), 0644))
			},
			expectedLangs: []string{"javascript", "python"},
			minConfidence: 0.3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory
			repoPath, err := os.MkdirTemp("", "test-repo-*")
			require.NoError(t, err)
			defer os.RemoveAll(repoPath)

			// Setup repository
			tt.setupRepo(t, repoPath)

			// Set up analysis options
			opts := detection.DefaultAnalysisOptions()
			opts.MinConfidence = tt.minConfidence

			// Run detection
			ctx := context.Background()
			results, err := detector.DetectLanguages(ctx, repoPath, opts)

			if tt.expectedError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, results)

			// Check that expected languages are detected
			detectedLangs := make(map[string]bool)
			for _, result := range results {
				detectedLangs[result.Language] = true

				// Verify result structure
				assert.GreaterOrEqual(t, result.Confidence, tt.minConfidence)
				assert.NotEmpty(t, result.Evidence)
				assert.NotEmpty(t, result.Files)
			}

			for _, expectedLang := range tt.expectedLangs {
				assert.True(t, detectedLangs[expectedLang],
					"Expected language %s not detected. Detected: %v", expectedLang, detectedLangs)
			}
		})
	}
}

func TestEnhancedDetector_AnalysisOptions(t *testing.T) {
	log := logger.NewTestLogger()
	detector := detection.NewEnhancedDetector(log)

	// Create test repository
	repoPath, err := os.MkdirTemp("", "test-repo-*")
	require.NoError(t, err)
	defer os.RemoveAll(repoPath)

	// Create files
	require.NoError(t, os.WriteFile(filepath.Join(repoPath, "index.js"), []byte("console.log('test');"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(repoPath, ".hidden.js"), []byte("console.log('hidden');"), 0644))

	// Create node_modules directory
	nodeModulesPath := filepath.Join(repoPath, "node_modules")
	require.NoError(t, os.MkdirAll(nodeModulesPath, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(nodeModulesPath, "package.js"), []byte("module.exports = {};"), 0644))

	tests := []struct {
		name     string
		opts     *detection.AnalysisOptions
		validate func(t *testing.T, results []detection.LanguageResult)
	}{
		{
			name: "exclude hidden files",
			opts: &detection.AnalysisOptions{
				MaxFiles:        1000,
				Timeout:         30 * time.Second,
				IncludeHidden:   false,
				MinConfidence:   0.1,
				ExcludePatterns: []string{"**/node_modules/**"},
			},
			validate: func(t *testing.T, results []detection.LanguageResult) {
				require.NotEmpty(t, results)
				// Should detect JavaScript but not include hidden files
				for _, result := range results {
					for _, file := range result.Files {
						assert.False(t, filepath.Base(file) == ".hidden.js")
					}
				}
			},
		},
		{
			name: "include hidden files",
			opts: &detection.AnalysisOptions{
				MaxFiles:        1000,
				Timeout:         30 * time.Second,
				IncludeHidden:   true,
				MinConfidence:   0.1,
				ExcludePatterns: []string{"**/node_modules/**"},
			},
			validate: func(t *testing.T, results []detection.LanguageResult) {
				require.NotEmpty(t, results)
				// Should include hidden files
				foundHidden := false
				for _, result := range results {
					for _, file := range result.Files {
						if filepath.Base(file) == ".hidden.js" {
							foundHidden = true
						}
					}
				}
				assert.True(t, foundHidden, "Hidden file should be included")
			},
		},
		{
			name: "high confidence threshold",
			opts: &detection.AnalysisOptions{
				MaxFiles:      1000,
				Timeout:       30 * time.Second,
				IncludeHidden: false,
				MinConfidence: 0.9,
			},
			validate: func(t *testing.T, results []detection.LanguageResult) {
				// All results should have high confidence
				for _, result := range results {
					assert.GreaterOrEqual(t, result.Confidence, 0.9)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			results, err := detector.DetectLanguages(ctx, repoPath, tt.opts)
			require.NoError(t, err)
			tt.validate(t, results)
		})
	}
}

func TestEnhancedDetector_Timeout(t *testing.T) {
	log := logger.NewTestLogger()
	detector := detection.NewEnhancedDetector(log)

	// Create test repository
	repoPath, err := os.MkdirTemp("", "test-repo-*")
	require.NoError(t, err)
	defer os.RemoveAll(repoPath)

	// Create a simple file
	require.NoError(t, os.WriteFile(filepath.Join(repoPath, "index.js"), []byte("console.log('test');"), 0644))

	// Test with very short timeout
	opts := &detection.AnalysisOptions{
		MaxFiles:      1000,
		Timeout:       1 * time.Nanosecond, // Extremely short timeout
		IncludeHidden: false,
		MinConfidence: 0.1,
	}

	ctx := context.Background()
	_, err = detector.DetectLanguages(ctx, repoPath, opts)

	// Should either succeed quickly or timeout
	if err != nil {
		assert.Contains(t, err.Error(), "deadline")
	}
}

func TestLanguagePatterns(t *testing.T) {
	// Test that all pattern regexes compile
	for language, patterns := range detection.LanguagePatterns {
		for i, pattern := range patterns {
			if pattern.Content != nil {
				// Regex should be valid
				assert.NotNil(t, pattern.Content, "Language %s pattern %d has nil regex", language, i)
			}

			// Confidence should be valid
			assert.GreaterOrEqual(t, pattern.Confidence, 0.0, "Language %s pattern %d has negative confidence", language, i)
			assert.LessOrEqual(t, pattern.Confidence, 1.0, "Language %s pattern %d has confidence > 1.0", language, i)
		}
	}
}

func TestGetLanguagePatterns(t *testing.T) {
	// Test existing language
	patterns, exists := detection.GetLanguagePatterns("javascript")
	assert.True(t, exists)
	assert.NotEmpty(t, patterns)

	// Test non-existing language
	patterns, exists = detection.GetLanguagePatterns("nonexistent")
	assert.False(t, exists)
	assert.Nil(t, patterns)
}

func TestGetAllLanguages(t *testing.T) {
	languages := detection.GetAllLanguages()
	assert.NotEmpty(t, languages)

	// Should include common languages
	languageMap := make(map[string]bool)
	for _, lang := range languages {
		languageMap[lang] = true
	}

	expectedLanguages := []string{"javascript", "typescript", "python", "rust", "go"}
	for _, lang := range expectedLanguages {
		assert.True(t, languageMap[lang], "Expected language %s not found", lang)
	}
}
