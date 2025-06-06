package functionbase

import (
	"context"
	"encoding/json"
	"time"
)

// FunctionParameter represents a parameter for an AI function
type FunctionParameter struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Required    bool        `json:"required"`
	Default     interface{} `json:"default,omitempty"`
	Enum        []string    `json:"enum,omitempty"`
	Pattern     string      `json:"pattern,omitempty"`
	MinLength   *int        `json:"minLength,omitempty"`
	MaxLength   *int        `json:"maxLength,omitempty"`
	Minimum     *float64    `json:"minimum,omitempty"`
	Maximum     *float64    `json:"maximum,omitempty"`
}

// FunctionSchema defines the JSON schema for an AI function
type FunctionSchema struct {
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Parameters  []FunctionParameter `json:"parameters"`
	Examples    []FunctionExample   `json:"examples,omitempty"`
}

// FunctionExample provides usage examples for the function
type FunctionExample struct {
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
	Expected    string                 `json:"expected"`
}

// FunctionCall represents a call to an AI function
type FunctionCall struct {
	Name       string                 `json:"name"`
	Parameters map[string]interface{} `json:"parameters"`
	Context    context.Context        `json:"-"`
	Timestamp  time.Time              `json:"timestamp"`
}

// FunctionResult represents the result of a function execution
type FunctionResult struct {
	Success   bool                   `json:"success"`
	Data      interface{}            `json:"data,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Duration  time.Duration          `json:"duration"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// FunctionExecution tracks the execution of a function
type FunctionExecution struct {
	ID        string         `json:"id"`
	Call      FunctionCall   `json:"call"`
	Result    FunctionResult `json:"result"`
	StartTime time.Time      `json:"start_time"`
	EndTime   time.Time      `json:"end_time"`
	Status    string         `json:"status"` // pending, running, completed, failed
}

// Progress represents the progress of a long-running function
type Progress struct {
	Current    int       `json:"current"`
	Total      int       `json:"total"`
	Percentage float64   `json:"percentage"`
	Message    string    `json:"message"`
	Stage      string    `json:"stage"`
	Timestamp  time.Time `json:"timestamp"`
}

// ProgressCallback is called during function execution to report progress
type ProgressCallback func(progress Progress)

// FunctionOptions contains options for function execution
type FunctionOptions struct {
	Timeout          time.Duration          `json:"timeout,omitempty"`
	ProgressCallback ProgressCallback       `json:"-"`
	Async            bool                   `json:"async"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// ValidationError represents a parameter validation error
type ValidationError struct {
	Parameter string      `json:"parameter"`
	Message   string      `json:"message"`
	Value     interface{} `json:"value,omitempty"`
}

func (e ValidationError) Error() string {
	return e.Message
}

// FunctionError represents an error during function execution
type FunctionError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Details   string `json:"details,omitempty"`
	Retryable bool   `json:"retryable"`
}

func (e FunctionError) Error() string {
	return e.Message
}

// ToJSON converts any struct to JSON for logging/debugging
func ToJSON(v interface{}) string {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "error marshaling to JSON"
	}
	return string(data)
}

// FromJSON converts JSON string to interface{}
func FromJSON(jsonStr string) (interface{}, error) {
	var result interface{}
	err := json.Unmarshal([]byte(jsonStr), &result)
	return result, err
}
