package ai

import (
	"bytes"
	"strings"
	"text/template"
)

type PromptContext struct {
	Question       string
	LogSnippet     string
	ConfigSnippet  string
	CommandOutput  string
	DocExcerpts    []string
	Intent         string // e.g. "troubleshoot", "explain", "generate-code"
	OutputFormat   string // e.g. "markdown", "code", "steps"
	Provider       string // e.g. "ollama", "openai", "gemini"
	SessionHistory []string
}

type PromptBuilder interface {
	BuildPrompt(ctx PromptContext) (string, error)
}

type DefaultPromptBuilder struct{}

const defaultPromptTemplate = `You are an expert NixOS assistant.
{{if eq .Intent "troubleshoot"}}Diagnose and solve the following problem:{{end}}
User question: {{.Question}}
{{if .LogSnippet}}Relevant log:
{{.LogSnippet}}
{{end}}{{if .ConfigSnippet}}Relevant config:
{{.ConfigSnippet}}
{{end}}{{if .CommandOutput}}Command output:
{{.CommandOutput}}
{{end}}{{if .DocExcerpts}}Relevant documentation:
{{range .DocExcerpts}}- {{.}}
{{end}}{{end}}{{if .SessionHistory}}Previous conversation:
{{range .SessionHistory}}{{.}}
{{end}}{{end}}Please provide a {{.OutputFormat}} answer for a Linux terminal user.`

func (b *DefaultPromptBuilder) BuildPrompt(ctx PromptContext) (string, error) {
	tmpl, err := template.New("prompt").Parse(defaultPromptTemplate)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, ctx); err != nil {
		return "", err
	}
	return strings.TrimSpace(buf.String()), nil
}
