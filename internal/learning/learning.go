// Package learning provides interactive learning modules, quizzes, and onboarding for NixOS users.
package learning

import (
	"fmt"
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v3"
)

// Module represents a learning module with steps and optional quiz.
type Module struct {
	ID          string
	Title       string
	Description string
	Steps       []Step
	Quiz        *Quiz
	Tags        []string
	Level       string // e.g. "basics", "advanced"
}

type Step struct {
	Title       string
	Instruction string
	Example     string
	Exercise    string
}

type Quiz struct {
	Questions []Question
}

type Question struct {
	Prompt   string
	Choices  []string
	Answer   int // index of correct answer
	Feedback string
}

// Progress tracks user progress through modules and quizzes.
type Progress struct {
	CompletedModules map[string]bool
	QuizScores       map[string]int
}

// LoadModules loads available learning modules (stub).
func LoadModules() ([]Module, error) {
	// TODO: Load from YAML or embed default modules
	return []Module{}, nil
}

// SaveProgress saves user progress persistently to ~/.config/nixai/learning.yaml.
func SaveProgress(progress Progress) error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	progressPath := filepath.Join(configDir, "nixai", "learning.yaml")
	if err := os.MkdirAll(filepath.Dir(progressPath), 0700); err != nil {
		return err
	}
	data, err := yaml.Marshal(&progress)
	if err != nil {
		return err
	}
	return os.WriteFile(progressPath, data, 0600)
}

// LoadProgress loads user progress from ~/.config/nixai/learning.yaml.
func LoadProgress() (Progress, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return Progress{}, err
	}
	progressPath := filepath.Join(configDir, "nixai", "learning.yaml")
	data, err := os.ReadFile(progressPath)
	if err != nil {
		if os.IsNotExist(err) {
			return Progress{CompletedModules: map[string]bool{}, QuizScores: map[string]int{}}, nil
		}
		return Progress{}, err
	}
	var progress Progress
	if err := yaml.Unmarshal(data, &progress); err != nil {
		return Progress{}, err
	}
	if progress.CompletedModules == nil {
		progress.CompletedModules = map[string]bool{}
	}
	if progress.QuizScores == nil {
		progress.QuizScores = map[string]int{}
	}
	return progress, nil
}

// RenderModule prints a module overview.
func RenderModule(m Module) {
	fmt.Printf("\n# %s\n\n%s\n", m.Title, m.Description)
	for i, step := range m.Steps {
		fmt.Printf("\n%d. %s\n%s\n", i+1, step.Title, step.Instruction)
		if step.Example != "" {
			fmt.Printf("Example:\n%s\n", step.Example)
		}
		if step.Exercise != "" {
			fmt.Printf("Exercise: %s\n", step.Exercise)
		}
	}
	if m.Quiz != nil {
		fmt.Println("\nQuiz available!")
	}
}
