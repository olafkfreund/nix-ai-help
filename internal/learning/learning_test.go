package learning

import (
	"testing"
)

func TestModuleStruct(t *testing.T) {
	m := Module{
		ID:    "basics",
		Title: "NixOS Basics",
		Steps: []Step{{Title: "Intro", Instruction: "Welcome!"}},
	}
	if m.ID != "basics" {
		t.Errorf("expected ID 'basics', got %s", m.ID)
	}
	if m.Title != "NixOS Basics" {
		t.Errorf("expected Title 'NixOS Basics', got %s", m.Title)
	}
	if len(m.Steps) != 1 {
		t.Errorf("expected 1 step, got %d", len(m.Steps))
	}
}

func TestProgressStruct(t *testing.T) {
	p := Progress{
		CompletedModules: map[string]bool{"basics": true},
		QuizScores:       map[string]int{"basics": 100},
	}
	if !p.CompletedModules["basics"] {
		t.Error("expected basics to be completed")
	}
	if p.QuizScores["basics"] != 100 {
		t.Errorf("expected quiz score 100, got %d", p.QuizScores["basics"])
	}
}
